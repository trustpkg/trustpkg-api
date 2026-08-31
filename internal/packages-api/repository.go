package packagesApi

import (
	"context"
	"fmt"
	"time"

	"github.com/trustpkg/trustpkg-api/db"
)

func redisAllowSearchRequest(ctx context.Context, clientIP string, window int64, limit int, expiration time.Duration) (bool, error) {
	if db.Redis == nil {
		return false, fmt.Errorf("Redis client is not initialized")
	}

	key := fmt.Sprintf("packages-api:search:rate:%s:%d", clientIP, window)
	count, err := db.Redis.Incr(ctx, key).Result()
	if err != nil {
		return false, err
	}
	if count == 1 {
		if err := db.Redis.Expire(ctx, key, expiration).Err(); err != nil {
			return false, err
		}
	}

	return count <= int64(limit), nil
}
