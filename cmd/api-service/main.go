package main

import (
	"context"
	"os"

	"github.com/trustpkg/trustpkg-api/db"
	"github.com/trustpkg/trustpkg-api/internal/router"

	"github.com/gofiber/fiber/v3"
)

func main() {
	ctx := context.Background()

	db.ConnectDb()
	db.ConnectRedis(ctx)

	app := fiber.New()

	router.Initialize(app)

	app.Listen(":" + os.Getenv("API_SERVICE_PORT"))
}
