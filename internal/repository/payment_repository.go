package repository

import (
	"errors"

	"github.com/Kiyopon176/restaurant-booking-backend/internal/domain"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type PaymentRepository interface {
	Create(tx *gorm.DB, p *domain.Payment) error
	GetByID(tx *gorm.DB, id uuid.UUID) (*domain.Payment, error)
	GetByExternalID(tx *gorm.DB, externalID string) (*domain.Payment, error)
	Update(tx *gorm.DB, p *domain.Payment) error
	GetByUserID(tx *gorm.DB, userID uuid.UUID, pagination Pagination) ([]domain.Payment, int64, error)
}

type paymentRepo struct {
	db *gorm.DB
}

func NewPaymentRepository(db *gorm.DB) PaymentRepository {
	return &paymentRepo{db: db}
}

func (r *paymentRepo) Create(tx *gorm.DB, p *domain.Payment) error {
	if tx == nil {
		tx = r.db
	}
	return tx.Create(p).Error
}

func (r *paymentRepo) GetByID(tx *gorm.DB, id uuid.UUID) (*domain.Payment, error) {
	if tx == nil {
		tx = r.db
	}
	var p domain.Payment
	if err := tx.Where("id = ?", id).First(&p).Error; err != nil {
		return nil, err
	}
	return &p, nil
}

func (r *paymentRepo) GetByExternalID(tx *gorm.DB, externalID string) (*domain.Payment, error) {
	if tx == nil {
		tx = r.db
	}
	var p domain.Payment
	if err := tx.Where("external_payment_id = ?", externalID).First(&p).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &p, nil
}

func (r *paymentRepo) Update(tx *gorm.DB, p *domain.Payment) error {
	if tx == nil {
		tx = r.db
	}
	return tx.Save(p).Error
}

func (r *paymentRepo) GetByUserID(tx *gorm.DB, userID uuid.UUID, pagination Pagination) ([]domain.Payment, int64, error) {
	if tx == nil {
		tx = r.db
	}
	var payments []domain.Payment
	var total int64
	query := tx.Where("user_id = ?", userID)
	query.Count(&total)
	if err := query.Offset(pagination.Offset()).Limit(pagination.Limit()).Find(&payments).Error; err != nil {
		return nil, 0, err
	}
	return payments, total, nil
}
