package middleware

import (
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/tison2810/be-go-tc/utils"
)

// FiberMiddleware provide Fiber's built-in middlewares.
// See: https://docs.gofiber.io/api/middleware
func FiberMiddleware(a *fiber.App) {
	a.Use(
		cors.New(cors.Config{
			AllowOrigins:     "http://localhost:3001",
			AllowMethods:     "GET,POST,PUT,DELETE,OPTIONS",
			AllowHeaders:     "Origin, Content-Type, Accept, Authorization, Access-Control-Allow-Origin",
			AllowCredentials: true,
		})
	)
}

func AuthMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Lấy header Authorization
		authHeader := c.Get("Authorization")
		if authHeader == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Authorization header required",
			})
		}

		// Kiểm tra định dạng "Bearer <token>"
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Invalid authorization header format",
			})
		}

		// Xác thực token
		tokenString := parts[1]
		email, role, err := utils.VerifyJWT(tokenString)
		if err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": err.Error(),
			})
		}

		// Lưu email vào Locals để sử dụng trong handler sau này
		c.Locals("email", email)
		c.Locals("role", role)
		c.Locals("token", tokenString)
		return c.Next()
	}
}
