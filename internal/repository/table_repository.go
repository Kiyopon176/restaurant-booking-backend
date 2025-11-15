package repository

import (
	"context"
	"restaurant-booking/internal/domain"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type TableRepository interface {
	Create(ctx context.Context, table *domain.Table) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Table, error)
	GetByRestaurantID(ctx context.Context, restaurantID uuid.UUID) ([]*domain.Table, error)
	GetAvailableTables(ctx context.Context, restaurantID uuid.UUID, minCapacity int) ([]*domain.Table, error)
	Update(ctx context.Context, table *domain.Table) error
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, limit, offset int) ([]*domain.Table, error)
}

type tableRepository struct {
	db *gorm.DB
}

func NewTableRepository(db *gorm.DB) TableRepository {
	return &tableRepository{db: db}
}

func (r *tableRepository) Create(ctx context.Context, table *domain.Table) error {
	return r.db.WithContext(ctx).Create(table).Error
}

func (r *tableRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Table, error) {
	var table domain.Table
	err := r.db.WithContext(ctx).
		Preload("Restaurant").
		First(&table, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &table, nil
}

func (r *tableRepository) GetByRestaurantID(ctx context.Context, restaurantID uuid.UUID) ([]*domain.Table, error) {
	var tables []*domain.Table
	err := r.db.WithContext(ctx).
		Where("restaurant_id = ? AND is_active = ?", restaurantID, true).
		Order("table_number ASC").
		Find(&tables).Error
	return tables, err
}

func (r *tableRepository) GetAvailableTables(ctx context.Context, restaurantID uuid.UUID, minCapacity int) ([]*domain.Table, error) {
	var tables []*domain.Table
	err := r.db.WithContext(ctx).
		Where("restaurant_id = ? AND is_active = ? AND max_capacity >= ?",
			restaurantID, true, minCapacity).
		Order("min_capacity ASC").
		Find(&tables).Error
	return tables, err
}

func (r *tableRepository) Update(ctx context.Context, table *domain.Table) error {
	return r.db.WithContext(ctx).Save(table).Error
}

func (r *tableRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&domain.Table{}, "id = ?", id).Error
}

func (r *tableRepository) List(ctx context.Context, limit, offset int) ([]*domain.Table, error) {
	var tables []*domain.Table
	err := r.db.WithContext(ctx).
		Where("is_active = ?", true).
		Limit(limit).
		Offset(offset).
		Find(&tables).Error
	return tables, err
}
