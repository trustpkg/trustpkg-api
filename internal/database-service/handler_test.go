package databaseService

import (
	"context"
	"errors"
	"fmt"
	"testing"
)

type fakeMigrationManager struct {
	status     MigrationStatus
	statusErr  error
	upErr      error
	upCalls    int
	downCalls  int
	forceCalls int
}

func (manager *fakeMigrationManager) Status() (MigrationStatus, error) {
	return manager.status, manager.statusErr
}

func (manager *fakeMigrationManager) Up() error {
	manager.upCalls++
	return manager.upErr
}

func (manager *fakeMigrationManager) Down() error {
	manager.downCalls++
	return nil
}

func (manager *fakeMigrationManager) Force(int) error {
	manager.forceCalls++
	return nil
}

type fakeInitializer struct {
	needed        bool
	neededErr     error
	initializeErr error
	calls         int
}

func (initializer *fakeInitializer) Needed(context.Context) (bool, error) {
	return initializer.needed, initializer.neededErr
}

func (initializer *fakeInitializer) Initialize(context.Context) error {
	initializer.calls++
	return initializer.initializeErr
}

type fakeLogger struct {
	messages []string
}

func (logger *fakeLogger) Printf(format string, args ...any) {
	logger.messages = append(logger.messages, fmt.Sprintf(format, args...))
}

func TestHandlerRunAppliesPendingMigrationsThenInitializes(t *testing.T) {
	migrations := &fakeMigrationManager{
		status: MigrationStatus{CurrentVersion: 2, LatestVersion: 4, HasCurrent: true, HasMigrations: true},
	}
	initializer := &fakeInitializer{needed: true}
	logger := &fakeLogger{}
	handler := NewHandler(migrations, initializer, logger)

	if err := handler.Run(context.Background()); err != nil {
		t.Fatalf("Run() error = %v", err)
	}
	if migrations.upCalls != 1 {
		t.Fatalf("migration Up() calls = %d, want 1", migrations.upCalls)
	}
	if initializer.calls != 1 {
		t.Fatalf("initializer calls = %d, want 1", initializer.calls)
	}
}

func TestHandlerRunSkipsCompletedWork(t *testing.T) {
	migrations := &fakeMigrationManager{
		status: MigrationStatus{CurrentVersion: 4, LatestVersion: 4, HasCurrent: true, HasMigrations: true},
	}
	initializer := &fakeInitializer{}
	handler := NewHandler(migrations, initializer, &fakeLogger{})

	if err := handler.Run(context.Background()); err != nil {
		t.Fatalf("Run() error = %v", err)
	}
	if migrations.upCalls != 0 || initializer.calls != 0 {
		t.Fatalf("Run() performed unnecessary work: migration calls=%d init calls=%d", migrations.upCalls, initializer.calls)
	}
}

func TestHandlerRunStopsOnDirtyMigration(t *testing.T) {
	migrations := &fakeMigrationManager{status: MigrationStatus{CurrentVersion: 3, Dirty: true}}
	initializer := &fakeInitializer{needed: true}
	handler := NewHandler(migrations, initializer, &fakeLogger{})

	if err := handler.Run(context.Background()); err == nil {
		t.Fatal("Run() error = nil, want dirty migration error")
	}
	if migrations.upCalls != 0 || initializer.calls != 0 {
		t.Fatalf("Run() continued after dirty migration: migration calls=%d init calls=%d", migrations.upCalls, initializer.calls)
	}
}

func TestHandlerRunStopsAfterMigrationFailure(t *testing.T) {
	migrationErr := errors.New("migration failed")
	migrations := &fakeMigrationManager{
		status: MigrationStatus{LatestVersion: 1, HasMigrations: true},
		upErr:  migrationErr,
	}
	initializer := &fakeInitializer{needed: true}
	handler := NewHandler(migrations, initializer, &fakeLogger{})

	if err := handler.Run(context.Background()); !errors.Is(err, migrationErr) {
		t.Fatalf("Run() error = %v, want %v", err, migrationErr)
	}
	if initializer.calls != 0 {
		t.Fatalf("initializer calls = %d, want 0", initializer.calls)
	}
}

func TestHandlerRunReturnsInitializationFailure(t *testing.T) {
	initializationErr := errors.New("initialization failed")
	migrations := &fakeMigrationManager{
		status: MigrationStatus{CurrentVersion: 1, LatestVersion: 1, HasCurrent: true, HasMigrations: true},
	}
	initializer := &fakeInitializer{needed: true, initializeErr: initializationErr}
	handler := NewHandler(migrations, initializer, &fakeLogger{})

	if err := handler.Run(context.Background()); !errors.Is(err, initializationErr) {
		t.Fatalf("Run() error = %v, want %v", err, initializationErr)
	}
}

func TestHandlerRunRejectsDatabaseAheadOfMigrations(t *testing.T) {
	migrations := &fakeMigrationManager{
		status: MigrationStatus{CurrentVersion: 3, LatestVersion: 2, HasCurrent: true, HasMigrations: true},
	}
	initializer := &fakeInitializer{needed: true}
	handler := NewHandler(migrations, initializer, &fakeLogger{})

	if err := handler.Run(context.Background()); err == nil {
		t.Fatal("Run() error = nil, want migration source mismatch")
	}
	if initializer.calls != 0 {
		t.Fatalf("initializer calls = %d, want 0", initializer.calls)
	}
}
