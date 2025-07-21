package service

import (
	"context"
	"docs-server/internal/model"
	"docs-server/internal/repository"
	"errors"
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
		return nil, errors.New("user ID cannot be empty")
	}

	user, err := s.userRepo.GetUserByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, errors.New("user not found")
	}

	return user, nil
}

// GetUserByLogin возвращает пользователя по логину
func (s *UserService) GetUserByLogin(ctx context.Context, login string) (*model.User, error) {
	if login == "" {
		return nil, errors.New("login cannot be empty")
	}

	user, err := s.userRepo.GetUserByLogin(ctx, login)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, errors.New("user not found")
	}

	return user, nil
}

// UpdateUser обновляет данные пользователя
func (s *UserService) UpdateUser(ctx context.Context, user *model.User) error {
	if user == nil {
		return errors.New("user cannot be nil")
	}
	if user.ID == "" {
		return errors.New("user ID cannot be empty")
	}

	// Проверяем, существует ли пользователь
	existingUser, err := s.userRepo.GetUserByID(ctx, user.ID)
	if err != nil {
		return err
	}
	if existingUser == nil {
		return errors.New("user not found")
	}

	// Обновляем только разрешенные поля
	existingUser.Login = user.Login

	return s.userRepo.UpdateUser(ctx, existingUser)
}

// DeleteUser удаляет пользователя по ID
func (s *UserService) DeleteUser(ctx context.Context, id string) error {
	if id == "" {
		return errors.New("user ID cannot be empty")
	}

	// Проверяем, существует ли пользователь
	user, err := s.userRepo.GetUserByID(ctx, id)
	if err != nil {
		return err
	}
	if user == nil {
		return errors.New("user not found")
	}

	return s.userRepo.DeleteUser(ctx, id)
}

// ListUsers возвращает список пользователей с пагинацией
func (s *UserService) ListUsers(ctx context.Context, limit, offset int) ([]*model.User, error) {
	if limit <= 0 {
		limit = 10
	}
	if offset < 0 {
		offset = 0
	}

	return s.userRepo.ListUsers(ctx, limit, offset)
}
