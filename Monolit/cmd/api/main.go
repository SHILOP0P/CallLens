package main

import (
	"calllens/monolit/internal/API/call"
	"calllens/monolit/internal/config"
	"calllens/monolit/internal/httpserver"
	"calllens/monolit/internal/migrator"
	callRepo "calllens/monolit/internal/repository/call"
	callService "calllens/monolit/internal/service/call"
	"calllens/monolit/internal/storage/audio"
	"context"
	"log"
	"net/http"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/stdlib"
)

const (
	configPath = "./.env"
)

func main() {

	err := config.Load(configPath)
	if err != nil {
		log.Printf("failed to load .env: %v", err)
		return
	}

	ctx := context.Background()
	//var cancel context.CancelFunc

	dbURI := config.AppConfig().Postgres.URI()
	if dbURI == "" {
		log.Printf("failed to load .env: postgres URI is empty")
		return
	}

	con, err := pgx.Connect(ctx, dbURI)
	if err != nil {
		log.Printf("failed to connect to postgres: %v", err)
		return
	}
	defer func() {
		if cerr := con.Close(ctx); cerr != nil {
			log.Printf("failed to close postgres connection: %v", cerr)
		}
	}()

	err = con.Ping(ctx)
	if err != nil {
		log.Printf("failed to ping postgres: %v", err)
	}

	sqlDB := stdlib.OpenDB(*con.Config().Copy())
	migrationsDIR := config.AppConfig().Postgres.MigrationDir()
	if migrationsDIR == "" {
		log.Printf("failed to load .env: migrations directory is empty")
		return
	}
	migratorRunner := migrator.NewMigrator(sqlDB, migrationsDIR)

	err = migratorRunner.Up()
	if err != nil {
		log.Printf("failed to run migrator: %v", err)
	}

	uploadPath := config.AppConfig().Upload.Path()

	audioStorage := audio.NewLocalStorage(uploadPath)
	repository := callRepo.NewRepository(sqlDB)
	service := callService.NewService(repository, audioStorage)
	callHandler := call.NewCallHandler(service)

	r := httpserver.NewRouter(callHandler)

	server := &http.Server{
		Addr:              config.AppConfig().HTTPConfig.Address(),
		Handler:           r,
		ReadHeaderTimeout: config.AppConfig().HTTPConfig.ReadTimeout(),
	}

	log.Printf("api server started on %s", server.Addr)

	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("api server stopped with error: %v", err)
	}
}
