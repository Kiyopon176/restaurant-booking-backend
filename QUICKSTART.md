# 🚀 Быстрый старт

## 1. Запустите PostgreSQL
Убедитесь что PostgreSQL запущен на порту 5432

## 2. Создайте базу данных (если нужно)
```sql
CREATE DATABASE dastarkhan_db;
```

## 3. Запустите сервер
```bash
go run cmd/api/main.go
```

## 4. Запустите тесты (в новом терминале)
```powershell
.\test-api.ps1
```

## Быстрая проверка через curl

### Регистрация
```bash
curl -X POST http://localhost:8080/api/auth/register \
  -H "Content-Type: application/json" \
  -d "{\"email\":\"test@test.com\",\"password\":\"pass12345\",\"name\":\"Test\",\"role\":\"client\"}"
```

### Логин
```bash
curl -X POST http://localhost:8080/api/auth/login \
  -H "Content-Type: application/json" \
  -d "{\"email\":\"test@test.com\",\"password\":\"pass12345\"}"
```

Сохраните `access_token` из ответа.

### Получить информацию о себе
```bash
curl -X GET http://localhost:8080/api/auth/me \
  -H "Authorization: Bearer ВАШ_ACCESS_TOKEN"
```

## 📚 Полная документация
Смотрите `AUTH_README.md` и `ГОТОВО.md`
