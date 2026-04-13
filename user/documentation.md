# User Service API Documentation

## Общая информация

User Service - микросервис для управления профилями пользователей. Хранит расширенную информацию о пользователях и синхронизируется с Auth Service.

### Технологический стек
- **Язык:** Go 1.26.1
- **База данных:** PostgreSQL 17
- **Контейнеризация:** Docker + Docker Compose
- **Драйвер БД:** lib/pq

### Порты
- **API:** `8069`
- **PostgreSQL:** `5430`

---

## Структура проекта

```
user-service/
├── cmd/
│   └── server/
│       └── main.go              # Точка входа
├── internal/
│   ├── config/
│   │   └── config.go            # Конфигурация сервиса
│   ├── handler/
│   │   └── handler.go           # HTTP обработчики
│   ├── model/
│   │   └── model.go             # Модели данных
│   └── storage/
│       └── postgres.go          # Слой работы с БД
├── migrations/
│   └── 001_create_users_table.sql  # SQL миграции
├── Dockerfile                    # Сборка Docker образа
├── docker-compose.yml            # Оркестрация сервисов
├── go.mod                        # Go модуль
└── go.sum                        # Контрольные суммы зависимостей
```

---

## API Endpoints

### 1. Health Check
Проверка работоспособности сервиса и подключения к БД.

**Endpoint:** `GET /health`

**Response:**
```json
{
  "status": "healthy",
  "users": 42,
  "timestamp": "2026-04-13T15:30:00Z",
  "database": "connected"
}
```

**Коды ответов:**
- `200` - Сервис работает
- `503` - Проблема с подключением к БД

---

### 2. Синхронизация пользователя
Создание нового пользователя или обновление email существующего. Вызывается Auth Service при регистрации или смене email.

**Endpoint:** `POST /sync`

**Request Body:**
```json
{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "email": "user@example.com"
}
```

**Валидация:**
- `id` - обязательное поле, формат UUID
- `email` - обязательное поле, содержит `@`, длина > 5 символов

**Response (создан):**
```json
{
  "status": "created",
  "user_id": "550e8400-e29b-41d4-a716-446655440000"
}
```

**Response (обновлен):**
```json
{
  "status": "updated",
  "user_id": "550e8400-e29b-41d4-a716-446655440000"
}
```

**Коды ответов:**
- `200` - Успешная синхронизация
- `400` - Ошибка валидации
- `500` - Внутренняя ошибка сервера

---

### 3. Получение своего профиля
Возвращает полную информацию о текущем пользователе, включая приватные поля.

**Endpoint:** `GET /profile`

**Headers:**
```
X-User-ID: 550e8400-e29b-41d4-a716-446655440000
```

**Response:**
```json
{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "email": "user@example.com",
  "username": "john_doe",
  "full_name": "John Doe",
  "age": 30,
  "created_at": "2026-04-13T10:00:00Z",
  "updated_at": "2026-04-13T15:30:00Z"
}
```

**Коды ответов:**
- `200` - Успешно
- `400` - Невалидный формат ID
- `401` - Отсутствует заголовок X-User-ID
- `404` - Пользователь не найден

---

### 4. Обновление своего профиля
Обновляет поля профиля текущего пользователя. Поддерживает частичное обновление.

**Endpoint:** `PUT /profile/update` или `PATCH /profile/update`

**Headers:**
```
X-User-ID: 550e8400-e29b-41d4-a716-446655440000
Content-Type: application/json
```

**Request Body (можно отправлять любую комбинацию полей):**
```json
{
  "username": "john_doe",
  "full_name": "John Doe",
  "age": 30
}
```

**Валидация полей:**
| Поле | Ограничения |
|------|-------------|
| `username` | 3-50 символов, не пустой |
| `full_name` | 1-100 символов, не пустой |
| `age` | 0-150 |

**Response:**
```json
{
  "status": "updated",
  "user": {
    "id": "550e8400-e29b-41d4-a716-446655440000",
    "email": "user@example.com",
    "username": "john_doe",
    "full_name": "John Doe",
    "age": 30,
    "created_at": "2026-04-13T10:00:00Z",
    "updated_at": "2026-04-13T15:35:00Z"
  }
}
```

**Коды ответов:**
- `200` - Успешно обновлено
- `400` - Ошибка валидации или пустое тело
- `401` - Отсутствует заголовок X-User-ID
- `404` - Пользователь не найден

---

### 5. Получение публичного профиля
Возвращает публичную информацию о любом пользователе по ID (без email).

**Endpoint:** `GET /user/{id}`

**Пример:** `GET /user/550e8400-e29b-41d4-a716-446655440000`

**Response:**
```json
{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "username": "john_doe",
  "full_name": "John Doe",
  "age": 30,
  "created_at": "2026-04-13T10:00:00Z"
}
```

**Коды ответов:**
- `200` - Успешно
- `400` - Невалидный ID
- `404` - Пользователь не найден

---

### 6. Удаление своего аккаунта
Удаляет профиль текущего пользователя.

**Endpoint:** `DELETE /profile/delete`

**Headers:**
```
X-User-ID: 550e8400-e29b-41d4-a716-446655440000
```

**Response:**
```json
{
  "status": "deleted",
  "user_id": "550e8400-e29b-41d4-a716-446655440000"
}
```

