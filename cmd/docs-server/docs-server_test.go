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
	"github.com/gofiber/fiber/v2/middleware/recover"
)

func TestDocsServer(t *testing.T) {
	app := setupDocsServer()
	defer cleanupTestData()

	t.Run("Регистрация пользователя - успешно", func(t *testing.T) {
		body := map[string]string{
			"token": "admin-secret-token",
			"login": "testuser123",
			"pswd":  "Secur3P@ss",
		}
		jsonBody, _ := json.Marshal(body)

		req := httptest.NewRequest("POST", "/api/register", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)
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
	})

	t.Run("Регистрация - неверный токен администратора", func(t *testing.T) {
		body := map[string]string{
			"token": "wrong-token",
			"login": "testuser123",
			"pswd":  "Secur3P@ss",
		}
		jsonBody, _ := json.Marshal(body)

		req := httptest.NewRequest("POST", "/api/register", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)
		if err != nil {
			t.Fatalf("Ошибка запроса: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusUnauthorized {
			t.Fatalf("Ожидался статус 401, получен %d", resp.StatusCode)
		}
	})

	t.Run("Аутентификация - успешно", func(t *testing.T) {
		body := map[string]string{
			"login": "testuser123",
			"pswd":  "Secur3P@ss",
		}
		jsonBody, _ := json.Marshal(body)

		req := httptest.NewRequest("POST", "/api/auth", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)
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
	})

	t.Run("Загрузка документа - успешно", func(t *testing.T) {
		token := getTestToken(t, app)

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

		resp, err := app.Test(req)
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
	})

	t.Run("Получение списка документов - успешно", func(t *testing.T) {
		token := getTestToken(t, app)

		req := httptest.NewRequest("GET", "/api/docs", nil)
		req.Header.Set("Authorization", "Bearer "+token)

		resp, err := app.Test(req)
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
	})

	t.Run("Получение документа - успешно", func(t *testing.T) {
		token := getTestToken(t, app)
		docID := uploadTestDocument(t, app, token)

		req := httptest.NewRequest("GET", "/api/docs/"+docID, nil)
		req.Header.Set("Authorization", "Bearer "+token)

		resp, err := app.Test(req)
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
	})

	t.Run("Удаление документа - успешно", func(t *testing.T) {
		token := getTestToken(t, app)
		docID := uploadTestDocument(t, app, token)

		req := httptest.NewRequest("DELETE", "/api/docs/"+docID, nil)
		req.Header.Set("Authorization", "Bearer "+token)

		resp, err := app.Test(req)
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
	})
}

// Вспомогательные функции

func setupDocsServer() *fiber.App {
	cfg, err := app.NewConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	userRepo := repository.NewUserRepository(cfg.Database.DSN)
	docRepo := repository.NewDocumentRepository(cfg.Database.DSN)
	cache := cache.NewMemoryCache()
	authService := service.NewAuthService(userRepo, cfg.Auth.AdminToken)
	userService := service.NewUserService(userRepo)
	docService := service.NewDocumentService(docRepo, userRepo, cache, cfg.Storage.UploadDir)

	app := fiber.New(fiber.Config{ErrorHandler: app.ErrorHandler})
	app.Use(recover.New())

	authCtrl := controller.NewAuthController(authService)
	docsCtrl := controller.NewDocsController(docService, userService)

	api := app.Group("/api")
	api.Post("/register", authCtrl.Register)
	api.Post("/auth", authCtrl.Authenticate)
	api.Delete("/auth/:token", authCtrl.Logout)

	docs := api.Group("/docs", controller.AuthMiddleware(*authService))
	docs.Post("/", docsCtrl.UploadDocument)
	docs.Get("/", docsCtrl.GetDocumentsList)
	docs.Get("/:id", docsCtrl.GetDocument)
	docs.Delete("/:id", docsCtrl.DeleteDocument)

	return app
}

func cleanupTestData() {
	os.RemoveAll("./test_uploads")
}

func getTestToken(t *testing.T, app *fiber.App) string {
	// Регистрация тестового пользователя
	regBody := map[string]string{
		"token": "admin-secret-token",
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

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("Ошибка аутентификации: %v", err)
	}

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		t.Fatalf("Ошибка декодирования токена: %v", err)
	}

	return result["response"].(map[string]interface{})["token"].(string)
}

func uploadTestDocument(t *testing.T, app *fiber.App, token string) string {
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

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("Ошибка загрузки тестового документа: %v", err)
	}

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		t.Fatalf("Ошибка декодирования ответа: %v", err)
	}

	data := result["data"].(map[string]interface{})
	return data["id"].(string)
}
