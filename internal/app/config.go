package app

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

func NewConfig() (*Config, error) {
	return &Config{
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
	}, nil
}
