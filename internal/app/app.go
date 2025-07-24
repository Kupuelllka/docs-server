package app

import (
	"docs-server/internal/cache"
	"docs-server/internal/controller"
	"docs-server/internal/repository"
	"docs-server/internal/service"

	"github.com/gofiber/fiber/v2"
)

type App struct {
	*fiber.App
	cfg *Config
}

func NewApp(fiberApp *fiber.App, cfg *Config) (*App, error) {
	app := &App{
		App: fiberApp,
		cfg: cfg,
	}

	// Инициализация репозиториев
	userRepo := repository.NewUserRepository(cfg.Database.DSN)
	docRepo := repository.NewDocumentRepository(cfg.Database.DSN)

	// Инициализация кеша
	cache := cache.NewMemoryCache()
	// Инициализация сервисов
	authService := service.NewAuthService(userRepo, cfg.Auth.AdminToken, []byte(cfg.Auth.JWTSecret))
	userService := service.NewUserService(userRepo)
	docService := service.NewDocumentService(docRepo, userRepo, cache, cfg.Storage.UploadDir)

	// Инициализация контроллеров
	authController := controller.NewAuthController(authService)
	docsController := controller.NewDocsController(docService, userService)

	// Настройка маршрутов
	app.setupRoutes(authController, docsController)

	return app, nil
}

func (a *App) setupRoutes(authCtrl *controller.AuthController, docsCtrl *controller.DocsController) {
	api := a.Group("/api")

	// Маршруты для авторизации
	api.Post("/register", authCtrl.Register)
	api.Post("/auth", authCtrl.Authenticate)
	api.Delete("/auth/:token", authCtrl.Logout)

	// Маршруты для документы
	docs := api.Group("/docs", controller.AuthMiddleware(*authCtrl.GetAuthService()))
	docs.Post("/", docsCtrl.UploadDocument)
	docs.Get("/", docsCtrl.GetDocumentsList)
	docs.Get("/:id", docsCtrl.GetDocument)
	docs.Delete("/:id", docsCtrl.DeleteDocument)
}

func ErrorHandler(ctx *fiber.Ctx, err error) error {
	// Обработка ошибок и формирование стандартного ответа
	code := fiber.StatusInternalServerError
	message := "Internal Server Error"

	if e, ok := err.(*fiber.Error); ok {
		code = e.Code
		message = e.Message
	}

	return ctx.Status(code).JSON(fiber.Map{
		"error": fiber.Map{
			"code": code,
			"text": message,
		},
	})
}
