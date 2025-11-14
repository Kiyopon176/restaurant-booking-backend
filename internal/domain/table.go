package domain

import "time"

type Table struct {
	ID            int64     `json:"id" db:"id"`
	RestaurantID  int64     `json:"restaurant_id" db:"restaurant_id"`
	TableNumber   int       `json:"table_number" db:"table_number"`
	MinCapacity   int       `json:"min_capacity" db:"min_capacity"`
	MaxCapacity   int       `json:"max_capacity" db:"max_capacity"`
	LocationType  string    `json:"location_type" db:"location_type"`
	IsAvailable   bool      `json:"is_available" db:"is_available"`
	IsActive      bool      `json:"is_active" db:"is_active"`
	MaxCombinable int       `json:"max_combinable_tables" db:"max_combinable_tables"`
	Notes         string    `json:"notes" db:"notes"`
	CreatedAt     time.Time `json:"created_at" db:"created_at"`
	UpdatedAt     time.Time `json:"updated_at" db:"updated_at"`
}
