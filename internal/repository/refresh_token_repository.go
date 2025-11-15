package repository

import (
	"restaurant-booking/internal/domain"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type RefreshTokenRepository interface {
	Create(token *domain.RefreshToken) error
	GetByToken(token string) (*domain.RefreshToken, error)
	DeleteByToken(token string) error
	DeleteAllByUserID(userID uuid.UUID) error
}

type refreshTokenRepository struct {
	db *gorm.DB
}

func NewRefreshTokenRepository(db *gorm.DB) RefreshTokenRepository {
	return &refreshTokenRepository{db: db}
}

func (r *refreshTokenRepository) Create(token *domain.RefreshToken) error {
	return r.db.Create(token).Error
}

func (r *refreshTokenRepository) GetByToken(token string) (*domain.RefreshToken, error) {
	var refreshToken domain.RefreshToken
	if err := r.db.Where("token = ?", token).First(&refreshToken).Error; err != nil {
		return nil, err
	}
	return &refreshToken, nil
}

func (r *refreshTokenRepository) DeleteByToken(token string) error {
	return r.db.Where("token = ?", token).Delete(&domain.RefreshToken{}).Error
}

func (r *refreshTokenRepository) DeleteAllByUserID(userID uuid.UUID) error {
	return r.db.Where("user_id = ?", userID).Delete(&domain.RefreshToken{}).Error
}
