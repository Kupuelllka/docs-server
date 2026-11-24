# Docs Server

Микросервис для хранения и управления документами с аутентификацией и авторизацией.

---

## ✨ Особенности

- 📄 Хранение документов (файлы и JSON)
- 🔐 JWT-аутентификация
- 👥 Управление правами доступа
- 🚀 Быстрые ответы благодаря кешированию
- 📁 Загрузка файлов с метаданными

---

## 📋 Требования

- Go `1.20+`
- MySQL `8.0+`

---

## 🚀 Установка

### 1. Клонировать репозиторий

```bash
git clone https://github.com/yourusername/docs-server.git
cd docs-server
```

### 2. Настроить конфигурацию

См. раздел [Конфигурация](#-конфигурация).

### 3. Установить зависимости

```bash
go mod tidy
```

### 4. Запустить сервер

```bash
go run cmd/main.go
```

---

## ⚙️ Конфигурация

Создайте файл `config.yml` в корне проекта или в `/etc/docs-server/`:

```yaml
server:
  host: "127.0.0.1"
  port: "8080"

database:
  dsn: "user:password@tcp(localhost:3306)/docs_db"

auth:
  admin_token: "your-secret-admin-token"
  jwt_secret: "base64-encoded-32-byte-secret"

storage:
  upload_dir: "uploads"
```

Либо используйте **переменные окружения**:

```bash
export DB_DSN="user:password@tcp(localhost:3306)/docs_db"
export JWT_SECRET="your-jwt-secret"
```

---

## 📡 API Endpoints

### 🔐 Аутентификация

| Метод | Endpoint | Описание |
|-------|----------|----------|
| `POST` | `/api/register` | Регистрация (только для админа) |
| `POST` | `/api/auth` | Вход (получение токена) |
| `DELETE` | `/api/auth/:token` | Выход (инвалидация токена) |

### 📄 Документы

| Метод | Endpoint | Описание |
|-------|----------|----------|
| `POST` | `/api/docs` | Загрузить документ |
| `GET` | `/api/docs` | Список документов |
| `GET` | `/api/docs/:id` | Получить документ |
| `DELETE` | `/api/docs/:id` | Удалить документ |

---

## 📚 Подробная документация

Актуальная спецификация API доступна в Swagger:

```
docs/swagger
```
