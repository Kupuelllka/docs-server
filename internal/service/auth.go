package service

import (
	"context"
	"errors"
	"time"

	"golang.org/x/crypto/bcrypt"

	"docs-server/internal/model"
	"docs-server/internal/repository"
)

type AuthService struct {
	userRepo    *repository.UserRepository
	adminToken  string
	tokenExpiry time.Duration
}

func NewAuthService(userRepo *repository.UserRepository, adminToken string) *AuthService {
	return &AuthService{
		userRepo:    userRepo,
		adminToken:  adminToken,
		tokenExpiry: 24 * time.Hour,
	}
}

func (s *AuthService) Register(adminToken, login, password string) error {
	if adminToken != s.adminToken {
		return errors.New("invalid admin token")
	}

	if err := validateLogin(login); err != nil {
		return err
	}

	if err := validatePassword(password); err != nil {
		return err
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	return s.userRepo.CreateUser(ctx, login, string(hashedPassword))
}

func (s *AuthService) Authenticate(login, password string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	user, err := s.userRepo.GetUserByLogin(ctx, login)
	if err != nil {
		return "", err
	}
	if user == nil {
		return "", errors.New("user not found")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		return "", errors.New("invalid credentials")
	}

	// Генерация токена
	token := generateToken()
	user.Token = token

	// Сохраняем токен в БД
	if err := s.userRepo.UpdateUserToken(ctx, user.ID, token, time.Now().Add(s.tokenExpiry)); err != nil {
		return "", err
	}

	return token, nil
}

func (s *AuthService) ValidateToken(token string) (*model.User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	user, err := s.userRepo.GetUserByToken(ctx, token)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, errors.New("invalid token")
	}

	if time.Now().After(user.TokenExpiry) {
		return nil, errors.New("token expired")
	}

	return user, nil
}

func (s *AuthService) Logout(token string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	user, err := s.userRepo.GetUserByToken(ctx, token)
	if err != nil {
		return err
	}
	if user == nil {
		return errors.New("invalid token")
	}

	return s.userRepo.UpdateUserToken(ctx, user.ID, "", time.Time{})
}

func validateLogin(login string) error {
	// Добавить валидацию логина
	return nil
}

func validatePassword(password string) error {
	// Добавить валидацию пароля
	return nil
}

func generateToken() string {
	// Добавить jwt
	return "generated-token-" + time.Now().Format("20060102150405")
}
