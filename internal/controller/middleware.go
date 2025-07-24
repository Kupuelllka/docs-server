package controller

import (
	"docs-server/internal/model"
	"docs-server/internal/service"
	"errors"

	"github.com/gofiber/fiber/v2"
)

// AuthMiddleware проверяет JWT токен и добавляет пользователя в контекст
func AuthMiddleware(authService *service.AuthService) fiber.Handler {
	return func(ctx *fiber.Ctx) error {
		// Получаем токен из заголовка Authorization
		token := ctx.Get("Authorization")
		if token == "" {
			return ctx.Status(fiber.StatusUnauthorized).JSON(model.Response{
				Data: fiber.Map{
					"code":    fiber.StatusUnauthorized,
					"message": "Authorization token required",
				},
			})
		}

		// Валидируем токен
		user, err := authService.ValidateToken(token)
		if err != nil {
			statusCode := fiber.StatusUnauthorized
			message := "Invalid token"

			if errors.Is(err, service.ErrTokenExpired) {
				message = "Token expired"
			}

			return ctx.Status(statusCode).JSON(model.Response{
				Data: fiber.Map{
					"code":    statusCode,
					"message": message,
				},
			})
		}

		// Добавляем пользователя в контекст
		ctx.Locals("user", user)

		return ctx.Next()
	}
}
