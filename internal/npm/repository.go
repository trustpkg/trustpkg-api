package npm

import (
	"context"
	"errors"

	_ "embed"

	"github.com/jackc/pgx/v5"
	"github.com/trustpkg/trustpkg-api/db"
)

var (
	//go:embed sql/delete_packages_batch.sql
	queryDeletePackagesBatch string

	//go:embed sql/insert_last_sequence.sql
	queryInsertLastSequence string

	//go:embed sql/insert_packages_batch.sql
	queryInsertPackagesBatch string

	//go:embed sql/select_last_sequence.sql
	querySelectLastSequence string
)

func insertLastSequence(ctx context.Context, transaction pgx.Tx, sequence int) error {
	_, err := transaction.Exec(ctx, queryInsertLastSequence, sequence)

	return err
}

func selectLastSequence(ctx context.Context) (int, error) {
	var sequence int

	err := db.Pool.QueryRow(ctx, querySelectLastSequence).Scan(&sequence)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return 0, nil
		}

		return 0, err
	}

	return sequence, nil
}

func dropPackagesBatch(ctx context.Context, transaction pgx.Tx, payload dbDropPackagesBatchPayload) error {
	_, err := transaction.Exec(ctx, queryDeletePackagesBatch, payload.packages)

	return err
}

func insertPackagesBatch(ctx context.Context, transaction pgx.Tx, payload dbInsertPackagesBatchPayload) error {
	_, err := transaction.Exec(ctx, queryInsertPackagesBatch, payload.packages)

	return err
}
