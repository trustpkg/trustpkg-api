package packagesApi

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	validation "github.com/go-ozzo/ozzo-validation/v4"
)

const packageDetailMaxIdentifierLen = 214

var npmPackageRegex = regexp.MustCompile(
	`^\s*@?[a-z0-9][a-z0-9._-]*(\s+[a-z0-9][a-z0-9._-]*)?(/[a-z0-9][a-z0-9._-]*)?\s*$`,
)

type SearchParams struct {
	Query string
	Limit int
}

func (params *SearchParams) validate() error {
	return validation.ValidateStruct(params,
		validation.Field(&params.Query,
			validation.Required,
			validation.Length(1, 214),
			validation.Match(npmPackageRegex),
		),
		validation.Field(&params.Limit, validation.Min(1), validation.Max(20)),
	)
}

type Response[T any] struct {
	Status    int    `json:"status" validate:"required,oneof=200 400 404 429 500"`
	Message   string `json:"message" validate:"required,min=1,max=64"`
	Documents []T    `json:"documents,omitempty"`
}

type PaginatedResponse[T any] struct {
	Response[T]
	LastPage   int `json:"lastPage"`
	TotalPages int `json:"totalPages,omitempty"`
}

type SearchResponse struct {
	Status    int              `json:"status"`
	Message   string           `json:"message"`
	Documents []SearchDocument `json:"documents,omitempty"`
}

type ListResponse struct {
	Status     int                   `json:"status"`
	Message    string                `json:"message"`
	LastPage   int                   `json:"lastPage"`
	TotalPages int                   `json:"totalPages"`
	Documents  []PackageListDocument `json:"documents,omitempty"`
}

type SearchDocument struct {
	ID              string         `json:"id" validate:"required,min=1,max=19"`     // Database package identifier.
	Name            string         `json:"name" validate:"required,min=1,max=214"`  // npm package name.
	Slug            string         `json:"slug" validate:"required,min=1"`          // URL-safe package identifier.
	Ecosystem       string         `json:"ecosystem" validate:"required,oneof=npm"` // Package ecosystem.
	Popularity      int64          `json:"popularity" validate:"required,min=0"`    // Download popularity score.
	IsDeprecated    bool           `json:"is_deprecated" validate:"required"`       // Whether npm marks the package as deprecated.
	Vulnerabilities map[string]int `json:"vulnerabilities" validate:"required"`     // Monthly vulnerability counts keyed by YYYY-MM.
}

type PackageListDocument struct {
	ID                  int64                      `json:"id"`
	Name                string                     `json:"name"`
	Slug                string                     `json:"slug"`
	Ecosystem           string                     `json:"ecosystem"`
	Popularity          int64                      `json:"popularity"`
	IsDeprecated        bool                       `json:"is_deprecated"`
	VulnerabilityCounts map[string]int             `json:"vulnerability_counts,omitempty"`
	Vulnerabilities     map[string][]Vulnerability `json:"vulnerabilities,omitempty"`
	CurrentStatuses     *[]PackageCurrentStatus    `json:"current_statuses,omitempty"`
}

type PackageCurrentStatus struct {
	Code          string            `json:"code"`
	Label         string            `json:"label"`
	OSVID         string            `json:"osv_id"`
	CVEID         *string           `json:"cve_id"`
	Severity      *string           `json:"severity"`
	Affected      []AffectedPackage `json:"affected"`
	FixedVersions []string          `json:"fixed_versions"`
}

type Vulnerability struct {
	ID          int64             `json:"id"`
	OSVID       string            `json:"osv_id"`
	CVEID       *string           `json:"cve_id"`
	Summary     *string           `json:"summary"`
	Description *string           `json:"description"`
	Severity    *string           `json:"severity"`
	CVSSScore   *float64          `json:"cvss_score"`
	CVSSVector  *string           `json:"cvss_vector"`
	PublishedAt *string           `json:"published_at"`
	ModifiedAt  *string           `json:"modified_at"`
	WithdrawnAt *string           `json:"withdrawn_at"`
	Affected    []AffectedPackage `json:"affected"`
	References  []string          `json:"references"`
}

type AffectedPackage struct {
	Ecosystem           string  `json:"ecosystem"`
	IntroducedVersion   *string `json:"introduced_version"`
	FixedVersion        *string `json:"fixed_version"`
	LastAffectedVersion *string `json:"last_affected_version"`
}

type ListParams struct {
	Limit                   int
	Page                    int
	SortBy                  string
	Order                   string
	Ecosystem               string
	WithCurrentStatus       bool
	OnlyWithVulnerabilities bool
	VulnerabilityMonths     int
}

func (params *ListParams) validate() error {
	return validation.ValidateStruct(params,
		validation.Field(&params.Limit, validation.Min(1), validation.Max(50)),
		validation.Field(&params.Page, validation.Min(1)),
		validation.Field(&params.SortBy, validation.In("name", "popularity", "ecosystem")),
		validation.Field(&params.Order, validation.In("asc", "desc")),
		validation.Field(&params.Ecosystem, validation.In("npm")),
		validation.Field(&params.VulnerabilityMonths, validation.Min(0), validation.Max(12)),
	)
}

type dbSelectPackagesPayload struct {
	ctx                     context.Context
	page                    int
	limit                   int
	sortBy                  string
	order                   string
	ecosystem               string
	withCurrentStatus       bool
	includeVulnerabilities  bool
	onlyWithVulnerabilities bool
	vulnerabilityMonths     int
}

