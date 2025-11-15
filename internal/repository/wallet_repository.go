package repository

import (
	"errors"

	"github.com/Kiyopon176/restaurant-booking-backend/internal/domain"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

type WalletRepository interface {
	Create(tx *sqlx.Tx, wallet *domain.Wallet) error
	GetByUserID(tx *sqlx.Tx, userID uuid.UUID) (*domain.Wallet, error)
	UpdateBalance(tx *sqlx.Tx, walletID uuid.UUID, newBalance int) error
	GetByID(tx *sqlx.Tx, id uuid.UUID) (*domain.Wallet, error)
}

type walletRepo struct {
	db *sqlx.DB
}

func NewWalletRepository(db *sqlx.DB) WalletRepository {
	return &walletRepo{db: db}
}

func (r *walletRepo) exec(tx *sqlx.Tx) sqlx.Ext {
	if tx != nil {
		return tx
	}
	return r.db
}

func (r *walletRepo) Create(tx *sqlx.Tx, wallet *domain.Wallet) error {
	q := `
		INSERT INTO wallets (
			id, user_id, balance, created_at, updated_at
		)
		VALUES (
			:id, :user_id, :balance, NOW(), NOW()
		)
	`
	_, err := sqlx.NamedExec(r.exec(tx), q, wallet)
	return err
}

func (r *walletRepo) GetByUserID(tx *sqlx.Tx, userID uuid.UUID) (*domain.Wallet, error) {
	var w domain.Wallet
	q := `SELECT * FROM wallets WHERE user_id = $1 LIMIT 1`
	err := sqlx.Get(r.exec(tx), &w, q, userID)
	if err != nil {
		if errors.Is(err, sqlx.ErrNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &w, nil
}

func (r *walletRepo) UpdateBalance(tx *sqlx.Tx, walletID uuid.UUID, newBalance int) error {
	q := `
		UPDATE wallets
		SET balance = $1, updated_at = NOW()
		WHERE id = $2
	`
	_, err := r.exec(tx).Exec(q, newBalance, walletID)
	return err
}

func (r *walletRepo) GetByID(tx *sqlx.Tx, id uuid.UUID) (*domain.Wallet, error) {
	var w domain.Wallet
	q := `SELECT * FROM wallets WHERE id = $1 LIMIT 1`
	err := sqlx.Get(r.exec(tx), &w, q, id)
	if err != nil {
		return nil, err
	}
	return &w, nil
}
