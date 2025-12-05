package repository

import (
	"context"
	"restaurant-booking/internal/domain"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type PaymentRepository interface {
	Create(ctx context.Context, payment *domain.Payment) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Payment, error)
	GetByExternalID(ctx context.Context, externalID string) (*domain.Payment, error)
	GetByUserID(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*domain.Payment, error)
	Update(ctx context.Context, payment *domain.Payment) error
}

type paymentRepository struct {
	db *gorm.DB
}

func NewPaymentRepository(db *gorm.DB) PaymentRepository {
	return &paymentRepository{db: db}
}

func (r *paymentRepository) Create(ctx context.Context, payment *domain.Payment) error {
	return r.db.WithContext(ctx).Create(payment).Error
}

func (r *paymentRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Payment, error) {
	var payment domain.Payment
	err := r.db.WithContext(ctx).
		Preload("User").
		Preload("Booking").
		First(&payment, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &payment, nil
}

func (r *paymentRepository) GetByExternalID(ctx context.Context, externalID string) (*domain.Payment, error) {
	var payment domain.Payment
	err := r.db.WithContext(ctx).
		Where("external_payment_id = ?", externalID).
		First(&payment).Error
	if err != nil {
		return nil, err
	}
	return &payment, nil
}

func (r *paymentRepository) GetByUserID(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*domain.Payment, error) {
	var payments []*domain.Payment
	err := r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&payments).Error
	return payments, err
}

func (r *paymentRepository) Update(ctx context.Context, payment *domain.Payment) error {
	return r.db.WithContext(ctx).Save(payment).Error
}
