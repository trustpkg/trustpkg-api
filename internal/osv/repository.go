package osv

import (
	"context"
	"embed"

	"github.com/trustpkg/trustpkg-api/db"
)

//go:embed sql/*.sql
var sqlFiles embed.FS

const (
	queryUpsertVulnerability = "sql/upsert_vulnerability.sql"
	queryUpsertPackage       = "sql/upsert_package.sql"
	queryDeleteAffected      = "sql/delete_vulnerability_children.sql"
	queryDeleteReferences    = "sql/delete_vulnerability_references.sql"
	queryInsertAffected      = "sql/insert_affected.sql"
	queryInsertReference     = "sql/insert_reference.sql"
)

func saveVulnerability(ctx context.Context, vulnerability ParsedVulnerability) error {
	transaction, err := db.Pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer transaction.Rollback(ctx)

	var vulnerabilityID int64
	err = transaction.QueryRow(ctx, readSQL(queryUpsertVulnerability),
		vulnerability.OSVID, vulnerability.CVEID, vulnerability.Summary,
		vulnerability.Description, vulnerability.Severity, vulnerability.CVSSScore,
		vulnerability.CVSSVector, vulnerability.PublishedAt, vulnerability.ModifiedAt,
	).Scan(&vulnerabilityID)
	if err != nil {
		return err
	}
	if _, err = transaction.Exec(ctx, readSQL(queryDeleteAffected), vulnerabilityID); err != nil {
		return err
	}
	if _, err = transaction.Exec(ctx, readSQL(queryDeleteReferences), vulnerabilityID); err != nil {
		return err
	}
	for _, affected := range vulnerability.Affected {
		var packageID int64
		err = transaction.QueryRow(ctx, readSQL(queryUpsertPackage), affected.PackageName, affected.Ecosystem).Scan(&packageID)
		if err != nil {
			return err
		}
		if _, err = transaction.Exec(ctx, readSQL(queryInsertAffected), vulnerabilityID, packageID,
			affected.IntroducedVersion, affected.FixedVersion, affected.LastAffectedVersion); err != nil {
			return err
		}
	}
	for _, reference := range vulnerability.References {
		if _, err = transaction.Exec(ctx, readSQL(queryInsertReference), vulnerabilityID, reference); err != nil {
			return err
		}
	}

	return transaction.Commit(ctx)
}

func readSQL(name string) string {
	contents, err := sqlFiles.ReadFile(name)
	if err != nil {
		panic(err)
	}
	return string(contents)
}
