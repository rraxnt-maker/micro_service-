# Microservices Platform Documentation

## Обзор системы

Платформа состоит из двух микросервисов:
- **User Service** (порт 8069) - управление профилями пользователей
- **Status Service** (порт 8070) - управление статусами пользователей

### Архитектура
```
┌─────────────────┐     ┌─────────────────┐
│   Auth Service  │────▶│   User Service  │
│   (будущий)     │     │     :8069       │
└─────────────────┘     └────────┬────────┘
                                 │
                                 ▼
                        ┌─────────────────┐
                        │  Status Service │
                        │     :8070       │
                        └─────────────────┘
```

### Технологический стек
- **Язык:** Go 1.26.1
- **База данных:** PostgreSQL 17
- **Контейнеризация:** Docker + Docker Compose
- **Драйвер БД:** lib/pq
- **Валидация:** UUID v4 для идентификаторов

---

# User Service

## Назначение
Управление профилями пользователей, хранение расширенной информации о пользователях, синхронизация с Auth Service.

## Порты
- **API:** `8069`
- **PostgreSQL:** `5430`

## Структура проекта
```
user-service/
├── cmd/server/main.go           # Точка входа
├── internal/
│   ├── config/config.go         # Конфигурация
│   ├── handler/handler.go       # HTTP обработчики
│   ├── model/model.go           # Модели данных
│   └── storage/postgres.go      # Слой работы с БД
├── migrations/
│   └── 001_create_users_table.sql
├── Dockerfile
├── docker-compose.yml
├── go.mod
└── go.sum
```

## Модель данных

### User
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

### PublicUser (публичная модель)
```json
{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "username": "john_doe",
  "full_name": "John Doe",
  "age": 30,
  "created_at": "2026-04-13T10:00:00Z"
}
```

## Схема БД
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
```

## API Endpoints

### 1. Health Check
```
GET /health
```
**Ответ:**
```json
{
  "status": "healthy",
  "users": 42,
  "timestamp": "2026-04-13T15:30:00Z",
  "database": "connected"
}
```

### 2. Синхронизация пользователя
```
POST /sync
Content-Type: application/json

{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "email": "user@example.com"
}
```
**Ответ:**
```json
{
  "status": "created|updated",
  "user_id": "550e8400-e29b-41d4-a716-446655440000"
}
```

### 3. Получение своего профиля
```
GET /profile
X-User-ID: 550e8400-e29b-41d4-a716-446655440000
```
**Ответ:** объект `User`

### 4. Обновление профиля
```
PUT /profile/update
PATCH /profile/update
X-User-ID: 550e8400-e29b-41d4-a716-446655440000
Content-Type: application/json

{
  "username": "new_username",
  "full_name": "New Name",
  "age": 25
}
```
**Валидация:**
- `username`: 3-50 символов
- `full_name`: 1-100 символов
- `age`: 0-150

**Ответ:**
```json
{
  "status": "updated",
  "user": { ... }
}
```

### 5. Получение публичного профиля
```
GET /user/{id}
```
**Ответ:** объект `PublicUser`

### 6. Удаление аккаунта
```
DELETE /profile/delete
X-User-ID: 550e8400-e29b-41d4-a716-446655440000
```
**Ответ:**
```json
{
  "status": "deleted",
  "user_id": "550e8400-e29b-41d4-a716-446655440000"
}
```

## Переменные окружения
| Переменная | Описание | По умолчанию |
|------------|----------|--------------|
| `DB_HOST` | Хост PostgreSQL | `localhost` |
| `DB_PORT` | Порт PostgreSQL | `5430` |
| `DB_USER` | Пользователь БД | `useradmin` |
| `DB_PASSWORD` | Пароль БД | `secretpassword` |
| `DB_NAME` | Имя БД | `userdb` |
| `STATUS_SERVICE_URL` | URL Status Service | `http://localhost:8070` |
| `INTERNAL_TOKEN` | Токен для внутренних запросов | `secret-internal-token-123` |

