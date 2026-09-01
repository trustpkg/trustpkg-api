package packagesApi

import (
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v3"
	"github.com/trustpkg/trustpkg-api/internal/typesense"
	"github.com/typesense/typesense-go/v4/typesense/api"
)

type handlers struct {
	typesenseClient *typesense.Controller
	rateLimiter     *ipRateLimiter
}

func newHandlers(typesenseClient *typesense.Controller, rateLimiter *ipRateLimiter) *handlers {
	return &handlers{
		typesenseClient: typesenseClient,
		rateLimiter:     rateLimiter,
	}
}

// search godoc
// @Summary Search packages
// @Description Searches npm package names using prefix matching and returns the complete package documents.
// @Tags packages
// @ID searchPackages
// @Produce json
// @Param query query string true "Package name or scoped package name" minlength(1) maxlength(214)
// @Param limit query int false "Maximum number of results" minimum(1) maximum(20) default(10)
// @Success 200 {object} packagesApi.SearchResponse
// @Failure 400 {object} packagesApi.SearchResponse
// @Failure 429 {object} packagesApi.SearchResponse
// @Failure 500 {object} packagesApi.SearchResponse
// @Router /packages/search [get]
func (handlers *handlers) search(ctx fiber.Ctx) error {
	ctx.Set("Access-Control-Allow-Origin", "*")

	for key := range ctx.Request().Header.All() {
		if !isAllowedSearchHeader(string(key)) {
			return respond(ctx, fiber.StatusBadRequest, badRequestMessage, nil)
		}
	}

	allowed, err := handlers.rateLimiter.allow(ctx)
	if err != nil {
		return respond(ctx, fiber.StatusInternalServerError, internalErrorMessage, nil)
	}
	if !allowed {
		return respond(ctx, fiber.StatusTooManyRequests, toManyRequest, nil)
	}

	if len(ctx.Request().Body()) > 0 {
		return respond(ctx, fiber.StatusBadRequest, badRequestMessage, nil)
	}

	queryArgs := ctx.Request().URI().QueryArgs()
	invalidQueryParameter := false
	for key := range queryArgs.All() {
		if string(key) != "query" && string(key) != "limit" {
			invalidQueryParameter = true
		}
	}
	if invalidQueryParameter {
		return respond(ctx, fiber.StatusBadRequest, badRequestMessage, nil)
	}

	queryParams := strings.ToLower(strings.TrimSpace(ctx.Query("query")))
	limitParams := ctx.Query("limit", "10")
	limit, err := strconv.Atoi(limitParams)
	if err != nil {
		return respond(ctx, fiber.StatusBadRequest, badRequestMessage, nil)
	}

	searchParams := SearchParams{
		Query: queryParams,
		Limit: limit,
	}
	if err := searchParams.validate(); err != nil {
		return respond(ctx, fiber.StatusBadRequest, badRequestMessage, nil)
	}

	queryBy := "name"
	prefix := "true"
	includeFields := "id,name,ecosystem,popularity,is_deprecated,vulnerabilities"
	searchResult, err := handlers.typesenseClient.SearchDocuments(typesense.PackagesCollection, &api.SearchCollectionParams{
		Q:             &queryParams,
		PerPage:       &limit,
		QueryBy:       &queryBy,
		Prefix:        &prefix,
		IncludeFields: &includeFields,
	})
	if err != nil {
		return respond(ctx, fiber.StatusInternalServerError, internalErrorMessage, nil)
	}
	if searchResult == nil {
		return respond(ctx, fiber.StatusInternalServerError, internalErrorMessage, nil)
	}
	documents := make([]SearchDocument, 0)
	if searchResult.Hits != nil {
		for _, hit := range *searchResult.Hits {
			if hit.Document == nil {
				continue
			}
			document := SearchDocument{
				ID:           stringValue((*hit.Document)["id"]),
				Name:         stringValue((*hit.Document)["name"]),
				Ecosystem:    stringValue((*hit.Document)["ecosystem"]),
				Popularity:   int64Value((*hit.Document)["popularity"]),
				IsDeprecated: boolValue((*hit.Document)["is_deprecated"]),
			}
			if vulnerabilities, ok := (*hit.Document)["vulnerabilities"].(map[string]any); ok {
				document.Vulnerabilities = intMap(vulnerabilities)
			}
			documents = append(documents, document)
		}
	}
	if searchResult.Found == nil || *searchResult.Found == 0 {
		return respond(ctx, fiber.StatusOK, successNotFoundMessage, documents)
	}

	return respond(ctx, fiber.StatusOK, successMessage, documents)
}
