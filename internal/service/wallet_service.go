package service

import (
	"context"
	"errors"
	"restaurant-booking/internal/domain"
	"restaurant-booking/internal/repository"
	"restaurant-booking/pkg/logger"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

var (
	ErrInsufficientBalance = errors.New("insufficient balance")
	ErrInvalidAmount       = errors.New("amount must be positive")
	ErrWalletNotFound      = errors.New("wallet not found")
)

type WalletService interface {
	GetOrCreateWallet(ctx context.Context, userID uuid.UUID) (*domain.Wallet, error)
	GetBalance(ctx context.Context, userID uuid.UUID) (int, error)
	Deposit(ctx context.Context, userID uuid.UUID, amount int, description string) error
	Withdraw(ctx context.Context, userID uuid.UUID, amount int, description string) error
	ChargeForBooking(ctx context.Context, userID uuid.UUID, amount int, bookingID uuid.UUID) error
	RefundBooking(ctx context.Context, userID uuid.UUID, amount int, bookingID uuid.UUID, reason string) error
	GetTransactions(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*domain.WalletTransaction, error)
}

type walletService struct {
	walletRepo repository.WalletRepository
	db         *gorm.DB
	log        logger.Logger
}

func NewWalletService(walletRepo repository.WalletRepository, db *gorm.DB, log logger.Logger) WalletService {
	return &walletService{
		walletRepo: walletRepo,
		db:         db,
		log:        log,
	}
}

func (s *walletService) GetOrCreateWallet(ctx context.Context, userID uuid.UUID) (*domain.Wallet, error) {
	wallet, err := s.walletRepo.GetByUserID(ctx, userID)
	if err == nil {
		return wallet, nil
	}

	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	wallet = &domain.Wallet{
		UserID:  userID,
		Balance: 0,
	}

	if err := s.walletRepo.Create(ctx, wallet); err != nil {
		return nil, err
	}

	return wallet, nil
}

func (s *walletService) GetBalance(ctx context.Context, userID uuid.UUID) (int, error) {
	wallet, err := s.walletRepo.GetByUserID(ctx, userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return 0, nil
		}
		return 0, err
	}
	return wallet.Balance, nil
}

func (s *walletService) Deposit(ctx context.Context, userID uuid.UUID, amount int, description string) error {
	if amount <= 0 {
		return ErrInvalidAmount
	}

	return s.db.Transaction(func(tx *gorm.DB) error {
		var wallet domain.Wallet
		err := tx.WithContext(ctx).
			Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("user_id = ?", userID).
			First(&wallet).Error

		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {

				wallet = domain.Wallet{
					UserID:  userID,
					Balance: 0,
				}
				if err := tx.WithContext(ctx).Create(&wallet).Error; err != nil {
					return err
				}
			} else {
				return err
			}
		}

		wallet.Balance += amount
		if err := tx.WithContext(ctx).Save(&wallet).Error; err != nil {
			return err
		}

		transaction := &domain.WalletTransaction{
			WalletID:    wallet.ID,
			Amount:      amount,
			Type:        domain.TransactionDeposit,
			Description: description,
		}

		return tx.WithContext(ctx).Create(transaction).Error
	})
}

func (s *walletService) Withdraw(ctx context.Context, userID uuid.UUID, amount int, description string) error {
	if amount <= 0 {
		return ErrInvalidAmount
	}

	return s.db.Transaction(func(tx *gorm.DB) error {
		var wallet domain.Wallet
		err := tx.WithContext(ctx).
			Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("user_id = ?", userID).
			First(&wallet).Error

		if err != nil {
			return err
		}

		if wallet.Balance < amount {
			return ErrInsufficientBalance
		}

		wallet.Balance -= amount
		if err := tx.WithContext(ctx).Save(&wallet).Error; err != nil {
			return err
		}

		transaction := &domain.WalletTransaction{
			WalletID:    wallet.ID,
			Amount:      amount,
			Type:        domain.TransactionWithdraw,
			Description: description,
		}

		return tx.WithContext(ctx).Create(transaction).Error
	})
}

func (s *walletService) ChargeForBooking(ctx context.Context, userID uuid.UUID, amount int, bookingID uuid.UUID) error {
	if amount <= 0 {
		return ErrInvalidAmount
	}

	return s.db.Transaction(func(tx *gorm.DB) error {
		var wallet domain.Wallet
		err := tx.WithContext(ctx).
			Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("user_id = ?", userID).
			First(&wallet).Error

		if err != nil {
			return err
		}

		if wallet.Balance < amount {
			return ErrInsufficientBalance
		}

		wallet.Balance -= amount
		if err := tx.WithContext(ctx).Save(&wallet).Error; err != nil {
			return err
		}

		transaction := &domain.WalletTransaction{
			WalletID:    wallet.ID,
			Amount:      amount,
			Type:        domain.TransactionBookingCharge,
			BookingID:   &bookingID,
			Description: "Charge for booking",
		}

		return tx.WithContext(ctx).Create(transaction).Error
	})
}

func (s *walletService) RefundBooking(ctx context.Context, userID uuid.UUID, amount int, bookingID uuid.UUID, reason string) error {
	if amount <= 0 {
		return ErrInvalidAmount
	}

	return s.db.Transaction(func(tx *gorm.DB) error {
		var wallet domain.Wallet
		err := tx.WithContext(ctx).
			Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("user_id = ?", userID).
			First(&wallet).Error

		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				wallet = domain.Wallet{
					UserID:  userID,
					Balance: 0,
				}
				if err := tx.WithContext(ctx).Create(&wallet).Error; err != nil {
					return err
				}
			} else {
				return err
			}
		}

		wallet.Balance += amount
		if err := tx.WithContext(ctx).Save(&wallet).Error; err != nil {
			return err
		}

		transaction := &domain.WalletTransaction{
			WalletID:    wallet.ID,
			Amount:      amount,
			Type:        domain.TransactionRefund,
			BookingID:   &bookingID,
			Description: reason,
		}

		return tx.WithContext(ctx).Create(transaction).Error
	})
}

func (s *walletService) GetTransactions(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*domain.WalletTransaction, error) {
	wallet, err := s.walletRepo.GetByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	return s.walletRepo.GetTransactions(ctx, wallet.ID, limit, offset)
}
