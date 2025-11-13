# 🍽️ Restaurant Booking Platform

## 🎯 Архитектура проекта

### Роли в системе:

- **Client** — бронирует столики
- **Restaurant Owner** — владелец ресторана, регистрирует заведение
- **Restaurant Manager** — управляет бронями конкретного ресторана
- **Admin** — супер-админ платформы

---

## 📊 Структура базы данных

### users — все пользователи системы

| Поле | Описание |
| --- | --- |
| id |  |
| email |  |
| password_hash |  |
| name |  |
| phone |  |
| role | client / owner / manager / admin |
| oauth_provider | google / apple / null |
| oauth_id |  |
| created_at, updated_at |  |

---

### wallets — кошельки клиентов

| Поле | Описание |
| --- | --- |
| id |  |
| user_id |  |
| balance |  |
| created_at, updated_at |  |

---

### wallet_transactions — история операций кошелька

| Поле | Описание |
| --- | --- |
| id |  |
| wallet_id |  |
| amount |  |
| type | deposit / withdraw / booking_charge / refund |
| description |  |
| created_at |  |

---

### restaurants — рестораны

| Поле | Описание |
| --- | --- |
| id |  |
| owner_id (user_id) |  |
| name |  |
| address |  |
| latitude, longitude |  |
| description |  |
| phone |  |
| instagram |  |
| website |  |
| cuisine_type |  |
| average_price |  |
| max_combinable_tables | макс объединяемых столиков |
| working_hours | JSON: график работы |
| created_at, updated_at |  |

---

### restaurant_images — фото ресторанов

| Поле | Описание |
| --- | --- |
| id |  |
| restaurant_id |  |
| cloudinary_url |  |
| is_main |  |
| created_at |  |

---

### tables — столики в ресторане

| Поле | Описание |
| --- | --- |
| id |  |
| restaurant_id |  |
| table_number |  |
| min_capacity |  |
| max_capacity |  |
| location_type | window / vip / regular / outdoor |
| x_position, y_position | для схемы |
| is_active |  |
| created_at, updated_at |  |

---

### bookings — брони

| Поле | Описание |
| --- | --- |
| id |  |
| user_id |  |
| restaurant_id |  |
| booking_date |  |
| start_time, end_time |  |
| guests_count |  |
| status | pending / confirmed / cancelled / completed / no_show |
| booking_type | standard / corporate |
| deposit_amount |  |
| is_deposit_paid |  |
| payment_method | wallet / halyk / kaspi |
| cancelled_at |  |
| created_at, updated_at |  |

---

### booking_tables — связь бронь-столики (many-to-many)

| Поле | Описание |
| --- | --- |
| id |  |
| booking_id |  |
| table_id |  |

---

### reviews — отзывы

| Поле | Описание |
| --- | --- |
| id |  |
| user_id |  |
| restaurant_id |  |
| booking_id |  |
| rating (1–5) |  |
| comment |  |
| created_at, updated_at |  |

---

### restaurant_managers — менеджеры ресторанов

| Поле | Описание |
| --- | --- |
| id |  |
| user_id |  |
| restaurant_id |  |
| assigned_at |  |

---

### audit_logs — история изменений (логирование)

| Поле | Описание |
| --- | --- |
| id |  |
| restaurant_id |  |
| user_id | кто изменил |
| action_type |  |
| entity_type | booking / restaurant / table |
| entity_id |  |
| old_value | JSON |
| new_value | JSON |
| description |  |
| created_at |  |

---

## 🏗️ Этапы разработки (Agile Sprints)

### **Sprint 1: Фундамент (1–1.5 недели)**

**Цель:** Базовая структура проекта + Auth

**Задачи:**

- Инициализация проекта (Clean Architecture структура)
    
    ```
    /cmd, /internal, /pkg, /migrations
    
    ```
    
- Подключение PostgreSQL + миграции (golang-migrate)
- Настройка env конфигов
- Auth система:
    - Регистрация / логин (email/password)
    - JWT tokens (access + refresh)
    - Middleware для проверки токенов и ролей
    - OAuth Google / Apple
- CRUD для Users

**API endpoints:**

```
POST /api/auth/register
POST /api/auth/login
POST /api/auth/refresh
POST /api/auth/google
POST /api/auth/apple
GET  /api/auth/me

```

