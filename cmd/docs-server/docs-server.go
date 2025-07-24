package main

import (
	"docs-server/internal/controller"
	"log"

	"docs-server/internal/app"
	"docs-server/internal/cache"
	"docs-server/internal/repository"
	"docs-server/internal/service"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
)

func main() {
	// Инициализация конфигурации
	cfg, err := app.NewConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Инициализация репозиториев
	userRepo := repository.NewUserRepository(cfg.Database.DSN)
	defer userRepo.Close()

	docRepo := repository.NewDocumentRepository(cfg.Database.DSN)
	defer docRepo.Close()

	// Инициализация кеша
	cache := cache.NewMemoryCache()

	// Инициализация сервисов
	authService := service.NewAuthService(userRepo, cfg.Auth.AdminToken, []byte(cfg.Auth.JWTSecret), cache)
	docService := service.NewDocumentService(docRepo, userRepo, cache, cfg.Storage.UploadDir)
	userService := service.NewUserService(userRepo)

	// Создание Fiber приложения
	application := fiber.New(fiber.Config{
		ErrorHandler: app.ErrorHandler,
	})

	// Middleware
	application.Use(recover.New())
	application.Use(logger.New())

	// Инициализация контроллеров
	authController := controller.NewAuthController(authService)

	docsController := controller.NewDocsController(docService, userService)

	// Настройка маршрутов
	api := application.Group("/api")
	api.Post("/register", authController.Register)
	api.Post("/auth", authController.Authenticate)
	api.Delete("/auth/:token", authController.Logout)

	docs := api.Group("/docs", controller.AuthMiddleware(*authService))
	docs.Post("/", docsController.UploadDocument)
	docs.Get("/", docsController.GetDocumentsList)
	docs.Get("/:id", docsController.GetDocument)
	docs.Delete("/:id", docsController.DeleteDocument)

	// Запуск сервера
	log.Fatal(application.Listen(":" + cfg.Server.Port))
}
