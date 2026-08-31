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

func (handlers *handlers) search(ctx fiber.Ctx) error {
	for key := range ctx.Request().Header.All() {
		if _, allowed := allowedSearchHeaders[strings.ToLower(string(key))]; !allowed {
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
	documents := make([]searchDocument, 0)
	if searchResult.Hits != nil {
		for _, hit := range *searchResult.Hits {
			if hit.Document == nil {
				continue
			}
			document := searchDocument{
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
