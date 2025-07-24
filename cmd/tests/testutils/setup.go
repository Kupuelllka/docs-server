package testutils

import (
	"bytes"
	"encoding/json"
	"fmt"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
)

func CreateMultipartRequest(meta string, filename string, content string) (*bytes.Buffer, string) {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	writer.WriteField("meta", meta)
	fileWriter, _ := writer.CreateFormFile("file", filename)
	fileWriter.Write([]byte(content))
	writer.Close()

	return body, writer.FormDataContentType()
}

func CreateJSONRequest(data interface{}) (*bytes.Buffer, error) {
	jsonBody, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}
	return bytes.NewBuffer(jsonBody), nil
}

func ParseResponse(resp *httptest.ResponseRecorder, v interface{}) error {
	return json.NewDecoder(resp.Body).Decode(v)
}

// Вспомогательная функция для получения ID документа
func GetDocumentIDFromResponse(resp *http.Response) (string, error) {
	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}

	data, ok := result["data"].(map[string]interface{})
	if !ok {
		return "", fmt.Errorf("invalid response format: missing data field")
	}

	id, ok := data["id"].(string)
	if !ok {
		return "", fmt.Errorf("invalid response format: missing id field")
	}

	return id, nil
}
