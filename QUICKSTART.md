# 🚀 Быстрый старт

## Запуск через Docker (рекомендуется)
1) Подготовьте `.env` (пример значений):
```
DB_HOST=db
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=postgres
DB_NAME=dastarkhan_db
JWT_SECRET=replace_with_strong_secret
JWT_ACCESS_EXPIRE=15m
JWT_REFRESH_EXPIRE=168h
PORT=8080
PGADMIN_EMAIL=admin@example.com
PGADMIN_PASSWORD=admin123
```
2) Соберите и поднимите контейнеры:
```powershell
docker compose build api
docker compose up -d
docker compose logs -n 80 api
```
3) Проверки:
- API: `http://localhost:8080/health`
- Swagger: `http://localhost:8080/swagger/index.html`
- pgAdmin: `http://localhost:5050` (сервер: host `db`, user `DB_USER`, password `DB_PASSWORD`)

Остановить/очистить:
```powershell
docker compose down
docker compose down -v  # полный сброс БД
```

## Локальный запуск (без Docker)
1) Установите PostgreSQL и создайте БД:
```sql
CREATE DATABASE dastarkhan_db;
```
2) Подготовьте `.env` (локально):
```
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=postgres
DB_NAME=dastarkhan_db
JWT_SECRET=replace_with_strong_secret
JWT_ACCESS_EXPIRE=15m
JWT_REFRESH_EXPIRE=168h
PORT=8080
```
3) Запустите API:
```powershell
go run cmd/api/main.go
```

## Быстрая проверка API

### Регистрация
```bash
curl -X POST http://localhost:8080/api/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "email": "test@test.com",
    "password": "P@ssw0rd!",
    "first_name": "Test",
    "last_name": "User",
    "phone": "1234567890",
    "role": "customer"
  }'
```

### Логин
```bash
curl -X POST http://localhost:8080/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"test@test.com","password":"P@ssw0rd!"}'
```

Сохраните `access_token` из ответа.

### Получить информацию о себе
```bash
curl -X GET http://localhost:8080/api/auth/me \
  -H "Authorization: Bearer ВАШ_ACCESS_TOKEN"
```

## 📚 Полная документация
Смотрите `AUTH_README.md` и `ГОТОВО.md`
