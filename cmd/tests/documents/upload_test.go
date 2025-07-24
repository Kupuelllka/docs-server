package documents_test

import (
	"docs-server/cmd/tests/testutils"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUploadDocument_Success(t *testing.T) {
	app := testutils.TestApp
	token := testutils.TestToken
	meta := `{
        "name": "testfile.txt",
        "public": false,
        "mime": "text/plain",
        "grant": ["user1", "user2"],
        "json": {"description": "test file", "version": 1}
    }`

	body, contentType := testutils.CreateMultipartRequest(meta, "testfile.txt", "test content")

	req := httptest.NewRequest("POST", "/api/docs", body)
	req.Header.Set("Content-Type", contentType)
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
	_, ok = data["file"].(string)
	assert.True(t, ok)
	_, ok = data["json"].(map[string]interface{})
	assert.True(t, ok)
}
