package repository

import (
	"context"
	"database/sql"
	"docs-server/internal/model"
	"errors"
	"fmt"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

type UserRepository struct {
	db *sql.DB
}

func NewUserRepository(dsn string) *UserRepository {
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		panic(err)
	}

	if err := db.Ping(); err != nil {
		panic(err)
	}

	// Устанавливаем настройки пула соединений
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(25)
	db.SetConnMaxLifetime(5 * time.Minute)
	return &UserRepository{db: db}
}

func (r *UserRepository) GetUserByID(ctx context.Context, id string) (*model.User, error) {
	user := &model.User{}
	var tokenExpiry []byte

	err := r.db.QueryRowContext(ctx,
		"SELECT UUID_TO_STRING(id), login, password, token, token_expiry FROM users WHERE id = UUID_TO_BIN(?)", id).
		Scan(&user.ID, &user.Login, &user.Password, &user.Token, &tokenExpiry)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	if len(tokenExpiry) > 0 {
		expiryStr := string(tokenExpiry)
		expiry, err := time.Parse("2006-01-02 15:04:05", expiryStr)
		if err != nil {
			return nil, fmt.Errorf("failed to parse token expiry: %v", err)
		}
		user.TokenExpiry = expiry
	}

	return user, nil
}

func (r *UserRepository) GetUserByLogin(ctx context.Context, login string) (*model.User, error) {
	user := &model.User{}
	var (
		token       sql.NullString
		tokenExpiry []byte
		password    sql.NullString
	)

	err := r.db.QueryRowContext(ctx,
		"SELECT UUID_TO_STRING(id), login, password, token, token_expiry FROM users WHERE login = ?", login).
		Scan(&user.ID, &user.Login, &password, &token, &tokenExpiry)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get user by login: %w", err)
	}

	if password.Valid {
		user.Password = password.String
	}
	if token.Valid {
		user.Token = token.String
	}

	if len(tokenExpiry) > 0 {
		expiryStr := string(tokenExpiry)
		expiry, err := time.Parse("2006-01-02 15:04:05", expiryStr)
		if err != nil {
			return nil, fmt.Errorf("failed to parse token expiry: %v", err)
		}
		user.TokenExpiry = expiry
	}

	return user, nil
}
func (r *UserRepository) UpdateUser(ctx context.Context, user *model.User) error {
	_, err := r.db.ExecContext(ctx,
		"UPDATE users SET login = ? WHERE id = UUID_TO_BIN(?)",
		user.Login, user.ID)
	return err
}

func (r *UserRepository) DeleteUser(ctx context.Context, id string) error {
	_, err := r.db.ExecContext(ctx, "DELETE FROM users WHERE id = UUID_TO_BIN(?)", id)
	return err
}

func (r *UserRepository) ListUsers(ctx context.Context, limit, offset int) ([]*model.User, error) {
	rows, err := r.db.QueryContext(ctx,
		"SELECT UUID_TO_STRING(id), login FROM users LIMIT ? OFFSET ?", limit, offset)
	if err != nil {
		return nil, err
	}

	var users []*model.User
	for rows.Next() {
		user := &model.User{}
		if err := rows.Scan(&user.ID, &user.Login); err != nil {
			rows.Close()
			return nil, err
		}
		users = append(users, user)
	}
	rows.Close()
	return users, nil
}

func (r *UserRepository) CreateUser(ctx context.Context, id string, login, hashedPassword string) error {
	_, err := r.db.ExecContext(ctx,
		"INSERT INTO users (id, login, password) VALUES (UUID_TO_BIN(?), ?, ?)",
		id, login, hashedPassword)
	return err
}

func (r *UserRepository) GetUserByToken(ctx context.Context, token string) (*model.User, error) {
	user := &model.User{}
	var tokenExpiry []byte

	err := r.db.QueryRowContext(ctx,
		"SELECT UUID_TO_STRING(id), login, password, token_expiry FROM users WHERE token = ?", token).
		Scan(&user.ID, &user.Login, &user.Password, &tokenExpiry)
	fmt.Println(user.ID)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get user by token: %w", err)
	}

	if len(tokenExpiry) > 0 {
		expiryStr := string(tokenExpiry)
		expiry, err := time.Parse("2006-01-02 15:04:05", expiryStr)
		if err != nil {
			return nil, fmt.Errorf("failed to parse token expiry: %v", err)
		}
		user.TokenExpiry = expiry
	}

	return user, nil
}

func (r *UserRepository) UpdateUserToken(ctx context.Context, userID string, token string, expiry time.Time) error {
	_, err := r.db.ExecContext(ctx,
		"UPDATE users SET token = ?, token_expiry = ? WHERE id = UUID_TO_BIN(?)",
		token, expiry.Format("2006-01-02 15:04:05"), userID)
	return err
}

func (r *UserRepository) Close() error {
	return r.db.Close()
}
