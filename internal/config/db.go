package config

import (
	"fmt"
	"log"
	_ "os"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func ConnectDB() *gorm.DB {
	host := "localhost"
	port := "5432"
	user := "postgres"
	password := "0806"
	dbname := "dastarkhan_db"

	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	return db
}
