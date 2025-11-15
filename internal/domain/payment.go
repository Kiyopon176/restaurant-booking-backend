package domain

import (
	"time"

	"github.com/google/uuid"
)

type Payment struct {
	ID                 uuid.UUID     `db:"id" json:"id"`
	UserID             uuid.UUID     `db:"user_id" json:"user_id"`
	BookingID          *uuid.UUID    `db:"booking_id" json:"booking_id,omitempty"`
	Amount             int           `db:"amount" json:"amount"`
	PaymentMethod      PaymentMethod `db:"payment_method" json:"payment_method"`
	PaymentStatus      PaymentStatus `db:"payment_status" json:"payment_status"`
	ExternalPaymentID  *string       `db:"external_payment_id" json:"external_payment_id,omitempty"`
	ExternalPaymentURL *string       `db:"external_payment_url" json:"external_payment_url,omitempty"`
	ErrorMessage       *string       `db:"error_message" json:"error_message,omitempty"`
	CreatedAt          time.Time     `db:"created_at" json:"created_at"`
	UpdatedAt          time.Time     `db:"updated_at" json:"updated_at"`

	User *User `json:"user,omitempty"`
}

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
