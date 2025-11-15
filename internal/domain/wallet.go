package domain

import (
	"time"

	"github.com/google/uuid"
)

type Wallet struct {
	ID        uuid.UUID `db:"id" json:"id"`
	UserID    uuid.UUID `db:"user_id" json:"user_id"`
	Balance   int       `db:"balance" json:"balance"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`
	UpdatedAt time.Time `db:"updated_at" json:"updated_at"`

	User         *User               `json:"user,omitempty"`
	Transactions []WalletTransaction `json:"transactions,omitempty"`
}

type WalletTransaction struct {
	ID          uuid.UUID       `db:"id" json:"id"`
	WalletID    uuid.UUID       `db:"wallet_id" json:"wallet_id"`
	Amount      int             `db:"amount" json:"amount"`
	Type        TransactionType `db:"type" json:"type"`
	Description string
