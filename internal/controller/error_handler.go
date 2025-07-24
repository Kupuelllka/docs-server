package controller

import (
	"docs-server/internal/model"
	"docs-server/internal/service"
	"errors"

	"github.com/gofiber/fiber/v2"
)

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
