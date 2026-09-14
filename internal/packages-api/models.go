package packagesApi

import (
	"context"
	"regexp"

	validation "github.com/go-ozzo/ozzo-validation/v4"
)

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
