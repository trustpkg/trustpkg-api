package databaseService

import "context"

type MigrationManager interface {
	Status() (MigrationStatus, error)
	Up() error
	Down() error
	Force(version int) error
}

type Initializer interface {
	Needed(ctx context.Context) (bool, error)
	Initialize(ctx context.Context) error
}

type Logger interface {
	Printf(format string, args ...any)
}
