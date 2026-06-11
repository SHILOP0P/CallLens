package main

import (
	instructionAPI "calllens/monolit/internal/API/analysis_instruction"
	authAPI "calllens/monolit/internal/API/auth"
	"calllens/monolit/internal/API/call"
	companyAPI "calllens/monolit/internal/API/company"
	departmentAPI "calllens/monolit/internal/API/department"
	"calllens/monolit/internal/config"
	"calllens/monolit/internal/httpserver"
	"calllens/monolit/internal/logger"
	"calllens/monolit/internal/migrator"
	analysisInstructionRepo "calllens/monolit/internal/repository/analysis_instruction"
	callRepo "calllens/monolit/internal/repository/call"
	companyRepo "calllens/monolit/internal/repository/company"
	departmentRepo "calllens/monolit/internal/repository/department"
	processingJobRepo "calllens/monolit/internal/repository/processing_job"
	refreshSessionRepo "calllens/monolit/internal/repository/refresh_session"
	transcriptionRepo "calllens/monolit/internal/repository/transcription"
	userRepo "calllens/monolit/internal/repository/user"
	analysisInstructionService "calllens/monolit/internal/service/analysis_instruction"
	authService "calllens/monolit/internal/service/auth"
	callService "calllens/monolit/internal/service/call"
	companyService "calllens/monolit/internal/service/company"
	departmentService "calllens/monolit/internal/service/department"
	processingService "calllens/monolit/internal/service/processing"
	"calllens/monolit/internal/storage/audio"
	"calllens/monolit/internal/storage/instruction"
	"calllens/monolit/internal/transcriber"
	"context"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/stdlib"
	"go.uber.org/zap"
)

const (
	configPath      = "./.env"
	shutdownTimeout = 10 * time.Second

	audioUploadDirName       = "audio"
	instructionUploadDirName = "instructions"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

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
		if cerr := con.Close(context.Background()); cerr != nil {
			appLogger.Error(context.Background(), "failed to close postgres connection", zap.Error(cerr))
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
	audioUploadPath := filepath.Join(uploadPath, audioUploadDirName)
	instructionUploadPath := filepath.Join(uploadPath, instructionUploadDirName)

	for _, dir := range []string{audioUploadPath, instructionUploadPath} {
		if err := os.MkdirAll(dir, 0755); err != nil {
			appLogger.Error(ctx, "failed to create upload directory", zap.String("path", dir), zap.Error(err))
			return
		}
	}

	audioStorage := audio.NewLocalStorage(audioUploadPath)
	instructionStorage := instruction.NewLocalStorage(instructionUploadPath)

	analysisInstructionRepository := analysisInstructionRepo.NewRepository(sqlDB)
	callRepository := callRepo.NewRepository(sqlDB)
	userRepository := userRepo.NewUserRepository(sqlDB)
	refreshRepository := refreshSessionRepo.NewRepository(sqlDB)
	companyRepository := companyRepo.NewRepository(sqlDB)
	departmentRepository := departmentRepo.NewRepository(sqlDB)
	transcriptionRepository := transcriptionRepo.NewRepository(sqlDB)
	processingJobRepository := processingJobRepo.NewRepository(sqlDB)

	transcriberProvider, err := transcriber.NewFromConfig(config.AppConfig().Transcriber)
	if err != nil {
		appLogger.Error(ctx, "failed to configure transcriber", zap.Error(err))
		return
	}

	processingSvc := processingService.NewService(callRepository, transcriptionRepository, processingJobRepository, audioStorage, transcriberProvider, appLogger)
	var workerDone <-chan struct{}
	if config.AppConfig().Worker.Enabled() {
		processingWorker := processingService.NewWorker(processingSvc, processingService.WorkerOptions{
			PollInterval: config.AppConfig().Worker.PollInterval(),
			Limit:        config.AppConfig().Worker.Limit(),
			RetryDelay:   config.AppConfig().Worker.RetryDelay(),
			StaleAfter:   config.AppConfig().Worker.StaleAfter(),
		}, appLogger)

		done := make(chan struct{})
		workerDone = done
		go func() {
			defer close(done)
			processingWorker.Run(ctx)
		}()
	} else {
		appLogger.Info(ctx, "processing worker disabled")
	}

	callSvc := callService.NewService(callRepository, companyRepository, departmentRepository, audioStorage, appLogger)
	callSvc.SetTranscriptionRepository(transcriptionRepository)
	callSvc.SetProcessingJobRepository(processingJobRepository)
	callSvc.SetProcessingJobMaxAttempts(config.AppConfig().Worker.MaxAttempts())
	callSvc.SetDurationDetector(audio.NewFFProbeDurationDetector(audioUploadPath, config.AppConfig().Upload.FFProbePath()))
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
	companySvc := companyService.NewService(companyRepository, appLogger)
	departmentSvc := departmentService.NewService(companyRepository, departmentRepository, appLogger)
	instructionSvc := analysisInstructionService.NewService(analysisInstructionRepository, companyRepository, departmentRepository, instructionStorage, appLogger)

	callHandler := call.NewCallHandler(callSvc)
	authHandler := authAPI.NewAuthHandler(authSvc)
	companyHandler := companyAPI.NewCompanyHandler(companySvc)
	departmentHandler := departmentAPI.NewDepartmentHandler(departmentSvc)
	instructionHandler := instructionAPI.NewHandler(instructionSvc)

	r := httpserver.NewRouter(callHandler, authHandler, companyHandler, departmentHandler, instructionHandler, config.AppConfig().Auth.JWTSecret(), refreshRepository, appLogger)

	server := &http.Server{
		Addr:              config.AppConfig().HTTPConfig.Address(),
		Handler:           r,
		ReadHeaderTimeout: config.AppConfig().HTTPConfig.ReadTimeout(),
	}

	appLogger.Info(ctx, "api server started", zap.String("address", server.Addr))

	serverErr := make(chan error, 1)
	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			serverErr <- err
			return
		}

		serverErr <- nil
	}()

	select {
	case <-ctx.Done():
		appLogger.Info(context.Background(), "shutdown signal received")
	case err := <-serverErr:
		if err != nil {
			appLogger.Error(context.Background(), "api server stopped with error", zap.Error(err))
		}
		stop()
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		appLogger.Error(context.Background(), "failed to shutdown api server", zap.Error(err))
	} else {
		appLogger.Info(context.Background(), "api server stopped")
	}

	if workerDone != nil {
		select {
		case <-workerDone:
			appLogger.Info(context.Background(), "processing worker shutdown completed")
		case <-shutdownCtx.Done():
			appLogger.Warn(context.Background(), "processing worker shutdown timed out", zap.Error(shutdownCtx.Err()))
		}
	}
}
