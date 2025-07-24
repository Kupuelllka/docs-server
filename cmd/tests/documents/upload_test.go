package documents_test

import (
	"docs-server/cmd/tests/testutils"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
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

func TestUploadDocument_ImageFile(t *testing.T) {
	app := testutils.TestApp
	token := testutils.TestToken

	// Читаем тестовое изображение из файла
	imgData, err := os.ReadFile("../data/test_image.jpg")
	if err != nil {
		t.Fatalf("Failed to read test image: %v", err)
	}

	meta := `{
        "name": "testimage.jpg",
        "public": true,
        "mime": "image/jpeg",
        "grant": ["user1"],
        "json": {"title": "test image", "category": "photos"}
    }`

	body, contentType := testutils.CreateMultipartRequest(meta, "testimage.jpg", string(imgData))

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

	fileName, ok := data["file"].(string)
	assert.True(t, ok)
	assert.Contains(t, fileName, ".jpg")

	jsonData, ok := data["json"].(map[string]interface{})
	assert.True(t, ok)
	assert.Equal(t, "test image", jsonData["title"])
}

func TestUploadDocument_PDFFile(t *testing.T) {
	app := testutils.TestApp
	token := testutils.TestToken

	// Читаем тестовый PDF из файла
	pdfData, err := os.ReadFile("../data/test_pdf.pdf")
	if err != nil {
		t.Fatalf("Failed to read test PDF: %v", err)
	}

	meta := `{
        "name": "testdoc.pdf",
        "public": false,
        "mime": "application/pdf",
        "grant": ["user1", "user2"],
        "json": {"title": "Test Document", "pages": 10}
    }`

	body, contentType := testutils.CreateMultipartRequest(meta, "testdoc.pdf", string(pdfData))

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

	data := result["data"].(map[string]interface{})
	fileName := data["file"].(string)
	assert.Contains(t, fileName, ".pdf")

	jsonData := data["json"].(map[string]interface{})
	assert.Equal(t, "Test Document", jsonData["title"])
}

func TestUploadDocument_ExcelFile(t *testing.T) {
	app := testutils.TestApp
	token := testutils.TestToken

	// Читаем тестовый Excel файл
	excelData, err := os.ReadFile("../data/testdata.xls")
	if err != nil {
		t.Fatalf("Failed to read test Excel file: %v", err)
	}

	meta := `{
        "name": "financial_report.xls",
        "public": false,
        "mime": "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet",
        "grant": ["accounting"],
        "json": {"report_type": "quarterly", "year": 2023}
    }`

	body, contentType := testutils.CreateMultipartRequest(meta, "financial_report.xls", string(excelData))

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

	data := result["data"].(map[string]interface{})
	fileName := data["file"].(string)
	fmt.Println(fileName)
	assert.Contains(t, fileName, ".xls")

	jsonData := data["json"].(map[string]interface{})
	assert.Equal(t, "quarterly", jsonData["report_type"])
}
