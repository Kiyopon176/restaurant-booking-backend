package domain

import (
	"time"

	"github.com/google/uuid"
)

type Table struct {
	ID           uuid.UUID    `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	RestaurantID uuid.UUID    `gorm:"type:uuid;not null" json:"restaurant_id"`
	TableNumber  string       `gorm:"not null" json:"table_number"`
	MinCapacity  int          `gorm:"not null" json:"min_capacity"`
	MaxCapacity  int          `gorm:"not null" json:"max_capacity"`
	LocationType LocationType `gorm:"type:location_type;not null;default:'regular'" json:"location_type"`
	XPosition    *int         `json:"x_position,omitempty"`
	YPosition    *int         `json:"y_position,omitempty"`
	IsActive     bool         `gorm:"default:true" json:"is_active"`
	CreatedAt    time.Time    `json:"created_at"`
	UpdatedAt    time.Time    `json:"updated_at"`
}

type LocationType string

const (
	LocationWindow  LocationType = "window"
	LocationVIP     LocationType = "vip"
	LocationRegular LocationType = "regular"
	LocationOutdoor LocationType = "outdoor"
)

type RestaurantManager struct {
	ID           uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	UserID       uuid.UUID `gorm:"type:uuid;not null" json:"user_id"`
	RestaurantID uuid.UUID `gorm:"type:uuid;not null" json:"restaurant_id"`
	AssignedAt   time.Time `json:"assigned_at"`

	User       *User       `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Restaurant *Restaurant `gorm:"foreignKey:RestaurantID" json:"restaurant,omitempty"`
}
