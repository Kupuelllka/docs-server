package service

import (
	"context"
	"docs-server/internal/model"
	"docs-server/internal/repository"
	"errors"
)

var (
	ErrUserIDEmpty   = errors.New("user ID cannot be empty")
	ErrLoginEmpty    = errors.New("login cannot be empty")
	ErrUserNil       = errors.New("user cannot be nil")
	ErrInvalidLimit  = errors.New("limit must be positive")
	ErrInvalidOffset = errors.New("offset cannot be negative")
)

type UserService struct {
	userRepo *repository.UserRepository
}

func NewUserService(userRepo *repository.UserRepository) *UserService {
	return &UserService{
		userRepo: userRepo,
	}
}

// GetUserByID возвращает пользователя по ID
func (s *UserService) GetUserByID(ctx context.Context, id string) (*model.User, error) {
	if id == "" {
		return nil, ErrUserIDEmpty
	}

	user, err := s.userRepo.GetUserByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, ErrUserNotFound
	}

	return user, nil
}

// GetUserByLogin возвращает пользователя по логину
func (s *UserService) GetUserByLogin(ctx context.Context, login string) (*model.User, error) {
	if login == "" {
		return nil, ErrLoginEmpty
	}

	user, err := s.userRepo.GetUserByLogin(ctx, login)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, ErrUserNotFound
	}

	return user, nil
}

// UpdateUser обновляет данные пользователя
func (s *UserService) UpdateUser(ctx context.Context, user *model.User) error {
	if user == nil {
		return ErrUserNil
	}
	if user.ID == "" {
		return ErrUserIDEmpty
	}

	// Проверяем, существует ли пользователь
	existingUser, err := s.userRepo.GetUserByID(ctx, user.ID)
	if err != nil {
		return err
	}
	if existingUser == nil {
		return ErrUserNotFound
	}

	// Обновляем только разрешенные поля
	existingUser.Login = user.Login

	return s.userRepo.UpdateUser(ctx, existingUser)
}

// DeleteUser удаляет пользователя по ID
func (s *UserService) DeleteUser(ctx context.Context, id string) error {
	if id == "" {
		return ErrUserIDEmpty
	}

	// Проверяем, существует ли пользователь
	user, err := s.userRepo.GetUserByID(ctx, id)
	if err != nil {
		return err
	}
	if user == nil {
		return ErrUserNotFound
	}

	return s.userRepo.DeleteUser(ctx, id)
}

// ListUsers возвращает список пользователей с пагинацией
func (s *UserService) ListUsers(ctx context.Context, limit, offset int) ([]*model.User, error) {
	if limit <= 0 {
		return nil, ErrInvalidLimit
	}
	if offset < 0 {
		return nil, ErrInvalidOffset
	}

	return s.userRepo.ListUsers(ctx, limit, offset)
}
