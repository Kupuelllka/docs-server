package main

import (
	"bytes"
	"docs-server/internal/app"
	"docs-server/internal/cache"
	"docs-server/internal/controller"
	"docs-server/internal/repository"
	"docs-server/internal/service"
	"encoding/json"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
)

var testApp *fiber.App
var testToken string

func TestMain(m *testing.M) {
	// Инициализация тестового приложения один раз для всех тестов
	testApp = setupTestApp()
	testToken = getTestToken(testApp)

	// Запуск тестов
	code := m.Run()

	// Очистка после тестов
	cleanupTestData()
	os.Exit(code)
}

func TestRegisterUser_Success(t *testing.T) {
	body := map[string]string{
		"token": "secure-admin-token-123",
		"login": "testuser123",
		"pswd":  "Secur3P@ss",
	}
	jsonBody, _ := json.Marshal(body)

	req := httptest.NewRequest("POST", "/api/register", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	resp, err := testApp.Test(req)
	if err != nil {
		t.Fatalf("Ошибка запроса: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		t.Fatalf("Ожидался статус 200, получен %d. Тело: %s", resp.StatusCode, body)
	}

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		t.Fatalf("Ошибка декодирования JSON: %v", err)
	}

	if response, ok := result["response"].(map[string]interface{}); !ok {
		t.Error("Отсутствует поле response в ответе")
	} else if _, ok := response["login"].(string); !ok {
		t.Error("Отсутствует поле login в ответе")
	}
}

func TestRegisterUser_InvalidAdminToken(t *testing.T) {
	body := map[string]string{
		"token": "wrong-token",
		"login": "testuser123",
		"pswd":  "Secur3P@ss",
	}
	jsonBody, _ := json.Marshal(body)

	req := httptest.NewRequest("POST", "/api/register", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	resp, err := testApp.Test(req)
	if err != nil {
		t.Fatalf("Ошибка запроса: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("Ожидался статус 401, получен %d", resp.StatusCode)
	}
}

func TestAuthenticateUser_Success(t *testing.T) {
	body := map[string]string{
		"login": "testuser123",
		"pswd":  "Secur3P@ss",
	}
	jsonBody, _ := json.Marshal(body)

	req := httptest.NewRequest("POST", "/api/auth", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	resp, err := testApp.Test(req)
	if err != nil {
		t.Fatalf("Ошибка запроса: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		t.Fatalf("Ожидался статус 200, получен %d. Тело: %s", resp.StatusCode, body)
	}

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		t.Fatalf("Ошибка декодирования JSON: %v", err)
	}

	if response, ok := result["response"].(map[string]interface{}); !ok {
		t.Error("Отсутствует поле response в ответе")
	} else if _, ok := response["token"].(string); !ok {
		t.Error("Отсутствует поле token в ответе")
	}
}

func TestUploadDocument_Success(t *testing.T) {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	meta := `{"name":"testfile.txt","public":false,"mime":"text/plain"}`
	writer.WriteField("meta", meta)

	fileWriter, _ := writer.CreateFormFile("file", "testfile.txt")
	fileWriter.Write([]byte("test content"))
	writer.Close()

	req := httptest.NewRequest("POST", "/api/docs", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("Authorization", "Bearer "+testToken)

	resp, err := testApp.Test(req)
	if err != nil {
		t.Fatalf("Ошибка запроса: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		t.Fatalf("Ожидался статус 200, получен %d. Тело: %s", resp.StatusCode, body)
	}

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		t.Fatalf("Ошибка декодирования JSON: %v", err)
	}

	if data, ok := result["data"].(map[string]interface{}); !ok {
		t.Error("Отсутствует поле data в ответе")
	} else {
		if _, ok := data["file"].(string); !ok {
			t.Error("Отсутствует поле file в ответе")
		}
		if _, ok := data["json"].(map[string]interface{}); !ok {
			t.Error("Отсутствует поле json в ответе")
		}
	}
}

func TestGetDocumentsList_Success(t *testing.T) {
	req := httptest.NewRequest("GET", "/api/docs", nil)
	req.Header.Set("Authorization", "Bearer "+testToken)

	resp, err := testApp.Test(req)
	if err != nil {
		t.Fatalf("Ошибка запроса: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		t.Fatalf("Ожидался статус 200, получен %d. Тело: %s", resp.StatusCode, body)
	}

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		t.Fatalf("Ошибка декодирования JSON: %v", err)
	}

	if data, ok := result["data"].(map[string]interface{}); !ok {
		t.Error("Отсутствует поле data в ответе")
	} else if _, ok := data["docs"].([]interface{}); !ok {
		t.Error("Отсутствует поле docs в ответе")
	}
}

func TestGetDocument_Success(t *testing.T) {
	docID := uploadTestDocument(testApp, testToken)

	req := httptest.NewRequest("GET", "/api/docs/"+docID, nil)
	req.Header.Set("Authorization", "Bearer "+testToken)

	resp, err := testApp.Test(req)
	if err != nil {
		t.Fatalf("Ошибка запроса: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		t.Fatalf("Ожидался статус 200, получен %d. Тело: %s", resp.StatusCode, body)
	}

	contentType := resp.Header.Get("Content-Type")
	if !strings.Contains(contentType, "application/octet-stream") {
		t.Errorf("Ожидался Content-Type application/octet-stream, получен %s", contentType)
	}
}

func TestDeleteDocument_Success(t *testing.T) {
	docID := uploadTestDocument(testApp, testToken)

	req := httptest.NewRequest("DELETE", "/api/docs/"+docID, nil)
	req.Header.Set("Authorization", "Bearer "+testToken)

	resp, err := testApp.Test(req)
	if err != nil {
		t.Fatalf("Ошибка запроса: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		t.Fatalf("Ожидался статус 200, получен %d. Тело: %s", resp.StatusCode, body)
	}

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		t.Fatalf("Ошибка декодирования JSON: %v", err)
	}

	if response, ok := result["response"].(map[string]interface{}); !ok {
		t.Error("Отсутствует поле response в ответе")
	} else if _, ok := response[docID].(bool); !ok {
		t.Error("Отсутствует ID документа в ответе")
	}
}

// Вспомогательные функции

func setupTestApp() *fiber.App {
	// Инициализация конфигурации
	cfg, err := app.NewConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Инициализация репозиториев
	userRepo := repository.NewUserRepository(cfg.Database.DSN)

	docRepo := repository.NewDocumentRepository(cfg.Database.DSN)

	// Инициализация кеша
	cache := cache.NewMemoryCache()

	// Инициализация сервисов
	authService := service.NewAuthService(userRepo, cfg.Auth.AdminToken)
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

	return application
}

func cleanupTestData() {
	os.RemoveAll("./test_uploads")
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

func uploadTestDocument(app *fiber.App, token string) string {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	meta := `{"name":"testfile.txt","public":false,"mime":"text/plain"}`
	writer.WriteField("meta", meta)

	fileWriter, _ := writer.CreateFormFile("file", "testfile.txt")
	fileWriter.Write([]byte("test content"))
	writer.Close()

	req := httptest.NewRequest("POST", "/api/docs", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("Authorization", "Bearer "+token)

	resp, _ := app.Test(req)

	var result map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&result)

	data := result["data"].(map[string]interface{})
	return data["id"].(string)
}
