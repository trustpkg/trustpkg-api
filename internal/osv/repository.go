package osv

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	_ "embed"

	"github.com/jackc/pgx/v5"
	"github.com/redis/go-redis/v9"
	"github.com/trustpkg/trustpkg-api/db"
)

const (
	osvReplicationRedisKeyPrefix  = "osv-replication"
	osvReplicationQueueRedisField = "queue"
)

var (
	//go:embed sql/upsert_vulnerability.sql
	queryUpsertVulnerability string
	//go:embed sql/upsert_package.sql
	queryUpsertPackage string
	//go:embed sql/delete_affected.sql
	queryDeleteAffected string
	//go:embed sql/delete_references.sql
	queryDeleteReferences string
	//go:embed sql/insert_affected.sql
	queryInsertAffected string
	//go:embed sql/insert_reference.sql
	queryInsertReference string
	//go:embed sql/select_replication_state.sql
	querySelectReplicationState string
	//go:embed sql/insert_replication_state.sql
	queryInsertReplicationState string
	//go:embed sql/update_replication_check.sql
	queryUpdateReplicationCheck string
	//go:embed sql/update_full_replication_state.sql
	queryUpdateFullReplicationState string
	//go:embed sql/update_incremental_replication_state.sql
	queryUpdateIncrementalReplicationState string
	//go:embed sql/update_withdrawn_vulnerability.sql
	queryUpdateWithdrawnVulnerability string
	//go:embed sql/select_vulnerability_history_id.sql
	querySelectVulnerabilityHistoryID string
	//go:embed sql/insert_vulnerability_history_affected.sql
	queryInsertVulnerabilityHistoryAffected string
	//go:embed sql/insert_withdrawn_vulnerability_history.sql
	queryInsertWithdrawnVulnerabilityHistory string
)

func dbGetOsvReplicationState(ctx context.Context, ecosystem string) (*OsvReplicationState, error) {
	if _, err := db.Pool.Exec(ctx, queryInsertReplicationState, ecosystem); err != nil {
		return nil, err
	}

	state := &OsvReplicationState{}
	err := db.Pool.QueryRow(ctx, querySelectReplicationState, ecosystem).Scan(
		&state.Ecosystem,
		&state.IsFirstSynchronised,
		&state.LastModifiedAt,
		&state.LastCheckedAt,
		&state.NextCheckAt,
		&state.LastSynchronizedAt,
	)
	if err != nil {
		return nil, err
	}

	return state, nil
}

func dbClaimOsvReplicationCheck(ctx context.Context, ecosystem string, nextCheckAt time.Time) (bool, error) {
	var claimed bool
	err := db.Pool.QueryRow(ctx, queryUpdateReplicationCheck, ecosystem, nextCheckAt).Scan(&claimed)
	return claimed, err
}

func dbCompleteOsvFullReplication(ctx context.Context, ecosystem string, modifiedAt *time.Time) error {
	_, err := db.Pool.Exec(ctx, queryUpdateFullReplicationState, ecosystem, modifiedAt)
	return err
}

func dbCompleteOsvIncrementalReplication(ctx context.Context, ecosystem string, modifiedAt *time.Time) error {
	_, err := db.Pool.Exec(ctx, queryUpdateIncrementalReplicationState, ecosystem, modifiedAt)
	return err
}

func dbMarkVulnerabilityWithdrawn(ctx context.Context, transaction pgx.Tx, osvID string, withdrawnAt time.Time) error {
	var historyID int64
	err := transaction.QueryRow(ctx, queryInsertWithdrawnVulnerabilityHistory, osvID, withdrawnAt).Scan(&historyID)
	if errors.Is(err, pgx.ErrNoRows) {
		_, updateErr := transaction.Exec(ctx, queryUpdateWithdrawnVulnerability, osvID, withdrawnAt)
		return updateErr
	}
	if err != nil {
		return err
	}

	_, err = transaction.Exec(ctx, queryUpdateWithdrawnVulnerability, osvID, withdrawnAt)
	return err
}

func dbSaveVulnerability(ctx context.Context, vulnerability ParsedVulnerability) error {
	transaction, err := db.Pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer transaction.Rollback(ctx)

	if err := dbSaveVulnerabilityTx(ctx, transaction, vulnerability); err != nil {
		return err
	}

	return transaction.Commit(ctx)
}

