package typesense

import (
	"context"
	"encoding/json"
	"errors"

	_ "embed"

	"github.com/jackc/pgx/v5"
	"github.com/redis/go-redis/v9"
	"github.com/trustpkg/trustpkg-api/db"
)

const pipelineRedisKey = "typesense-packages-pipeline"

var (
	//go:embed sql/select_typesense_state.sql
	querySelectTypesenseState string
	//go:embed sql/upsert_typesense_state.sql
	queryUpsertTypesenseState string
	//go:embed sql/select_watchlist_documents.sql
	querySelectWatchlistDocuments string
	//go:embed sql/select_database_documents.sql
	querySelectDatabaseDocuments string
)

func redisGetPipelineState(ctx context.Context) (*pipelineState, error) {
	data, err := db.Redis.Get(ctx, pipelineRedisKey).Bytes()
	if err == nil {
		state := &pipelineState{}
		if err := json.Unmarshal(data, state); err == nil {
			return state, nil
		}
	} else if !errors.Is(err, redis.Nil) {
		return nil, err
	}

	state := &pipelineState{}
	err = db.Pool.QueryRow(ctx, querySelectTypesenseState).Scan(
		&state.Phase, &state.WatchlistOffset, &state.LastPackageID,
		&state.Cycle, &state.DocumentsUploaded,
		&state.WatchlistDocumentsUploaded, &state.DatabaseDocumentsUploaded,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		state.Phase = pipelinePhaseWatchlist
		return state, nil
	}
	return state, err
}

func redisSetPipelineState(ctx context.Context, state *pipelineState, status, lastError string) error {
	state.Status = status
	state.LastError = lastError
	data, err := json.Marshal(state)
	if err != nil {
		return err
	}
	return db.Redis.Set(ctx, pipelineRedisKey, data, 0).Err()
}

func dbSavePipelineState(ctx context.Context, state *pipelineState, status, lastError string) error {
	if err := redisSetPipelineState(ctx, state, status, lastError); err != nil {
		return err
	}
	_, err := db.Pool.Exec(ctx, queryUpsertTypesenseState,
		state.Phase, state.WatchlistOffset, state.LastPackageID, state.Cycle,
		state.DocumentsUploaded, state.WatchlistDocumentsUploaded,
		state.DatabaseDocumentsUploaded, status, lastError,
	)
	return err
}

func dbSelectWatchlistDocuments(ctx context.Context, watchlist []string, offset, limit int) ([]packageDocument, error) {
	rows, err := db.Pool.Query(ctx, querySelectWatchlistDocuments, watchlist, offset, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return dbScanPackageDocuments(rows)
}

func dbSelectDatabaseDocuments(ctx context.Context, lastID int64, limit int) ([]packageDocument, error) {
	rows, err := db.Pool.Query(ctx, querySelectDatabaseDocuments, lastID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return dbScanPackageDocuments(rows)
}

func dbScanPackageDocuments(rows pgx.Rows) ([]packageDocument, error) {
	var documents []packageDocument
	for rows.Next() {
		var document packageDocument
		var vulnerabilities []byte
		if err := rows.Scan(&document.ID, &document.Name, &document.Ecosystem, &document.Popularity, &document.IsDeprecated, &vulnerabilities); err != nil {
			return nil, err
		}
		document.Vulnerabilities = make(vulnerabilitiesMap)
		if len(vulnerabilities) > 0 {
			if err := json.Unmarshal(vulnerabilities, &document.Vulnerabilities); err != nil {
				return nil, err
			}
		}
		documents = append(documents, document)
	}
	return documents, rows.Err()
}
