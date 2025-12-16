package service

import (
	"context"
	"errors"
	"restaurant-booking/internal/domain"
	"restaurant-booking/internal/repository"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

var (
	ErrOldPasswordIncorrect = errors.New("old password is incorrect")
)

type UserService interface {
	GetUserByID(id uuid.UUID) (*domain.User, error)
	UpdateUser(id uuid.UUID, name string, phone *string) (*domain.User, error)
	ChangePassword(id uuid.UUID, oldPassword, newPassword string) error
}

type userService struct {
	userRepo repository.UserRepository
}

func NewUserService(userRepo repository.UserRepository) UserService {
	return &userService{
		userRepo: userRepo,
	}
}

func (s *userService) GetUserByID(id uuid.UUID) (*domain.User, error) {
	ctx := context.Background()
	user, err := s.userRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}
	return user, nil
}

func (s *userService) UpdateUser(id uuid.UUID, name string, phone *string) (*domain.User, error) {
	ctx := context.Background()
	user, err := s.userRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}

	// Update FirstName (use name as FirstName for simplicity)
	user.FirstName = name
	if phone != nil {
		user.Phone = *phone
	}

	if err := s.userRepo.Update(ctx, user); err != nil {
		return nil, err
	}

	return user, nil
}

func (s *userService) ChangePassword(id uuid.UUID, oldPassword, newPassword string) error {
	ctx := context.Background()
	user, err := s.userRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrUserNotFound
		}
		return err
	}

	// Verify old password
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(oldPassword)); err != nil {
		return ErrOldPasswordIncorrect
	}

	// Validate new password length
	if len(newPassword) < 8 {
		return ErrInvalidPassword
	}

	// Hash new password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	user.Password = string(hashedPassword)

	return s.userRepo.Update(ctx, user)
}
