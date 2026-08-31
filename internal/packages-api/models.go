package packagesApi

import (
	"regexp"

	validation "github.com/go-ozzo/ozzo-validation/v4"
)

const npmPackageRegex = `^@?[a-z0-9][a-z0-9._-]*(/[a-z0-9][a-z0-9._-]*)?$`

type SearchParams struct {
	Query string
	Limit int
}

func (params *SearchParams) validate() error {
	return validation.ValidateStruct(params,
		validation.Field(&params.Query,
			validation.Required,
			validation.Length(1, 214),
			validation.Match(regexp.MustCompile(npmPackageRegex)),
		),
		validation.Field(&params.Limit, validation.Min(1), validation.Max(20)),
	)
}

type SearchResponse struct {
	Status    int              `json:"status" validate:"required,oneof=200 400 429 500"`
	Message   string           `json:"message" validate:"required,min=1,max=64"`
	Documents []SearchDocument `json:"documents,omitempty"`
}

type SearchDocument struct {
	ID              string         `json:"id" validate:"required,min=1,max=19"`     // Database package identifier.
	Name            string         `json:"name" validate:"required,min=1,max=214"`  // npm package name.
	Ecosystem       string         `json:"ecosystem" validate:"required,oneof=npm"` // Package ecosystem.
	Popularity      int64          `json:"popularity" validate:"required,min=0"`    // Download popularity score.
	IsDeprecated    bool           `json:"is_deprecated" validate:"required"`       // Whether npm marks the package as deprecated.
	Vulnerabilities map[string]int `json:"vulnerabilities" validate:"required"`     // Monthly vulnerability counts keyed by YYYY-MM.
}
