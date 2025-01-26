package main

import (
	"github.com/gofiber/fiber/v2"
	"github.com/tison2810/be-go-tc/database"
	"github.com/tison2810/be-go-tc/middleware"
)

func main() {
	database.ConnectDb()
	app := fiber.New()
	middleware.FiberMiddleware(app)
	publicRoutes(app)

	app.Listen(":3000")
}