---

### **Sprint 2: Рестораны и столики (1 неделя)**

**Цель:** Владелец может создать ресторан и добавить столики

**Задачи:**

- CRUD для Restaurants (только Owner)
- Upload фото в Cloudinary
- CRUD для Tables
- График работы (validation)
- Добавление менеджеров к ресторану

**API endpoints:**

```
POST   /api/restaurants
GET    /api/restaurants
GET    /api/restaurants/:id
PUT    /api/restaurants/:id
DELETE /api/restaurants/:id

POST   /api/restaurants/:id/tables
GET    /api/restaurants/:id/tables
PUT    /api/restaurants/:id/tables/:table_id
DELETE /api/restaurants/:id/tables/:table_id

POST   /api/restaurants/:id/managers
DELETE /api/restaurants/:id/managers/:user_id

```

---

### **Sprint 3: Система бронирования (1.5 недели)**

**Цель:** Клиент может забронировать столик

**Задачи:**

- Проверка доступности столиков по времени
- Логика бронирования:
    - Стандартная (2 часа, 500тг)
    - Корпоративная (больше времени, 1500тг)
- Валидация: минимум 30 мин до брони
- Автоматическое назначение столиков или выбор клиентом
- Статусы брони
- Отмена / изменение (до 1 часа)

**API endpoints:**

```
GET  /api/restaurants/:id/availability?date=&time=&guests=
POST /api/bookings
GET  /api/bookings
GET  /api/bookings/:id
PUT  /api/bookings/:id
DELETE /api/bookings/:id

GET  /api/restaurants/:id/bookings
PUT  /api/restaurants/:id/bookings/:booking_id

```

---

### **Sprint 4: Платежная система (1 неделя)**

**Цель:** Оплата депозита за бронь

**Задачи:**

- Система кошельков (Wallets)
- Пополнение / вывод денег
- Списание депозита при брони
- Интеграция Halyk / Kaspi (mock или базовая)
- Логика возврата:
    - Отмена >1 часа = возврат
    - Отмена <1 часа = сгорает
    - No-show = сгорает

**API endpoints:**

```
GET  /api/wallet
POST /api/wallet/deposit
POST /api/wallet/withdraw
GET  /api/wallet/transactions

POST /api/payments/halyk
POST /api/payments/kaspi

```

---

### **Sprint 5: Отзывы и рейтинги (3–4 дня)**

**Цель:** Клиенты могут оставлять отзывы

**Задачи:**

- Создание отзыва (только после завершённой брони)
- Расчёт среднего рейтинга ресторана
- Список отзывов ресторана

**API endpoints:**

```
POST /api/restaurants/:id/reviews
GET  /api/restaurants/:id/reviews

```

---

### **Sprint 6: Audit Logs и история (2–3 дня)**

**Цель:** Логирование всех изменений

**Логируем:**

- Создание / изменение / отмена брони (кто, когда, что изменил)
- Изменение данных ресторана
- Изменение столиков
- Добавление / удаление менеджеров

**API endpoints:**

```
GET /api/restaurants/:id/audit-logs

```

---

### **Sprint 7: Доп. функционал (если успеем)**

**Опционально:**

- Предзаказ блюд после брони
- Система напоминаний (cron job)
- Расширенная фильтрация
- Статистика для владельцев

---

## 🛠️ Технологический стек

**Backend:**

- Go 1.23
- Gin / Echo (REST API framework)
- GORM / sqlx (ORM)
- PostgreSQL
- JWT (golang-jwt/jwt)
- OAuth2 (golang.org/x/oauth2)
- Cloudinary SDK
- golang-migrate (миграции)

**Инфраструктура:**

- Docker + Docker Compose
- .env для конфигов

---

## 📁 Структура проекта

```
restaurant-booking/
├── cmd/
│   └── api/
│       └── main.go
├── internal/
│   ├── config/
│   ├── domain/          # entities
│   ├── repository/      # database layer
│   ├── service/         # business logic
│   ├── handler/         # http handlers
│   ├── middleware/      # auth, logging
│   └── utils/
├── pkg/
│   ├── jwt/
│   ├── cloudinary/
│   └── oauth/
├── migrations/
├── .env.example
├── docker-compose.yml
├── go.mod
└── go.sum

```
