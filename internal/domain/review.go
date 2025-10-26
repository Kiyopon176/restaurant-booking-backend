package domain

import (
	"time"

	"github.com/google/uuid"
)

type Review struct {
	ID           uuid.UUID  `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	RestaurantID uuid.UUID  `gorm:"type:uuid;not null" json:"restaurant_id"`
	UserID       uuid.UUID  `gorm:"type:uuid;not null" json:"user_id"`
	BookingID    *uuid.UUID `gorm:"type:uuid" json:"booking_id,omitempty"`
	Rating       int        `gorm:"not null;check:rating >= 1 AND rating <= 5" json:"rating"`
	Comment      string     `gorm:"type:text" json:"comment"`
	IsVisible    bool       `gorm:"default:true" json:"is_visible"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`

	Restaurant *Restaurant `gorm:"foreignKey:RestaurantID" json:"restaurant,omitempty"`
	User       *User       `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Booking    *Booking    `gorm:"foreignKey:BookingID" json:"booking,omitempty"`
}
