package documents_test

import (
	"docs-server/cmd/tests/testutils"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDeleteDocument_Success(t *testing.T) {
	app := testutils.TestApp
	token := testutils.TestToken

	// Сначала загружаем документ для теста
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

	docID, err := testutils.GetDocumentIDFromResponse(uploadResp)
	assert.NoError(t, err)

	// Теперь удаляем документ
	req := httptest.NewRequest("DELETE", "/api/docs/"+docID, nil)
	req.Header.Set("Authorization", token)

	resp, err := app.Test(req)
	assert.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var result map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&result)
	assert.NoError(t, err)

	response, ok := result["response"].(map[string]interface{})
	assert.True(t, ok)
	_, ok = response[docID].(bool)
	assert.True(t, ok)
}
