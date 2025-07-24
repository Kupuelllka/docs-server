package repository

import (
	"context"
	"database/sql"
	"docs-server/internal/model"
	"encoding/json"
	"fmt"
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
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %v", err)
	}
	defer tx.Rollback()

	// Если есть jsondata записываем
	var jsonData []byte
	if doc.JSONData != nil {
		jsonData, err = json.Marshal(doc.JSONData)
		if err != nil {
			return fmt.Errorf("failed to marshal JSON data: %v", err)
		}
	}

	// Добавляем документ
	_, err = tx.ExecContext(ctx, `
        INSERT INTO documents 
        (id, name, mime, is_file, is_public, created_at, owner_id, file_path, json_data) 
        VALUES (UUID_TO_BIN(?), ?, ?, ?, ?, ?, UUID_TO_BIN(?), ?, ?)`,
		doc.ID, doc.Name, doc.Mime, doc.File, doc.Public,
		doc.Created, doc.Owner, doc.FilePath, jsonData)
	if err != nil {
		return fmt.Errorf("failed to insert document: %v", err)
	}

	// Добавляем разрешения, если они есть
	if len(doc.Grant) > 0 {
		// Получаем список ID пользователей
		stmtSelect, err := tx.PrepareContext(ctx, `
            SELECT UUID_TO_STRING(id) FROM users WHERE login = ?`)
		if err != nil {
			return fmt.Errorf("failed to prepare select user statement: %v", err)
		}
		defer stmtSelect.Close()

		stmtInsert, err := tx.PrepareContext(ctx, `
            INSERT INTO document_grants 
            (document_id, user_id) 
            VALUES (UUID_TO_BIN(?), UUID_TO_BIN(?))`)
		if err != nil {
			return fmt.Errorf("failed to prepare grants statement: %v", err)
		}
		defer stmtInsert.Close()

		for _, username := range doc.Grant {
			var userID string
			err = stmtSelect.QueryRowContext(ctx, username).Scan(&userID)
			if err != nil {
				return fmt.Errorf("failed to find user with username %s: %v", username, err)
			}

			_, err = stmtInsert.ExecContext(ctx, doc.ID, userID)
			if err != nil {
				return fmt.Errorf("failed to insert grant for user %s: %v", username, err)
			}
		}
	}
	// Пишем транзакцию
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %v", err)
	}

	return nil
}
func (r *DocumentRepository) GetDocumentByID(ctx context.Context, id string) (*model.Document, error) {
	doc := &model.Document{}
	var createdAtBytes []byte

	// Получаем основные данные документа
	err := r.db.QueryRowContext(ctx, `
        SELECT 
            UUID_TO_STRING(id), name, mime, is_file, is_public, 
            created_at, file_path, json_data, UUID_TO_STRING(owner_id)
        FROM documents WHERE id = UUID_TO_BIN(?)`, id).
		Scan(&doc.ID, &doc.Name, &doc.Mime, &doc.File, &doc.Public,
			&createdAtBytes, &doc.FilePath, &doc.JSONData, &doc.Owner)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	// Парсим дату создания
	createdAtStr := string(createdAtBytes)
	createdAt, err := time.Parse("2006-01-02 15:04:05", createdAtStr)
	if err != nil {
		return nil, fmt.Errorf("failed to parse created_at: %v", err)
	}
	doc.Created = createdAt

	// Получаем список прав доступа из document_grants
	rows, err := r.db.QueryContext(ctx, `
        SELECT UUID_TO_STRING(user_id) 
        FROM document_grants 
        WHERE document_id = UUID_TO_BIN(?)`, id)
	if err != nil {
		return nil, fmt.Errorf("failed to query grants: %v", err)
	}
	defer rows.Close()

	var grants []string
	for rows.Next() {
		var userID string
		if err := rows.Scan(&userID); err != nil {
			return nil, fmt.Errorf("failed to scan grant user_id: %v", err)
		}
		grants = append(grants, userID)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating grants: %v", err)
	}

	doc.Grant = grants

	return doc, nil
}

func (r *DocumentRepository) GetUserDocuments(ctx context.Context, userid string, limit int) ([]*model.Document, error) {
	rows, err := r.db.QueryContext(ctx, `
        SELECT 
            UUID_TO_STRING(id), name, mime, is_file, is_public, created_at
        FROM documents 
        WHERE owner_id = UUID_TO_BIN(?)
        ORDER BY name, created_at
        LIMIT ?`, userid, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var docs []*model.Document
	for rows.Next() {
		doc := &model.Document{}
		var createdAtBytes []byte

		if err := rows.Scan(
			&doc.ID,
			&doc.Name,
			&doc.Mime,
			&doc.File,
			&doc.Public,
			&createdAtBytes,
		); err != nil {
			return nil, err
		}

		createdAtStr := string(createdAtBytes)
		createdAt, err := time.Parse("2006-01-02 15:04:05", createdAtStr)
		if err != nil {
			return nil, fmt.Errorf("failed to parse created_at: %v", err)
		}
		doc.Created = createdAt

		docs = append(docs, doc)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return docs, nil
}
func (r *DocumentRepository) GetSharedDocuments(ctx context.Context, currentUserID, ownerID string, limit int) ([]*model.Document, error) {
	rows, err := r.db.QueryContext(ctx, `
        SELECT 
            UUID_TO_STRING(d.id), 
            d.name, 
            d.mime, 
            d.is_file, 
            d.is_public, 
            d.created_at,
            d.file_path,
            d.json_data,
            UUID_TO_STRING(d.owner_id)
        FROM documents d
        LEFT JOIN document_grants g ON d.id = g.document_id
        WHERE d.owner_id = UUID_TO_BIN(?)
        AND (d.is_public = TRUE OR g.user_id = UUID_TO_BIN(?))
        ORDER BY d.name, d.created_at
        LIMIT ?`, ownerID, currentUserID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var docs []*model.Document
	for rows.Next() {
		doc := &model.Document{}
		var createdAtBytes []byte

		err := rows.Scan(
			&doc.ID,
			&doc.Name,
			&doc.Mime,
			&doc.File,
			&doc.Public,
			&createdAtBytes,
			&doc.FilePath,
			&doc.JSONData,
			&doc.Owner,
		)
		if err != nil {
			return nil, err
		}

		createdAtStr := string(createdAtBytes)
		createdAt, err := time.Parse("2006-01-02 15:04:05", createdAtStr)
		if err != nil {
			return nil, fmt.Errorf("failed to parse created_at: %v", err)
		}
		doc.Created = createdAt

		docs = append(docs, doc)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return docs, nil
}
func (r *DocumentRepository) DeleteDocument(ctx context.Context, id string) error {
	_, err := r.db.ExecContext(ctx, "DELETE FROM documents WHERE id = UUID_TO_BIN(?)", id)
	return err
}

func (r *DocumentRepository) Close() error {
	return r.db.Close()
}
