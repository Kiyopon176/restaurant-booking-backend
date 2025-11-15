package repository

import (
	"errors"

	"github.com/Kiyopon176/restaurant-booking-backend/internal/domain"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type WalletRepository interface {
	Create(tx *gorm.DB, wallet *domain.Wallet) error
	GetByUserID(tx *gorm.DB, userID uuid.UUID) (*domain.Wallet, error)
	UpdateBalance(tx *gorm.DB, walletID uuid.UUID, newBalance int) error
	GetByID(tx *gorm.DB, id uuid.UUID) (*domain.Wallet, error)
}

type walletRepo struct {
	db *gorm.DB
}

func NewWalletRepository(db *gorm.DB) WalletRepository {
	return &walletRepo{db: db}
}

func (r *walletRepo) Create(tx *gorm.DB, wallet *domain.Wallet) error {
	if tx == nil {
		tx = r.db
	}
	return tx.Create(wallet).Error
}

func (r *walletRepo) GetByUserID(tx *gorm.DB, userID uuid.UUID) (*domain.Wallet, error) {
	if tx == nil {
		tx = r.db
	}
	var w domain.Wallet
	if err := tx.Where("user_id = ?", userID).First(&w).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &w, nil
}

func (r *walletRepo) UpdateBalance(tx *gorm.DB, walletID uuid.UUID, newBalance int) error {
	if tx == nil {
		tx = r.db
	}
	return tx.Model(&domain.Wallet{}).Where("id = ?", walletID).Update("balance", newBalance).Error
}

func (r *walletRepo) GetByID(tx *gorm.DB, id uuid.UUID) (*domain.Wallet, error) {
	if tx == nil {
		tx = r.db
	}
	var w domain.Wallet
	if err := tx.Where("id = ?", id).First(&w).Error; err != nil {
		return nil, err
	}
	return &w, nil
}
