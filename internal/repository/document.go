package repository

import (
	"context"
	"database/sql"
	"docs-server/internal/model"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

type DocumentRepository struct {
	db *sql.DB
}

func NewDocumentRepository(dsn string) *DocumentRepository {
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		panic(err)
	}

	// Проверяем соединение
	if err := db.Ping(); err != nil {
		panic(err)
	}

	// Устанавливаем настройки пула соединений
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(25)
	db.SetConnMaxLifetime(5 * time.Minute)

	return &DocumentRepository{db: db}
}

func (r *DocumentRepository) CreateDocument(ctx context.Context, doc *model.Document) error {
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO documents 
		(id, name, mime, is_file, is_public, created_at, owner_id, file_path, json_data) 
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		doc.ID, doc.Name, doc.Mime, doc.File, doc.Public,
		doc.Created, doc.Owner, doc.FilePath, doc.JSONData)
	return err
}

func (r *DocumentRepository) GetDocumentByID(ctx context.Context, id string) (*model.Document, error) {
	doc := &model.Document{}
	err := r.db.QueryRowContext(ctx, `
		SELECT id, name, mime, is_file, is_public, created_at, file_path, json_data, owner_id
		FROM documents WHERE id = ?`, id).
		Scan(&doc.ID, &doc.Name, &doc.Mime, &doc.File, &doc.Public,
			&doc.Created, &doc.FilePath, &doc.JSONData, &doc.Owner)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // Документ не найден - не ошибка
		}
		return nil, err
	}
	return doc, nil
}

func (r *DocumentRepository) GetUserDocuments(ctx context.Context, userID string, limit int) ([]*model.Document, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, name, mime, is_file, is_public, created_at
		FROM documents 
		WHERE owner_id = ?
		ORDER BY name, created_at
		LIMIT ?`, userID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var docs []*model.Document
	for rows.Next() {
		doc := &model.Document{}
		if err := rows.Scan(&doc.ID, &doc.Name, &doc.Mime, &doc.File, &doc.Public, &doc.Created); err != nil {
			return nil, err
		}
		docs = append(docs, doc)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return docs, nil
}

func (r *DocumentRepository) DeleteDocument(ctx context.Context, id string) error {
	_, err := r.db.ExecContext(ctx, "DELETE FROM documents WHERE id = ?", id)
	return err
}

func (r *DocumentRepository) Close() error {
	return r.db.Close()
}
