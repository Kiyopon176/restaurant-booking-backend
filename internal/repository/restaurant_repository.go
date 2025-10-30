package repository

import (
	"context"
	"restaurant-booking/internal/domain"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type RestaurantRepository interface {
	Create(ctx context.Context, restaurant *domain.Restaurant) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Restaurant, error)
	GetByOwnerID(ctx context.Context, ownerID uuid.UUID) ([]*domain.Restaurant, error)
	Update(ctx context.Context, restaurant *domain.Restaurant) error
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, limit, offset int) ([]*domain.Restaurant, error)
	Search(ctx context.Context, cuisineType *domain.CuisineType, minRating float64, limit, offset int) ([]*domain.Restaurant, error)
}

type restaurantRepository struct {
	db *gorm.DB
}

func NewRestaurantRepository(db *gorm.DB) RestaurantRepository {
	return &restaurantRepository{db: db}
}

func (r *restaurantRepository) Create(ctx context.Context, restaurant *domain.Restaurant) error {
	return r.db.WithContext(ctx).Create(restaurant).Error
}

func (r *restaurantRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Restaurant, error) {
	var restaurant domain.Restaurant
	err := r.db.WithContext(ctx).
		Preload("Images").
		Preload("Tables").
		First(&restaurant, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &restaurant, nil
}

func (r *restaurantRepository) GetByOwnerID(ctx context.Context, ownerID uuid.UUID) ([]*domain.Restaurant, error) {
	var restaurants []*domain.Restaurant
	err := r.db.WithContext(ctx).
		Where("owner_id = ?", ownerID).
		Find(&restaurants).Error
	return restaurants, err
}

func (r *restaurantRepository) Update(ctx context.Context, restaurant *domain.Restaurant) error {
	return r.db.WithContext(ctx).Save(restaurant).Error
}

func (r *restaurantRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&domain.Restaurant{}, "id = ?", id).Error
}

func (r *restaurantRepository) List(ctx context.Context, limit, offset int) ([]*domain.Restaurant, error) {
	var restaurants []*domain.Restaurant
	err := r.db.WithContext(ctx).
		Where("is_active = ?", true).
		Limit(limit).
		Offset(offset).
		Find(&restaurants).Error
	return restaurants, err
}

func (r *restaurantRepository) Search(ctx context.Context, cuisineType *domain.CuisineType, minRating float64, limit, offset int) ([]*domain.Restaurant, error) {
	var restaurants []*domain.Restaurant
	query := r.db.WithContext(ctx).Where("is_active = ?", true)

	if cuisineType != nil {
		query = query.Where("cuisine_type = ?", *cuisineType)
	}

	if minRating > 0 {
		query = query.Where("rating >= ?", minRating)
	}

	err := query.Limit(limit).Offset(offset).Find(&restaurants).Error
	return restaurants, err
}
