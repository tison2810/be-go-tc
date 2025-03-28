package main

import (
	"github.com/gofiber/fiber/v2"
	"github.com/tison2810/be-go-tc/handlers"
	"github.com/tison2810/be-go-tc/middleware"
)

func privateRoutes(app *fiber.App) {
	private := app.Group("/", middleware.AuthMiddleware())
	private.Get("/create", handlers.CreatePost)
}
