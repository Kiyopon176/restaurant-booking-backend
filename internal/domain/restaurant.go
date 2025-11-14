package domain

import "time"

type Restaurant struct {
	ID          int64     `json:"id" db:"id"`
	OwnerID     int64     `json:"owner_id" db:"owner_id"`
	Name        string    `json:"name" db:"name"`
	Description string    `json:"description" db:"description"`
	City        string    `json:"city" db:"city"`
	Address     string    `json:"address" db:"address"`
	Longitude   string    `json:"longitude" db:"longitude"`
	Latitude    string    `json:"latitude" db:"latitude"`
	Phone       string    `json:"phone" db:"phone"`
	Rating      float32   `json:"rating" db:"rating"`
	Instagram   string    `json:"instagram" db:"instagram"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
}
