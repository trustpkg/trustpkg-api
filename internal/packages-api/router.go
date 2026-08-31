package packagesApi

import (
	"github.com/gofiber/fiber/v3"
	"github.com/trustpkg/trustpkg-api/internal/typesense"
)

func RegisterRoutes(app fiber.Router, typesenseClient *typesense.Controller) {
	packages := app.Group("/packages")
	rateLimiter := newIPRateLimiter()
	handlers := newHandlers(typesenseClient, rateLimiter)

	packages.Get("/search", handlers.search)
}
