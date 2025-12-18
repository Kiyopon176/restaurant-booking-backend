package service

import (
	"context"
	"errors"
	"restaurant-booking/internal/domain"
	"restaurant-booking/internal/repository"
	"restaurant-booking/pkg/logger"
	"strings"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

var (
	ErrRestaurantNotFound    = errors.New("restaurant not found")
	ErrUnauthorized          = errors.New("unauthorized: not the owner")
	ErrInvalidRestaurantName = errors.New("restaurant name cannot be empty")
	ErrImageNotFound         = errors.New("image not found")
)

type CreateRestaurantRequest struct {
	Name                string
	Address             string
	Latitude            *float64
	Longitude           *float64
	Description         string
	Phone               string
	Instagram           *string
	Website             *string
	CuisineType         domain.CuisineType
	AveragePrice        int
	MaxCombinableTables int
	WorkingHours        domain.WorkingHours
}

type UpdateRestaurantRequest struct {
	Name                *string
	Address             *string
	Latitude            *float64
	Longitude           *float64
	Description         *string
	Phone               *string
	Instagram           *string
	Website             *string
	CuisineType         *domain.CuisineType
	AveragePrice        *int
	MaxCombinableTables *int
	WorkingHours        *domain.WorkingHours
	IsActive            *bool
}

type AddImageRequest struct {
	CloudinaryURL      string
	CloudinaryPublicID string
	IsMain             bool
}

type RestaurantService interface {
	CreateRestaurant(ctx context.Context, ownerID uuid.UUID, req CreateRestaurantRequest) (*domain.Restaurant, error)
	GetRestaurant(ctx context.Context, id uuid.UUID) (*domain.Restaurant, error)
	GetRestaurants(ctx context.Context, limit, offset int) ([]*domain.Restaurant, error)
	UpdateRestaurant(ctx context.Context, id uuid.UUID, ownerID uuid.UUID, req UpdateRestaurantRequest) (*domain.Restaurant, error)
	DeleteRestaurant(ctx context.Context, id uuid.UUID, ownerID uuid.UUID) error
	AddImage(ctx context.Context, restaurantID uuid.UUID, ownerID uuid.UUID, req AddImageRequest) (*domain.RestaurantImage, error)
	DeleteImage(ctx context.Context, imageID uuid.UUID, restaurantID uuid.UUID, ownerID uuid.UUID) error
}

type restaurantService struct {
	restaurantRepo repository.RestaurantRepository
	db             *gorm.DB
	log            logger.Logger
}

func NewRestaurantService(restaurantRepo repository.RestaurantRepository, db *gorm.DB, log logger.Logger) RestaurantService {
	return &restaurantService{
		restaurantRepo: restaurantRepo,
		db:             db,
		log:            log,
	}
}

func (s *restaurantService) CreateRestaurant(ctx context.Context, ownerID uuid.UUID, req CreateRestaurantRequest) (*domain.Restaurant, error) {

	if strings.TrimSpace(req.Name) == "" {
		return nil, ErrInvalidRestaurantName
	}

	restaurant := &domain.Restaurant{
		OwnerID:             ownerID,
		Name:                req.Name,
		Address:             req.Address,
		Latitude:            req.Latitude,
		Longitude:           req.Longitude,
		Description:         req.Description,
		Phone:               req.Phone,
		Instagram:           req.Instagram,
		Website:             req.Website,
		CuisineType:         req.CuisineType,
		AveragePrice:        req.AveragePrice,
		MaxCombinableTables: req.MaxCombinableTables,
		WorkingHours:        req.WorkingHours,
		IsActive:            true,
	}

	if err := s.restaurantRepo.Create(ctx, restaurant); err != nil {
		return nil, err
	}

	return restaurant, nil
}

func (s *restaurantService) GetRestaurant(ctx context.Context, id uuid.UUID) (*domain.Restaurant, error) {
	restaurant, err := s.restaurantRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrRestaurantNotFound
		}
		return nil, err
	}
	return restaurant, nil
}