## Коды ошибок
| Код | Описание |
|-----|----------|
| 200 | Успешно |
| 400 | Ошибка валидации |
| 401 | Не авторизован |
| 404 | Не найдено |
| 405 | Метод не поддерживается |
| 500 | Внутренняя ошибка |
| 503 | Сервис недоступен |

---

# Status Service

## Назначение
Управление статусами пользователей, поддержка временных статусов, режим "Не беспокоить", история изменений.

## Порты
- **API:** `8070`
- **PostgreSQL:** `5431`

## Структура проекта
```
status-service/
├── cmd/server/main.go
├── internal/
│   ├── config/config.go
│   ├── handler/handler.go
│   ├── model/model.go
│   ├── storage/postgres.go
│   └── cleaner/cleaner.go      # Очистка истекших статусов
├── migrations/
│   └── 001_create_statuses.sql
├── Dockerfile
├── docker-compose.yml
├── go.mod
└── go.sum
```

## Модели данных

### Status
```json
{
  "user_id": "550e8400-e29b-41d4-a716-446655440000",
  "text": "Working on a project",
  "emoji": "💻",
  "type": "custom",
  "activity": "working",
  "expires_at": "2026-04-13T17:30:00Z",
  "created_at": "2026-04-13T15:30:00Z",
  "updated_at": "2026-04-13T15:30:00Z"
}
```

### StatusHistory
```json
{
  "id": "660e8400-e29b-41d4-a716-446655440000",
  "user_id": "550e8400-e29b-41d4-a716-446655440000",
  "text": "Previous status",
  "emoji": "😴",
  "created_at": "2026-04-13T10:00:00Z"
}
```

## Типы статусов
| Тип | Описание |
|-----|----------|
| `normal` | Обычный статус |
| `dnd` | Не беспокоить |
| `custom` | Кастомный с активностью |

## Активности (для custom)
- `working` - Работаю
- `meeting` - На встрече
- `commuting` - В пути
- `vacation` - В отпуске
- `gaming` - Играю
- `sleeping` - Сплю
- `studying` - Учусь

## Схема БД
```sql
-- Текущие статусы
CREATE TABLE statuses (
    user_id UUID PRIMARY KEY,
    text VARCHAR(140) NOT NULL,
    emoji VARCHAR(10),
    type VARCHAR(20) DEFAULT 'normal',
    activity VARCHAR(50),
    expires_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- История статусов
CREATE TABLE status_history (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL,
    text VARCHAR(140) NOT NULL,
    emoji VARCHAR(10),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);
```

## API Endpoints

### 1. Health Check
```
GET /health
```
**Ответ:**
```json
{
  "status": "healthy",
  "timestamp": "2026-04-13T15:30:00Z",
  "database": "connected"
}
```

### 2. Установка статуса
```
PUT /status
X-User-ID: {user_id}
Content-Type: application/json

{
  "text": "Working on a project",
  "emoji": "💻",
  "type": "custom",
  "activity": "working",
  "expires_in": 3600
}
```
**Валидация:**
- `text`: 1-140 символов, обязательное
- `emoji`: до 10 символов
- `type`: normal, dnd, custom
- `expires_in`: секунды, ≥0

**Ответ:**
```json
{
  "status": "set",
  "data": { ... }
}
```

### 3. Получение своего статуса
```
GET /status
X-User-ID: {user_id}
```
**Ответ:** объект `Status`

### 4. Получение статуса пользователя
```
GET /status/{user_id}
```
**Ответ:** объект `PublicStatus`

### 5. Получение нескольких статусов
```
POST /status/batch
Content-Type: application/json

{
  "user_ids": ["id1", "id2", "id3"]
}
```
**Ограничение:** до 100 ID

**Ответ:**
```json
{
  "statuses": [
    { "user_id": "id1", "text": "...", "emoji": "..." },
    { "user_id": "id2", "text": "...", "emoji": "..." }
  ]
}
```

