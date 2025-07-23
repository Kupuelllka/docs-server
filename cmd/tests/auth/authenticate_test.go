package auth_test

import (
	"docs-server/cmd/tests/testutils"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAuthenticateUser_Success(t *testing.T) {
	app := testutils.TestApp

	body := map[string]string{
		"login": "testuser",
		"pswd":  "Secur3P@ss",
	}
	jsonBody, _ := testutils.CreateJSONRequest(body)

	req := httptest.NewRequest("POST", "/api/auth", jsonBody)
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	assert.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var result map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&result)
	assert.NoError(t, err)

	response, ok := result["response"].(map[string]interface{})
	assert.True(t, ok)
	_, ok = response["token"].(string)
	assert.True(t, ok)
}
