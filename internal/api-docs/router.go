package apiDocs

import "github.com/gofiber/fiber/v3"

func RegisterRoutes(api fiber.Router, service *Service) {
	api.Get("/swagger/*", service.handlers.swaggerHandler)
	api.Get("/openapi.json", service.handlers.openAPIHandler)
}
