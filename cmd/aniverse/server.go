package aniverse

import (
	"aniverse/internal/controller"
	"os"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
)

func Start() {
	app := fiber.New()
	app.Use(logger.New())
	app.Use(cors.New())

	// Initialize Providers
	controller := controller.NewBaseController()

	// Routes
	app.Get("/search", controller.Search)
	app.Get("/info", controller.GetAnimeInfo)
	app.Get("/watch", controller.WatchEpisode)

	port := os.Getenv("PORT")
	if port == "" {
		port = "3000"
	}
	app.Listen(":" + port)
}
