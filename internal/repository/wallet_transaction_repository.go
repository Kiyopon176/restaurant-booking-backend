package repository

import (
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"

	"github.com/Kiyopon176/restaurant-booking-backend/internal/domain"
)

type WalletTransactionRepository interface {
	Create(transaction *domain.WalletTransaction) error
	GetByWalletID(walletID uuid.UUID, pagination Pagination) ([]domain.WalletTransaction, int64, error)
	GetByBookingID(bookingID uuid.UUID) ([]domain.WalletTransaction, error)
}

type walletTransactionRepo struct{ db *sqlx.DB }

func NewWalletTransactionRepository(db *sqlx.DB) WalletTransactionRepository {
	return &walletTransactionRepo{db}
}

func (r *walletTransactionRepo) Create(transaction *domain.WalletTransaction) error {
	q := `
		INSERT INTO wallet_transactions (
			id, wallet_id, amount, type, description, booking_id, created_at
		)
		VALUES (
			:id, :wallet_id, :amount, :type, :description, :booking_id, NOW()
		)
	`
	_, err := sqlx.NamedExec(r.db, q, transaction)
	return err
}

func (r *walletTransactionRepo) GetByWalletID(walletID uuid.UUID, pagination Pagination) ([]domain.WalletTransaction, int64, error) {
	var list []domain.WalletTransaction
	var total int64
	q := `SELECT COUNT(*) FROM wallet_transactions WHERE wallet_id = $1`
	err := r.db.Get(&total, q, walletID)
	if err != nil {
		return nil, 0, err
	}

	q = `SELECT * FROM wallet_transactions WHERE wallet_id = $1 ORDER BY created_at DESC LIMIT $2 OFFSET $3`
	err = r.db.Select(&list, q, walletID, pagination.Limit(), pagination.Offset())
	if err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

func (r *walletTransactionRepo) GetByBookingID(bookingID uuid.UUID) ([]domain.WalletTransaction, error) {
	var list []domain.WalletTransaction
	q := `SELECT * FROM wallet_transactions WHERE booking_id = $1`
	err := r.db.Select(&list, q, bookingID)
	if err != nil {
		return nil, err
	}
	return list, nil
}
