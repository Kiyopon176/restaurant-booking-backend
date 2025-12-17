package domain

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID          uuid.UUID  `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	Email       string     `gorm:"unique;not null" json:"email"`
	Password    string     `gorm:"not null" json:"-"`
	FirstName   string     `gorm:"not null" json:"first_name"`
	LastName    string     `gorm:"not null" json:"last_name"`
	Phone       string     `gorm:"unique" json:"phone"`
	Role        UserRole   `gorm:"type:user_role;not null;default:'customer'" json:"role"`
	Avatar      *string    `json:"avatar,omitempty"`
	IsActive    bool       `gorm:"default:true" json:"is_active"`
	LastLoginAt *time.Time `json:"last_login_at,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`

	OwnedRestaurants   []Restaurant        `gorm:"foreignKey:OwnerID" json:"owned_restaurants,omitempty"`
	ManagedRestaurants []RestaurantManager `gorm:"foreignKey:UserID" json:"managed_restaurants,omitempty"`
	Bookings           []Booking           `gorm:"foreignKey:UserID" json:"bookings,omitempty"`
	Reviews            []Review            `gorm:"foreignKey:UserID" json:"reviews,omitempty"`
}

type UserRole string

const (
	UserRoleCustomer UserRole = "customer"
	UserRoleOwner    UserRole = "owner"
	UserRoleManager  UserRole = "manager"
	UserRoleAdmin    UserRole = "admin"
)

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