func (s *restaurantService) GetRestaurants(ctx context.Context, limit, offset int) ([]*domain.Restaurant, error) {
	return s.restaurantRepo.List(ctx, limit, offset)
}

func (s *restaurantService) UpdateRestaurant(ctx context.Context, id uuid.UUID, ownerID uuid.UUID, req UpdateRestaurantRequest) (*domain.Restaurant, error) {
	restaurant, err := s.restaurantRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrRestaurantNotFound
		}
		return nil, err
	}

	if restaurant.OwnerID != ownerID {
		return nil, ErrUnauthorized
	}

	if req.Name != nil {
		if strings.TrimSpace(*req.Name) == "" {
			return nil, ErrInvalidRestaurantName
		}
		restaurant.Name = *req.Name
	}
	if req.Address != nil {
		restaurant.Address = *req.Address
	}
	if req.Latitude != nil {
		restaurant.Latitude = req.Latitude
	}
	if req.Longitude != nil {
		restaurant.Longitude = req.Longitude
	}
	if req.Description != nil {
		restaurant.Description = *req.Description
	}
	if req.Phone != nil {
		restaurant.Phone = *req.Phone
	}
	if req.Instagram != nil {
		restaurant.Instagram = req.Instagram
	}
	if req.Website != nil {
		restaurant.Website = req.Website
	}
	if req.CuisineType != nil {
		restaurant.CuisineType = *req.CuisineType
	}
	if req.AveragePrice != nil {
		restaurant.AveragePrice = *req.AveragePrice
	}
	if req.MaxCombinableTables != nil {
		restaurant.MaxCombinableTables = *req.MaxCombinableTables
	}
	if req.WorkingHours != nil {
		restaurant.WorkingHours = *req.WorkingHours
	}
	if req.IsActive != nil {
		restaurant.IsActive = *req.IsActive
	}

	if err := s.restaurantRepo.Update(ctx, restaurant); err != nil {
		return nil, err
	}

	return restaurant, nil
}

func (s *restaurantService) DeleteRestaurant(ctx context.Context, id uuid.UUID, ownerID uuid.UUID) error {
	restaurant, err := s.restaurantRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrRestaurantNotFound
		}
		return err
	}
	if restaurant.OwnerID != ownerID {
		return ErrUnauthorized
	}
	restaurant.IsActive = false
	return s.restaurantRepo.Update(ctx, restaurant)
}

func (s *restaurantService) AddImage(ctx context.Context, restaurantID uuid.UUID, ownerID uuid.UUID, req AddImageRequest) (*domain.RestaurantImage, error) {
	restaurant, err := s.restaurantRepo.GetByID(ctx, restaurantID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrRestaurantNotFound
		}
		return nil, err
	}

	if restaurant.OwnerID != ownerID {
		return nil, ErrUnauthorized
	}

	image := &domain.RestaurantImage{
		RestaurantID:       restaurantID,
		CloudinaryURL:      req.CloudinaryURL,
		CloudinaryPublicID: req.CloudinaryPublicID,
		IsMain:             req.IsMain,
	}

	if err := s.db.WithContext(ctx).Create(image).Error; err != nil {
		return nil, err
	}

	return image, nil
}

func (s *restaurantService) DeleteImage(ctx context.Context, imageID uuid.UUID, restaurantID uuid.UUID, ownerID uuid.UUID) error {
	restaurant, err := s.restaurantRepo.GetByID(ctx, restaurantID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrRestaurantNotFound
		}
		return err
	}

	if restaurant.OwnerID != ownerID {
		return ErrUnauthorized
	}

	var image domain.RestaurantImage
	if err := s.db.WithContext(ctx).
		Where("id = ? AND restaurant_id = ?", imageID, restaurantID).
		First(&image).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrImageNotFound
		}
		return err
	}

	return s.db.WithContext(ctx).Delete(&image).Error
}
