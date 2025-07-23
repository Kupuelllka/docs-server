package service

import (
	"context"
	"errors"
	"regexp"
	"time"
	"unicode"

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
	uuidUser, err := generateID()
	if err != nil {
		return err
	}
	if adminToken != s.adminToken {
		return ErrInvalidAdminToken
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

	return s.userRepo.CreateUser(ctx, uuidUser, login, string(hashedPassword))
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
	// Проверка минимальной длины
	if len(login) < 8 {
		return errors.New("login must be at least 8 characters long")
	}

	// Проверка на латиницу и цифры
	matched, err := regexp.MatchString(`^[a-zA-Z0-9]+$`, login)
	if err != nil {
		return err
	}
	if !matched {
		return errors.New("login must contain only latin letters and digits")
	}

	return nil
}

func validatePassword(password string) error {
	// Проверка минимальной длины
	if len(password) < 8 {
		return errors.New("password must be at least 8 characters long")
	}

	var (
		hasUpper   bool
		hasLower   bool
		hasDigit   bool
		hasSpecial bool
	)

	for _, char := range password {
		switch {
		case unicode.IsUpper(char):
			hasUpper = true
		case unicode.IsLower(char):
			hasLower = true
		case unicode.IsDigit(char):
			hasDigit = true
		case !unicode.IsLetter(char) && !unicode.IsDigit(char):
			hasSpecial = true
		}
	}

	// Проверка требований
	if !hasUpper || !hasLower {
		return errors.New("password must contain at least 2 letters in different cases (upper and lower)")
	}

	if !hasDigit {
		return errors.New("password must contain at least 1 digit")
	}

	if !hasSpecial {
		return errors.New("password must contain at least 1 special character")
	}

	return nil
}

func generateToken() string {
	// Добавить jwt
	return "generated-token-" + time.Now().Format("20060102150405")
}
