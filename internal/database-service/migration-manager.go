package databaseService

import (
	"errors"
	"fmt"
	"os"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/source"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
)

type migrationManager struct {
	pool          *pgxpool.Pool
	migrationsURL string
}

func NewMigrationManager(pool *pgxpool.Pool, migrationsURL string) MigrationManager {
	return &migrationManager{pool: pool, migrationsURL: migrationsURL}
}

func (manager *migrationManager) Status() (MigrationStatus, error) {
	migrationSource, err := source.Open(manager.migrationsURL)
	if err != nil {
		return MigrationStatus{}, fmt.Errorf("open migration source: %w", err)
	}
	latestVersion, hasMigrations, sourceErr := latestMigrationVersion(migrationSource)
	closeErr := migrationSource.Close()
	if sourceErr != nil || closeErr != nil {
		return MigrationStatus{}, errors.Join(sourceErr, closeErr)
	}

	status := MigrationStatus{LatestVersion: latestVersion, HasMigrations: hasMigrations}
	err = manager.withMigrator(func(migrator *migrate.Migrate) error {
		version, dirty, err := migrator.Version()
		if errors.Is(err, migrate.ErrNilVersion) {
			return nil
		}
		if err != nil {
			return fmt.Errorf("read current migration version: %w", err)
		}
		status.CurrentVersion = version
		status.HasCurrent = true
		status.Dirty = dirty
		return nil
	})
	if err != nil {
		return MigrationStatus{}, err
	}
	return status, nil
}

func (manager *migrationManager) Up() error {
	return manager.withMigrator(func(migrator *migrate.Migrate) error {
		err := migrator.Up()
		if errors.Is(err, migrate.ErrNoChange) {
			return nil
		}
		return err
	})
}

func (manager *migrationManager) Down() error {
	return manager.withMigrator(func(migrator *migrate.Migrate) error {
		err := migrator.Down()
		if errors.Is(err, migrate.ErrNoChange) {
			return nil
		}
		return err
	})
}

func (manager *migrationManager) Force(version int) error {
	return manager.withMigrator(func(migrator *migrate.Migrate) error {
		return migrator.Force(version)
	})
}

func (manager *migrationManager) withMigrator(operation func(*migrate.Migrate) error) (resultErr error) {
	if manager.pool == nil {
		return fmt.Errorf("PostgreSQL connection pool is nil")
	}

	sqlDB := stdlib.OpenDBFromPool(manager.pool)
	defer func() {
		if err := sqlDB.Close(); err != nil {
			resultErr = errors.Join(resultErr, fmt.Errorf("close PostgreSQL connection: %w", err))
		}
	}()

	databaseDriver, err := postgres.WithInstance(sqlDB, &postgres.Config{})
	if err != nil {
		return fmt.Errorf("initialize PostgreSQL migration driver: %w", err)
	}
	migrator, err := migrate.NewWithDatabaseInstance(manager.migrationsURL, "postgres", databaseDriver)
	if err != nil {
		return fmt.Errorf("initialize migration engine: %w", errors.Join(err, databaseDriver.Close()))
	}
	defer func() {
		sourceErr, databaseErr := migrator.Close()
		if sourceErr != nil || databaseErr != nil {
			resultErr = errors.Join(resultErr, fmt.Errorf("close migration resources: %w", errors.Join(sourceErr, databaseErr)))
		}
	}()

	return operation(migrator)
}

func latestMigrationVersion(migrationSource source.Driver) (uint, bool, error) {
	version, err := migrationSource.First()
	if errors.Is(err, os.ErrNotExist) {
		return 0, false, nil
	}
	if err != nil {
		return 0, false, fmt.Errorf("find first migration: %w", err)
	}

	for {
		next, err := migrationSource.Next(version)
		if errors.Is(err, os.ErrNotExist) {
			return version, true, nil
		}
		if err != nil {
			return 0, false, fmt.Errorf("find next migration after %d: %w", version, err)
		}
		version = next
	}
}
