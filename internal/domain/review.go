package domain

import "time"

type Review struct {
	ID           int64     `json:"id" db:"id"`
	RestaurantID int64     `json:"restaurant_id" db:"restaurant_id"`
	UserID       int64     `json:"user_id" db:"user_id"`
	Rating       int       `json:"rating" db:"rating"`
	Comment      string    `json:"comment" db:"comment"`
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time `json:"updated_at" db:"updated_at"`
}
