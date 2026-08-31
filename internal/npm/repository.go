package npm

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	_ "embed"

	"github.com/jackc/pgx/v5"
	"github.com/redis/go-redis/v9"
	"github.com/trustpkg/trustpkg-api/db"
)

var (
	//go:embed sql/insert_mark_deprecated_batch.sql
	queryInsertMarkDeprecatedBatch string

	//go:embed sql/insert_last_sequence.sql
	queryInsertLastSequence string

	//go:embed sql/insert_packages_batch.sql
	queryInsertPackagesBatch string

	//go:embed sql/select_last_sequence.sql
	querySelectLastSequence string

	//go:embed sql/insert_packages_popularity_batch.sql
	queryInsertPackagesPopularityBatch string

	//go:embed sql/select_scoped_packages.sql
	querySelectScopedPackages string

	//go:embed sql/select_unscoped_packages.sql
	querySelectUnscopedPackages string
	//go:embed sql/mark_packages_need_update.sql
	queryMarkPackagesNeedUpdate string
)

const (
	oneWeekDuration = time.Hour * 24 * 7
	oneHourDuration = time.Hour

	cooldownOneDayDuration = time.Hour * 24

	packagesWatchlistRedisKey     string = "packages-watchlist"
	areGatheredRedisField         string = "are-gathered"
	lastPackageRedisField         string = "last-package"
	queueRedisField               string = "queue"
	allUnscopedGatheredRedisField string = "all-unscoped-gathered"

	basicPackagesRedisKey     string = "basic-packages"
	lastBasicPackageFieldKey  string = "last-package"
	basicQueueRedisField      string = "queue"
	shouldRunScopedRedisField string = "should-run-scoped"

	collectorCooldown string = "collector-cooldown"
)

func dbInsertLastSequence(ctx context.Context, transaction pgx.Tx, sequence int) error {
	_, err := transaction.Exec(ctx, queryInsertLastSequence, sequence)

	return err
}

