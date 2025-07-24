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

// ErrorHandler middleware для обработки ошибок сервиса аутентификации
func AuthErrorHandler(ctx *fiber.Ctx) error {
	err := ctx.Next()

	if err == nil {
		return nil
	}

	// Обрабатываем ошибки сервиса аутентификации
	switch {
	case errors.Is(err, service.ErrInvalidAdminToken):
		return ctx.Status(fiber.StatusForbidden).JSON(model.Response{
			Data: fiber.Map{
				"code":    fiber.StatusForbidden,
				"message": "Invalid admin token",
			},
		})
	case errors.Is(err, service.ErrInvalidCredentials):
		return ctx.Status(fiber.StatusUnauthorized).JSON(model.Response{
			Data: fiber.Map{
				"code":    fiber.StatusUnauthorized,
				"message": "Invalid credentials",
			},
		})
	case errors.Is(err, service.ErrTokenExpired):
		return ctx.Status(fiber.StatusUnauthorized).JSON(model.Response{
			Data: fiber.Map{
				"code":    fiber.StatusUnauthorized,
				"message": "Token expired",
			},
		})
	}

	// Обрабатываем Fiber ошибки
	if e, ok := err.(*fiber.Error); ok {
		return ctx.Status(e.Code).JSON(model.Response{
			Data: fiber.Map{
				"code":    e.Code,
				"message": e.Message,
			},
		})
	}

	// Все остальные ошибки
	return ctx.Status(fiber.StatusInternalServerError).JSON(model.Response{
		Data: fiber.Map{
			"code":    fiber.StatusInternalServerError,
			"message": "Internal server error",
		},
	})
}

// ErrorHandler middleware для обработки ошибок контроллера документов
func DocsErrorHandler(ctx *fiber.Ctx) error {
	// Продолжаем цепочку middleware/обработчиков
	err := ctx.Next()

	// Если ошибок не было, просто выходим
	if err == nil {
		return nil
	}

	// Обрабатываем Fiber ошибки
	if e, ok := err.(*fiber.Error); ok {

		return ctx.Status(e.Code).JSON(model.Response{
			Data: fiber.Map{
				"code":    e.Code,
				"message": e.Message,
			},
		})
	}

	// Обрабатываем кастомные ошибки сервиса
	switch err.Error() {
	case "document not found":
		return ctx.Status(fiber.StatusNotFound).JSON(model.Response{
			Data: fiber.Map{
				"code":    fiber.StatusNotFound,
				"message": "Document not found",
			},
		})
	case "permission denied":
		return ctx.Status(fiber.StatusNotFound).JSON(model.Response{
			Data: fiber.Map{
				"code":    fiber.StatusForbidden,
				"message": "Permission denied",
			},
		})
	case "invalid document data":
		return ctx.Status(fiber.StatusNotFound).JSON(model.Response{
			Data: fiber.Map{
				"code":    fiber.StatusBadRequest,
				"message": "Invalid document data",
			},
		})
	}

	// Все остальные ошибки считаем внутренними серверными
	return ctx.Status(fiber.StatusInternalServerError).JSON(model.Response{
		Data: fiber.Map{
			"code":    fiber.StatusInternalServerError,
			"message": "Invalid document data",
		},
	})
}
