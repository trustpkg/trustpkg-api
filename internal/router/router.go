package router

import (
	"github.com/gofiber/fiber/v3"
	apiDocs "github.com/trustpkg/trustpkg-api/internal/api-docs"
	packagesApi "github.com/trustpkg/trustpkg-api/internal/packages-api"
	"github.com/trustpkg/trustpkg-api/internal/typesense"
)

func Initialize(router *fiber.App, typesenseClient *typesense.Controller, apiDocsService *apiDocs.Service) {
	api := router.Group("/api")
	apiDocs.RegisterRoutes(api, apiDocsService)

	api.Get("/hello", func(ctx fiber.Ctx) error {
		return ctx.SendString("Hello, World!")
	})

	packagesApi.RegisterRoutes(api, typesenseClient)
}
