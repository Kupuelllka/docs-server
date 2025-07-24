package service

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"regexp"
	"time"
	"unicode"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"

	"docs-server/internal/cache"
	"docs-server/internal/model"
	"docs-server/internal/repository"
)

var (
	ErrInvalidAdminToken  = errors.New("invalid admin token")
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrTokenExpired       = errors.New("token expired")
)

// JWT секретный ключ (в продакшене должен храниться в безопасном месте)
var jwtSecret = generateSecureKey(32) // 256-bit key

type AuthService struct {
	userRepo    *repository.UserRepository
	adminToken  string
	jwtSecret   []byte
	tokenExpiry time.Duration
	tokenCache  *cache.MemoryCache // Кеш для токенов
}

// Claims - структура для хранения данных в токене
type Claims struct {
	UserID string `json:"user_id"`
	Login  string `json:"login"`
	jwt.RegisteredClaims
}
type jwtClaims struct {
	UserID string `json:"user_id"`
	Login  string `json:"login"`
	jwt.RegisteredClaims
}

func NewAuthService(
	userRepo *repository.UserRepository,
	adminToken string,
	jwtSecret []byte,
	tokenCache *cache.MemoryCache,
) *AuthService {
	if len(jwtSecret) == 0 {
		panic("jwt secret cannot be empty")
	}

	return &AuthService{
		userRepo:    userRepo,
		adminToken:  adminToken,
		jwtSecret:   jwtSecret,
		tokenExpiry: 24 * time.Hour,
		tokenCache:  tokenCache,
	}
}

func (s *AuthService) Register(adminToken, login, password string) error {
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
		return fmt.Errorf("failed to hash password: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	userID, err := generateID()
	if err != nil {
		return fmt.Errorf("failed to generate user ID: %w", err)
	}

	return s.userRepo.CreateUser(ctx, userID, login, string(hashedPassword))
}

func (s *AuthService) Authenticate(login, password string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Проверяем кеш перед обращением к БД
	cacheKey := "auth_" + login
	if cachedToken, found := s.tokenCache.Get(cacheKey); found {
		if token, ok := cachedToken.(string); ok {
			return token, nil
		}
	}

	user, err := s.userRepo.GetUserByLogin(ctx, login)
	if err != nil {
		return "", fmt.Errorf("failed to get user: %w", err)
	}
	if user == nil {
		return "", ErrInvalidCredentials
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		return "", ErrInvalidCredentials
	}

	token, err := s.generateJWTToken(user.ID, user.Login)
	if err != nil {
		return "", fmt.Errorf("failed to generate token: %w", err)
	}

	expiryTime := time.Now().Add(s.tokenExpiry)
	if err := s.userRepo.UpdateUserToken(ctx, user.ID, token, expiryTime); err != nil {
		return "", fmt.Errorf("failed to update user token: %w", err)
	}

	// Сохраняем токен в кеш
	s.tokenCache.Set(cacheKey, token, s.tokenExpiry)

	return token, nil
}
func (s *AuthService) ValidateToken(tokenString string) (*model.User, error) {
	// Проверяем кеш перед валидацией токена
	cacheKey := "token_" + tokenString
	if cachedUser, found := s.tokenCache.Get(cacheKey); found {
		if user, ok := cachedUser.(*model.User); ok {
			return user, nil
		}
	}

	claims, err := s.parseJWTToken(tokenString)
	if err != nil {
		return nil, fmt.Errorf("invalid token: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	user, err := s.userRepo.GetUserByID(ctx, claims.UserID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	if user == nil {
		return nil, errors.New("user not found")
	}

	if user.Token != tokenString {
		return nil, errors.New("token mismatch")
	}

	if time.Now().After(user.TokenExpiry) {
		return nil, ErrTokenExpired
	}

	// Сохраняем пользователя в кеш
	s.tokenCache.Set(cacheKey, user, time.Until(user.TokenExpiry))

	return user, nil
}
func (s *AuthService) Logout(tokenString string) error {
	claims, err := s.parseJWTToken(tokenString)
	if err != nil {
		return fmt.Errorf("invalid token: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Инвалидируем кеш
	s.tokenCache.Delete("auth_" + claims.Login)
	s.tokenCache.Delete("token_" + tokenString)

	return s.userRepo.UpdateUserToken(ctx, claims.UserID, "", time.Time{})
}

func (s *AuthService) generateJWTToken(userID, login string) (string, error) {
	claims := &jwtClaims{
		UserID: userID,
		Login:  login,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(s.tokenExpiry)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    "docs-server",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(s.jwtSecret)
}

func (s *AuthService) parseJWTToken(tokenString string) (*jwtClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &jwtClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return s.jwtSecret, nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*jwtClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, errors.New("invalid token")
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
func generateSecureKey(length int) []byte {
	key := make([]byte, length)
	_, err := rand.Read(key)
	if err != nil {
		panic(fmt.Sprintf("failed to generate secure key: %v", err))
	}
	return key
}

func GenerateToken(userID, login string) (string, error) {
	// Создаем claims с данными пользователя
	claims := &Claims{
		UserID: userID,
		Login:  login,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)), // Токен действителен 24 часа
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    "docs-server",
		},
	}

	// Создаем токен с методом подписи HS256
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// Подписываем токен секретным ключом
	signedToken, err := token.SignedString(jwtSecret)
	if err != nil {
		return "", fmt.Errorf("failed to sign token: %w", err)
	}

	return signedToken, nil
}

func ParseToken(tokenString string) (*Claims, error) {
	// Парсим токен с нашими claims
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		// Проверяем метод подписи
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return jwtSecret, nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to parse token: %w", err)
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}

	return nil, fmt.Errorf("invalid token")
}

// Вспомогательная функция для генерации случайного секрета (можно использовать для инициализации)
func GenerateRandomSecret() string {
	return base64.StdEncoding.EncodeToString(generateSecureKey(32))
}
