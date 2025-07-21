package controller

import (
	"docs-server/internal/model"
	"docs-server/internal/service"
	"io"

	"github.com/gofiber/fiber/v2"
)

type DocsController struct {
	docService  *service.DocumentService
	userService *service.UserService
}

func NewDocsController(docService *service.DocumentService, userService *service.UserService) *DocsController {
	return &DocsController{
		docService:  docService,
		userService: userService,
	}
}

// UploadDocument Метод загрузки документа
func (c *DocsController) UploadDocument(ctx *fiber.Ctx) error {
	// Получение токена из заголовков
	token := ctx.Get("Authorization")
	if token == "" {
		return fiber.NewError(fiber.StatusUnauthorized, "Authorization token required")
	}

	// Парсинг multipart формы
	form, err := ctx.MultipartForm()
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid form data")
	}

	// Получение метаданных
	meta := form.Value["meta"]
	if len(meta) == 0 {
		return fiber.NewError(fiber.StatusBadRequest, "Meta data required")
	}

	// Преобразование файлов в нужный формат
	var uploadedFiles []*model.UploadedFile
	if files, ok := form.File["file"]; ok {
		for _, fileHeader := range files {
			file, err := fileHeader.Open()
			if err != nil {
				return fiber.NewError(fiber.StatusInternalServerError, "Failed to open file")
			}
			defer file.Close()

			data, err := io.ReadAll(file)
			if err != nil {
				return fiber.NewError(fiber.StatusInternalServerError, "Failed to read file")
			}

			uploadedFiles = append(uploadedFiles, &model.UploadedFile{
				Filename: fileHeader.Filename,
				Data:     data,
				Size:     fileHeader.Size,
			})
		}
	}

	// Вызов сервиса с преобразованными файлами
	doc, err := c.docService.UploadDocument(token, meta[0], uploadedFiles)
	if err != nil {
		return err
	}

	return ctx.JSON(model.Response{
		Data: fiber.Map{
			"json": doc.JSONData,
			"file": doc.Name,
		},
	})
}

// GetDocumentsList Получить список документов
func (c *DocsController) GetDocumentsList(ctx *fiber.Ctx) error {
	token := ctx.Get("Authorization")
	if token == "" {
		return fiber.NewError(fiber.StatusUnauthorized, "Authorization token required")
	}

	// Получение параметров запроса
	login := ctx.Query("login")
	key := ctx.Query("key")
	value := ctx.Query("value")
	limit := ctx.QueryInt("limit", 10)

	docs, err := c.docService.GetDocumentsList(token, login, key, value, limit)
	if err != nil {
		return err
	}

	return ctx.JSON(model.Response{
		Data: fiber.Map{
			"docs": docs,
		},
	})
}

// GetDocument Получить документ
func (c *DocsController) GetDocument(ctx *fiber.Ctx) error {
	token := ctx.Get("Authorization")
	if token == "" {
		return fiber.NewError(fiber.StatusUnauthorized, "Authorization token required")
	}

	id := ctx.Params("id")
	if id == "" {
		return fiber.NewError(fiber.StatusBadRequest, "Document ID required")
	}

	doc, err := c.docService.GetDocument(token, id)
	if err != nil {
		return err
	}

	if doc.File {
		// Если это файл, отправить его
		return ctx.SendFile(doc.FilePath)
	}

	// Если это JSON, вернуть данные
	return ctx.JSON(model.Response{
		Data: doc.JSONData,
	})
}

// DeleteDocument Удалить документ
func (c *DocsController) DeleteDocument(ctx *fiber.Ctx) error {
	token := ctx.Get("Authorization")
	if token == "" {
		return fiber.NewError(fiber.StatusUnauthorized, "Authorization token required")
	}

	id := ctx.Params("id")
	if id == "" {
		return fiber.NewError(fiber.StatusBadRequest, "Document ID required")
	}

	success, err := c.docService.DeleteDocument(token, id)
	if err != nil {
		return err
	}

	return ctx.JSON(model.Response{
		Response: fiber.Map{
			id: success,
		},
	})
}
