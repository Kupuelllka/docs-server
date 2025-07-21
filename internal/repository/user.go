package repository

import (
	"context"
	"database/sql"
	"docs-server/internal/model"
	"errors"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/google/uuid"
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
	err := r.db.QueryRowContext(ctx,
		"SELECT id, login, password, token, token_expiry FROM users WHERE id = ?", id).
		Scan(&user.ID, &user.Login, &user.Password, &user.Token, &user.TokenExpiry)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (r *UserRepository) GetUserByLogin(ctx context.Context, login string) (*model.User, error) {
	user := &model.User{}
	err := r.db.QueryRowContext(ctx,
		"SELECT id, login, password, token, token_expiry FROM users WHERE login = ?", login).
		Scan(&user.ID, &user.Login, &user.Password, &user.Token, &user.TokenExpiry)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (r *UserRepository) UpdateUser(ctx context.Context, user *model.User) error {
	_, err := r.db.ExecContext(ctx,
		"UPDATE users SET login = ? WHERE id = ?",
		user.Login, user.ID)
	return err
}

func (r *UserRepository) DeleteUser(ctx context.Context, id string) error {
	_, err := r.db.ExecContext(ctx, "DELETE FROM users WHERE id = ?", id)
	return err
}

func (r *UserRepository) ListUsers(ctx context.Context, limit, offset int) ([]*model.User, error) {
	rows, err := r.db.QueryContext(ctx,
		"SELECT id, login FROM users LIMIT ? OFFSET ?", limit, offset)
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

func (r *UserRepository) CreateUser(ctx context.Context, uuid uuid.UUID, login, hashedPassword string) error {
	// Конвертируем UUID в бинарный формат (16 байт)
	uuidBinary, err := uuid.MarshalBinary()
	if err != nil {
		return err
	}
	_, err = r.db.ExecContext(ctx,
		"INSERT INTO users (id, login, password) VALUES (?, ?, ?)",
		uuidBinary, login, hashedPassword)
	return err
}

func (r *UserRepository) GetUserByToken(ctx context.Context, token string) (*model.User, error) {
	user := &model.User{}
	err := r.db.QueryRowContext(ctx,
		"SELECT id, login, password, token_expiry FROM users WHERE token = ?", token).
		Scan(&user.ID, &user.Login, &user.Password, &user.TokenExpiry)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (r *UserRepository) UpdateUserToken(ctx context.Context, userID uuid.UUID, token string, expiry time.Time) error {
	_, err := r.db.ExecContext(ctx,
		"UPDATE users SET token = ?, token_expiry = ? WHERE id = ?",
		token, expiry, userID)
	return err
}

func (r *UserRepository) Close() error {
	return r.db.Close()
}
