package npm

import (
	"context"

	"github.com/trustpkg/trustpkg-api/db"
)

var pool = db.Pool

const (
	queryInsertLastSequence = `INSERT INTO last_sequence (sequence) VALUES ($1)`
	querySelectLastSequence = `SELECT sequence FROM last_sequence ORDER BY sequence DESC LIMIT 1`
)

func insertLastSequence(ctx context.Context, sequence int) error {
	_, err := pool.Exec(ctx, queryInsertLastSequence, sequence)
	
	return err
}

func selectLastSequence(ctx context.Context) (int, error) {
	var sequence int

	err := pool.QueryRow(ctx, querySelectLastSequence).Scan(&sequence)
	if err != nil {
		return 0, err
	}

	return sequence, nil
}