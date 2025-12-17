package service

import (
	"errors"
	"regexp"
	"restaurant-booking/internal/domain"
	"restaurant-booking/internal/repository"
	"restaurant-booking/pkg/jwt"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

var (
	ErrInvalidEmail        = errors.New("invalid email format")
	ErrInvalidPassword     = errors.New("password must be at least 8 characters")
	ErrEmailExists         = errors.New("email already exists")
	ErrInvalidCredentials  = errors.New("invalid email or password")
	ErrUserNotFound        = errors.New("user not found")
	ErrInvalidRefreshToken = errors.New("invalid refresh token")
	ErrExpiredRefreshToken = errors.New("refresh token has expired")
)

type AuthService interface {
	Register(email, password, firstName, lastName string, phone string, role domain.UserRole) (*domain.User, string, string, error)
	Login(email, password string) (string, string, *domain.User, error)
	RefreshToken(refreshToken string) (string, string, error)
	Logout(refreshToken string) error
}

type authService struct {
	userRepo         repository.UserRepository
	refreshTokenRepo repository.RefreshTokenRepository
	jwtManager       *jwt.Manager
}

func NewAuthService(
	userRepo repository.UserRepository,
	refreshTokenRepo repository.RefreshTokenRepository,
	jwtManager *jwt.Manager,
) AuthService {
	return &authService{
		userRepo:         userRepo,
		refreshTokenRepo: refreshTokenRepo,
		jwtManager:       jwtManager,
	}
}

func (s *authService) Register(email, password, firstName, lastName string, phone string, role domain.UserRole) (*domain.User, string, string, error) {

	if !isValidEmail(email) {
		return nil, "", "", ErrInvalidEmail
	}

	if len(password) < 8 {
		return nil, "", "", ErrInvalidPassword
	}

	_, err := s.userRepo.GetByEmail(email)
	if err == nil {
		return nil, "", "", ErrEmailExists
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, "", "", err
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, "", "", err
	}

	user := &domain.User{
		ID:        uuid.New(),
		Email:     email,
		Password:  string(hashedPassword),
		FirstName: firstName,
		LastName:  lastName,
		Phone:     phone,
		Role:      role,
		IsActive:  true,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := s.userRepo.Create(user); err != nil {
		return nil, "", "", err
	}

	accessToken, err := s.jwtManager.GenerateAccessToken(user.ID, user.Role)
	if err != nil {
		return nil, "", "", err
	}

	refreshToken, err := s.jwtManager.GenerateRefreshToken()
	if err != nil {
		return nil, "", "", err
	}

	refreshTokenEntity := &domain.RefreshToken{
		ID:        uuid.New(),
		UserID:    user.ID,
		Token:     refreshToken,
		ExpiresAt: time.Now().Add(s.jwtManager.GetRefreshExpire()),
		CreatedAt: time.Now(),
	}

	if err := s.refreshTokenRepo.Create(refreshTokenEntity); err != nil {
		return nil, "", "", err
	}

	return user, accessToken, refreshToken, nil
}

func (s *authService) Login(email, password string) (string, string, *domain.User, error) {

	user, err := s.userRepo.GetByEmail(email)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return "", "", nil, ErrInvalidCredentials
		}
		return "", "", nil, err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		return "", "", nil, ErrInvalidCredentials
	}

	accessToken, err := s.jwtManager.GenerateAccessToken(user.ID, user.Role)
	if err != nil {
		return "", "", nil, err
	}

	refreshToken, err := s.jwtManager.GenerateRefreshToken()
	if err != nil {
		return "", "", nil, err
	}

	refreshTokenEntity := &domain.RefreshToken{
		ID:        uuid.New(),
		UserID:    user.ID,
		Token:     refreshToken,
		ExpiresAt: time.Now().Add(s.jwtManager.GetRefreshExpire()),
		CreatedAt: time.Now(),
	}

	if err := s.refreshTokenRepo.Create(refreshTokenEntity); err != nil {
		return "", "", nil, err
	}

	return accessToken, refreshToken, user, nil
}

func (s *authService) RefreshToken(refreshToken string) (string, string, error) {

	tokenEntity, err := s.refreshTokenRepo.GetByToken(refreshToken)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return "", "", ErrInvalidRefreshToken
		}
		return "", "", err
	}

	if time.Now().After(tokenEntity.ExpiresAt) {

		_ = s.refreshTokenRepo.DeleteByToken(refreshToken)
		return "", "", ErrExpiredRefreshToken
	}

	user, err := s.userRepo.GetByID(tokenEntity.UserID)
	if err != nil {
		return "", "", err
	}

	newAccessToken, err := s.jwtManager.GenerateAccessToken(user.ID, user.Role)
	if err != nil {
		return "", "", err
	}

	newRefreshToken, err := s.jwtManager.GenerateRefreshToken()
	if err != nil {
		return "", "", err
	}

	if err := s.refreshTokenRepo.DeleteByToken(refreshToken); err != nil {
		return "", "", err
	}

	newTokenEntity := &domain.RefreshToken{
		ID:        uuid.New(),
		UserID:    user.ID,
		Token:     newRefreshToken,
		ExpiresAt: time.Now().Add(s.jwtManager.GetRefreshExpire()),
		CreatedAt: time.Now(),
	}

	if err := s.refreshTokenRepo.Create(newTokenEntity); err != nil {
		return "", "", err
	}

	return newAccessToken, newRefreshToken, nil
}

func (s *authService) Logout(refreshToken string) error {
	return s.refreshTokenRepo.DeleteByToken(refreshToken)
}

func isValidEmail(email string) bool {
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)
	return emailRegex.MatchString(email)
}