func dbSaveVulnerabilityTx(ctx context.Context, transaction pgx.Tx, vulnerability ParsedVulnerability) error {
	var vulnerabilityID int64
	err := transaction.QueryRow(ctx, queryUpsertVulnerability,
		vulnerability.OSVID, vulnerability.CVEID, vulnerability.Summary,
		vulnerability.Description, vulnerability.Severity, vulnerability.CVSSScore,
		vulnerability.CVSSVector, vulnerability.PublishedAt, vulnerability.ModifiedAt, vulnerability.WithdrawnAt,
	).Scan(&vulnerabilityID)
	if err != nil {
		return err
	}
	if err := dbInsertVulnerabilityHistory(ctx, transaction, vulnerabilityID, vulnerability); err != nil {
		return err
	}
	if _, err = transaction.Exec(ctx, queryDeleteAffected, vulnerabilityID); err != nil {
		return err
	}
	if _, err = transaction.Exec(ctx, queryDeleteReferences, vulnerabilityID); err != nil {
		return err
	}
	for _, affected := range vulnerability.Affected {
		var packageID int64
		err = transaction.QueryRow(ctx, queryUpsertPackage, affected.PackageName, affected.Ecosystem).Scan(&packageID)
		if err != nil {
			return err
		}
		if _, err = transaction.Exec(ctx, queryInsertAffected, vulnerabilityID, packageID,
			affected.IntroducedVersion, affected.FixedVersion, affected.LastAffectedVersion); err != nil {
			return err
		}
	}
	for _, reference := range vulnerability.References {
		if _, err = transaction.Exec(ctx, queryInsertReference, vulnerabilityID, reference); err != nil {
			return err
		}
	}

	return nil
}

func dbInsertVulnerabilityHistory(ctx context.Context, transaction pgx.Tx, vulnerabilityID int64, vulnerability ParsedVulnerability) error {
	affected, err := json.Marshal(vulnerability.Affected)
	if err != nil {
		return err
	}

	sourceModifiedAt := time.Now().UTC()
	if vulnerability.ModifiedAt != nil {
		sourceModifiedAt = vulnerability.ModifiedAt.UTC()
	}

	var historyID int64
	err = transaction.QueryRow(ctx, querySelectVulnerabilityHistoryID,
		vulnerabilityID,
		vulnerability.OSVID,
		vulnerability.CVEID,
		vulnerability.Summary,
		vulnerability.Description,
		vulnerability.Severity,
		vulnerability.CVSSScore,
		vulnerability.CVSSVector,
		vulnerability.PublishedAt,
		vulnerability.ModifiedAt,
		vulnerability.WithdrawnAt,
		sourceModifiedAt,
		affected,
	).Scan(&historyID)
	if err != nil {
		return err
	}

	for _, affectedPackage := range vulnerability.Affected {
		var packageID int64
		err = transaction.QueryRow(ctx, queryUpsertPackage, affectedPackage.PackageName, affectedPackage.Ecosystem).Scan(&packageID)
		if err != nil {
			return err
		}
		_, err = transaction.Exec(ctx, queryInsertVulnerabilityHistoryAffected,
			historyID,
			packageID,
			affectedPackage.Ecosystem,
			affectedPackage.IntroducedVersion,
			affectedPackage.FixedVersion,
			affectedPackage.LastAffectedVersion,
		)
		if err != nil {
			return err
		}
	}

	return nil
}

func redisGetOsvQueue(ctx context.Context, ecosystem string) ([]ModifiedVulnerability, error) {
	data, err := db.Redis.HGet(ctx, osvReplicationRedisKey(ecosystem), osvReplicationQueueRedisField).Result()
	if errors.Is(err, redis.Nil) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	var queue []ModifiedVulnerability
	if err := json.Unmarshal([]byte(data), &queue); err != nil {
		return nil, fmt.Errorf("decode OSV Redis queue: %w", err)
	}

	return queue, nil
}

func redisSetOsvQueue(ctx context.Context, ecosystem string, queue []ModifiedVulnerability) error {
	data, err := json.Marshal(queue)
	if err != nil {
		return err
	}

	pipe := db.Redis.Pipeline()
	key := osvReplicationRedisKey(ecosystem)
	pipe.HSet(ctx, key, osvReplicationQueueRedisField, data)
	pipe.Expire(ctx, key, time.Hour*24)
	_, err = pipe.Exec(ctx)
	return err
}

func osvReplicationRedisKey(ecosystem string) string {
	return fmt.Sprintf("%s:%s", osvReplicationRedisKeyPrefix, ecosystem)
}
