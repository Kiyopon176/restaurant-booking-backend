package repository

import (
	"context"
	"restaurant-booking/internal/domain"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type BookingRepository interface {
	Create(ctx context.Context, booking *domain.Booking) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Booking, error)
	GetByUserID(ctx context.Context, userID uuid.UUID) ([]*domain.Booking, error)
	GetByRestaurantID(ctx context.Context, restaurantID uuid.UUID, date time.Time) ([]*domain.Booking, error)
	Update(ctx context.Context, booking *domain.Booking) error
	Delete(ctx context.Context, id uuid.UUID) error
	CheckTableAvailability(ctx context.Context, tableID uuid.UUID, startTime, endTime time.Time) (bool, error)
}

type bookingRepository struct {
	db *gorm.DB
}

func NewBookingRepository(db *gorm.DB) BookingRepository {
	return &bookingRepository{db: db}
}

func (r *bookingRepository) Create(ctx context.Context, booking *domain.Booking) error {
	return r.db.WithContext(ctx).Create(booking).Error
}

func (r *bookingRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Booking, error) {
	var booking domain.Booking
	err := r.db.WithContext(ctx).
		Preload("Restaurant").
		Preload("Table").
		Preload("User").
		First(&booking, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &booking, nil
}

func (r *bookingRepository) GetByUserID(ctx context.Context, userID uuid.UUID) ([]*domain.Booking, error) {
	var bookings []*domain.Booking
	err := r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Order("booking_date DESC, start_time DESC").
		Find(&bookings).Error
	return bookings, err
}

func (r *bookingRepository) GetByRestaurantID(ctx context.Context, restaurantID uuid.UUID, date time.Time) ([]*domain.Booking, error) {
	var bookings []*domain.Booking
	err := r.db.WithContext(ctx).
		Where("restaurant_id = ? AND booking_date = ? AND status != ?",
			restaurantID, date, domain.BookingStatusCancelled).
		Find(&bookings).Error
	return bookings, err
}

func (r *bookingRepository) Update(ctx context.Context, booking *domain.Booking) error {
	return r.db.WithContext(ctx).Save(booking).Error
}

func (r *bookingRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&domain.Booking{}, "id = ?", id).Error
}

func (r *bookingRepository) CheckTableAvailability(ctx context.Context, tableID uuid.UUID, startTime, endTime time.Time) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&domain.Booking{}).
		Where("table_id = ? AND status NOT IN (?, ?) AND ((start_time < ? AND end_time > ?) OR (start_time < ? AND end_time > ?))",
			tableID,
			domain.BookingStatusCancelled,
			domain.BookingStatusCompleted,
			endTime, startTime,
			endTime, endTime,
		).
		Count(&count).Error

	return count == 0, err
}
