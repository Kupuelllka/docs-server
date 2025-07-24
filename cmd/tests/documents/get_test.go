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

func TestGetDocument_Success(t *testing.T) {
	app := testutils.TestApp
	token := testutils.TestToken

	meta := `{
        "name": "testfile.txt",
        "file": true,
        "public": false,
        "mime": "text/plain",
        "grant": ["testuser", "user2"],
        "json": {"description": "test file", "version": 1},
        "token": "` + token + `"
    }`

	body, contentType := testutils.CreateMultipartRequest(meta, "testfile.txt", "test content")
	uploadReq := httptest.NewRequest("POST", "/api/docs", body)
	uploadReq.Header.Set("Content-Type", contentType)
	uploadReq.Header.Set("Authorization", token)

	uploadResp, err := app.Test(uploadReq)
	assert.NoError(t, err)
	defer uploadResp.Body.Close()
	assert.Equal(t, http.StatusOK, uploadResp.StatusCode)

	listReq := httptest.NewRequest("GET", "/api/docs", nil)
	listReq.Header.Set("Authorization", token)

	listResp, err := app.Test(listReq)
	assert.NoError(t, err)
	defer listResp.Body.Close()
	assert.Equal(t, http.StatusOK, listResp.StatusCode)

	var listResult map[string]interface{}
	err = json.NewDecoder(listResp.Body).Decode(&listResult)
	assert.NoError(t, err)

	data, ok := listResult["data"].(map[string]interface{})
	assert.True(t, ok)

	docsList, ok := data["docs"].([]interface{})
	assert.True(t, ok)
	assert.Greater(t, len(docsList), 0)

	firstDoc, ok := docsList[0].(map[string]interface{})
	assert.True(t, ok)

	docID, ok := firstDoc["id"].(string)
	assert.True(t, ok)
	assert.NotEmpty(t, docID)

	getReq := httptest.NewRequest("GET", "/api/docs/"+docID, nil)
	getReq.Header.Set("Authorization", token)

	getResp, err := app.Test(getReq)
	assert.NoError(t, err)
	defer getResp.Body.Close()

	assert.Equal(t, http.StatusOK, getResp.StatusCode)
	assert.Equal(t, "text/plain; charset=utf-8", getResp.Header.Get("Content-Type"))

	fileContent, err := io.ReadAll(getResp.Body)
	assert.NoError(t, err)
	assert.Equal(t, "test content", string(fileContent))
}
