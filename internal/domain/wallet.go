package domain

import (
	"time"

	"github.com/google/uuid"
)

type Wallet struct {
	ID        uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	UserID    uuid.UUID `gorm:"type:uuid;uniqueIndex;not null" json:"user_id"`
	Balance   int       `gorm:"not null;default:0;check:balance >= 0" json:"balance"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	User         *User               `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Transactions []WalletTransaction `gorm:"foreignKey:WalletID" json:"transactions,omitempty"`
}

type WalletTransaction struct {
	ID          uuid.UUID       `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	WalletID    uuid.UUID       `gorm:"type:uuid;not null" json:"wallet_id"`
	Amount      int             `gorm:"not null" json:"amount"`
	Type        TransactionType `gorm:"type:transaction_type;not null" json:"type"`
	Description string          `gorm:"type:text" json:"description"`
	BookingID   *uuid.UUID      `gorm:"type:uuid" json:"booking_id,omitempty"`
	CreatedAt   time.Time       `json:"created_at"`
}

type TransactionType string

const (
	TransactionDeposit             TransactionType = "deposit"
	TransactionWithdraw            TransactionType = "withdraw"
	TransactionBookingCharge       TransactionType = "booking_charge"
	TransactionRefund              TransactionType = "refund"
	TransactionPaymentToRestaurant TransactionType = "payment_to_restaurant"
)
