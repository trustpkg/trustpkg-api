package databaseService

import (
	"errors"
	"fmt"
	"io"
)

func WriteHelp(output io.Writer) error {
	if output == nil {
		return errors.New("help output is nil")
	}

	_, err := fmt.Fprintln(output, `Usage: database-service [command]

With no command, the service checks the database migration state and applies all
pending migrations. It then checks whether the npm OSV dataset has been initialized
and imports the full dataset if it has not. This is the recommended startup mode.

Commands:
  help
      Show this help without connecting to the database.

  migrate version
      Show the database's current migration version, the latest migration available
      in this service, and whether the current migration state is dirty.

  migrate up
      Apply all migrations newer than the database's current migration version.
      Does not run the OSV first initialization.

  migrate down
      Revert all applied migrations. This runs every available down migration and
      may remove database tables and data. Does not run the OSV first initialization.

  migrate force <version>
      Set the recorded migration version without executing migration SQL. Use this
      only to repair a migration state after verifying the actual database schema.
      Use -1 to mark the database as having no applied migrations.

  first-init
      Apply pending migrations, then check the npm OSV replication state. If the
      initial OSV import is not complete, download and import the full npm dataset.
      If initialization is already complete, no import is run. This command does
      not run Redis-dependent replication jobs.

Examples:
  podman exec database_service help
  podman exec database_service migrate version
  podman exec database_service migrate up
  podman exec database_service first-init`)
	return err
}
