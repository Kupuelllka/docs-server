package documents_test

import (
	"docs-server/cmd/tests/testutils"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetDocumentsList_Success(t *testing.T) {
	app := testutils.TestApp
	token := testutils.TestToken

	t.Logf("Using test token: %s", token) // Добавьте это для отладки

	req := httptest.NewRequest("GET", "/api/docs", nil)
	req.Header.Set("Authorization", token)

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("Request error: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		t.Fatalf("Expected 200, got %d. Body: %s", resp.StatusCode, string(body))
	}

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var result map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&result)
	assert.NoError(t, err)

	data, ok := result["data"].(map[string]interface{})
	assert.True(t, ok)
	_, ok = data["docs"].([]interface{})
	assert.True(t, ok)
}
func TestGetDocumentWorkflow(t *testing.T) {
	app := testutils.TestApp
	token := testutils.TestToken

	// 1. Создаем тестовый документ
	t.Run("Create document", func(t *testing.T) {
		meta := `{
            "name": "workflow_test.txt",
            "file": true,
            "public": false,
            "mime": "text/plain",
            "grant": ["testuser1"],
            "json": {"purpose": "test workflow"},
            "token": "` + token + `"
        }`

		body, contentType := testutils.CreateMultipartRequest(meta, "workflow_test.txt", "test content 123")
		req := httptest.NewRequest("POST", "/api/docs", body)
		req.Header.Set("Content-Type", contentType)
		req.Header.Set("Authorization", token)

		resp, err := app.Test(req)
		assert.NoError(t, err)
		defer resp.Body.Close()
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var result struct {
			Data struct {
				File string `json:"file"`
			} `json:"data"`
		}
		err = json.NewDecoder(resp.Body).Decode(&result)
		assert.NoError(t, err)
		assert.NotEmpty(t, result.Data.File)
	})

	// 2. Получаем список документов и находим наш
	var docID string
	t.Run("Find document in list", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/docs", nil)
		req.Header.Set("Authorization", token)

		resp, err := app.Test(req)
		assert.NoError(t, err)
		defer resp.Body.Close()
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var list struct {
			Data struct {
				Docs []struct {
					ID   string `json:"id"`
					Name string `json:"name"`
				} `json:"docs"`
			} `json:"data"`
		}

		err = json.NewDecoder(resp.Body).Decode(&list)
		assert.NoError(t, err)
		assert.Greater(t, len(list.Data.Docs), 0)

		// Ищем наш документ по имени
		for _, doc := range list.Data.Docs {
			if doc.Name == "workflow_test.txt" {
				docID = doc.ID
				break
			}
		}

		assert.NotEmpty(t, docID, "Document not found in list")
	})

	// 3. Получаем документ по найденному ID
	t.Run("Get document by ID", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/docs/"+docID, nil)
		req.Header.Set("Authorization", token)

		resp, err := app.Test(req)
		assert.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)
		assert.Equal(t, "text/plain; charset=utf-8", resp.Header.Get("Content-Type"))

		content, err := io.ReadAll(resp.Body)
		assert.NoError(t, err)
		assert.Equal(t, "test content 123", string(content))
	})
}

func TestDocumentWithGrants(t *testing.T) {
	app := testutils.TestApp
	token := testutils.TestToken
	tokenOtherUser := testutils.TestToken2

	// 1. Создаем тестовый документ
	t.Run("Create document", func(t *testing.T) {
		meta := `{
            "name": "workflow_test.txt",
            "file": true,
            "public": false,
            "mime": "text/plain",
            "grant": ["testuser1"],
            "json": {"purpose": "test workflow"},
            "token": "` + token + `"
        }`

		body, contentType := testutils.CreateMultipartRequest(meta, "workflow_test.txt", "test content 123")
		req := httptest.NewRequest("POST", "/api/docs", body)
		req.Header.Set("Content-Type", contentType)
		req.Header.Set("Authorization", token)

		resp, err := app.Test(req)
		assert.NoError(t, err)
		defer resp.Body.Close()
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var result struct {
			Data struct {
				File string `json:"file"`
			} `json:"data"`
		}
		err = json.NewDecoder(resp.Body).Decode(&result)
		assert.NoError(t, err)
		assert.NotEmpty(t, result.Data.File)
	})

	// 2. Получаем список документов и находим наш
	var docID string
	t.Run("Find document in list", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/docs", nil)
		req.Header.Set("Authorization", token)

		resp, err := app.Test(req)
		assert.NoError(t, err)
		defer resp.Body.Close()
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var list struct {
			Data struct {
				Docs []struct {
					ID   string `json:"id"`
					Name string `json:"name"`
				} `json:"docs"`
			} `json:"data"`
		}

		err = json.NewDecoder(resp.Body).Decode(&list)
		assert.NoError(t, err)
		assert.Greater(t, len(list.Data.Docs), 0)

		// Ищем наш документ по имени
		for _, doc := range list.Data.Docs {
			if doc.Name == "workflow_test.txt" {
				docID = doc.ID
				break
			}
		}

		assert.NotEmpty(t, docID, "Document not found in list")
	})

	// 3. Получаем документ по найденному ID
	t.Run("Get document by ID", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/docs/"+docID, nil)
		req.Header.Set("Authorization", tokenOtherUser)

		resp, err := app.Test(req)
		assert.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)
		assert.Equal(t, "text/plain; charset=utf-8", resp.Header.Get("Content-Type"))

		content, err := io.ReadAll(resp.Body)
		assert.NoError(t, err)
		assert.Equal(t, "test content 123", string(content))
	})
}