### 6. Удаление статуса
```
DELETE /status
X-User-ID: {user_id}
```
**Ответ:**
```json
{
  "status": "cleared"
}
```

### 7. Режим "Не беспокоить"
```
POST /status/dnd
X-User-ID: {user_id}
Content-Type: application/json

{
  "duration": 7200
}
```
**Ответ:**
```json
{
  "status": "dnd_enabled",
  "data": { ... },
  "until": "2026-04-13T17:30:00Z"
}
```

### 8. История статусов
```
GET /status/history?limit=10&offset=0
X-User-ID: {user_id}
```
**Параметры:**
- `limit`: 1-100 (по умолчанию 10)
- `offset`: ≥0 (по умолчанию 0)

**Ответ:**
```json
{
  "history": [ ... ],
  "total": 25,
  "limit": 10,
  "offset": 0
}
```

### 9. Внутренние эндпоинты

#### Синхронизация пользователя
```
POST /internal/sync
X-Internal-Token: {token}
Content-Type: application/json

{
  "user_id": "550e8400-e29b-41d4-a716-446655440000"
}
```

#### Удаление пользователя
```
DELETE /internal/user/{user_id}
X-Internal-Token: {token}
```

## Переменные окружения
| Переменная | Описание | По умолчанию |
|------------|----------|--------------|
| `DB_HOST` | Хост PostgreSQL | `localhost` |
| `DB_PORT` | Порт PostgreSQL | `5431` |
| `DB_USER` | Пользователь БД | `statusadmin` |
| `DB_PASSWORD` | Пароль БД | `statuspass` |
| `DB_NAME` | Имя БД | `statusdb` |
| `INTERNAL_TOKEN` | Токен для внутренних запросов | `secret-internal-token-123` |

## Фоновые процессы
- **Cleaner:** удаляет истекшие статусы каждую минуту

---

# Интеграция сервисов

## Сетевое взаимодействие
Сервисы общаются через общую Docker-сеть `microservices-network`:
```bash
docker network create microservices-network
```

## Сценарии интеграции

### 1. Создание пользователя
```
Auth Service → User Service → Status Service
```
1. Auth Service вызывает `POST /sync` в User Service
2. User Service создает/обновляет пользователя
3. User Service асинхронно вызывает `POST /internal/sync` в Status Service

### 2. Удаление пользователя
```
User Service → Status Service
```
1. User Service получает `DELETE /profile/delete`
2. User Service удаляет пользователя из своей БД
3. User Service асинхронно вызывает `DELETE /internal/user/{id}` в Status Service
4. Status Service удаляет статус и историю пользователя

## Формат внутренних запросов
Все внутренние запросы требуют заголовок:
```
X-Internal-Token: secret-internal-token-123
```

---

# Запуск системы

## 1. Создание общей сети
```bash
docker network create microservices-network
```

## 2. Запуск Status Service
```bash
cd status-service
docker-compose up -d
```

## 3. Запуск User Service
```bash
cd ../user-service
docker-compose up -d
```

## 4. Проверка
```bash
# Проверить контейнеры
docker ps

# Проверить сеть
docker network inspect microservices-network

# Проверить health
curl http://localhost:8069/health
curl http://localhost:8070/health
```

## 5. Остановка
```bash
cd status-service && docker-compose down
cd ../user-service && docker-compose down
```

---

# Тестирование

