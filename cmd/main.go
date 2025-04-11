package main

import (
	"github.com/gofiber/fiber/v2"
	_ "github.com/tison2810/be-go-tc/cmd/docs"
	"github.com/tison2810/be-go-tc/database"
	"github.com/tison2810/be-go-tc/middleware"
)

// @title           My API
// @version         1.0
// @description     API using Fiber + Swagger
// @contact.name    Your Name
// @host            localhost:3000
// @BasePath        /
// @schemes         http

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
func main() {
	database.ConnectDb()
	app := fiber.New()
	middleware.FiberMiddleware(app)
	publicRoutes(app)
	// app.Get("/swagger/*", swagger.HandlerDefault)
	privateRoutes(app)
	app.Listen(":3000")
}
