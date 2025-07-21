package controller

import (
	"docs-server/internal/service"

	"github.com/gofiber/fiber/v2"
)

func AuthMiddleware(authService service.AuthService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		token := c.Get("Authorization")
		if token == "" {
			return fiber.NewError(fiber.StatusUnauthorized, "Authorization token required")
		}

		user, err := authService.ValidateToken(token)
		if err != nil {
			return fiber.NewError(fiber.StatusUnauthorized, "Invalid token")
		}

		// Сохраняем пользователя в контексте
		c.Locals("user", user)
		return c.Next()
	}
}
