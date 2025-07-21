package app

import (
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Server struct {
		Port string `yaml:"port"`
	} `yaml:"server"`
	Database struct {
		DSN string `yaml:"dsn"` // Формат: "user:password@tcp(host:port)/dbname"
	} `yaml:"database"`
	Auth struct {
		AdminToken string `yaml:"admin_token"`
	} `yaml:"auth"`
	Storage struct {
		UploadDir string `yaml:"upload_dir"`
	} `yaml:"storage"`
}

// NewConfig загружает конфигурацию из файла или использует значения по умолчанию
func NewConfig() (*Config, error) {
	// Значения по умолчанию
	config := &Config{
		Server: struct {
			Port string `yaml:"port"`
		}{Port: "8080"},
		Database: struct {
			DSN string `yaml:"dsn"`
		}{DSN: "root:password@tcp(localhost:3306)/documents_db"},
		Auth: struct {
			AdminToken string `yaml:"admin_token"`
		}{AdminToken: "admin-secret-token"},
		Storage: struct {
			UploadDir string `yaml:"upload_dir"`
		}{UploadDir: "uploads"},
	}

	// Пути к возможным расположениям конфигурационных файлов
	configPaths := []string{
		"prod.yml",                            // В текущей директории
		"/etc/docs-server/config.yml",         // Общий системный конфиг
		filepath.Join("config", "config.yml"), // В папке config
	}

	// Попробуем найти и загрузить конфигурационный файл
	var configFile *os.File
	var err error

	for _, path := range configPaths {
		configFile, err = os.Open(path)
		if err == nil {
			defer configFile.Close()
			break
		}
	}

	// Если файл конфигурации найден - загружаем его
	if configFile != nil {
		decoder := yaml.NewDecoder(configFile)
		if err := decoder.Decode(config); err != nil {
			return nil, err
		}
	}

	return config, nil
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
