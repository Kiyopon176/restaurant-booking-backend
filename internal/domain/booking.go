package domain

import (
	"time"

	"github.com/google/uuid"
)

type Booking struct {
	ID           uuid.UUID     `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	RestaurantID uuid.UUID     `gorm:"type:uuid;not null" json:"restaurant_id"`
	TableID      uuid.UUID     `gorm:"type:uuid;not null" json:"table_id"`
	UserID       uuid.UUID     `gorm:"type:uuid;not null" json:"user_id"`
	BookingDate  time.Time     `gorm:"not null" json:"booking_date"`
	StartTime    time.Time     `gorm:"not null" json:"start_time"`
	EndTime      time.Time     `gorm:"not null" json:"end_time"`
	GuestsCount  int           `gorm:"not null" json:"guests_count"`
	Status       BookingStatus `gorm:"type:booking_status;not null;default:'pending'" json:"status"`
	SpecialNote  string        `gorm:"type:text" json:"special_note,omitempty"`
	CreatedAt    time.Time     `json:"created_at"`
	UpdatedAt    time.Time     `json:"updated_at"`

	Restaurant *Restaurant `gorm:"foreignKey:RestaurantID" json:"restaurant,omitempty"`
	Table      *Table      `gorm:"foreignKey:TableID" json:"table,omitempty"`
	User       *User       `gorm:"foreignKey:UserID" json:"user,omitempty"`
}

type BookingStatus string

const (
	BookingStatusPending   BookingStatus = "pending"
	BookingStatusConfirmed BookingStatus = "confirmed"
	BookingStatusCancelled BookingStatus = "cancelled"
	BookingStatusCompleted BookingStatus = "completed"
	BookingStatusNoShow    BookingStatus = "no_show"
)
