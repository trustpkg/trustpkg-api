package packagesApi

import (
	"bytes"
	"encoding/json"
	"errors"
	"log"
	"net/url"
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
	detailRateLimiter RateLimiter
}

func newHandlers(typesenseClient *typesense.Controller, rateLimits InitRateLimitingConfig) *handlers {
	return &handlers{
		typesenseClient:   typesenseClient,
		searchRateLimiter: NewIPRateLimiter(rateLimits.Search),
		listRateLimiter:   NewIPRateLimiter(rateLimits.List),
		detailRateLimiter: NewIPRateLimiter(rateLimits.Detail),
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
	searchCollectionParams := &api.SearchCollectionParams{
		Q:             &queryParams,
		PerPage:       &limit,
		QueryBy:       &queryBy,
		Prefix:        &prefix,
		IncludeFields: &includeFields,
	}
	searchResult, err := handlers.typesenseClient.SearchDocuments(typesense.PackagesCollection, searchCollectionParams)
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
// @Param includeVulnerabilities query bool false "Include full vulnerability details" default(false)
// @Param onlyWithVulnerabilities query bool false "Only return packages with vulnerabilities" default(false)
// @Param vulnerabilityMonths query int false "Limit vulnerabilities to the last N months (0 for all available months, 1-12 for a window)" minimum(0) maximum(12) default(0)
// @Param withCurrentStatus query bool false "Include current vulnerability status details" default(false)
// @Success 200 {object} packagesApi.ListResponse
// @Failure 400 {object} packagesApi.ListResponse
// @Failure 404 {object} packagesApi.ListResponse
// @Failure 429 {object} packagesApi.ListResponse
// @Failure 500 {object} packagesApi.ListResponse
// @Router /api/packages/list [get]
func (handlers *handlers) list(ctx fiber.Ctx) error {
	ctx.Set("Access-Control-Allow-Origin", "*")
	pageParam := 1

	for key := range ctx.Request().Header.All() {
		if !isAllowedSearchHeader(string(key)) {
			return listRespond(ctx, fiber.StatusBadRequest, badRequestMessage, pageParam, nil)
		}
	}

	allowed, err := handlers.listRateLimiter.Allow(ctx)
	if err != nil {
		return listRespond(ctx, fiber.StatusInternalServerError, internalErrorMessage, pageParam, nil)
	}
	if !allowed {
		return listRespond(ctx, fiber.StatusTooManyRequests, toManyRequest, pageParam, nil)
	}

	queryArgs := ctx.Request().URI().QueryArgs()
	invalidQueryParameter := false
	for key := range queryArgs.All() {
		switch string(key) {
		case "limit", "page", "sortBy", "order", "ecosystem", "includeVulnerabilities", "onlyWithVulnerabilities", "vulnerabilityMonths", "withCurrentStatus":
			continue
		default:
			invalidQueryParameter = true
		}
	}
	if invalidQueryParameter {
		return listRespond(ctx, fiber.StatusBadRequest, badRequestMessage, pageParam, nil)
	}

	limitParam := 10
	if queryArgs.Has("limit") {
		limitParam, err = strconv.Atoi(string(queryArgs.Peek("limit")))
		if err != nil {
			return listRespond(ctx, fiber.StatusBadRequest, badRequestMessage, pageParam, nil)
		}
	}

	if queryArgs.Has("page") {
		pageParam, err = strconv.Atoi(string(queryArgs.Peek("page")))
		if err != nil {
			return listRespond(ctx, fiber.StatusBadRequest, badRequestMessage, pageParam, nil)
		}
	}

	includeVulnerabilities := false
	if queryArgs.Has("includeVulnerabilities") {
		includeVulnerabilities, err = strconv.ParseBool(string(queryArgs.Peek("includeVulnerabilities")))
		if err != nil {
			return listRespond(ctx, fiber.StatusBadRequest, badRequestMessage, pageParam, nil)
		}
	}

	onlyWithVulnerabilities := false
	if queryArgs.Has("onlyWithVulnerabilities") {
		onlyWithVulnerabilities, err = strconv.ParseBool(string(queryArgs.Peek("onlyWithVulnerabilities")))
		if err != nil {
			return listRespond(ctx, fiber.StatusBadRequest, badRequestMessage, pageParam, nil)
		}
	}

	vulnerabilityMonths := 0
	if queryArgs.Has("vulnerabilityMonths") {
		vulnerabilityMonths, err = strconv.Atoi(string(queryArgs.Peek("vulnerabilityMonths")))
		if err != nil || vulnerabilityMonths < 0 || vulnerabilityMonths > 12 {
			return listRespond(ctx, fiber.StatusBadRequest, badRequestMessage, pageParam, nil)
		}
	}

	withCurrentStatus := false
	if queryArgs.Has("withCurrentStatus") {
		withCurrentStatus, err = strconv.ParseBool(string(queryArgs.Peek("withCurrentStatus")))
		if err != nil {
			return listRespond(ctx, fiber.StatusBadRequest, badRequestMessage, pageParam, nil)
		}
	}

	listParams := ListParams{
		Limit:                   limitParam,
		Page:                    pageParam,
		SortBy:                  string(queryArgs.Peek("sortBy")),
		Order:                   string(queryArgs.Peek("order")),
		Ecosystem:               string(queryArgs.Peek("ecosystem")),
		WithCurrentStatus:       withCurrentStatus,
		OnlyWithVulnerabilities: onlyWithVulnerabilities,
		VulnerabilityMonths:     vulnerabilityMonths,
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
		return listRespond(ctx, fiber.StatusBadRequest, badRequestMessage, pageParam, nil)
	}

	dbPayload := dbSelectPackagesPayload{
		ctx:                     ctx.Context(),
		page:                    listParams.Page,
		limit:                   listParams.Limit,
		sortBy:                  listParams.SortBy,
		order:                   listParams.Order,
		ecosystem:               listParams.Ecosystem,
		withCurrentStatus:       listParams.WithCurrentStatus,
		includeVulnerabilities:  includeVulnerabilities,
		onlyWithVulnerabilities: listParams.OnlyWithVulnerabilities,
		vulnerabilityMonths:     listParams.VulnerabilityMonths,
	}
	totalPackages, err := dbCountPackages(dbPayload)
	if err != nil {
		log.Printf("[PACKAGES] count list query failed: %v", err)
		return listRespond(ctx, fiber.StatusInternalServerError, internalErrorMessage, pageParam, nil)
	}

	list, err := dbSelectPackages(dbPayload)
	if err != nil {
		log.Printf("[PACKAGES] select list query failed: %v", err)
		if errors.Is(err, pgx.ErrNoRows) {
			return listRespond(ctx, fiber.StatusNotFound, successNotFoundMessage, pageParam, nil)
		}

		return listRespond(ctx, fiber.StatusInternalServerError, internalErrorMessage, pageParam, nil)
	}
	if len(list) == 0 {
		return listRespond(ctx, fiber.StatusNotFound, successNotFoundMessage, pageParam, nil)
	}

	return listRespond(ctx, fiber.StatusOK, successMessage, pageParam, list, calculateTotalPages(totalPackages, listParams.Limit))
}

// packageDetail godoc
// @Summary Get a single package
// @Description Get a single package by slug or exact name (scoped names: /api/packages/@scope/name) with optional vulnerability details selected in the JSON request body (query parameters are not accepted).
// @Tags packages
// @ID getPackage
// @Accept json
// @Produce json
// @Param name path string true "Package slug or name"
// @Param request body packagesApi.PackageDetailRequest false "Optional detail options; an empty body returns the base package data"
// @Success 200 {object} packagesApi.PackageDetailResponse
// @Failure 400 {object} packagesApi.PackageDetailResponse
// @Failure 404 {object} packagesApi.PackageDetailResponse
// @Failure 429 {object} packagesApi.PackageDetailResponse
// @Failure 500 {object} packagesApi.PackageDetailResponse
// @Router /api/packages/{name} [post]
func (handlers *handlers) detail(ctx fiber.Ctx) error {
	ctx.Set("Access-Control-Allow-Origin", "*")

	for key := range ctx.Request().Header.All() {
		if !isAllowedSearchHeader(string(key)) {
			return packageDetailRespond(ctx, fiber.StatusBadRequest, badRequestMessage, nil)
		}
	}

	allowed, err := handlers.detailRateLimiter.Allow(ctx)
	if err != nil {
		return packageDetailRespond(ctx, fiber.StatusInternalServerError, internalErrorMessage, nil)
	}
	if !allowed {
		return packageDetailRespond(ctx, fiber.StatusTooManyRequests, toManyRequest, nil)
	}

	if ctx.Request().URI().QueryArgs().Len() > 0 {
		return packageDetailRespond(ctx, fiber.StatusBadRequest, badRequestMessage, nil)
	}

	var request PackageDetailRequest
	rawBody := ctx.Body()
	if len(rawBody) > packageDetailMaxBodyBytes {
		return packageDetailRespond(ctx, fiber.StatusBadRequest, badRequestMessage, nil)
	}
	if len(bytes.TrimSpace(rawBody)) > 0 {
		decoder := json.NewDecoder(bytes.NewReader(rawBody))
		decoder.DisallowUnknownFields()
		if err := decoder.Decode(&request); err != nil || decoder.More() {
			return packageDetailRespond(ctx, fiber.StatusBadRequest, badRequestMessage, nil)
		}
	}

	identifier := ctx.Params("name")
	if scope := ctx.Params("scope"); scope != "" {
		identifier = scope + "/" + identifier
	}
	if strings.Contains(identifier, "%") {
		decoded, err := url.PathUnescape(identifier)
		if err != nil {
			return packageDetailRespond(ctx, fiber.StatusBadRequest, badRequestMessage, nil)
		}
		identifier = decoded
	}

	params := request.params(identifier)
	if err := params.validate(); err != nil {
		return packageDetailRespond(ctx, fiber.StatusBadRequest, badRequestMessage, nil)
	}

	status, body, err := dbGetPackageDetailResponse(ctx.Context(), params)
	if err != nil {
		log.Printf("[PACKAGES] select package query failed: %v", err)
		return packageDetailRespond(ctx, fiber.StatusInternalServerError, internalErrorMessage, nil)
	}

	return packageDetailRespondRaw(ctx, status, body)
}
