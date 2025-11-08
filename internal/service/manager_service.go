package service

import (
	"context"
	"errors"
	"restaurant-booking/internal/domain"
	"restaurant-booking/internal/repository"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

var (
	ErrManagerAlreadyExists = errors.New("user is already a manager for this restaurant")
	ErrManagerNotFound      = errors.New("manager not found")
	ErrUserNotFound         = errors.New("user not found")
)

type AddManagerRequest struct {
	UserID uuid.UUID
}

type ManagerService interface {
	AddManager(ctx context.Context, restaurantID uuid.UUID, ownerID uuid.UUID, req AddManagerRequest) (*domain.RestaurantManager, error)
	RemoveManager(ctx context.Context, restaurantID uuid.UUID, ownerID uuid.UUID, userID uuid.UUID) error
	GetManagers(ctx context.Context, restaurantID uuid.UUID) ([]*domain.RestaurantManager, error)
}

type managerService struct {
	managerRepo    repository.RestaurantManagerRepository
	restaurantRepo repository.RestaurantRepository
	userRepo       repository.UserRepository
}

func NewManagerService(
	managerRepo repository.RestaurantManagerRepository,
	restaurantRepo repository.RestaurantRepository,
	userRepo repository.UserRepository,
) ManagerService {
	return &managerService{
		managerRepo:    managerRepo,
		restaurantRepo: restaurantRepo,
		userRepo:       userRepo,
	}
}

func (s *managerService) AddManager(ctx context.Context, restaurantID uuid.UUID, ownerID uuid.UUID, req AddManagerRequest) (*domain.RestaurantManager, error) {

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
	user, err := s.userRepo.GetByID(ctx, req.UserID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}

	isManager, err := s.managerRepo.IsManager(ctx, user.ID, restaurantID)
	if err != nil {
		return nil, err
	}

	if isManager {
		return nil, ErrManagerAlreadyExists
	}
	manager := &domain.RestaurantManager{
		UserID:       user.ID,
		RestaurantID: restaurantID,
		AssignedAt:   time.Now(),
	}

	if err := s.managerRepo.Create(ctx, manager); err != nil {
		return nil, err
	}

	manager.User = user
	manager.Restaurant = restaurant

	return manager, nil
}

func (s *managerService) RemoveManager(ctx context.Context, restaurantID uuid.UUID, ownerID uuid.UUID, userID uuid.UUID) error {

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

	isManager, err := s.managerRepo.IsManager(ctx, userID, restaurantID)
	if err != nil {
		return err
	}

	if !isManager {
		return ErrManagerNotFound
	}

	return s.managerRepo.Delete(ctx, userID, restaurantID)
}

func (s *managerService) GetManagers(ctx context.Context, restaurantID uuid.UUID) ([]*domain.RestaurantManager, error) {

	_, err := s.restaurantRepo.GetByID(ctx, restaurantID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrRestaurantNotFound
		}
		return nil, err
	}

	return s.managerRepo.GetManagersByRestaurant(ctx, restaurantID)
}
