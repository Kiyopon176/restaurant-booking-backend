package repository

import (
	"context"
	"restaurant-booking/internal/domain"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type RestaurantManagerRepository interface {
	Create(ctx context.Context, manager *domain.RestaurantManager) error
	Delete(ctx context.Context, userID, restaurantID uuid.UUID) error
	GetManagersByRestaurant(ctx context.Context, restaurantID uuid.UUID) ([]*domain.RestaurantManager, error)
	GetRestaurantsByManager(ctx context.Context, userID uuid.UUID) ([]*domain.RestaurantManager, error)
	IsManager(ctx context.Context, userID, restaurantID uuid.UUID) (bool, error)
}

type restaurantManagerRepository struct {
	db *gorm.DB
}

func NewRestaurantManagerRepository(db *gorm.DB) RestaurantManagerRepository {
	return &restaurantManagerRepository{db: db}
}

func (r *restaurantManagerRepository) Create(ctx context.Context, manager *domain.RestaurantManager) error {
	return r.db.WithContext(ctx).Create(manager).Error
}

func (r *restaurantManagerRepository) Delete(ctx context.Context, userID, restaurantID uuid.UUID) error {
	return r.db.WithContext(ctx).
		Where("user_id = ? AND restaurant_id = ?", userID, restaurantID).
		Delete(&domain.RestaurantManager{}).Error
}

func (r *restaurantManagerRepository) GetManagersByRestaurant(ctx context.Context, restaurantID uuid.UUID) ([]*domain.RestaurantManager, error) {
	var managers []*domain.RestaurantManager
	err := r.db.WithContext(ctx).
		Preload("User").
		Preload("Restaurant").
		Where("restaurant_id = ?", restaurantID).
		Find(&managers).Error
	return managers, err
}

func (r *restaurantManagerRepository) GetRestaurantsByManager(ctx context.Context, userID uuid.UUID) ([]*domain.RestaurantManager, error) {
	var managers []*domain.RestaurantManager
	err := r.db.WithContext(ctx).
		Preload("User").
		Preload("Restaurant").
		Where("user_id = ?", userID).
		Find(&managers).Error
	return managers, err
}

func (r *restaurantManagerRepository) IsManager(ctx context.Context, userID, restaurantID uuid.UUID) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&domain.RestaurantManager{}).
		Where("user_id = ? AND restaurant_id = ?", userID, restaurantID).
		Count(&count).Error
	return count > 0, err
}
