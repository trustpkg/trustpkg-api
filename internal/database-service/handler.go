package databaseService

import (
	"context"
	"fmt"
	"log"
)

type Handler struct {
	migrations  MigrationManager
	initializer Initializer
	logger      Logger
}

func NewHandler(migrations MigrationManager, initializer Initializer, logger Logger) *Handler {
	if logger == nil {
		logger = log.Default()
	}
	return &Handler{
		migrations:  migrations,
		initializer: initializer,
		logger:      logger,
	}
}

func (handler *Handler) Run(ctx context.Context) error {
	if handler == nil || handler.migrations == nil || handler.initializer == nil {
		return fmt.Errorf("database startup handler is not fully configured")
	}

	status, err := handler.migrations.Status()
	if err != nil {
		return fmt.Errorf("check migration status: %w", err)
	}
	if status.Dirty {
		return fmt.Errorf("database migration state is dirty at version %d; repair it before startup", status.CurrentVersion)
	}
	if status.HasCurrent && status.HasMigrations && status.CurrentVersion > status.LatestVersion {
		return fmt.Errorf("database migration version %d is newer than the latest available migration %d", status.CurrentVersion, status.LatestVersion)
	}
	if status.Pending() {
		handler.logger.Printf("pending migrations detected: current=%d latest=%d", status.CurrentVersion, status.LatestVersion)
		if err := handler.migrations.Up(); err != nil {
			return fmt.Errorf("apply pending migrations: %w", err)
		}
		handler.logger.Printf("pending migrations applied")
	} else {
		handler.logger.Printf("database migrations are up to date at version %d", status.CurrentVersion)
	}

	needed, err := handler.initializer.Needed(ctx)
	if err != nil {
		return fmt.Errorf("check whether first init is needed: %w", err)
	}
	if !needed {
		handler.logger.Printf("first init is already complete")
		return nil
	}

	handler.logger.Printf("first init is needed; starting initialization")
	if err := handler.initializer.Initialize(ctx); err != nil {
		return fmt.Errorf("run first init: %w", err)
	}
	handler.logger.Printf("first init completed")
	return nil
}
