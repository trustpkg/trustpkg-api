package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/trustpkg/trustpkg-api/db"
	databaseService "github.com/trustpkg/trustpkg-api/internal/database-service"
)

// @title TrustPkg Database Service
// @version 1.7.2
// @description Database service for managing migrations and initializations.
func main() {
	command, err := databaseService.ParseCommand(os.Args[1:])
	if err != nil {
		log.Printf("database service failed: %v", err)
		os.Exit(1)
	}
	if command.Action == databaseService.ActionHelp {
		if err := databaseService.Execute(context.Background(), command, nil, nil, os.Stdout); err != nil {
			log.Printf("database service failed: %v", err)
			os.Exit(1)
		}
		return
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	db.ConnectDb()
	defer db.Pool.Close()

	migrationsURL := os.Getenv("MIGRATIONS_URL")
	if migrationsURL == "" {
		migrationsURL = "file://migrations"
	}

	migrations := databaseService.NewMigrationManager(db.Pool, migrationsURL)
	initializer := databaseService.NewOSVInitializer("npm")
	handler := databaseService.NewHandler(migrations, initializer, log.Default())
	if err := databaseService.Execute(ctx, command, migrations, handler, os.Stdout); err != nil {
		log.Printf("database service failed: %v", err)
		os.Exit(1)
	}
}
