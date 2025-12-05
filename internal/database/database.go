package database

import (
	"fmt"
	"log"
	"restaurant-booking/internal/config"
	"restaurant-booking/internal/domain"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func InitDB(cfg *config.Config) (*gorm.DB, error) {
	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		cfg.DBHost, cfg.DBPort, cfg.DBUser, cfg.DBPassword, cfg.DBName)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Configure connection pool
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get database instance: %w", err)
	}

	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)

	// Auto migrate tables
	// Ensure required extensions and enum types exist
	if err := db.Exec("CREATE EXTENSION IF NOT EXISTS pgcrypto;").Error; err != nil {
		return nil, fmt.Errorf("failed to ensure pgcrypto extension: %w", err)
	}

	if err := db.Exec(`DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'user_role') THEN
        CREATE TYPE user_role AS ENUM ('customer','owner','manager','admin');
    END IF;
END$$;`).Error; err != nil {
		return nil, fmt.Errorf("failed to ensure user_role type: %w", err)
	}

	// Additional enum types used by domain models
	if err := db.Exec(`DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'cuisine_type') THEN
        CREATE TYPE cuisine_type AS ENUM (
            'Italian','Chinese','Mexican','Japanese','Indian',
            'French','Kazakh','Turkish','Thai','American',
            'Korean','Cafe','Bar','Fast Food','Vegetarian','Other'
        );
    END IF;
END$$;`).Error; err != nil {
		return nil, fmt.Errorf("failed to ensure cuisine_type type: %w", err)
	}

	if err := db.Exec(`DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'booking_status') THEN
        CREATE TYPE booking_status AS ENUM ('pending','confirmed','cancelled','completed','no_show');
    END IF;
END$$;`).Error; err != nil {
		return nil, fmt.Errorf("failed to ensure booking_status type: %w", err)
	}

	if err := db.Exec(`DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'location_type') THEN
        CREATE TYPE location_type AS ENUM ('window','vip','regular','outdoor');
    END IF;
END$$;`).Error; err != nil {
		return nil, fmt.Errorf("failed to ensure location_type type: %w", err)
	}

	// Auto migrate tables
	if err := db.AutoMigrate(
		&domain.User{},
		&domain.RefreshToken{},
		&domain.Restaurant{},
		&domain.Table{},
		&domain.Booking{},
		&domain.Review{},
	); err != nil {
		return nil, fmt.Errorf("failed to migrate database: %w", err)
	}

	log.Println("Database connected and migrated successfully")
	return db, nil
}
