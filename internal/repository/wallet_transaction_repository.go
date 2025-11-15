package repository

import (
	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/Kiyopon176/restaurant-booking-backend/internal/domain"
)

type WalletTransactionRepository interface {
	Create(transaction *domain.WalletTransaction) error
	GetByWalletID(walletID uuid.UUID, pagination Pagination) ([]domain.WalletTransaction, int64, error)
	GetByBookingID(bookingID uuid.UUID) ([]domain.WalletTransaction, error)
}

type walletTransactionRepo struct{ db *gorm.DB }

func NewWalletTransactionRepository(db *gorm.DB) WalletTransactionRepository {
	return &walletTransactionRepo{db}
}

func (r *walletTransactionRepo) Create(transaction *domain.WalletTransaction) error {
	return r.db.Create(transaction).Error
}

func (r *walletTransactionRepo) GetByWalletID(walletID uuid.UUID, pagination Pagination) ([]domain.WalletTransaction, int64, error) {
	var list []domain.WalletTransaction
	var total int64
	q := r.db.Where("wallet_id = ?", walletID)
	q.Model(&domain.WalletTransaction{}).Count(&total)
	if err := q.Order("created_at desc").Offset(pagination.Offset()).Limit(pagination.PageSize).Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

func (r *walletTransactionRepo) GetByBookingID(bookingID uuid.UUID) ([]domain.WalletTransaction, error) {
	var list []domain.WalletTransaction
	if err := r.db.Where("booking_id = ?", bookingID).Find(&list).Error; err != nil {
		return nil, err
	}
	return list, nil
}
