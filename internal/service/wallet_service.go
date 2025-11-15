package service

import (
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"

	"github.com/Kiyopon176/restaurant-booking-backend/internal/domain"
	"github.com/Kiyopon176/restaurant-booking-backend/internal/repository"
)

type WalletService struct {
	DB         *sqlx.DB
	WalletRepo repository.WalletRepository
	TxnRepo    repository.PaymentRepository // unused here but commonly needed
}

func NewWalletService(db *sqlx.DB, wr repository.WalletRepository) *WalletService {
	return &WalletService{
		DB:         db,
		WalletRepo: wr,
	}
}

func (s *WalletService) GetOrCreateWallet(userID uuid.UUID) (*domain.Wallet, error) {
	wallet, err := s.WalletRepo.GetByUserID(nil, userID)
	if err != nil {
		return nil, err
	}
	if wallet != nil {
		return wallet, nil
	}
	w := &domain.Wallet{UserID: userID, Balance: 0}
	if err := s.WalletRepo.Create(nil, w); err != nil {
		return nil, err
	}
	return w, nil
}

func (s *WalletService) GetBalance(userID uuid.UUID) (int, error) {
	w, err := s.WalletRepo.GetByUserID(nil, userID)
	if err != nil {
		return 0, err
	}
	if w == nil {
		return 0, nil
	}
	return w.Balance, nil
}

func (s *WalletService) Deposit(userID uuid.UUID, amount int, description string) error {
	if amount <= 0 {
		return errors.New("invalid_amount")
	}
	return s.DB.Transaction(func(tx *sqlx.Tx) error {
		var wallet domain.Wallet
		if err := tx.Get(&wallet, `SELECT * FROM wallets WHERE user_id = $1 FOR UPDATE`, userID); err != nil {
			if err.Error() == "sql: no rows in result set" {
				wallet = domain.Wallet{UserID: userID, Balance: 0}
				if err := tx.Exec(`INSERT INTO wallets (user_id, balance) VALUES ($1, $2)`, userID, wallet.Balance); err != nil {
					return err
				}
			} else {
				return err
			}
		}
		newBal := wallet.Balance + amount
		if err := tx.Exec(`UPDATE wallets SET balance = $1 WHERE id = $2`, newBal, wallet.ID); err != nil {
			return err
		}
		wt := domain.WalletTransaction{
			WalletID:    wallet.ID,
			Amount:      amount,
			Type:        domain.TransactionDeposit,
			Description: description,
		}
		if err := tx.Exec(`INSERT INTO wallet_transactions (wallet_id, amount, type, description) VALUES ($1, $2, $3, $4)`, wt.WalletID, wt.Amount, wt.Type, wt.Description); err != nil {
			return err
		}
		return nil
	})
}

func (s *WalletService) Withdraw(userID uuid.UUID, amount int, description string) error {
	if amount <= 0 {
		return errors.New("invalid_amount")
	}
	return s.DB.Transaction(func(tx *sqlx.Tx) error {
		var wallet domain.Wallet
		if err := tx.Get(&wallet, `SELECT * FROM wallets WHERE user_id = $1 FOR UPDATE`, userID); err != nil {
			return err
		}
		if wallet.Balance < amount {
			return errors.New("insufficient_balance")
		}
		newBal := wallet.Balance - amount
		if err := tx.Exec(`UPDATE wallets SET balance = $1 WHERE id = $2`, newBal, wallet.ID); err != nil {
			return err
		}
		wt := domain.WalletTransaction{
			WalletID:    wallet.ID,
			Amount:      amount,
			Type:        domain.TransactionWithdraw,
			Description: description,
		}
		if err := tx.Exec(`INSERT INTO wallet_transactions (wallet_id, amount, type, description) VALUES ($1, $2, $3, $4)`, wt.WalletID, wt.Amount, wt.Type, wt.Description); err != nil {
			return err
		}
		return nil
	})
}

func (s *WalletService) ChargeForBooking(userID uuid.UUID, amount int, bookingID uuid.UUID) error {
	if amount <= 0 {
		return errors.New("invalid_amount")
	}
	return s.DB.Transaction(func(tx *sqlx.Tx) error {
		var wallet domain.Wallet
		if err := tx.Get(&wallet, `SELECT * FROM wallets WHERE user_id = $1 FOR UPDATE`, userID); err != nil {
			return err
		}
		if wallet.Balance < amount {
			return errors.New("insufficient_balance")
		}
		newBal := wallet.Balance - amount
		if err := tx.Exec(`UPDATE wallets SET balance = $1 WHERE id = $2`, newBal, wallet.ID); err != nil {
			return err
		}
		wt := domain.WalletTransaction{
			WalletID:    wallet.ID,
			Amount:      amount,
			Type:        domain.TransactionBookingCharge,
			BookingID:   &bookingID,
			Description: "Charge for booking",
		}
		if err := tx.Exec(`INSERT INTO wallet_transactions (wallet_id, amount, type, booking_id, description) VALUES ($1, $2, $3, $4, $5)`, wt.WalletID, wt.Amount, wt.Type, wt.BookingID, wt.Description); err != nil {
			return err
		}
		return nil
	})
}

func (s *WalletService) RefundBooking(userID uuid.UUID, amount int, bookingID uuid.UUID, reason string) error {
	if amount <= 0 {
		return errors.New("invalid_amount")
	}
	return s.DB.Transaction(func(tx *sqlx.Tx) error {
		var wallet domain.Wallet
		if err := tx.Get(&wallet, `SELECT * FROM wallets WHERE user_id = $1 FOR UPDATE`, userID); err != nil {
			if err.Error() == "sql: no rows in result set" {
				wallet = domain.Wallet{UserID: userID, Balance: 0}
				if err := tx.Exec(`INSERT INTO wallets (user_id, balance) VALUES ($1, $2)`, userID, wallet.Balance); err != nil {
					return err
				}
			} else {
				return err
			}
		}
		newBal := wallet.Balance + amount
		if err := tx.Exec(`UPDATE wallets SET balance = $1 WHERE id = $2`, newBal, wallet.ID); err != nil {
			return err
		}
		wt := domain.WalletTransaction{
			WalletID:    wallet.ID,
			Amount:      amount,
			Type:        domain.TransactionRefund,
			BookingID:   &bookingID,
			Description: reason,
		}
		if err := tx.Exec(`INSERT INTO wallet_transactions (wallet_id, amount, type, booking_id, description) VALUES ($1, $2, $3, $4, $5)`, wt.WalletID, wt.Amount, wt.Type, wt.BookingID, wt.Description); err != nil {
			return err
		}
		return nil
	})
}
