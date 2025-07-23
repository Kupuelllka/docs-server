package documents_test

import (
	"docs-server/cmd/tests/testutils"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetDocumentsList_Success(t *testing.T) {
	app := testutils.TestApp
	token := testutils.TestToken

	req := httptest.NewRequest("GET", "/api/docs", nil)
	req.Header.Set("Authorization", token)

	resp, err := app.Test(req)
	assert.NoError(t, err)
	defer resp.Body.Close()

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

	// Сначала загружаем документ для теста
	meta := `{"name":"testfile.txt","public":false,"mime":"text/plain"}`
	body, contentType := testutils.CreateMultipartRequest(meta, "testfile.txt", "test content")

	uploadReq := httptest.NewRequest("POST", "/api/docs", body)
	uploadReq.Header.Set("Content-Type", contentType)
	uploadReq.Header.Set("Authorization", token)

	uploadResp, err := app.Test(uploadReq)
	assert.NoError(t, err)
	defer uploadResp.Body.Close()

	docID, err := testutils.GetDocumentIDFromResponse(uploadResp)
	assert.NoError(t, err)

	// Теперь получаем документ
	req := httptest.NewRequest("GET", "/api/docs/"+docID, nil)
	req.Header.Set("Authorization", token)

	resp, err := app.Test(req)
	assert.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	contentType = resp.Header.Get("Content-Type")
	assert.True(t, strings.Contains(contentType, "application/octet-stream"))
}
