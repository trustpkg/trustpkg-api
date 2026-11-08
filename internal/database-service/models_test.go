package databaseService

import "testing"

func TestMigrationStatusPending(t *testing.T) {
	tests := []struct {
		name   string
		status MigrationStatus
		want   bool
	}{
		{name: "no migrations", status: MigrationStatus{}},
		{name: "no migration applied", status: MigrationStatus{LatestVersion: 1, HasMigrations: true}, want: true},
		{name: "older migration applied", status: MigrationStatus{CurrentVersion: 1, LatestVersion: 2, HasCurrent: true, HasMigrations: true}, want: true},
		{name: "at latest migration", status: MigrationStatus{CurrentVersion: 2, LatestVersion: 2, HasCurrent: true, HasMigrations: true}},
		{name: "current exceeds latest", status: MigrationStatus{CurrentVersion: 3, LatestVersion: 2, HasCurrent: true, HasMigrations: true}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.status.Pending(); got != tt.want {
				t.Fatalf("Pending() = %t, want %t", got, tt.want)
			}
		})
	}
}
