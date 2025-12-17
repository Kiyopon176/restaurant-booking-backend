# Руководство по использованию Горутин и Каналов в проекте

## 📚 Содержание

1. [Notification Service](#1-notification-service)
2. [Background Cleaner](#2-background-cleaner)
3. [Booking Service](#3-booking-service)
4. [API Endpoints](#4-api-endpoints)
5. [Примеры использования](#5-примеры-использования)

---

## 1. Notification Service

### Описание
Сервис для асинхронной отправки уведомлений с использованием worker pool паттерна.

### Особенности
- ✅ **Worker Pool**: 5 воркеров обрабатывают уведомления параллельно
- ✅ **Channel Buffer**: Буфер на 100 уведомлений
- ✅ **Graceful Shutdown**: Корректное завершение всех горутин
- ✅ **Statistics**: Отслеживание успешных и неудачных отправок

### Файл
`internal/service/notification_service.go`

### Ключевые элементы конкурентности

```go
// Канал для уведомлений
notifications chan Notification

// Worker pool - 5 горутин обрабатывают уведомления
for i := 0; i < workers; i++ {
    wg.Add(1)
    go ns.worker(i)
}

// Context для graceful shutdown
ctx, cancel := context.WithCancel(context.Background())
```

### Пример использования

```go
// Инициализация
notificationSvc := service.NewNotificationService(5, 100)

// Отправка одного уведомления
notificationSvc.SendEmail(
    "user@example.com",
    "Welcome!",
    "Thanks for joining",
)

// Массовая отправка
notifications := []service.Notification{...}
notificationSvc.SendBulk(notifications)

// Получение статистики
sent, failed := notificationSvc.GetStats()

// Завершение работы
notificationSvc.Shutdown()
```

---

## 2. Background Cleaner

### Описание
Фоновая служба для периодической очистки данных с использованием горутин.

### Особенности
- ✅ **Periodic Execution**: Запуск задач по расписанию (каждый час)
- ✅ **Parallel Tasks**: Несколько задач выполняются параллельно
- ✅ **Done Channel**: Сигнализация о завершении работы
- ✅ **Task Scheduler**: Гибкая система планирования задач

### Файл
`internal/service/background_cleaner.go`

### Ключевые элементы конкурентности

```go
// Ticker для периодического выполнения
ticker := time.NewTicker(bc.interval)

// Done channel для завершения
done chan struct{}

// Параллельное выполнение задач
tasksDone := make(chan string, 3)
go func() { /* Task 1 */ tasksDone <- "tokens:success" }()
go func() { /* Task 2 */ tasksDone <- "notifications:success" }()
go func() { /* Task 3 */ tasksDone <- "bookings:success" }()
```

### Пример использования

```go
// Инициализация
cleaner := service.NewBackgroundCleaner(
    refreshTokenRepo,
    1*time.Hour, // каждый час
)

// Запуск фоновой горутины
cleaner.Start()

// Ручной запуск очистки
cleaner.RunNow()

// Остановка
cleaner.Stop()
```

---

## 3. Booking Service

### Описание
Сервис бронирования с параллельной обработкой операций.

### Особенности
- ✅ **Parallel Availability Check**: Проверка доступности нескольких столиков одновременно
- ✅ **Concurrent Search**: Поиск по нескольким ресторанам параллельно
- ✅ **Rate Limiting**: Ограничение количества одновременных операций
- ✅ **WaitGroup**: Синхронизация завершения горутин

### Файл
`internal/service/booking_service.go`

### Ключевые элементы конкурентности

```go
// WaitGroup для синхронизации
var wg sync.WaitGroup

// Channel для результатов
resultsChan := make(chan Result, len(items))

// Semaphore для rate limiting
semaphore := make(chan struct{}, maxConcurrent)

// Параллельная обработка
for _, item := range items {
    wg.Add(1)
    go func(i Item) {
        defer wg.Done()
        // обработка
    }(item)
}
wg.Wait()
```

### Примеры использования

```go
// 1. Проверка доступности нескольких столиков
tableIDs := []uuid.UUID{id1, id2, id3}
results := bookingSvc.CheckMultipleTablesAvailability(
    ctx,
    tableIDs,
    startTime,
    endTime,
)
// Результат: map[uuid.UUID]bool

// 2. Поиск по нескольким ресторанам
restaurantIDs := []uuid.UUID{r1, r2, r3}
tables := bookingSvc.SearchAvailableTablesParallel(
    ctx,
    restaurantIDs,
    startTime,
    endTime,
    guestCount,
)
// Результат: map[uuid.UUID][]uuid.UUID

// 3. Массовая обработка бронирований
bookings := []domain.Booking{...}
results := bookingSvc.ProcessBulkBookings(
    ctx,
    bookings,
    10, // максимум 10 одновременно
)

// 4. Получение статистики параллельно
stats, err := bookingSvc.GetBookingStatistics(ctx, restaurantID)
// Результат: map[string]int
```

---

## 4. API Endpoints

### Demo Routes (примеры работы с конкурентностью)

#### 1. Массовая отправка уведомлений
```bash
POST /api/demo/bulk-notifications
Content-Type: application/json

{
  "recipients": [
    "user1@example.com",
    "user2@example.com",
    "user3@example.com"
  ],
  "subject": "Special Offer",
  "message": "Check out our new menu!"
}
```

#### 2. Статистика уведомлений
```bash
GET /api/demo/notification-stats

Response:
{
  "sent": 150,
  "failed": 5
}
```

#### 3. Проверка доступности столиков
```bash
POST /api/demo/check-availability
Content-Type: application/json

{
  "table_ids": [
    "uuid-1",
    "uuid-2",
    "uuid-3"
  ],
  "start_time": "2024-12-20T19:00:00Z",
  "end_time": "2024-12-20T21:00:00Z"
}

Response:
{
  "availability": {
    "uuid-1": true,
    "uuid-2": false,
    "uuid-3": true
  }
}
```

#### 4. Статистика бронирований
```bash
GET /api/demo/booking-stats/{restaurant_id}

Response:
{
  "restaurant_id": "uuid",
  "stats": {
    "total_bookings": 150,
    "active_bookings": 25,
    "completed_bookings": 100,
    "cancelled_bookings": 25
  }
}
```

#### 5. Поиск доступных столиков
```bash
POST /api/demo/search-tables
Content-Type: application/json

{
  "restaurant_ids": ["uuid-1", "uuid-2"],
  "start_time": "2024-12-20T19:00:00Z",
  "end_time": "2024-12-20T21:00:00Z",
  "guest_count": 4
}

Response:
{
  "results": {
    "uuid-1": ["table-1", "table-2"],
    "uuid-2": ["table-3"]
  }
}
```

---

## 5. Примеры использования

### Запуск сервера

```bash
go run ./cmd/api/main.go
```

При запуске вы увидите:

```
Setting up concurrent services...
Notification service started with 5 workers
Notification worker 0 started
Notification worker 1 started
...
Starting background cleaner with interval: 1h0m0s
✅ All concurrent services initialized successfully

🎯 Demonstrating concurrent features...
📧 Demo 1: Sending bulk notifications...
Worker 0: Successfully sent email notification to user1@example.com
...
```

### Тестирование с curl

```bash
# 1. Отправить массовые уведомления
curl -X POST http://localhost:8080/api/demo/bulk-notifications \
  -H "Content-Type: application/json" \
  -d '{
    "recipients": ["test1@example.com", "test2@example.com"],
    "subject": "Test",
    "message": "Hello from concurrent service!"
  }'

# 2. Получить статистику
curl http://localhost:8080/api/demo/notification-stats

# 3. Проверить доступность
curl -X POST http://localhost:8080/api/demo/check-availability \
  -H "Content-Type: application/json" \
  -d '{
    "table_ids": ["550e8400-e29b-41d4-a716-446655440000"],
    "start_time": "2024-12-20T19:00:00Z",
    "end_time": "2024-12-20T21:00:00Z"
  }'
```

---

## 🔧 Ключевые паттерны конкурентности

### 1. Worker Pool Pattern
```go
// Создание worker pool
for i := 0; i < numWorkers; i++ {
    wg.Add(1)
    go worker(i, jobs, results)
}
```

### 2. Fan-Out Pattern
```go
// Распределение работы по горутинам
for _, task := range tasks {
    go processTask(task, resultChan)
}
```

### 3. Fan-In Pattern
```go
// Сбор результатов из нескольких каналов
for i := 0; i < numWorkers; i++ {
    result := <-resultChan
    results = append(results, result)
}
```

### 4. Rate Limiting
```go
// Ограничение количества одновременных горутин
semaphore := make(chan struct{}, maxConcurrent)
for _, item := range items {
    semaphore <- struct{}{}
    go func() {
        defer func() { <-semaphore }()
        process(item)
    }()
}
```

### 5. Context для отмены
```go
ctx, cancel := context.WithCancel(context.Background())
defer cancel()

select {
case <-ctx.Done():
    return ctx.Err()
case result := <-resultChan:
    return result
}
```

---

## 📊 Мониторинг и логирование

Все concurrent операции логируются:

```
2024/12/16 22:00:00 Notification worker 0 started
2024/12/16 22:00:01 Notification uuid queued for sending
2024/12/16 22:00:01 Worker 0: Successfully sent email notification
2024/12/16 22:00:05 Running cleanup tasks...
2024/12/16 22:00:05 Cleanup task completed: tokens:success
2024/12/16 22:00:10 📊 Notification Stats - Sent: 150, Failed: 5
```

---

## 🛑 Graceful Shutdown

Приложение корректно завершает все горутины при получении сигнала:

```go
// Ctrl+C или SIGTERM
^C
🛑 Received signal: interrupt. Starting graceful shutdown...
Stopping task scheduler...
Stopping background cleaner...
Stopping notification service...
Notification service shutdown complete. Sent: 150, Failed: 5
✅ All services stopped gracefully
```

---

## 📝 Резюме

В проекте реализованы следующие concurrent features:

1. ✅ **Worker Pool** - для обработки уведомлений
2. ✅ **Background Jobs** - для периодических задач
3. ✅ **Parallel Processing** - для одновременной обработки данных
4. ✅ **Rate Limiting** - для контроля нагрузки
5. ✅ **Graceful Shutdown** - для корректного завершения
6. ✅ **Channels** - для коммуникации между горутинами
7. ✅ **Context** - для отмены операций
8. ✅ **WaitGroup** - для синхронизации
9. ✅ **Mutex** - для защиты shared state

Все это позволяет эффективно использовать многоядерные процессоры и обрабатывать множество запросов одновременно!
