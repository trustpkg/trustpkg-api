package packagesApi

import (
	"encoding/json"

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

func respond(ctx fiber.Ctx, status int, message string, documents []searchDocument) error {
	return ctx.Status(status).JSON(searchResponse{
		Status:    status,
		Message:   message,
		Documents: documents,
	})
}
