package databaseService

import (
	"context"
	"errors"
	"fmt"
	"io"
	"strconv"
)

const (
	ActionHelp    = "help"
	actionStartup = "startup"
	actionUp      = "up"
	actionDown    = "down"
	actionForce   = "force"
	actionVersion = "version"
	actionInit    = "first-init"
)

func ParseCommand(args []string) (Command, error) {
	if len(args) == 0 {
		return Command{Action: actionStartup}, nil
	}
	if len(args) == 1 && args[0] == actionInit {
		return Command{Action: actionInit}, nil
	}
	if len(args) == 1 && args[0] == ActionHelp {
		return Command{Action: ActionHelp}, nil
	}
	if len(args) < 2 || args[0] != "migrate" {
		return Command{}, usageError()
	}

	switch args[1] {
	case actionUp, actionDown, actionVersion:
		if len(args) != 2 {
			return Command{}, usageError()
		}
		return Command{Action: args[1]}, nil
	case actionForce:
		if len(args) != 3 {
			return Command{}, usageError()
		}
		version, err := strconv.Atoi(args[2])
		if err != nil || version < -1 {
			return Command{}, fmt.Errorf("invalid migration version %q: expected an integer greater than or equal to -1", args[2])
		}
		return Command{Action: actionForce, Version: version}, nil
	default:
		return Command{}, usageError()
	}
}

func Execute(ctx context.Context, command Command, migrations MigrationManager, handler *Handler, output io.Writer) error {
	switch command.Action {
	case ActionHelp:
		return WriteHelp(output)
	case actionStartup, actionInit:
		if handler == nil {
			return errors.New("database startup handler is nil")
		}
		return handler.Run(ctx)
	case actionUp:
		if migrations == nil {
			return errors.New("migration manager is nil")
		}
		return migrations.Up()
	case actionDown:
		if migrations == nil {
			return errors.New("migration manager is nil")
		}
		return migrations.Down()
	case actionForce:
		if migrations == nil {
			return errors.New("migration manager is nil")
		}
		return migrations.Force(command.Version)
	case actionVersion:
		if migrations == nil {
			return errors.New("migration manager is nil")
		}
		return WriteMigrationStatus(output, migrations)
	default:
		return usageError()
	}
}

func WriteMigrationStatus(output io.Writer, migrations MigrationManager) error {
	if output == nil {
		return errors.New("migration status output is nil")
	}
	if migrations == nil {
		return errors.New("migration manager is nil")
	}
	status, err := migrations.Status()
	if err != nil {
		return fmt.Errorf("read migration status: %w", err)
	}
	if !status.HasMigrations {
		_, err = fmt.Fprintln(output, "No migrations are available.")
		return err
	}
	if !status.HasCurrent {
		_, err = fmt.Fprintf(output, "No migrations have been applied. Latest available version: %d (dirty: false)\n", status.LatestVersion)
		return err
	}
	_, err = fmt.Fprintf(output, "Current migration version: %d (latest: %d, dirty: %t)\n", status.CurrentVersion, status.LatestVersion, status.Dirty)
	return err
}

func usageError() error {
	return errors.New("usage: database-service [help | migrate <up|down|version|force <version>> | first-init]")
}
