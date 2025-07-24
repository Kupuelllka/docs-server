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

var (
	ErrInvalidToken         = errors.New("unauthorized: invalid token")
	ErrDocumentNameRequired = errors.New("document name is required")
	ErrUserNotFound         = errors.New("user not found")
	ErrForbidden            = errors.New("forbidden")
	ErrNotDocumentOwner     = errors.New("forbidden: not document owner")
	ErrFailedToCreateDir    = errors.New("failed to create upload directory")
	ErrFailedToSaveFile     = errors.New("failed to save file")
	ErrFailedToDeleteFile   = errors.New("failed to delete file")
	ErrInvalidMetaFormat    = errors.New("invalid meta format")
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

// getUserFromToken - внутренний метод для получения пользователя по токену
func (s *DocumentService) getUserFromToken(token string) (*model.User, error) {
	user, err := s.userRepo.GetUserByToken(context.Background(), token)
	if err != nil {
		return nil, fmt.Errorf("failed to get user from token: %w", err)
	}
	if user == nil {
		return nil, ErrInvalidToken
	}
	return user, nil
}

func (s *DocumentService) UploadDocument(token string, meta string, files []*model.UploadedFile) (*model.Document, error) {
	// Получаем пользователя из токена
	user, err := s.getUserFromToken(token)
	if err != nil {
		return nil, err
	}

	// Парсинг метаданных
	var metaData struct {
		Name   string      `json:"name"`
		Public bool        `json:"public"`
		Mime   string      `json:"mime"`
		Grant  []string    `json:"grant"`
		JSON   interface{} `json:"json"`
	}

	if err := json.Unmarshal([]byte(meta), &metaData); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInvalidMetaFormat, err)
	}

	// Валидация
	if metaData.Name == "" {
		return nil, ErrDocumentNameRequired
	}

	// Обработка файла
	var filePath string
	if len(files) > 0 {
		// Определение MIME-типа
		if metaData.Mime == "" {
			metaData.Mime = mime.TypeByExtension(filepath.Ext(files[0].Filename))
			if metaData.Mime == "" {
				metaData.Mime = "application/octet-stream"
			}
		}

		// Создание директории
		if err := os.MkdirAll(s.uploadDir, 0755); err != nil {
			return nil, fmt.Errorf("%w: %v", ErrFailedToCreateDir, err)
		}

		// Генерация имени файла
		fileExt := filepath.Ext(files[0].Filename)
		fileName := uuid.New().String() + fileExt
		filePath = filepath.Join(s.uploadDir, fileName)

		// Сохранение файла
		if err := os.WriteFile(filePath, files[0].Data, 0644); err != nil {
			return nil, fmt.Errorf("%w: %v", ErrFailedToSaveFile, err)
		}
	} else if metaData.Mime == "" {
		metaData.Mime = "application/json"
	}

	// Создание документа
	docID, err := generateID()
	if err != nil {
		return nil, err
	}

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

	// Сохранение в БД
	if err := s.docRepo.CreateDocument(context.Background(), doc); err != nil {
		if filePath != "" {
			os.Remove(filePath)
		}
		return nil, fmt.Errorf("failed to create document: %v", err)
	}

	// Инвалидация кеша
	s.cache.Delete("docs_" + user.ID)

	return doc, nil
}

func (s *DocumentService) GetDocumentsList(token, login, key, value string, limit int) ([]*model.Document, error) {
	user, err := s.getUserFromToken(token)
	if err != nil {
		return nil, err
	}

	// Проверка кеша
	cacheKey := "docs_" + user.ID
	if cached, found := s.cache.Get(cacheKey); found {
		return cached.([]*model.Document), nil
	}

	var docs []*model.Document
	if login == "" || login == user.Login {
		// Собственные документы
		docs, err = s.docRepo.GetUserDocuments(context.Background(), user.ID, limit)
		if err != nil {
			return nil, err
		}
		s.cache.Set(cacheKey, docs, 5*time.Minute)
	} else {
		// Документы другого пользователя
		otherUser, err := s.userRepo.GetUserByLogin(context.Background(), login)
		if err != nil {
			return nil, ErrUserNotFound
		}

		docs, err = s.docRepo.GetSharedDocuments(context.Background(), user.ID, otherUser.ID, limit)
		if err != nil {
			return nil, err
		}
	}

	return docs, nil
}

func (s *DocumentService) GetDocument(token, id string) (*model.Document, error) {
	user, err := s.getUserFromToken(token)
	if err != nil {
		return nil, err
	}

	// Проверка кеша
	cacheKey := "doc_" + id
	if cached, found := s.cache.Get(cacheKey); found {
		return cached.(*model.Document), nil
	}

	// Получение документа
	doc, err := s.docRepo.GetDocumentByID(context.Background(), id)
	if err != nil {
		return nil, err
	}
	// Проверка прав доступа
	hasAccess := doc.Owner == user.ID || doc.Public
	if !hasAccess {
		for _, grant := range doc.Grant {
			if grant == user.ID {
				hasAccess = true
				break
			}
		}
	}

	if !hasAccess {
		return nil, ErrForbidden
	}

	// Сохранение в кеш
	s.cache.Set(cacheKey, doc, 10*time.Minute)

	return doc, nil
}

func (s *DocumentService) DeleteDocument(token, id string) (bool, error) {
	user, err := s.getUserFromToken(token)
	if err != nil {
		return false, err
	}

	// Проверка владельца документа
	doc, err := s.docRepo.GetDocumentByID(context.Background(), id)
	if err != nil {
		return false, err
	}
	if doc.Owner != user.ID {
		return false, ErrNotDocumentOwner
	}

	// Удаление файла
	if doc.File && doc.FilePath != "" {
		if err := os.Remove(doc.FilePath); err != nil {
			return false, fmt.Errorf("%w: %v", ErrFailedToDeleteFile, err)
		}
	}

	// Удаление из БД
	if err := s.docRepo.DeleteDocument(context.Background(), id); err != nil {
		return false, fmt.Errorf("failed to delete document: %w", err)
	}

	// Инвалидация кеша
	s.cache.Delete("doc_" + id)
	s.cache.Delete("docs_" + user.ID)

	return true, nil
}

func generateID() (string, error) {
	id, err := uuid.NewRandom()
	if err != nil {
		return "", fmt.Errorf("failed to generate UUID: %w", err)
	}
	return id.String(), nil
}
