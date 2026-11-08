package databaseService

import (
	"bytes"
	"context"
	"testing"
)

func TestExecuteStartupRunsHandler(t *testing.T) {
	migrations := &fakeMigrationManager{
		status: MigrationStatus{CurrentVersion: 1, LatestVersion: 1, HasCurrent: true, HasMigrations: true},
	}
	initializer := &fakeInitializer{}
	handler := NewHandler(migrations, initializer, &fakeLogger{})

	err := Execute(context.Background(), Command{Action: actionStartup}, migrations, handler, &bytes.Buffer{})
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if migrations.upCalls != 0 || initializer.calls != 0 {
		t.Fatalf("startup executed unexpected work: migrations=%d init=%d", migrations.upCalls, initializer.calls)
	}
}

func TestExecuteMigrationVersionReportsCurrentAndLatest(t *testing.T) {
	migrations := &fakeMigrationManager{
		status: MigrationStatus{CurrentVersion: 2, LatestVersion: 3, HasCurrent: true, HasMigrations: true, Dirty: true},
	}
	var output bytes.Buffer

	err := Execute(context.Background(), Command{Action: actionVersion}, migrations, nil, &output)
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	want := "Current migration version: 2 (latest: 3, dirty: true)\n"
	if output.String() != want {
		t.Fatalf("Execute() output = %q, want %q", output.String(), want)
	}
}

func TestExecuteHelpDoesNotRequireDatabase(t *testing.T) {
	var output bytes.Buffer

	err := Execute(context.Background(), Command{Action: ActionHelp}, nil, nil, &output)
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if !bytes.Contains(output.Bytes(), []byte("migrate version")) ||
		!bytes.Contains(output.Bytes(), []byte("first-init")) ||
		!bytes.Contains(output.Bytes(), []byte("may remove database tables and data")) ||
		!bytes.Contains(output.Bytes(), []byte("without executing migration SQL")) {
		t.Fatalf("Execute() help output is incomplete: %q", output.String())
	}
}
