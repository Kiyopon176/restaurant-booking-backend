package main

import (
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"

	"github.com/Kiyopon176/restaurant-booking-backend/internal/domain"
	"github.com/Kiyopon176/restaurant-booking-backend/internal/handler"
	"github.com/Kiyopon176/restaurant-booking-backend/internal/repository"
	"github.com/Kiyopon176/restaurant-booking-backend/internal/service"
	"github.com/joho/godotenv" // импортируем библиотеку для работы с .env
)

func main() {
	// Загружаем .env
	if err := godotenv.Load(); err != nil {
		log.Println(".env file not found or failed to load")
	}

	// Теперь считываем переменную окружения
	dsn := os.Getenv("DATABASE_DSN")
	if dsn == "" {
		log.Println("DATABASE_DSN env empty — using fallback DSN")
		dsn = "host=localhost user=postgres password=0806 dbname=postgres port=5432 sslmode=disable TimeZone=UTC"
	}

	// Открываем подключение к базе
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		NamingStrategy: schema.NamingStrategy{
			TablePrefix: "public.", // Все таблицы в public
		},
	})

	if err != nil {
		log.Fatalf("failed to connect db: %v", err)
	}

	// Создаём типы вручную (если не существуют)
	if err := db.Exec(`
		DO $$ BEGIN
			-- Попытка создать тип transaction_type
			CREATE TYPE transaction_type AS ENUM ('deposit', 'withdraw', 'booking_charge', 'refund', 'payment_to_restaurant');
		EXCEPTION
			WHEN duplicate_object THEN null; -- Если тип уже существует, пропускаем ошибку
		END $$;
	`).Error; err != nil {
		log.Fatalf("failed to create transaction_type: %v", err)
	}

	if err := db.Exec(`
		DO $$ BEGIN
			-- Попытка создать тип payment_method
			CREATE TYPE payment_method AS ENUM ('wallet', 'halyk', 'kaspi');
		EXCEPTION
			WHEN duplicate_object THEN null; -- Если тип уже существует, пропускаем ошибку
		END $$;
	`).Error; err != nil {
		log.Fatalf("failed to create payment_method type: %v", err)
	}

	if err := db.Exec(`
		DO $$ BEGIN
			-- Попытка создать тип payment_status
			CREATE TYPE payment_status AS ENUM ('pending', 'completed', 'failed', 'refunded');
		EXCEPTION
			WHEN duplicate_object THEN null; -- Если тип уже существует, пропускаем ошибку
		END $$;
	`).Error; err != nil {
		log.Fatalf("failed to create payment_status type: %v", err)
	}

	// auto-migrate domain tables for demo (in prod use migrations)
	if err := db.AutoMigrate(&domain.User{}, &domain.Wallet{}, &domain.WalletTransaction{}, &domain.Payment{}); err != nil {
		log.Fatalf("auto migrate failed: %v", err)
	}

	// Инициализация репозиториев, сервисов и хендлеров
	wr := repository.NewWalletRepository(db)
	pr := repository.NewPaymentRepository(db)

	ws := service.NewWalletService(db, wr)
	ps := service.NewPaymentService(db, pr, ws)

	wh := handler.NewWalletHandler(ws)
	ph := handler.NewPaymentHandler(ps)

	// Запуск HTTP сервера
	r := gin.Default()

	api := r.Group("/api")
	{
		w := api.Group("/wallet")
		w.GET("", wh.GetWallet)
		w.POST("/deposit", wh.Deposit)
		w.POST("/withdraw", wh.Withdraw)
	}

	p := api.Group("/payments")
	{
		p.POST("/wallet", ph.CreateWalletPayment)
		p.POST("/halyk", ph.CreateHalykPayment)
		p.POST("/webhook/halyk", ph.HalykWebhook)
		p.POST("/webhook/kaspi", ph.KaspiWebhook)
		p.POST("/:id/refund", ph.RefundPayment)
	}

	if err := r.Run(":8080"); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
