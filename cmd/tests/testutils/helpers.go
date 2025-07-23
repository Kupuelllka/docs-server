package testutils

import (
	"bytes"
	"docs-server/internal/app"
	"docs-server/internal/cache"
	"docs-server/internal/controller"
	"docs-server/internal/repository"
	"docs-server/internal/service"
	"encoding/json"
	"log"
	"net/http/httptest"
	"os"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
)

var TestApp *fiber.App
var TestToken string

func init() {
	// Инициализация тестового приложения
	TestApp = createTestApp()
	TestToken = getTestToken(TestApp)
}

func Cleanup() {
	os.RemoveAll("./test_uploads")
}

func createTestApp() *fiber.App {
	cfg, err := app.NewConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	userRepo := repository.NewUserRepository(cfg.Database.DSN)
	docRepo := repository.NewDocumentRepository(cfg.Database.DSN)
	cache := cache.NewMemoryCache()

	authService := service.NewAuthService(userRepo, cfg.Auth.AdminToken)
	docService := service.NewDocumentService(docRepo, userRepo, cache, cfg.Storage.UploadDir)
	userService := service.NewUserService(userRepo)

	application := fiber.New(fiber.Config{
		ErrorHandler: app.ErrorHandler,
	})

	application.Use(recover.New())
	application.Use(logger.New())

	authController := controller.NewAuthController(authService)
	docsController := controller.NewDocsController(docService, userService)

	api := application.Group("/api")
	api.Post("/register", authController.Register)
	api.Post("/auth", authController.Authenticate)
	api.Delete("/auth/:token", authController.Logout)

	docs := api.Group("/docs", controller.AuthMiddleware(*authService))
	docs.Post("/", docsController.UploadDocument)
	docs.Get("/", docsController.GetDocumentsList)
	docs.Get("/:id", docsController.GetDocument)
	docs.Delete("/:id", docsController.DeleteDocument)

	return application
}

func getTestToken(app *fiber.App) string {
	// Регистрация тестового пользователя
	regBody := map[string]string{
		"token": "secure-admin-token-123",
		"login": "testuser",
		"pswd":  "Secur3P@ss",
	}
	jsonBody, _ := json.Marshal(regBody)

	req := httptest.NewRequest("POST", "/api/register", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	app.Test(req)

	// Аутентификация
	authBody := map[string]string{
		"login": "testuser",
		"pswd":  "Secur3P@ss",
	}
	jsonBody, _ = json.Marshal(authBody)

	req = httptest.NewRequest("POST", "/api/auth", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	resp, _ := app.Test(req)

	var result map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&result)
	return result["response"].(map[string]interface{})["token"].(string)
}
