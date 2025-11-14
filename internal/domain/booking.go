package domain

import "time"

type Booking struct {
	ID           int64     `json:"id" db:"id"`
	RestaurantID int64     `json:"restaurant_id" db:"restaurant_id"`
	TableID      int64     `json:"table_id" db:"table_id"`
	UserID       int64     `json:"user_id" db:"user_id"`
	BookingDate  time.Time `json:"booking_date" db:"booking_date"`
	StartTime    time.Time `json:"start_time" db:"start_time"`
	EndTime      time.Time `json:"end_time" db:"end_time"`
	CountGuests  int       `json:"count_guests" db:"count_guests"`
	Status       string    `json:"status" db:"status"`
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time `json:"updated_at" db:"updated_at"`
}
