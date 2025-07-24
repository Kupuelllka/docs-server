package controller

import (
	"docs-server/internal/model"
	"docs-server/internal/service"
	"errors"

	"github.com/gofiber/fiber/v2"
)

// AuthController авторизация
type AuthController struct {
	authService *service.AuthService
}

func NewAuthController(authService *service.AuthService) *AuthController {
	return &AuthController{authService: authService}
}

func (c *AuthController) GetAuthService() *service.AuthService {
	return c.authService
}

// Register устанавливает маршруты для регистрации
func (c *AuthController) Register(ctx *fiber.Ctx) error {
	type RegisterRequest struct {
		Token string `json:"token"`
		Login string `json:"login"`
		Pswd  string `json:"pswd"`
	}

	var req RegisterRequest
	if err := ctx.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid request body")
	}

	err := c.authService.Register(req.Token, req.Login, req.Pswd)
	if err != nil {
		// Проверяем, является ли ошибка ошибкой неверного токена
		if errors.Is(err, service.ErrInvalidAdminToken) {
			return fiber.NewError(fiber.StatusUnauthorized, "Invalid admin token")
		}
		return err // Возвращаем другие ошибки как есть
	}

	return ctx.JSON(model.Response{
		Response: fiber.Map{
			"login": req.Login,
		},
	})
}

// Authenticate установить авторизацию
func (c *AuthController) Authenticate(ctx *fiber.Ctx) error {
	type AuthRequest struct {
		Login string `json:"login"`
		Pswd  string `json:"pswd"`
	}

	var req AuthRequest
	if err := ctx.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid request body")
	}

	token, err := c.authService.Authenticate(req.Login, req.Pswd)
	if err != nil {
		return err
	}

	return ctx.JSON(model.Response{
		Response: fiber.Map{
			"token": token,
		},
	})
}

// Logout выйти из системы
func (c *AuthController) Logout(ctx *fiber.Ctx) error {
	token := ctx.Params("token")
	if token == "" {
		return fiber.NewError(fiber.StatusBadRequest, "Token required")
	}

	if err := c.authService.Logout(token); err != nil {
		return err
	}

	return ctx.JSON(model.Response{
		Response: fiber.Map{
			token: true,
		},
	})
}
