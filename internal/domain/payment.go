// internal/domain/payment.go
package domain

import (
	"time"

	"github.com/google/uuid"
)

type PaymentMethod string

const (
	PaymentMethodWallet PaymentMethod = "wallet"
	PaymentMethodHalyk  PaymentMethod = "halyk"
	PaymentMethodKaspi  PaymentMethod = "kaspi"
)

type PaymentStatus string

const (
	PaymentStatusPending   PaymentStatus = "pending"
	PaymentStatusCompleted PaymentStatus = "completed"
	PaymentStatusFailed    PaymentStatus = "failed"
	PaymentStatusRefunded  PaymentStatus = "refunded"
)

type Payment struct {
	ID                 uuid.UUID     `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	UserID             uuid.UUID     `gorm:"type:uuid;not null" json:"user_id"`
	BookingID          *uuid.UUID    `gorm:"type:uuid" json:"booking_id,omitempty"`
	Amount             int           `gorm:"not null" json:"amount"`
	PaymentMethod      PaymentMethod `gorm:"type:payment_method;not null" json:"payment_method"`
	PaymentStatus      PaymentStatus `gorm:"type:payment_status;not null;default:'pending'" json:"payment_status"`
	ExternalPaymentID  *string       `gorm:"type:varchar(255)" json:"external_payment_id,omitempty"`
	ExternalPaymentURL *string       `gorm:"type:text" json:"external_payment_url,omitempty"`
	ErrorMessage       *string       `gorm:"type:text" json:"error_message,omitempty"`
	CreatedAt          time.Time     `json:"created_at"`
	UpdatedAt          time.Time     `json:"updated_at"`

	User    *User    `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Booking *Booking `gorm:"foreignKey:BookingID" json:"booking,omitempty"`
}
