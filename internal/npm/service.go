package npm

import (
	"context"

	"github.com/trustpkg/trustpkg-api/db"
)

func ReplicationPipeline() error {
	ctx := context.Background()

	since, err := dbSelectLastSequence(ctx)
	if err != nil {
		return err
	}

	response, err := fetchNpmPackages(since)
	if err != nil {
		return err
	}

	length, err := savePackagesPage(ctx, response.Results)
	if err != nil {
		return err
	}

	npmReplicatorLog.Printf("changes synchronized: packages=%d, response_sequence=%d", length, response.LastSeq)

	transaction, err := db.Pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer transaction.Rollback(ctx)

	if err := dbInsertLastSequence(ctx, transaction, response.LastSeq); err != nil {
		return err
	}

	return transaction.Commit(ctx)
}
