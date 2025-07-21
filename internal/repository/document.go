package repository

import (
	"context"
	"database/sql"
	"docs-server/internal/model"
)

type DocumentRepository struct {
	db *sql.DB
}

func NewDocumentRepository(dsn string) *DocumentRepository {
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		panic(err)
	}
	return &DocumentRepository{db: db}
}

func (r *DocumentRepository) CreateDocument(ctx context.Context, doc *model.Document) error {
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO documents 
		(id, name, mime, file, public, created, owner_id, file_path, json_data) 
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`,
		doc.ID, doc.Name, doc.Mime, doc.File, doc.Public,
		doc.Created, doc.Owner, doc.FilePath, doc.JSONData)
	return err
}

func (r *DocumentRepository) GetDocumentByID(ctx context.Context, id string) (*model.Document, error) {
	doc := &model.Document{}
	err := r.db.QueryRowContext(ctx, `
		SELECT id, name, mime, file, public, created, file_path, json_data, owner_id
		FROM documents WHERE id = $1`, id).
		Scan(&doc.ID, &doc.Name, &doc.Mime, &doc.File, &doc.Public,
			&doc.Created, &doc.FilePath, &doc.JSONData, &doc.Owner)
	if err != nil {
		return nil, err
	}
	return doc, nil
}

func (r *DocumentRepository) GetUserDocuments(ctx context.Context, userID string, limit int) ([]*model.Document, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, name, mime, file, public, created
		FROM documents 
		WHERE owner_id = $1
		ORDER BY name, created
		LIMIT $2`, userID, limit)
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
	return docs, nil
}

func (r *DocumentRepository) DeleteDocument(ctx context.Context, id string) error {
	_, err := r.db.ExecContext(ctx, "DELETE FROM documents WHERE id = $1", id)
	return err
}

func (r *DocumentRepository) Close() error {
	return r.db.Close()
}
