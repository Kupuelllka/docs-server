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

	// 1. Upload test document
	meta := `{
        "name": "testfile.txt",
        "file": true,
        "public": false,
        "mime": "text/plain",
        "grant": ["testuser1", "testuser2"],
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

	// 2. Get document ID from list
	listReq := httptest.NewRequest("GET", "/api/docs", nil)
	listReq.Header.Set("Authorization", token)

	listResp, err := app.Test(listReq)
	assert.NoError(t, err)
	defer listResp.Body.Close()
	assert.Equal(t, http.StatusOK, listResp.StatusCode)

	// Parse response to find our test document
	var listResponse struct {
		Data struct {
			Docs []struct {
				ID   string `json:"id"`
				Name string `json:"name"`
			} `json:"docs"`
		} `json:"data"`
	}
	err = json.NewDecoder(listResp.Body).Decode(&listResponse)
	assert.NoError(t, err)

	var docID string
	for _, doc := range listResponse.Data.Docs {
		if doc.Name == "testfile.txt" {
			docID = doc.ID
			break
		}
	}
	assert.NotEmpty(t, docID, "Document ID not found in list")

	// 3. Delete the document
	deleteReq := httptest.NewRequest("DELETE", "/api/docs/"+docID, nil)
	deleteReq.Header.Set("Authorization", token)

	deleteResp, err := app.Test(deleteReq)
	assert.NoError(t, err)
	defer deleteResp.Body.Close()
	assert.Equal(t, http.StatusOK, deleteResp.StatusCode)

	// 4. Verify deletion by checking the list again
	listReq = httptest.NewRequest("GET", "/api/docs", nil)
	listReq.Header.Set("Authorization", token)

	listResp, err = app.Test(listReq)
	assert.NoError(t, err)
	defer listResp.Body.Close()
	assert.Equal(t, http.StatusOK, listResp.StatusCode)

	// Parse response again
	var postDeleteResponse struct {
		Data struct {
			Docs []struct {
				ID string `json:"id"`
			} `json:"docs"`
		} `json:"data"`
	}
	err = json.NewDecoder(listResp.Body).Decode(&postDeleteResponse)
	assert.NoError(t, err)

	// Verify document is no longer in the list
	found := false
	for _, doc := range postDeleteResponse.Data.Docs {
		if doc.ID == docID {
			found = true
			break
		}
	}
	assert.False(t, found, "Document should be deleted but still exists in list")
}
