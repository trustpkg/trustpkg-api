package packagesApi

import (
	"context"
	_ "embed"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/jackc/pgx/v5"
	"github.com/trustpkg/trustpkg-api/db"
)

var (
	//go:embed sql/select_packages.sql
	querySelectPackages string
	//go:embed sql/count_packages.sql
	queryCountPackages string
	//go:embed sql/select_packages_vulnerabilities.sql
	querySelectPackagesVulnerabilities string
	//go:embed sql/select_packages_current_status.sql
	querySelectPackagesCurrentStatus string
	//go:embed sql/select_package.sql
	querySelectPackage string
	//go:embed sql/package_cte_vulnerabilities.sql
	queryPackageCTEVulnerabilities string
	//go:embed sql/package_column_summary.sql
	queryPackageColumnSummary string
	//go:embed sql/package_column_vulnerabilities.sql
	queryPackageColumnVulnerabilities string
	//go:embed sql/package_column_current_status.sql
	queryPackageColumnCurrentStatus string
	//go:embed sql/package_join_vulnerabilities.sql
	queryPackageJoinVulnerabilities string
)

const (
	packageCountCacheTTL          = 5 * time.Minute
	packageDetailCacheTTL         = 60 * time.Second
	packageDetailNotFoundCacheTTL = 30 * time.Second
)

func WarmPackageListCountCache(ctx context.Context, ecosystem string) error {
	_, err := dbCountPackages(dbSelectPackagesPayload{
		ctx:       ctx,
		ecosystem: ecosystem,
	})
	return err
}

func WarmPackageListVulnerabilityCountCache(ctx context.Context, ecosystem string) error {
	_, err := dbCountPackages(dbSelectPackagesPayload{
		ctx:                     ctx,
		ecosystem:               ecosystem,
		onlyWithVulnerabilities: true,
	})
	return err
}

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

	if payload.sortBy == "popularity" && payload.order == "desc" {
		documents, err := runSelectPackages(payload, "packages_data AS pd JOIN packages AS p ON p.id = pd.package_id", "p.ecosystem = $3 AND pd.popularity > 0")
		if err != nil {
			return nil, err
		}
		if len(documents) == payload.limit {
			return documents, nil
		}
	}

	return runSelectPackages(payload, "packages AS p LEFT JOIN packages_data AS pd ON pd.package_id = p.id", "p.ecosystem = $3")
}

func dbCountPackages(payload dbSelectPackagesPayload) (int64, error) {
	cacheKey := fmt.Sprintf(
		"packages-api:list:count:%s:%t:%d",
		payload.ecosystem,
		payload.onlyWithVulnerabilities,
		payload.vulnerabilityMonths,
	)
	if db.Redis != nil {
		cachedCount, err := db.Redis.Get(payload.ctx, cacheKey).Int64()
		if err == nil {
			return cachedCount, nil
		}
	}

	packageFilter := "p.ecosystem = $1"
	if payload.onlyWithVulnerabilities {
		packageFilter += vulnerabilityPackageFilter("p.id", payload.vulnerabilityMonths, "$2")
	}
	query := strings.NewReplacer(
		"/*PACKAGE_SOURCE*/", "packages AS p LEFT JOIN packages_data AS pd ON pd.package_id = p.id",
		"/*PACKAGE_FILTER*/", packageFilter,
	).Replace(queryCountPackages)
	queryArguments := []any{payload.ecosystem}
	if payload.onlyWithVulnerabilities && payload.vulnerabilityMonths > 0 {
		queryArguments = append(queryArguments, payload.vulnerabilityMonths)
	}

	var totalPackages int64
	if err := db.Pool.QueryRow(payload.ctx, query, queryArguments...).Scan(&totalPackages); err != nil {
		return 0, err
	}
	if db.Redis != nil {
		_ = db.Redis.Set(payload.ctx, cacheKey, totalPackages, packageCountCacheTTL).Err()
	}

	return totalPackages, nil
}

