package router

import "github.com/gofiber/fiber/v3"

func Initialize(router *fiber.App) {
	api := router.Group("/api")

	api.Get("/hello", func(ctx fiber.Ctx) error {
		return ctx.SendString("Hello, World!")
	})

}
