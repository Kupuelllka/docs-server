package app

import (
	"crypto/rand"
	"encoding/base64"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Server struct {
		Host string `yaml:"host"`
		Port string `yaml:"port"`
	} `yaml:"server"`
	Database struct {
		DSN string `yaml:"dsn"` // Формат: "user:password@tcp(host:port)/dbname"
	} `yaml:"database"`
	Auth struct {
		AdminToken string `yaml:"admin_token"`
		JWTSecret  string `yaml:"jwt_secret"` // base64-encoded 32-byte secret
	} `yaml:"auth"`
	Storage struct {
		UploadDir string `yaml:"upload_dir"`
	} `yaml:"storage"`
}

// NewConfig загружает конфигурацию из файла или использует значения по умолчанию
func NewConfig() (*Config, error) {
	// Генерация случайного JWT секрета для дефолтной конфигурации
	jwtSecret := generateRandomSecret()

	// Значения по умолчанию
	config := &Config{
		Server: struct {
			Host string `yaml:"host"`
			Port string `yaml:"port"`
		}{
			Host: "127.0.0.1",
			Port: "8080",
		},
		Database: struct {
			DSN string `yaml:"dsn"`
		}{
			DSN: "root:password@tcp(localhost:3306)/documents_db",
		},
		Auth: struct {
			AdminToken string `yaml:"admin_token"`
			JWTSecret  string `yaml:"jwt_secret"`
		}{
			AdminToken: "admin-secret-token",
			JWTSecret:  jwtSecret,
		},
		Storage: struct {
			UploadDir string `yaml:"upload_dir"`
		}{
			UploadDir: "uploads",
		},
	}

	// Пути к возможным расположениям конфигурационных файлов
	configPaths := []string{
		"../../docs-server/prod.yml",        // В текущей директории
		"/etc/docs-server/prod.yml",         // Общий системный конфиг
		filepath.Join("config", "prod.yml"), // В папке config
	}

	// Попробуем найти и загрузить конфигурационный файл
	for _, path := range configPaths {
		if configFile, err := os.Open(path); err == nil {
			defer configFile.Close()
			if err := yaml.NewDecoder(configFile).Decode(config); err != nil {
				return nil, err
			}
			break
		}
	}

	return config, nil
}

// generateRandomSecret генерирует случайный base64-encoded 32-byte secret
func generateRandomSecret() string {
	key := make([]byte, 32)
	if _, err := rand.Read(key); err != nil {
		panic("failed to generate random secret: " + err.Error())
	}
	return base64.StdEncoding.EncodeToString(key)
}

// GetJWTSecret возвращает декодированный JWT секрет
func (c *Config) GetJWTSecret() ([]byte, error) {
	return base64.StdEncoding.DecodeString(c.Auth.JWTSecret)
}

// LoadConfigFromFile явно загружает конфигурацию из указанного файла
func LoadConfigFromFile(path string) (*Config, error) {
	config := &Config{}

	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	decoder := yaml.NewDecoder(file)
	if err := decoder.Decode(config); err != nil {
		return nil, err
	}

	return config, nil
}
