package typesense

import (
	typesenseApi "github.com/typesense/typesense-go/v4/typesense/api"
	"github.com/typesense/typesense-go/v4/typesense/api/pointer"
)

type vulnerabilitiesMap map[string]int

type PackageCollectionType struct {
	ID              string             `json:"id"`
	Name            string             `json:"name"`
	Slug            string             `json:"slug"`
	Ecosystem       string             `json:"ecosystem"`
	Popularity      int64              `json:"popularity"`
	IsDeprecated    bool               `json:"is_deprecated"`
	Vulnerabilities vulnerabilitiesMap `json:"vulnerabilities"`
}

type packageDocument = PackageCollectionType

type pipelineState struct {
	Phase                      string `json:"phase"`
	WatchlistOffset            int    `json:"watchlist_offset"`
	LastPackageID              int64  `json:"last_package_id"`
	Cycle                      int64  `json:"cycle"`
	DocumentsUploaded          int64  `json:"documents_uploaded"`
	WatchlistDocumentsUploaded int64  `json:"watchlist_documents_uploaded"`
	DatabaseDocumentsUploaded  int64  `json:"database_documents_uploaded"`
	Status                     string `json:"status"`
	LastError                  string `json:"last_error"`
}

var PackageSchema = &typesenseApi.CollectionSchema{
	Name:               "packages",
	EnableNestedFields: pointer.True(),
	Fields: []typesenseApi.Field{
		{
			Name: "id",
			Type: "string",
		},
		{
			Name: "name",
			Type: "string",
		},
		{
			Name: "slug",
			Type: "string",
		},
		{
			Name: "ecosystem",
			Type: "string",
		},
		{
			Name: "popularity",
			Type: "int64",
		},
		{
			Name: "is_deprecated",
			Type: "bool",
		},
		{
			Name: "vulnerabilities",
			Type: "object",
		},
	},
	DefaultSortingField: pointer.String("popularity"),
}