var packageIdentifierRegex = regexp.MustCompile(`^@?[A-Za-z0-9][A-Za-z0-9._~-]*(/[A-Za-z0-9][A-Za-z0-9._~-]*)?$`)

type PackageDetailResponse struct {
	Status   int                    `json:"status"`
	Message  string                 `json:"message"`
	Document *PackageDetailDocument `json:"document,omitempty"`
}

type PackageDetailDocument struct {
	ID                  int64           `json:"id"`
	Name                string          `json:"name"`
	Slug                string          `json:"slug"`
	Ecosystem           string          `json:"ecosystem"`
	Popularity          int64           `json:"popularity"`
	IsDeprecated        bool            `json:"is_deprecated"`
	Summary             json.RawMessage `json:"summary,omitempty" swaggertype:"object"`
	CurrentStatuses     json.RawMessage `json:"current_statuses,omitempty" swaggertype:"array,object"`
	VulnerabilityCounts json.RawMessage `json:"vulnerability_counts,omitempty" swaggertype:"object"`
	Vulnerabilities     json.RawMessage `json:"vulnerabilities,omitempty" swaggertype:"object"`
}

const packageDetailMaxBodyBytes = 1024

// VulnerabilityMonths accepts "all" (or 0) for the entire history, or 1-12 as a number or numeric string.
type VulnerabilityMonths int

func (months *VulnerabilityMonths) UnmarshalJSON(data []byte) error {
	raw := strings.Trim(strings.TrimSpace(string(data)), `"`)
	if strings.EqualFold(raw, "all") {
		*months = 0
		return nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil || value < 0 || value > 12 {
		return fmt.Errorf("invalid vulnerabilityMonths")
	}
	*months = VulnerabilityMonths(value)

	return nil
}

type PackageDetailRequest struct {
	Ecosystem              string              `json:"ecosystem" default:"npm" enums:"npm"`                                                             // Package ecosystem.
	IncludeVulnerabilities bool                `json:"includeVulnerabilities"`                                                                          // Include vulnerabilities grouped by month.
	VulnerabilityMonths    VulnerabilityMonths `json:"vulnerabilityMonths" swaggertype:"string" default:"all" enums:"all,0,1,2,3,4,5,6,7,8,9,10,11,12"` // "all" returns the entire history, 1-12 limits it to the last N months.
	Severity               string              `json:"severity" enums:"LOW,MODERATE,MEDIUM,HIGH,CRITICAL"`                                              // Only vulnerabilities with this severity.
	WithSummary            bool                `json:"withSummary"`                                                                                     // Include aggregated vulnerability summary.
	WithCurrentStatus      bool                `json:"withCurrentStatus"`                                                                               // Include current vulnerability status details.
	IncludeDescription     bool                `json:"includeDescription"`                                                                              // Include full vulnerability descriptions.
	IncludeReferences      bool                `json:"includeReferences"`                                                                               // Include vulnerability reference URLs.
	IncludeWithdrawn       bool                `json:"includeWithdrawn"`                                                                                // Include withdrawn vulnerabilities.
}

func (request PackageDetailRequest) params(identifier string) PackageDetailParams {
	params := PackageDetailParams{
		Identifier:             identifier,
		Ecosystem:              request.Ecosystem,
		IncludeVulnerabilities: request.IncludeVulnerabilities,
		VulnerabilityMonths:    int(request.VulnerabilityMonths),
		WithCurrentStatus:      request.WithCurrentStatus,
		WithSummary:            request.WithSummary,
		IncludeDescription:     request.IncludeDescription,
		IncludeReferences:      request.IncludeReferences,
		IncludeWithdrawn:       request.IncludeWithdrawn,
		Severity:               strings.ToUpper(request.Severity),
	}
	if params.Ecosystem == "" {
		params.Ecosystem = "npm"
	}

	return params
}

type PackageDetailParams struct {
	Identifier             string
	Ecosystem              string
	IncludeVulnerabilities bool
	VulnerabilityMonths    int
	WithCurrentStatus      bool
	WithSummary            bool
	IncludeDescription     bool
	IncludeReferences      bool
	IncludeWithdrawn       bool
	Severity               string
}

func (params *PackageDetailParams) validate() error {
	return validation.ValidateStruct(params,
		validation.Field(&params.Identifier,
			validation.Required,
			validation.Length(1, packageDetailMaxIdentifierLen),
			validation.Match(packageIdentifierRegex),
		),
		validation.Field(&params.Ecosystem, validation.In("npm")),
		validation.Field(&params.VulnerabilityMonths, validation.Min(0), validation.Max(12)),
		validation.Field(&params.Severity, validation.In("", "LOW", "MODERATE", "MEDIUM", "HIGH", "CRITICAL")),
	)
}

func (params *PackageDetailParams) needsVulnerabilityData() bool {
	return params.IncludeVulnerabilities || params.WithSummary
}

func (params *PackageDetailParams) cacheKey() string {
	flags := [...]bool{
		params.IncludeVulnerabilities, params.WithCurrentStatus, params.WithSummary,
		params.IncludeDescription, params.IncludeReferences, params.IncludeWithdrawn,
	}
	var builder strings.Builder
	for _, flag := range flags {
		if flag {
			builder.WriteByte('1')
		} else {
			builder.WriteByte('0')
		}
	}

	return fmt.Sprintf("packages-api:detail:%s:%s:%s:%d:%s", params.Ecosystem, params.Identifier, builder.String(), params.VulnerabilityMonths, params.Severity)
}
