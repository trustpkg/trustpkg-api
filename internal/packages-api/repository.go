package packagesApi

import (
	"context"
	_ "embed"
	"encoding/json"
	"fmt"
	"time"

	"github.com/trustpkg/trustpkg-api/db"
)

var (
	//go:embed sql/select_packages.sql
	querySelectPackages string
)

func redisAllowRequest(ctx context.Context, keyPrefix string, clientIP string, window int64, limit int, expiration time.Duration) (bool, error) {
	if db.Redis == nil {
		return false, fmt.Errorf("Redis client is not initialized")
	}

	key := rateLimitKey(keyPrefix, clientIP, window)
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

func dbSelectPackages(payload dbSelectPackagesPayload) ([]PackageListDocument, error) {
	if db.Pool == nil {
		return nil, fmt.Errorf("Postgres pool is not initialized")
	}

	rows, err := db.Pool.Query(payload.ctx, querySelectPackages, payload.page, payload.limit, payload.sortBy, payload.order, payload.ecosystem)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	documents := make([]PackageListDocument, 0, payload.limit)
	for rows.Next() {
		var document PackageListDocument
		var vulnerabilityCounts []byte
		var vulnerabilities []byte

		if err := rows.Scan(
			&document.ID,
			&document.Name,
			&document.Slug,
			&document.Ecosystem,
			&document.Popularity,
			&document.IsDeprecated,
			&vulnerabilityCounts,
			&vulnerabilities,
		); err != nil {
			return nil, err
		}

		document.VulnerabilityCounts = make(map[string]int)
		if len(vulnerabilityCounts) > 0 {
			if err := json.Unmarshal(vulnerabilityCounts, &document.VulnerabilityCounts); err != nil {
				return nil, err
			}
		}

		document.Vulnerabilities = make(map[string][]Vulnerability)
		if len(vulnerabilities) > 0 {
			if err := json.Unmarshal(vulnerabilities, &document.Vulnerabilities); err != nil {
				return nil, err
			}
		}

		documents = append(documents, document)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return documents, nil
}
