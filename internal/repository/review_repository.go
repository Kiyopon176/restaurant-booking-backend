package repository

import (
	"context"
	"restaurant-booking/internal/domain"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type ReviewRepository interface {
	Create(ctx context.Context, review *domain.Review) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Review, error)
	GetByRestaurantID(ctx context.Context, restaurantID uuid.UUID, limit, offset int) ([]*domain.Review, error)
	GetByUserID(ctx context.Context, userID uuid.UUID) ([]*domain.Review, error)
	Update(ctx context.Context, review *domain.Review) error
	Delete(ctx context.Context, id uuid.UUID) error
}

type reviewRepository struct {
	db *gorm.DB
}

func NewReviewRepository(db *gorm.DB) ReviewRepository {
	return &reviewRepository{db: db}
}

func (r *reviewRepository) Create(ctx context.Context, review *domain.Review) error {
	return r.db.WithContext(ctx).Create(review).Error
}

func (r *reviewRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Review, error) {
	var review domain.Review
	err := r.db.WithContext(ctx).
		Preload("User").
		Preload("Restaurant").
		First(&review, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &review, nil
}

func (r *reviewRepository) GetByRestaurantID(ctx context.Context, restaurantID uuid.UUID, limit, offset int) ([]*domain.Review, error) {
	var reviews []*domain.Review
	err := r.db.WithContext(ctx).
		Where("restaurant_id = ? AND is_visible = ?", restaurantID, true).
		Preload("User").
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&reviews).Error
	return reviews, err
}

func (r *reviewRepository) GetByUserID(ctx context.Context, userID uuid.UUID) ([]*domain.Review, error) {
	var reviews []*domain.Review
	err := r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Preload("Restaurant").
		Order("created_at DESC").
		Find(&reviews).Error
	return reviews, err
}

func (r *reviewRepository) Update(ctx context.Context, review *domain.Review) error {
	return r.db.WithContext(ctx).Save(review).Error
}

func (r *reviewRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&domain.Review{}, "id = ?", id).Error
}
