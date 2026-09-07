package packagesApi

import (
	"github.com/gofiber/fiber/v3"
	"github.com/trustpkg/trustpkg-api/internal/typesense"
)

type InitRateLimitingConfig struct {
	Search RateLimitConfig
	List   RateLimitConfig
}

func RegisterRoutes(app fiber.Router, typesenseClient *typesense.Controller) {
	RegisterRoutesWithRateLimits(app, typesenseClient, InitRateLimitingConfig{
		Search: RateLimitConfig{
			Key:               "packages-search",
			RequestsPerWindow: DefaultRequestsPerWindow,
			RequestWindow:     DefaultRequestWindow,
		},
		List: RateLimitConfig{
			Key:               "packages-list",
			RequestsPerWindow: DefaultRequestsPerWindow,
			RequestWindow:     DefaultRequestWindow,
		},
	})
}

func RegisterRoutesWithRateLimits(app fiber.Router, typesenseClient *typesense.Controller, rateLimits InitRateLimitingConfig) {
	packages := app.Group("/packages")
	handlers := newHandlers(typesenseClient, rateLimits)

	packages.Get("/search", handlers.search)
	packages.Get("/list", handlers.list)
}
