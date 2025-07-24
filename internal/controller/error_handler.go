package controller

import (
	"docs-server/internal/model"
	"docs-server/internal/service"
	"errors"

	"github.com/gofiber/fiber/v2"
)

var (
	// Общие ошибки
	ErrInternalServer = errors.New("internal server error")
)

// UnifiedErrorHandler универсальный обработчик ошибок для всех сервисов
func UnifiedErrorHandler(ctx *fiber.Ctx) error {
	err := ctx.Next()
	if err == nil {
		return nil
	}

	// Определяем HTTP статус и сообщение по умолчанию
	status, message := fiber.StatusInternalServerError, ErrInternalServer.Error()

	// Обрабатываем стандартные Fiber ошибки
	var fiberErr *fiber.Error
	if errors.As(err, &fiberErr) {
		status = fiberErr.Code
		message = fiberErr.Message
	} else {
		// Обрабатываем кастомные ошибки
		switch {
		// Аутентификация
		case errors.Is(err, service.ErrInvalidAdminToken):
			status, message = fiber.StatusForbidden, err.Error()
		case errors.Is(err, service.ErrInvalidCredentials),
			errors.Is(err, service.ErrTokenExpired),
			errors.Is(err, service.ErrInvalidToken):
			status, message = fiber.StatusUnauthorized, err.Error()

		// Документы
		case errors.Is(err, service.ErrDocumentNameRequired),
			errors.Is(err, service.ErrInvalidMetaFormat),
			errors.Is(err, service.ErrInvalidDocumentData):
			status, message = fiber.StatusBadRequest, err.Error()
		case errors.Is(err, service.ErrDocumentNotFound):
			status, message = fiber.StatusNotFound, err.Error()
		case errors.Is(err, service.ErrPermissionDenied),
			errors.Is(err, service.ErrNotDocumentOwner),
			errors.Is(err, service.ErrForbidden):
			status, message = fiber.StatusForbidden, err.Error()
		case errors.Is(err, service.ErrFailedToCreateDir),
			errors.Is(err, service.ErrFailedToSaveFile),
			errors.Is(err, service.ErrFailedToDeleteFile):
			status, message = fiber.StatusInternalServerError, err.Error()

		// Пользователи
		case errors.Is(err, service.ErrUserIDEmpty),
			errors.Is(err, service.ErrLoginEmpty),
			errors.Is(err, service.ErrUserNil):
			status, message = fiber.StatusBadRequest, err.Error()
		case errors.Is(err, service.ErrUserNotFound):
			status, message = fiber.StatusNotFound, err.Error()
		case errors.Is(err, service.ErrInvalidLimit),
			errors.Is(err, service.ErrInvalidOffset):
			status, message = fiber.StatusBadRequest, err.Error()
		}
	}

	// Формируем ответ
	return ctx.Status(status).JSON(model.Response{
		Data: fiber.Map{
			"code":    status,
			"message": message,
		},
	})
}

// Deprecated: используйте UnifiedErrorHandler вместо этих обработчиков
func AuthErrorHandler(ctx *fiber.Ctx) error {
	return UnifiedErrorHandler(ctx)
}

// Deprecated: используйте UnifiedErrorHandler вместо этих обработчиков
func DocsErrorHandler(ctx *fiber.Ctx) error {
	return UnifiedErrorHandler(ctx)
}
