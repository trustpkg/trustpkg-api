package packagesApi

import (
	"errors"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v3"
	"github.com/jackc/pgx/v5"
	"github.com/trustpkg/trustpkg-api/internal/typesense"
	"github.com/typesense/typesense-go/v4/typesense/api"
)

type handlers struct {
	typesenseClient   *typesense.Controller
	searchRateLimiter RateLimiter
	listRateLimiter   RateLimiter
}

func newHandlers(typesenseClient *typesense.Controller, rateLimits InitRateLimitingConfig) *handlers {
	return &handlers{
		typesenseClient:   typesenseClient,
		searchRateLimiter: NewIPRateLimiter(rateLimits.Search),
		listRateLimiter:   NewIPRateLimiter(rateLimits.List),
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
// @Router /api/packages/search [get]
func (handlers *handlers) search(ctx fiber.Ctx) error {
	ctx.Set("Access-Control-Allow-Origin", "*")

	for key := range ctx.Request().Header.All() {
		if !isAllowedSearchHeader(string(key)) {
			return searchRespond(ctx, fiber.StatusBadRequest, badRequestMessage, nil)
		}
	}

	allowed, err := handlers.searchRateLimiter.Allow(ctx)
	if err != nil {
		return searchRespond(ctx, fiber.StatusInternalServerError, internalErrorMessage, nil)
	}
	if !allowed {
		return searchRespond(ctx, fiber.StatusTooManyRequests, toManyRequest, nil)
	}

	if len(ctx.Request().Body()) > 0 {
		return searchRespond(ctx, fiber.StatusBadRequest, badRequestMessage, nil)
	}

	queryArgs := ctx.Request().URI().QueryArgs()
	invalidQueryParameter := false
	for key := range queryArgs.All() {
		if string(key) != "query" && string(key) != "limit" {
			invalidQueryParameter = true
		}
	}
	if invalidQueryParameter {
		return searchRespond(ctx, fiber.StatusBadRequest, badRequestMessage, nil)
	}

	queryParams := strings.ToLower(strings.TrimSpace(ctx.Query("query")))
	limitParams := ctx.Query("limit", "10")
	limit, err := strconv.Atoi(limitParams)
	if err != nil {
		return searchRespond(ctx, fiber.StatusBadRequest, badRequestMessage, nil)
	}

	searchParams := SearchParams{
		Query: queryParams,
		Limit: limit,
	}
	if err := searchParams.validate(); err != nil {
		return searchRespond(ctx, fiber.StatusBadRequest, badRequestMessage, nil)
	}

	queryBy := "name"
	prefix := "true"
	includeFields := "id,name,slug,ecosystem,popularity,is_deprecated,vulnerabilities"
	searchResult, err := handlers.typesenseClient.SearchDocuments(typesense.PackagesCollection, &api.SearchCollectionParams{
		Q:             &queryParams,
		PerPage:       &limit,
		QueryBy:       &queryBy,
		Prefix:        &prefix,
		IncludeFields: &includeFields,
	})
	if err != nil {
		return searchRespond(ctx, fiber.StatusInternalServerError, internalErrorMessage, nil)
	}
	if searchResult == nil {
		return searchRespond(ctx, fiber.StatusInternalServerError, internalErrorMessage, nil)
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
				Slug:         stringValue((*hit.Document)["slug"]),
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
		return searchRespond(ctx, fiber.StatusOK, successNotFoundMessage, documents)
	}

	return searchRespond(ctx, fiber.StatusOK, successMessage, documents)
}

// list godoc
// @Summary List packages
// @Description List packages with pagination and sorting.
// @Tags packages
// @ID listPackages
// @Produce json
// @Param page query int false "Page number" minimum(1) default(1)
// @Param limit query int false "Maximum number of results" minimum(1) maximum(50) default(10)
// @Param sortBy query string false "Field to sort by (e.g., name, popularity or ecosystem)" default(popularity)
// @Param order query string false "Sort order (asc or desc)" default(desc)
// @Param ecosystem query string false "Package ecosystem" default(npm)
// @Success 200 {object} packagesApi.ListResponse
// @Failure 400 {object} packagesApi.ListResponse
// @Failure 404 {object} packagesApi.ListResponse
// @Failure 429 {object} packagesApi.ListResponse
// @Failure 500 {object} packagesApi.ListResponse
// @Router /api/packages/list [get]
func (handlers *handlers) list(ctx fiber.Ctx) error {
	ctx.Set("Access-Control-Allow-Origin", "*")

	for key := range ctx.Request().Header.All() {
		if !isAllowedSearchHeader(string(key)) {
			return listRespond(ctx, fiber.StatusBadRequest, badRequestMessage, nil)
		}
	}

	allowed, err := handlers.listRateLimiter.Allow(ctx)
	if err != nil {
		return listRespond(ctx, fiber.StatusInternalServerError, internalErrorMessage, nil)
	}
	if !allowed {
		return listRespond(ctx, fiber.StatusTooManyRequests, toManyRequest, nil)
	}

	queryArgs := ctx.Request().URI().QueryArgs()
	invalidQueryParameter := false
	for key := range queryArgs.All() {
		switch string(key) {
		case "limit", "page", "sortBy", "order", "ecosystem":
			continue
		default:
			invalidQueryParameter = true
		}
	}
	if invalidQueryParameter {
		return listRespond(ctx, fiber.StatusBadRequest, badRequestMessage, nil)
	}

	limitParam := 10
	if queryArgs.Has("limit") {
		limitParam, err = strconv.Atoi(string(queryArgs.Peek("limit")))
		if err != nil {
			return listRespond(ctx, fiber.StatusBadRequest, badRequestMessage, nil)
		}
	}

	pageParam := 1
	if queryArgs.Has("page") {
		pageParam, err = strconv.Atoi(string(queryArgs.Peek("page")))
		if err != nil {
			return listRespond(ctx, fiber.StatusBadRequest, badRequestMessage, nil)
		}
	}

	listParams := ListParams{
		Limit:     limitParam,
		Page:      pageParam,
		SortBy:    string(queryArgs.Peek("sortBy")),
		Order:     string(queryArgs.Peek("order")),
		Ecosystem: string(queryArgs.Peek("ecosystem")),
	}
	if listParams.SortBy == "" {
		listParams.SortBy = "popularity"
	}
	if listParams.Order == "" {
		listParams.Order = "desc"
	}
	if listParams.Ecosystem == "" {
		listParams.Ecosystem = "npm"
	}
	if err := listParams.validate(); err != nil {
		return listRespond(ctx, fiber.StatusBadRequest, badRequestMessage, nil)
	}

	list, err := dbSelectPackages(dbSelectPackagesPayload{
		ctx:       ctx.Context(),
		page:      listParams.Page,
		limit:     listParams.Limit,
		sortBy:    listParams.SortBy,
		order:     listParams.Order,
		ecosystem: listParams.Ecosystem,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return listRespond(ctx, fiber.StatusNotFound, successNotFoundMessage, nil)
		}

		return listRespond(ctx, fiber.StatusInternalServerError, internalErrorMessage, nil)
	}
	if len(list) == 0 {
		return listRespond(ctx, fiber.StatusNotFound, successNotFoundMessage, nil)
	}

	return listRespond(ctx, fiber.StatusOK, successMessage, list)
}
