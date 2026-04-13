# 1. Создать пользователя
curl -X POST "http://localhost:8069/sync?id=123e4567-e89b-12d3-a456-426614174000&email=test@example.com"

# 2. Получить свой профиль
curl "http://localhost:8069/profile?user_id=123e4567-e89b-12d3-a456-426614174000"

# 3. Обновить профиль
curl -X PUT "http://localhost:8069/profile/update?user_id=123e4567-e89b-12d3-a456-426614174000" \
  -H "Content-Type: application/json" \
  -d '{"username":"john_doe","full_name":"John Doe","age":30}'

# 4. Получить публичный профиль другого пользователя
curl "http://localhost:8069/user?user_id=123e4567-e89b-12d3-a456-426614174000"

# 5. Health check
curl "http://localhost:8069/health"

# 6. Удалить пользователя
curl -X DELETE "http://localhost:8069/profile/delete?user_id=123e4567-e89b-12d3-a456-426614174000"