func dbSelectLastSequence(ctx context.Context) (int, error) {
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

func dbInsertMarkDeprecatedBatch(ctx context.Context, transaction pgx.Tx, payload dbInsertMarkDeprecatedBatchPayload) error {
	_, err := transaction.Exec(ctx, queryInsertMarkDeprecatedBatch, payload.packages)

	return err
}

func dbInsertPackagesBatch(ctx context.Context, transaction pgx.Tx, payload dbInsertPackagesBatchPayload) error {
	_, err := transaction.Exec(ctx, queryInsertPackagesBatch, payload.packages)

	return err
}

func dbInsertPackagesPopularity(ctx context.Context, transaction pgx.Tx, packagesDownloads map[string]int64) error {
	packagesDownloadsJSON, err := json.Marshal(packagesDownloads)
	if err != nil {
		return err
	}

	_, err = transaction.Exec(ctx, queryInsertPackagesPopularityBatch, packagesDownloadsJSON)

	return err
}

func dbMarkPackagesNeedUpdate(ctx context.Context, packages []string) error {
	_, err := db.Pool.Exec(ctx, queryMarkPackagesNeedUpdate, packages)
	return err
}

func dbSelectPackages(ctx context.Context, query string) ([]string, error) {
	var packages []string

	rows, err := db.Pool.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var pkg string
		if err := rows.Scan(&pkg); err != nil {
			return nil, err
		}
		packages = append(packages, pkg)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return packages, nil
}

func dbSelectScopedPackages(ctx context.Context) ([]string, error) {
	return dbSelectPackages(ctx, querySelectScopedPackages)
}

func dbSelectUnscopedPackages(ctx context.Context) ([]string, error) {
	return dbSelectPackages(ctx, querySelectUnscopedPackages)
}

func redisSetWatchlistGathered(
	ctx context.Context,
	pipe redis.Pipeliner,
) {
	pipe.HSet(ctx, packagesWatchlistRedisKey, areGatheredRedisField, true)
	pipe.Expire(ctx, packagesWatchlistRedisKey, oneWeekDuration)
}

func redisGetWatchlistGathered(
	ctx context.Context,
	pipe redis.Pipeliner,
) *redis.StringCmd {
	return pipe.HGet(ctx, packagesWatchlistRedisKey, areGatheredRedisField)
}

func redisSetLastPackage(ctx context.Context, pipe redis.Pipeliner, value string) {
	pipe.HSet(ctx, packagesWatchlistRedisKey, lastPackageRedisField, value)
}

func redisGetLastPackage(ctx context.Context, pipe redis.Pipeliner) *redis.StringCmd {
	return pipe.HGet(ctx, packagesWatchlistRedisKey, lastPackageRedisField)
}

func redisSetPackagesQueue(ctx context.Context, pipe redis.Pipeliner, queue []string) error {
	data, err := json.Marshal(queue)
	if err != nil {
		return err
	}

	pipe.HSet(ctx, packagesWatchlistRedisKey, queueRedisField, data)

	return nil
}

func redisGetPackagesQueue(ctx context.Context, pipe redis.Pipeliner) *redis.StringCmd {
	return pipe.HGet(ctx, packagesWatchlistRedisKey, queueRedisField)
}

func redisSetAllUnscopedGathered(ctx context.Context, pipe redis.Pipeliner) {
	pipe.HSet(ctx, packagesWatchlistRedisKey, allUnscopedGatheredRedisField, true)
	pipe.Expire(ctx, packagesWatchlistRedisKey, oneWeekDuration)
}

func redisGetAllUnscopedGathered(ctx context.Context, pipe redis.Pipeliner) *redis.StringCmd {
	return pipe.HGet(ctx, packagesWatchlistRedisKey, allUnscopedGatheredRedisField)
}

func redisGetBasicState(ctx context.Context) baseCollectionState {
	pipe := db.Redis.Pipeline()

	lastPackageCmd := pipe.HGet(ctx, basicPackagesRedisKey, lastBasicPackageFieldKey)
	basicQueueCmd := pipe.HGet(ctx, basicPackagesRedisKey, basicQueueRedisField)
	shouldRunScopedCmd := pipe.HGet(ctx, basicPackagesRedisKey, shouldRunScopedRedisField)

	if _, err := pipe.Exec(ctx); err != nil && !errors.Is(err, redis.Nil) {
		return baseCollectionState{}
	}

	var lastBasePackage *string
	if value, err := lastPackageCmd.Result(); err == nil && value != "" {
		lastBasePackage = &value
	}

	var scopedQueue []string
	if data, err := basicQueueCmd.Result(); err == nil && data != "" {
		if err := json.Unmarshal([]byte(data), &scopedQueue); err != nil {
			return baseCollectionState{}
		}
	}

	shouldRunScoped, err := shouldRunScopedCmd.Bool()
	if errors.Is(err, redis.Nil) {
		shouldRunScoped = false
	} else if err != nil {
		return baseCollectionState{}
	}

	return baseCollectionState{
		lastBasePackage: lastBasePackage,
		scopedQueue:     scopedQueue,
		shouldRunScoped: shouldRunScoped,
	}
}

func redisSetBasicLastPackage(ctx context.Context, pipe redis.Pipeliner, value string) {
	pipe.HSet(ctx, basicPackagesRedisKey, lastBasicPackageFieldKey, value)
}

func redisSetBasicQueue(ctx context.Context, pipe redis.Pipeliner, queue []string) error {
	data, err := json.Marshal(queue)
	if err != nil {
		return err
	}

	pipe.HSet(ctx, basicPackagesRedisKey, basicQueueRedisField, data)
	pipe.Expire(ctx, basicPackagesRedisKey, oneHourDuration)

	return nil
}

func redisSetShouldRunScoped(ctx context.Context, pipe redis.Pipeliner, value bool) {
	pipe.HSet(ctx, basicPackagesRedisKey, shouldRunScopedRedisField, value)
}

func redisSetCollectorCooldown(ctx context.Context, pipe redis.Pipeliner) {
	pipe.Set(ctx, collectorCooldown, true, cooldownOneDayDuration)
}

func redisGetCollectorCooldown(ctx context.Context) bool {
	isCooldown, err := db.Redis.Get(ctx, collectorCooldown).Bool()
	if err != nil {
		return false
	}

	return isCooldown
}
