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
	"sync"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
)

var (
	TestApp    *fiber.App
	TestToken  string
	TestToken2 string
	TestLogin  = "testuser" // Фиксированный логин для тестов
	TestLogin2 = "testuser1"
	TestPass   = "Secur3P@ss"

	initOnce sync.Once
)

func init() {
	initOnce.Do(func() {
		// Инициализация тестового приложения
		TestApp = createTestApp()
		// Регистрация и аутентификация тестового пользователя
		registerTestUser(TestApp)
		registerTestListUser(TestApp)
		TestToken = getTestToken(TestLogin, TestApp)
		TestToken2 = getTestToken(TestLogin2, TestApp)
	})
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

	authService := service.NewAuthService(userRepo, cfg.Auth.AdminToken, []byte(cfg.Auth.JWTSecret), cache)
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

	docs := api.Group("/docs", controller.AuthMiddleware(authService))
	docs.Post("/", docsController.UploadDocument)
	docs.Get("/", docsController.GetDocumentsList)
	docs.Get("/:id", docsController.GetDocument)
	docs.Delete("/:id", docsController.DeleteDocument)

	return application
}

func registerTestUser(app *fiber.App) {
	// Регистрация тестового пользователя (если еще не существует)
	regBody := map[string]string{
		"token": "secure-admin-token-123",
		"login": TestLogin,
		"pswd":  TestPass,
	}
	jsonBody, _ := json.Marshal(regBody)

	req := httptest.NewRequest("POST", "/api/register", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	app.Test(req)
}
func registerTestListUser(app *fiber.App) {
	// Регистрация тестового пользователя (если еще не существует)
	regBody1 := map[string]string{
		"token": "secure-admin-token-123",
		"login": "testuser1",
		"pswd":  TestPass,
	}
	// Регистрация тестового пользователя (если еще не существует)
	regBody2 := map[string]string{
		"token": "secure-admin-token-123",
		"login": "testuser2",
		"pswd":  TestPass,
	}
	jsonBody, _ := json.Marshal(regBody1)

	req := httptest.NewRequest("POST", "/api/register", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	app.Test(req)

	jsonBody, _ = json.Marshal(regBody2)

	req = httptest.NewRequest("POST", "/api/register", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	app.Test(req)
}
func getTestToken(login string, app *fiber.App) string {
	// Аутентификация
	authBody := map[string]string{
		"login": login,
		"pswd":  TestPass,
	}
	jsonBody, _ := json.Marshal(authBody)

	req := httptest.NewRequest("POST", "/api/auth", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	resp, _ := app.Test(req)

	var result map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&result)
	return result["response"].(map[string]interface{})["token"].(string)
}
