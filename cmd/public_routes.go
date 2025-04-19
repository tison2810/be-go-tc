package main

import (
	"github.com/gofiber/fiber/v2"
	"github.com/tison2810/be-go-tc/handlers"
)

func publicRoutes(app *fiber.App) {
	app.Post("/api/auth/google", handlers.GoogleAuthHandler)
}