func runSelectPackages(payload dbSelectPackagesPayload, packageSource string, packageFilter string) ([]PackageListDocument, error) {
	if payload.onlyWithVulnerabilities {
		packageFilter += vulnerabilityPackageFilter("p.id", payload.vulnerabilityMonths, "$4")
	}
	vulnerabilityColumns := ""
	vulnerabilityDetails := ""
	if payload.includeVulnerabilities {
		vulnerabilityColumns = `,
	COALESCE(details.vulnerability_counts, '{}'::jsonb) AS vulnerability_counts,
	COALESCE(details.vulnerabilities, '{}'::jsonb) AS vulnerabilities`
		vulnerabilityDetails = querySelectPackagesVulnerabilities
	}
	currentStatusColumn := ""
	currentStatusDetails := ""
	if payload.withCurrentStatus {
		currentStatusColumn = ", COALESCE(current_status.current_statuses, '[]'::jsonb) AS current_statuses"
		currentStatusDetails = querySelectPackagesCurrentStatus
	}
	selectOrder := selectPackageOrder(payload.sortBy, payload.order)
	if payload.sortBy == "popularity" && payload.order == "desc" && strings.HasPrefix(packageSource, "packages_data AS pd JOIN") {
		selectOrder = "pd.popularity DESC, pd.package_id ASC"
	}
	query := strings.NewReplacer(
		"/*PACKAGE_SOURCE*/", packageSource,
		"/*PACKAGE_FILTER*/", packageFilter,
		"/*SELECT_ORDER*/", selectOrder,
		"/*RESULT_ORDER*/", resultPackageOrder(payload.sortBy, payload.order),
		"/*VULNERABILITY_COLUMNS*/", vulnerabilityColumns,
		"/*VULNERABILITY_DETAILS*/", vulnerabilityDetails,
		"/*CURRENT_STATUS_COLUMN*/", currentStatusColumn,
		"/*CURRENT_STATUS_DETAILS*/", currentStatusDetails,
	).Replace(querySelectPackages)
	queryArguments := []any{payload.page, payload.limit, payload.ecosystem}
	if payload.includeVulnerabilities || (payload.onlyWithVulnerabilities && payload.vulnerabilityMonths > 0) {
		queryArguments = append(queryArguments, payload.vulnerabilityMonths)
	}
	rows, err := db.Pool.Query(payload.ctx, query, queryArguments...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	documents := make([]PackageListDocument, 0, payload.limit)
	for rows.Next() {
		var document PackageListDocument
		var vulnerabilityCounts []byte
		var vulnerabilities []byte
		var currentStatuses []byte

		scanTargets := []any{
			&document.ID,
			&document.Name,
			&document.Slug,
			&document.Ecosystem,
			&document.Popularity,
			&document.IsDeprecated,
		}
		if payload.withCurrentStatus {
			scanTargets = append(scanTargets, &currentStatuses)
		}
		if payload.includeVulnerabilities {
			scanTargets = append(scanTargets, &vulnerabilityCounts, &vulnerabilities)
		}

		if err := rows.Scan(scanTargets...); err != nil {
			return nil, err
		}

		if payload.withCurrentStatus {
			statuses := make([]PackageCurrentStatus, 0)
			if len(currentStatuses) > 0 {
				if err := json.Unmarshal(currentStatuses, &statuses); err != nil {
					return nil, err
				}
			}
			document.CurrentStatuses = &statuses
		}

		if payload.includeVulnerabilities {
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
		}

		documents = append(documents, document)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return documents, nil
}

func vulnerabilityPackageFilter(packageID string, vulnerabilityMonths int, monthsPlaceholder string) string {
	if vulnerabilityMonths == 0 {
		return fmt.Sprintf(` AND EXISTS (
			SELECT 1
			FROM vulnerability_affected AS va
			WHERE va.package_id = %s
		)`, packageID)
	}

	return fmt.Sprintf(` AND EXISTS (
		SELECT 1
		FROM vulnerability_affected AS va
		JOIN vulnerabilities AS v ON v.id = va.vulnerability_id
		WHERE va.package_id = %s
		  AND COALESCE(v.published_at, v.modified_at) >=
		       date_trunc('month', CURRENT_DATE) - ((%s::int - 1) * INTERVAL '1 month')
	)`, packageID, monthsPlaceholder)
}

func selectPackageOrder(sortBy string, order string) string {
	switch {
	case sortBy == "name" && order == "asc":
		return "LOWER(p.name) ASC, p.id ASC"
	case sortBy == "name" && order == "desc":
		return "LOWER(p.name) DESC, p.id ASC"
	case sortBy == "popularity" && order == "asc":
		return "COALESCE(pd.popularity, 0) ASC, p.id ASC"
	case sortBy == "popularity" && order == "desc":
		return "COALESCE(pd.popularity, 0) DESC, p.id ASC"
	case sortBy == "ecosystem" && order == "asc":
		return "p.ecosystem ASC, p.id ASC"
	default:
		return "p.ecosystem DESC, p.id ASC"
	}
}

func resultPackageOrder(sortBy string, order string) string {
	switch {
	case sortBy == "name" && order == "asc":
		return "LOWER(s.name) ASC, s.id ASC"
	case sortBy == "name" && order == "desc":
		return "LOWER(s.name) DESC, s.id ASC"
	case sortBy == "popularity" && order == "asc":
		return "s.popularity ASC, s.id ASC"
	case sortBy == "popularity" && order == "desc":
		return "s.popularity DESC, s.id ASC"
	case sortBy == "ecosystem" && order == "asc":
		return "s.ecosystem ASC, s.id ASC"
	default:
		return "s.ecosystem DESC, s.id ASC"
	}
}

func buildPackageDetailQuery(params PackageDetailParams) (string, []any) {
	args := make([]any, 0, 6)
	bind := func(value any) string {
		args = append(args, value)
		return "$" + strconv.Itoa(len(args))
	}

	ecosystem := bind(params.Ecosystem)
	identifier := bind(params.Identifier)

	ctes := ""
	columns := ""
	joins := ""

	if params.needsVulnerabilityData() {
		filters := ""
		if !params.IncludeWithdrawn {
			filters += "\n\tAND v.withdrawn_at IS NULL"
		}
		if params.Severity != "" {
			filters += "\n\tAND UPPER(v.severity) = " + bind(params.Severity)
		}
		if params.VulnerabilityMonths > 0 {
			filters += "\n\tAND COALESCE(v.published_at, v.modified_at) >= date_trunc('month', CURRENT_DATE) - ((" +
				bind(params.VulnerabilityMonths) + "::int - 1) * INTERVAL '1 month')"
		}
		descriptionColumn := "NULL::TEXT"
		if params.IncludeDescription {
			descriptionColumn = "v.description"
		}
		ctes = strings.NewReplacer(
			"/*DESCRIPTION_COLUMN*/", descriptionColumn,
			"/*VULNERABILITY_FILTERS*/", filters,
		).Replace(queryPackageCTEVulnerabilities)
	}

	if params.WithSummary {
		columns += queryPackageColumnSummary
	}
	if params.WithCurrentStatus {
		columns += queryPackageColumnCurrentStatus
	}
	if params.IncludeVulnerabilities {
		columns += queryPackageColumnVulnerabilities
		references := "'[]'::jsonb"
		if params.IncludeReferences {
			references = `COALESCE((
						SELECT jsonb_agg(reference.url ORDER BY reference.id)
						FROM vulnerability_references AS reference
						WHERE reference.vulnerability_id = vul.id
					), '[]'::jsonb)`
		}
		joins = strings.Replace(queryPackageJoinVulnerabilities, "/*REFERENCES*/", references, 1)
	}

	query := strings.NewReplacer(
		"/*ECOSYSTEM*/", ecosystem,
		"/*IDENTIFIER*/", identifier,
		"/*CTES*/", ctes,
		"/*COLUMNS*/", columns,
		"/*JOINS*/", joins,
	).Replace(querySelectPackage)

	return query, args
}

func dbSelectPackage(ctx context.Context, params PackageDetailParams) (*PackageDetailDocument, error) {
	if db.Pool == nil {
		return nil, fmt.Errorf("Postgres pool is not initialized")
	}

	query, args := buildPackageDetailQuery(params)
	var document PackageDetailDocument
	targets := []any{
		&document.ID,
		&document.Name,
		&document.Slug,
		&document.Ecosystem,
		&document.Popularity,
		&document.IsDeprecated,
	}
	if params.WithSummary {
		targets = append(targets, &document.Summary)
	}
	if params.WithCurrentStatus {
		targets = append(targets, &document.CurrentStatuses)
	}
	if params.IncludeVulnerabilities {
		targets = append(targets, &document.VulnerabilityCounts, &document.Vulnerabilities)
	}

	if err := db.Pool.QueryRow(ctx, query, args...).Scan(targets...); err != nil {
		return nil, err
	}

	return &document, nil
}

func dbGetPackageDetailResponse(ctx context.Context, params PackageDetailParams) (int, []byte, error) {
	cacheKey := params.cacheKey()
	if db.Redis != nil {
		if cached, err := db.Redis.Get(ctx, cacheKey).Bytes(); err == nil && len(cached) > 1 {
			if cached[0] == '0' {
				return fiber.StatusNotFound, cached[1:], nil
			}
			return fiber.StatusOK, cached[1:], nil
		}
	}

	document, err := dbSelectPackage(ctx, params)
	status, ttl, marker := fiber.StatusOK, packageDetailCacheTTL, byte('1')
	response := PackageDetailResponse{Status: fiber.StatusOK, Message: successMessage, Document: document}
	if err != nil {
		if !errors.Is(err, pgx.ErrNoRows) {
			return 0, nil, err
		}
		status, ttl, marker = fiber.StatusNotFound, packageDetailNotFoundCacheTTL, '0'
		response = PackageDetailResponse{Status: fiber.StatusNotFound, Message: successNotFoundMessage}
	}

	body, err := json.Marshal(response)
	if err != nil {
		return 0, nil, err
	}
	if db.Redis != nil {
		_ = db.Redis.Set(ctx, cacheKey, append([]byte{marker}, body...), ttl).Err()
	}

	return status, body, nil
}
