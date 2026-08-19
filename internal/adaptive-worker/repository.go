package adaptiveWorker

import (
	"context"
	"time"

	"github.com/trustpkg/trustpkg-api/db"
)

var (
	dayDuration = 24 * time.Hour
)

func getWorkerCycle(ctx context.Context) (int, error) {
	if cycle, err := db.Redis.HGet(ctx, adaptiveWorkerRedisKey, "cycle").Int(); err != nil {
		return 0, err
	} else {
		return cycle, nil
	}
}

func setWorkerCycle(ctx context.Context, cycle int) error {
	pipe := db.Redis.TxPipeline()

	pipe.HSet(ctx, adaptiveWorkerRedisKey, "cycle", cycle)
	pipe.Expire(ctx, adaptiveWorkerRedisKey, dayDuration)

	_, err := pipe.Exec(ctx)

	return err
}
