package service

import (
	"context"
	"docs-server/internal/cache"
	"docs-server/internal/model"
	"docs-server/internal/repository"
	"errors"
	"os"
	"path/filepath"
	"time"
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
	// (Добавить парсинга JSON из строки meta)

	// 3. Сохранение файла (если есть)
	var filePath string
	if len(files) > 0 {
		uploadDir := "uploads"
		if err := os.MkdirAll(uploadDir, 0755); err != nil {
			return nil, err
		}

		file := files[0]
		filePath = filepath.Join(uploadDir, file.Filename)
		if err := os.WriteFile(filePath, file.Data, 0644); err != nil {
			return nil, err
		}
	}

	// 4. Создание документа
	doc := &model.Document{
		ID:       generateID(),
		Name:     "example",                  // Добавить из meta
		Mime:     "application/octet-stream", // Добавить из meta или файла
		File:     len(files) > 0,
		Public:   false, // Добавить из meta
		Created:  time.Now(),
		Owner:    user.ID,
		FilePath: filePath,
		JSONData: nil, // Добавить из meta или отдельного поля
	}

	// 5. Сохранение в БД
	if err := s.docRepo.CreateDocument(context.Background(), doc); err != nil {
		return nil, err
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
		// Документы пользователя
		docs, err = s.docRepo.GetUserDocuments(context.Background(), user.ID, limit)
	} else {
		// Добавить обработку документов другого пользователя (с проверкой прав доступа)
		// (дополнительная реализация)
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
	if doc.Owner != user.ID && !doc.Public {
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

func generateID() string {
	// Добавить реализацию генерации UUID
	return "generated-id"
}
