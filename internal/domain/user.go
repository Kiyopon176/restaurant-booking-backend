package domain

import (
	"time"

	"github.com/google/uuid"
)

type UserRole string

const (
	RoleClient  UserRole = "client"
	RoleOwner   UserRole = "owner"
	RoleManager UserRole = "manager"
	RoleAdmin   UserRole = "admin"
)

type OAuthProvider string

const (
	OAuthGoogle OAuthProvider = "google"
	OAuthApple  OAuthProvider = "apple"
)

type User struct {
	ID            uuid.UUID      `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	Email         string         `gorm:"uniqueIndex;not null" json:"email"`
	PasswordHash  string         `gorm:"column:password_hash" json:"-"`
	Name          string         `gorm:"not null" json:"name"`
	Phone         *string        `json:"phone,omitempty"`
	Role          UserRole       `gorm:"type:user_role;not null;default:'client'" json:"role"`
	OAuthProvider *OAuthProvider `gorm:"type:oauth_provider" json:"oauth_provider,omitempty"`
	OAuthID       *string        `gorm:"column:oauth_id" json:"-"`
	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at"`
}

func (User) TableName() string {
	return "users"
}

type RefreshToken struct {
	ID        uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	UserID    uuid.UUID `gorm:"type:uuid;not null" json:"user_id"`
	Token     string    `gorm:"uniqueIndex;not null" json:"token"`
	ExpiresAt time.Time `gorm:"not null" json:"expires_at"`
	CreatedAt time.Time `json:"created_at"`
}

func (RefreshToken) TableName() string {
	return "refresh_tokens"
}
