package main

import (
	authAPI "calllens/monolit/internal/API/auth"
	"calllens/monolit/internal/API/call"
	companyAPI "calllens/monolit/internal/API/company"
	"calllens/monolit/internal/config"
	"calllens/monolit/internal/httpserver"
	"calllens/monolit/internal/logger"
	"calllens/monolit/internal/migrator"
	callRepo "calllens/monolit/internal/repository/call"
	companyRepo "calllens/monolit/internal/repository/company"
	departmentRepo "calllens/monolit/internal/repository/department"
	refreshSessionRepo "calllens/monolit/internal/repository/refresh_session"
	userRepo "calllens/monolit/internal/repository/user"
	authService "calllens/monolit/internal/service/auth"
	callService "calllens/monolit/internal/service/call"
	companyService "calllens/monolit/internal/service/company"
	"calllens/monolit/internal/storage/audio"
	"context"
	"net/http"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/stdlib"
	"go.uber.org/zap"
)

const (
	configPath = "./.env"
)

func main() {
	ctx := context.Background()
	startupLogger := logger.New("info", false)

	err := config.Load(configPath)
	if err != nil {
		startupLogger.Error(ctx, "failed to load config", zap.Error(err))
		return
	}

	appLogger := logger.New(config.AppConfig().Logger.Level(), config.AppConfig().Logger.AsJSON())
	//var cancel context.CancelFunc

	dbURI := config.AppConfig().Postgres.URI()
	if dbURI == "" {
		appLogger.Error(ctx, "postgres uri is empty")
		return
	}

	con, err := pgx.Connect(ctx, dbURI)
	if err != nil {
		appLogger.Error(ctx, "failed to connect to postgres", zap.Error(err))
		return
	}
	defer func() {
		if cerr := con.Close(ctx); cerr != nil {
			appLogger.Error(ctx, "failed to close postgres connection", zap.Error(cerr))
		}
	}()

	err = con.Ping(ctx)
	if err != nil {
		appLogger.Error(ctx, "failed to ping postgres", zap.Error(err))
	}

	sqlDB := stdlib.OpenDB(*con.Config().Copy())
	migrationsDIR := config.AppConfig().Postgres.MigrationDir()
	if migrationsDIR == "" {
		appLogger.Error(ctx, "migrations directory is empty")
		return
	}
	migratorRunner := migrator.NewMigrator(sqlDB, migrationsDIR)

	err = migratorRunner.Up()
	if err != nil {
		appLogger.Error(ctx, "failed to run migrator", zap.Error(err))
	}

	uploadPath := config.AppConfig().Upload.Path()

	audioStorage := audio.NewLocalStorage(uploadPath)

	callRepository := callRepo.NewRepository(sqlDB)
	userRepository := userRepo.NewUserRepository(sqlDB)
	refreshRepository := refreshSessionRepo.NewRepository(sqlDB)
	companyRepository := companyRepo.NewRepository(sqlDB)
	departmentRepository := departmentRepo.NewRepository(sqlDB)

	callSvc := callService.NewService(callRepository, companyRepository, departmentRepository, audioStorage, appLogger)
	authSvc := authService.NewService(
		userRepository,
		refreshRepository,
		config.AppConfig().Auth.PasswordPepper(),
		config.AppConfig().Auth.JWTSecret(),
		config.AppConfig().Auth.AccessTokenTTL(),
		config.AppConfig().Auth.RefreshTokenSecret(),
		config.AppConfig().Auth.RefreshTokenTTL(),
		appLogger,
	)
	companySvc := companyService.NewService(companyRepository, departmentRepository, appLogger)

	callHandler := call.NewCallHandler(callSvc)
	authHandler := authAPI.NewAuthHandler(authSvc)
	companyHandler := companyAPI.NewCompanyHandler(companySvc)

	r := httpserver.NewRouter(callHandler, authHandler, companyHandler, config.AppConfig().Auth.JWTSecret(), refreshRepository, appLogger)

	server := &http.Server{
		Addr:              config.AppConfig().HTTPConfig.Address(),
		Handler:           r,
		ReadHeaderTimeout: config.AppConfig().HTTPConfig.ReadTimeout(),
	}

	appLogger.Info(ctx, "api server started", zap.String("address", server.Addr))

	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		appLogger.Error(ctx, "api server stopped with error", zap.Error(err))
	}
}
