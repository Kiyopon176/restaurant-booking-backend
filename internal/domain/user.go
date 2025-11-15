package domain

import (
	"github.com/google/uuid"
	"time"
)

type User struct {
	ID           uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	Username     string    `gorm:"not null" json:"username"`
	Email        string    `gorm:"not null" json:"email"`
	PasswordHash string    `gorm:"not null" json:"password_hash"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}
