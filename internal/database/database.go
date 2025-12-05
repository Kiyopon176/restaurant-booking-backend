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

	// Create custom enum types if they don't exist
	if err := createEnumTypes(db); err != nil {
		return nil, fmt.Errorf("failed to create enum types: %w", err)
	}

	// Auto migrate tables
	if err := db.AutoMigrate(
		&domain.User{},
		&domain.RefreshToken{},
		&domain.Restaurant{},
		&domain.RestaurantImage{},
		&domain.RestaurantManager{},
		&domain.Table{},
		&domain.Booking{},
		&domain.Review{},
	); err != nil {
		return nil, fmt.Errorf("failed to migrate database: %w", err)
	}

	log.Println("Database connected and migrated successfully")
	return db, nil
}

func createEnumTypes(db *gorm.DB) error {
	enumTypes := []string{
		`DO $$ BEGIN
			CREATE TYPE user_role AS ENUM ('customer', 'owner', 'manager', 'admin');
		EXCEPTION
			WHEN duplicate_object THEN null;
		END $$;`,
		`DO $$ BEGIN
			CREATE TYPE booking_status AS ENUM ('pending', 'confirmed', 'cancelled', 'completed', 'no_show');
		EXCEPTION
			WHEN duplicate_object THEN null;
		END $$;`,
		`DO $$ BEGIN
			CREATE TYPE cuisine_type AS ENUM ('Italian', 'Chinese', 'Mexican', 'Japanese', 'Indian', 'French', 'Kazakh', 'Turkish', 'Thai', 'American', 'Korean', 'Cafe', 'Bar', 'Fast Food', 'Vegetarian', 'Other');
		EXCEPTION
			WHEN duplicate_object THEN null;
		END $$;`,
		`DO $$ BEGIN
			CREATE TYPE location_type AS ENUM ('window', 'vip', 'regular', 'outdoor');
		EXCEPTION
			WHEN duplicate_object THEN null;
		END $$;`,
	}

	for _, enumSQL := range enumTypes {
		if err := db.Exec(enumSQL).Error; err != nil {
			return err
		}
	}

	return nil
}