**Коды ответов:**
- `200` - Успешно удален
- `400` - Невалидный ID
- `401` - Отсутствует заголовок X-User-ID
- `404` - Пользователь не найден

---

## Модели данных

### User (полная модель)
```go
type User struct {
    ID        string    `json:"id"`         // UUID пользователя
    Email     string    `json:"email"`      // Email (приватное поле)
    Username  string    `json:"username"`   // Имя пользователя
    FullName  string    `json:"full_name"`  // Полное имя
    Age       int       `json:"age"`        // Возраст (0-150)
    CreatedAt time.Time `json:"created_at"` // Дата создания
    UpdatedAt time.Time `json:"updated_at"` // Дата обновления
}
```

### PublicUser (публичная модель)
```go
type PublicUser struct {
    ID        string    `json:"id"`
    Username  string    `json:"username"`
    FullName  string    `json:"full_name"`
    Age       int       `json:"age"`
    CreatedAt time.Time `json:"created_at"`
}
```

---

## Схема базы данных

```sql
CREATE TABLE users (
    id UUID PRIMARY KEY,
    email VARCHAR(255) NOT NULL,
    username VARCHAR(50),
    full_name VARCHAR(100),
    age INTEGER CHECK (age >= 0 AND age <= 150),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_users_email ON users(email);
CREATE INDEX idx_users_username ON users(username);
```

---

## Переменные окружения

| Переменная | Описание | Значение по умолчанию |
|------------|----------|----------------------|
| `DB_HOST` | Хост PostgreSQL | `localhost` (в Docker: `postgres`) |
| `DB_PORT` | Порт PostgreSQL | `5430` |
| `DB_USER` | Пользователь БД | `useradmin` |
| `DB_PASSWORD` | Пароль БД | `secretpassword` |
| `DB_NAME` | Имя БД | `userdb` |

---

## Запуск сервиса

### Локальная разработка
```bash
# 1. Запустить только PostgreSQL
docker-compose up -d postgres

# 2. Запустить приложение
export DB_HOST=localhost DB_PORT=5430
go run ./cmd/server
```

### Полный запуск через Docker
```bash
# Собрать и запустить всё
docker-compose up -d

# Посмотреть логи
docker-compose logs -f

# Остановить
docker-compose down
```

---

## Интеграция с другими сервисами

### Для Auth Service
При регистрации нового пользователя или смене email, Auth Service должен вызвать:

```http
POST /sync
Content-Type: application/json

{
  "id": "{{user_id}}",
  "email": "{{user_email}}"
}
```

### Для других сервисов
Для получения публичной информации о пользователе:

```http
GET /user/{{user_id}}
```

Для проверки авторизации (требуется заголовок):
```http
GET /profile
X-User-ID: {{user_id}}
```

---

## Обработка ошибок

Все ошибки возвращаются в формате:
```json
{
  "error": "описание ошибки"
}
```

### Коды ошибок
| Код | Описание |
|-----|----------|
| 400 | Ошибка валидации данных |
| 401 | Отсутствует или неверный заголовок авторизации |
| 404 | Ресурс не найден |
| 405 | Метод не поддерживается |
| 500 | Внутренняя ошибка сервера |
| 503 | Сервис недоступен |

---

## Мониторинг

### Health Check эндпоинт
```bash
curl http://localhost:8069/health
```

Используется для:
- Проверки доступности сервиса
- Мониторинга количества пользователей
- Проверки подключения к БД

### Логи
Логи содержат:
- Информацию о запуске сервиса
- Ошибки подключения к БД
- Ошибки выполнения запросов
- Информацию о создании/обновлении пользователей

---

## Примеры использования

### Полный цикл работы с пользователем

```bash
# 1. Auth Service создает пользователя
curl -X POST http://localhost:8069/sync \
  -H "Content-Type: application/json" \
  -d '{"id":"550e8400-e29b-41d4-a716-446655440000","email":"new@example.com"}'

# 2. Пользователь заполняет профиль
curl -X PUT http://localhost:8069/profile/update \
  -H "X-User-ID: 550e8400-e29b-41d4-a716-446655440000" \
  -H "Content-Type: application/json" \
  -d '{"username":"alice","full_name":"Alice Smith","age":28}'

# 3. Другой сервис получает публичный профиль
curl http://localhost:8069/user/550e8400-e29b-41d4-a716-446655440000

# 4. Пользователь меняет возраст
curl -X PATCH http://localhost:8069/profile/update \
  -H "X-User-ID: 550e8400-e29b-41d4-a716-446655440000" \
  -H "Content-Type: application/json" \
  -d '{"age":29}'

# 5. Пользователь удаляет аккаунт
curl -X DELETE http://localhost:8069/profile/delete \
  -H "X-User-ID: 550e8400-e29b-41d4-a716-446655440000"
```

---

## Создание нового микросервиса по образу

Для создания нового сервиса на основе этой архитектуры:

1. Скопировать структуру проекта
2. Изменить название модуля в `go.mod`
3. Обновить модели в `internal/model/`
4. Изменить SQL миграции под свои таблицы
5. Обновить `internal/storage/` под новую схему
6. Добавить обработчики в `internal/handler/`
7. Обновить порты в `docker-compose.yml`
8. Создать документацию по API

---

## Контакты
DS #stupidrushian