## Полный тестовый сценарий
```bash
# 1. Создать пользователя
curl -X POST http://localhost:8069/sync \
  -H "Content-Type: application/json" \
  -d '{"id":"123e4567-e89b-12d3-a456-426614174000","email":"test@example.com"}'

# 2. Обновить профиль
curl -X PUT http://localhost:8069/profile/update \
  -H "X-User-ID: 123e4567-e89b-12d3-a456-426614174000" \
  -H "Content-Type: application/json" \
  -d '{"username":"john","full_name":"John Doe","age":30}'

# 3. Установить статус
curl -X PUT http://localhost:8070/status \
  -H "X-User-ID: 123e4567-e89b-12d3-a456-426614174000" \
  -H "Content-Type: application/json" \
  -d '{"text":"Hello world!","emoji":"👋"}'

# 4. Получить профиль
curl http://localhost:8069/user/123e4567-e89b-12d3-a456-426614174000

# 5. Получить статус
curl http://localhost:8070/status/123e4567-e89b-12d3-a456-426614174000

# 6. Включить DND на 30 минут
curl -X POST http://localhost:8070/status/dnd \
  -H "X-User-ID: 123e4567-e89b-12d3-a456-426614174000" \
  -H "Content-Type: application/json" \
  -d '{"duration":1800}'

# 7. История статусов
curl "http://localhost:8070/status/history?limit=5" \
  -H "X-User-ID: 123e4567-e89b-12d3-a456-426614174000"

# 8. Удалить пользователя
curl -X DELETE http://localhost:8069/profile/delete \
  -H "X-User-ID: 123e4567-e89b-12d3-a456-426614174000"
```

## Проверка интеграции
```bash
# Посмотреть логи интеграции
docker logs user_app | grep "Status Service"
# Должно быть:
# ✅ Status Service notified: sync for user 123e4567-e89b-12d3-a456-426614174000
# ✅ Status Service notified: delete for user 123e4567-e89b-12d3-a456-426614174000
```

---

# Мониторинг

## Health Check эндпоинты
- User Service: `GET http://localhost:8069/health`
- Status Service: `GET http://localhost:8070/health`

## Метрики (доступные в ответах)
- Количество пользователей
- Статус подключения к БД
- Timestamp последней проверки

## Логи
```bash
# Все логи
docker-compose logs -f

# Только User Service
docker logs user_app -f

# Только Status Service
docker logs status_app -f

# Только PostgreSQL User
docker logs user_postgres -f

# Только PostgreSQL Status
docker logs status_postgres -f
```

---

# Безопасность

## Внутренние запросы
- Требуют заголовок `X-Internal-Token`
- Токен передается через переменную окружения `INTERNAL_TOKEN`
- В production использовать сложный случайный токен

## Пользовательские запросы
- Требуют заголовок `X-User-ID`
- ID должен быть в формате UUID v4
- Валидация всех входных данных

## Рекомендации для production
1. Использовать HTTPS
2. Настроить rate limiting
3. Добавить JWT аутентификацию вместо X-User-ID
4. Использовать секреты Docker/Kubernetes для паролей
5. Настроить репликацию БД
6. Добавить сбор метрик (Prometheus)
7. Настроить алерты при падении сервисов

---

# Создание нового сервиса (шаблон)

Для создания нового микросервиса в экосистеме:

1. Скопировать структуру `status-service`
2. Изменить название модуля в `go.mod`
3. Обновить порты:
   - API: следующий свободный (8071, 8072...)
   - PostgreSQL: следующий свободный (5432, 5433...)
4. Создать модели в `internal/model/`
5. Создать SQL миграции
6. Реализовать `internal/storage/`
7. Добавить обработчики в `internal/handler/`
8. Обновить `docker-compose.yml`
9. Добавить сервис в общую сеть `microservices-network`
10. При необходимости добавить интеграцию с существующими сервисами

---

# ЧаВо

**Q: Почему статус не создается автоматически при создании пользователя?**
A: Статус - опциональная функция. Пользователь сам решает когда его установить.

**Q: Что происходит при удалении пользователя?**
A: User Service удаляет пользователя и асинхронно уведомляет Status Service, который удаляет все связанные данные.

**Q: Как долго хранится история статусов?**
A: Бессрочно, пока пользователь не будет удален.

**Q: Можно ли использовать сервисы без Docker?**
A: Да, нужно запустить PostgreSQL локально и установить правильные переменные окружения.

**Q: Как добавить WebSocket для real-time уведомлений?**
A: Нужно создать отдельный WebSocket сервис или добавить в существующие.

---

# Контакты

discord @stupidrushian

**Версия документации:** 1.0.0  
**Последнее обновление:** 2026-04-13