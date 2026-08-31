package apiDocs

import (
	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/adaptor"
	httpSwagger "github.com/swaggo/http-swagger"
)

type handlers struct {
	service          *Service
	swaggerUIHandler fiber.Handler
}

func newHandlers(service *Service) *handlers {
	return &handlers{
		service: service,
		swaggerUIHandler: adaptor.HTTPHandler(httpSwagger.Handler(
			httpSwagger.URL("/api/openapi.json"),
		)),
	}
}

func (handler *handlers) swaggerHandler(ctx fiber.Ctx) error {
	return handler.swaggerUIHandler(ctx)
}

func (handler *handlers) openAPIHandler(ctx fiber.Ctx) error {
	ctx.Type("json")
	return ctx.Send(handler.service.openAPIJSON)
}
