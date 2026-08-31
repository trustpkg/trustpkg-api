package main

import (
	"context"
	"os"

	"github.com/trustpkg/trustpkg-api/db"
	apiDocs "github.com/trustpkg/trustpkg-api/internal/api-docs"
	"github.com/trustpkg/trustpkg-api/internal/router"
	"github.com/trustpkg/trustpkg-api/internal/typesense"

	"github.com/gofiber/fiber/v3"
)

// @title TrustPkg API
// @version 1.4.0
// @description API for searching npm packages and their vulnerability data.
// @BasePath /api
// @schemes http https
func main() {
	ctx := context.Background()

	db.ConnectDb()
	db.ConnectRedis(ctx)
	typesenseClient := typesense.NewInstance(ctx)

	app := fiber.New()

	apiDocsService, err := apiDocs.NewService(os.Getenv("API_SERVICE_URL"))
	if err != nil {
		panic(err)
	}

	router.Initialize(app, typesenseClient, apiDocsService)

	app.Listen(":" + os.Getenv("API_SERVICE_PORT"))
}
