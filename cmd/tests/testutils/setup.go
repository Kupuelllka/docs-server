package testutils

import (
	"bytes"
	"encoding/json"
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

func GetDocumentIDFromResponse(resp *http.Response) (string, error) {
	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}
	data := result["data"].(map[string]interface{})
	return data["id"].(string), nil
}
