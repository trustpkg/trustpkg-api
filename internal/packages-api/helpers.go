package packagesApi

import (
	"encoding/json"
	"fmt"

	"github.com/gofiber/fiber/v3"
)

func stringValue(value any) string {
	result, _ := value.(string)
	return result
}

func int64Value(value any) int64 {
	switch value := value.(type) {
	case int:
		return int64(value)
	case int64:
		return value
	case float64:
		return int64(value)
	case json.Number:
		result, _ := value.Int64()
		return result
	default:
		return 0
	}
}

func boolValue(value any) bool {
	result, _ := value.(bool)
	return result
}

func intMap(values map[string]any) map[string]int {
	result := make(map[string]int, len(values))
	for key, value := range values {
		result[key] = int(int64Value(value))
	}
	return result
}

func respond[T any](ctx fiber.Ctx, status int, message string, documents []T, pageInfo ...int) error {
	response := Response[T]{
		Status:    status,
		Message:   message,
		Documents: documents,
	}
	if len(pageInfo) == 0 {
		return ctx.Status(status).JSON(response)
	}

	paginatedResponse := PaginatedResponse[T]{
		Response: response,
		LastPage: pageInfo[0],
	}
	if len(pageInfo) > 1 {
		paginatedResponse.TotalPages = pageInfo[1]
	}

	return ctx.Status(status).JSON(paginatedResponse)
}

func searchRespond(ctx fiber.Ctx, status int, message string, documents []SearchDocument) error {
	return respond(ctx, status, message, documents)
}

func listRespond(ctx fiber.Ctx, status int, message string, lastPage int, documents []PackageListDocument, totalPages ...int) error {
	return respond(ctx, status, message, documents, append([]int{lastPage}, totalPages...)...)
}

func calculateTotalPages(totalPackages int64, limit int) int {
	if totalPackages == 0 {
		return 0
	}

	return int((totalPackages + int64(limit) - 1) / int64(limit))
}

func rateLimitKey(keyPrefix string, clientIP string, window int64) string {
	return fmt.Sprintf("packages-api:%s:rate:%s:%d", keyPrefix, clientIP, window)
}
