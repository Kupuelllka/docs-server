package service

import (
	"context"
	"docs-server/internal/cache"
	"docs-server/internal/model"
	"docs-server/internal/repository"
	"encoding/json"
	"errors"
	"fmt"
	"mime"
	"os"
	"path/filepath"
	"time"

	"github.com/google/uuid"
)

type DocumentService struct {
	docRepo   *repository.DocumentRepository
	userRepo  *repository.UserRepository
	cache     *cache.MemoryCache
	uploadDir string
}

func NewDocumentService(
	docRepo *repository.DocumentRepository,
	userRepo *repository.UserRepository,
	cache *cache.MemoryCache,
	uploadDir string,
) *DocumentService {
	return &DocumentService{
		docRepo:   docRepo,
		userRepo:  userRepo,
		cache:     cache,
		uploadDir: uploadDir,
	}
}

func (s *DocumentService) UploadDocument(token string, meta string, files []*model.UploadedFile) (*model.Document, error) {
	// 1. Валидация токена и получение пользователя
	user, err := s.userRepo.GetUserByToken(context.Background(), token)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, errors.New("unauthorized")
	}

	// 2. Парсинг метаданных
	var metaData struct {
		Name   string      `json:"name"`
		Public bool        `json:"public"`
		Mime   string      `json:"mime"`
		Grant  []string    `json:"grant"`
		JSON   interface{} `json:"json"`
	}

	if err := json.Unmarshal([]byte(meta), &metaData); err != nil {
		return nil, fmt.Errorf("invalid meta format: %v", err)
	}

	// Валидация обязательных полей
	if metaData.Name == "" {
		return nil, errors.New("document name is required")
	}

	// 3. Сохранение файла (если есть)
	var filePath string
	var fileName string
	if len(files) > 0 {
		// Автоматическое определение MIME-типа для файлов
		if metaData.Mime == "" {
			metaData.Mime = mime.TypeByExtension(filepath.Ext(files[0].Filename))
			if metaData.Mime == "" {
				metaData.Mime = "application/octet-stream"
			}
		}

		// Создаем директорию для загрузки
		if err := os.MkdirAll(s.uploadDir, 0755); err != nil {
			return nil, fmt.Errorf("failed to create upload directory: %v", err)
		}

		// Генерируем уникальное имя файла
		fileExt := filepath.Ext(files[0].Filename)
		fileName = uuid.New().String() + fileExt
		filePath = filepath.Join(s.uploadDir, fileName)

		// Сохраняем файл
		if err := os.WriteFile(filePath, files[0].Data, 0644); err != nil {
			return nil, fmt.Errorf("failed to save file: %v", err)
		}
	} else {
		// Для JSON-документов
		if metaData.Mime == "" {
			metaData.Mime = "application/json"
		}
	}

	// 4. Генерация id
	docID, err := generateID()
	if err != nil {
		return nil, err
	}

	// 5. Создание документа
	doc := &model.Document{
		ID:       docID,
		Name:     metaData.Name,
		Mime:     metaData.Mime,
		File:     len(files) > 0,
		Public:   metaData.Public,
		Created:  time.Now(),
		Owner:    user.ID,
		FilePath: filePath,
		JSONData: metaData.JSON,
		Grant:    metaData.Grant,
	}

	// 5. Сохранение в БД
	if err := s.docRepo.CreateDocument(context.Background(), doc); err != nil {
		// Удаляем сохраненный файл в случае ошибки
		if filePath != "" {
			os.Remove(filePath)
		}
		return nil, fmt.Errorf("failed to create document: %v", err)
	}

	// 6. Инвалидация кеша
	s.cache.Delete("docs_" + user.ID)

	return doc, nil
}

func (s *DocumentService) GetDocumentsList(token, login, key, value string, limit int) ([]*model.Document, error) {
	// 1. Валидация токена
	user, err := s.userRepo.GetUserByToken(context.Background(), token)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, errors.New("unauthorized")
	}

	// 2. Проверка кеша
	cacheKey := "docs_" + user.ID
	if cached, found := s.cache.Get(cacheKey); found {
		return cached.([]*model.Document), nil
	}

	// 3. Получение документов из БД
	var docs []*model.Document
	if login == "" || login == user.Login {
		// Собственные документы пользователя
		docs, err = s.docRepo.GetUserDocuments(context.Background(), user.ID, limit)
		if err != nil {
			return nil, err
		}

		// Сохраняем в кеш только свои документы
		s.cache.Set(cacheKey, docs, 5*time.Minute)
	} else {
		// Документы другого пользователя
		otherUser, err := s.userRepo.GetUserByLogin(context.Background(), login)
		if err != nil {
			return nil, fmt.Errorf("user not found")
		}

		// Получаем только публичные документы или документы с доступом для текущего пользователя
		docs, err = s.docRepo.GetSharedDocuments(context.Background(), user.ID, otherUser.ID, limit)
		if err != nil {
			return nil, err
		}
	}

	if err != nil {
		return nil, err
	}

	// 4. Сохранение в кеш
	s.cache.Set(cacheKey, docs, 5*time.Minute)

	return docs, nil
}

func (s *DocumentService) GetDocument(token, id string) (*model.Document, error) {
	// 1. Валидация токена
	user, err := s.userRepo.GetUserByToken(context.Background(), token)
	if err != nil {
		return nil, err
	}

	if user == nil {
		return nil, errors.New("unauthorized")
	}

	// 2. Проверка кеша
	cacheKey := "doc_" + id
	if cached, found := s.cache.Get(cacheKey); found {
		return cached.(*model.Document), nil
	}

	// 3. Получение документа из БД
	doc, err := s.docRepo.GetDocumentByID(context.Background(), id)
	if err != nil {
		return nil, err
	}

	// 4. Проверка прав доступа
	hasAccess := false
	if doc.Owner == user.ID {
		hasAccess = true
	} else if doc.Public {
		hasAccess = true
	} else {
		for _, grant := range doc.Grant {
			if grant == user.Login {
				hasAccess = true
				break
			}
		}
	}

	if !hasAccess {
		return nil, errors.New("forbidden")
	}

	// 5. Сохранение в кеш
	s.cache.Set(cacheKey, doc, 10*time.Minute)

	return doc, nil
}

func (s *DocumentService) DeleteDocument(token, id string) (bool, error) {
	// 1. Валидация токена
	user, err := s.userRepo.GetUserByToken(context.Background(), token)
	if err != nil {
		return false, err
	}
	if user == nil {
		return false, errors.New("unauthorized")
	}

	// 2. Получение документа для проверки владельца
	doc, err := s.docRepo.GetDocumentByID(context.Background(), id)
	if err != nil {
		return false, err
	}
	if doc.Owner != user.ID {
		return false, errors.New("forbidden")
	}

	// 3. Удаление файла (если есть)
	if doc.File && doc.FilePath != "" {
		if err := os.Remove(doc.FilePath); err != nil {
			return false, err
		}
	}

	// 4. Удаление из БД
	if err := s.docRepo.DeleteDocument(context.Background(), id); err != nil {
		return false, err
	}

	// 5. Инвалидация кеша
	s.cache.Delete("doc_" + id)
	s.cache.Delete("docs_" + user.ID)

	return true, nil
}

// generateID создает новый UUID версии 7
func generateID() (string, error) {
	id, err := uuid.NewV7()
	if err != nil {
		return "", fmt.Errorf("failed to generate UUID: %w", err)
	}
	return id.String(), nil
}
