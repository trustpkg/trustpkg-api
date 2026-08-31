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

type searchResponse struct {
	Status    int              `json:"status"`
	Message   string           `json:"message"`
	Documents []searchDocument `json:"documents,omitempty"`
}

type searchDocument struct {
	ID              string         `json:"id"`
	Name            string         `json:"name"`
	Ecosystem       string         `json:"ecosystem"`
	Popularity      int64          `json:"popularity"`
	IsDeprecated    bool           `json:"is_deprecated"`
	Vulnerabilities map[string]int `json:"vulnerabilities"`
}
