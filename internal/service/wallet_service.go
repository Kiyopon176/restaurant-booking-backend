package service

import (
	"errors"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/Kiyopon176/restaurant-booking-backend/internal/domain"
	"github.com/Kiyopon176/restaurant-booking-backend/internal/repository"
)

type WalletService struct {
	DB         *gorm.DB
	WalletRepo repository.WalletRepository
	TxnRepo    repository.PaymentRepository // unused here but commonly needed
}

func NewWalletService(db *gorm.DB, wr repository.WalletRepository) *WalletService {
	return &WalletService{
		DB:         db,
		WalletRepo: wr,
	}
}

func (s *WalletService) GetOrCreateWallet(userID uuid.UUID) (*domain.Wallet, error) {
	// non-transactional get; can be used in registration flow
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
	return s.DB.Transaction(func(tx *gorm.DB) error {
		// lock wallet row
		var wallet domain.Wallet
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("user_id = ?", userID).First(&wallet).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				// create wallet
				wallet = domain.Wallet{UserID: userID, Balance: 0}
				if err := tx.Create(&wallet).Error; err != nil {
					return err
				}
			} else {
				return err
			}
		}
		// update balance
		newBal := wallet.Balance + amount
		if err := tx.Model(&domain.Wallet{}).Where("id = ?", wallet.ID).Update("balance", newBal).Error; err != nil {
			return err
		}
		// create transaction record
		wt := domain.WalletTransaction{
			WalletID:    wallet.ID,
			Amount:      amount,
			Type:        domain.TransactionDeposit,
			Description: description,
		}
		if err := tx.Create(&wt).Error; err != nil {
			return err
		}
		return nil
	})
}

func (s *WalletService) Withdraw(userID uuid.UUID, amount int, description string) error {
	if amount <= 0 {
		return errors.New("invalid_amount")
	}
	return s.DB.Transaction(func(tx *gorm.DB) error {
		var wallet domain.Wallet
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("user_id = ?", userID).First(&wallet).Error; err != nil {
			return err
		}
		if wallet.Balance < amount {
			return errors.New("insufficient_balance")
		}
		if err := tx.Model(&domain.Wallet{}).Where("id = ?", wallet.ID).
			Update("balance", gorm.Expr("balance - ?", amount)).Error; err != nil {
			return err
		}
		wt := domain.WalletTransaction{
			WalletID:    wallet.ID,
			Amount:      amount,
			Type:        domain.TransactionWithdraw,
			Description: description,
		}
		if err := tx.Create(&wt).Error; err != nil {
			return err
		}
		return nil
	})
}

func (s *WalletService) ChargeForBooking(userID uuid.UUID, amount int, bookingID uuid.UUID) error {
	if amount <= 0 {
		return errors.New("invalid_amount")
	}
	return s.DB.Transaction(func(tx *gorm.DB) error {
		var wallet domain.Wallet
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("user_id = ?", userID).First(&wallet).Error; err != nil {
			return err
		}
		if wallet.Balance < amount {
			return errors.New("insufficient_balance")
		}
		if err := tx.Model(&domain.Wallet{}).Where("id = ?", wallet.ID).
			Update("balance", gorm.Expr("balance - ?", amount)).Error; err != nil {
			return err
		}
		wt := domain.WalletTransaction{
			WalletID:    wallet.ID,
			Amount:      amount,
			Type:        domain.TransactionBookingCharge,
			BookingID:   &bookingID,
			Description: "Charge for booking",
		}
		if err := tx.Create(&wt).Error; err != nil {
			return err
		}
		return nil
	})
}

func (s *WalletService) RefundBooking(userID uuid.UUID, amount int, bookingID uuid.UUID, reason string) error {
	if amount <= 0 {
		return errors.New("invalid_amount")
	}
	return s.DB.Transaction(func(tx *gorm.DB) error {
		var wallet domain.Wallet
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("user_id = ?", userID).First(&wallet).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				wallet = domain.Wallet{UserID: userID, Balance: 0}
				if err := tx.Create(&wallet).Error; err != nil {
					return err
				}
			} else {
				return err
			}
		}
		newBal := wallet.Balance + amount
		if err := tx.Model(&domain.Wallet{}).Where("id = ?", wallet.ID).Update("balance", newBal).Error; err != nil {
			return err
		}
		wt := domain.WalletTransaction{
			WalletID:    wallet.ID,
			Amount:      amount,
			Type:        domain.TransactionRefund,
			BookingID:   &bookingID,
			Description: reason,
		}
		if err := tx.Create(&wt).Error; err != nil {
			return err
		}
		return nil
	})
}
