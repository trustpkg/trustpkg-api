package router

import (
	"github.com/gofiber/fiber/v3"
	packagesApi "github.com/trustpkg/trustpkg-api/internal/packages-api"
	"github.com/trustpkg/trustpkg-api/internal/typesense"
)

func Initialize(router *fiber.App, typesenseClient *typesense.Controller) {
	api := router.Group("/api")

	api.Get("/hello", func(ctx fiber.Ctx) error {
		return ctx.SendString("Hello, World!")
	})

	packagesApi.RegisterRoutes(api, typesenseClient)
}
