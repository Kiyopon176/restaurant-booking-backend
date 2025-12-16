package service

import (
	"context"
	_ "errors"
	"restaurant-booking/internal/domain"
	"restaurant-booking/pkg/jwt"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	tmock "github.com/stretchr/testify/mock"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// Mock repositories
type MockUserRepository struct {
	tmock.Mock
}

func (m *MockUserRepository) Create(ctx context.Context, user *domain.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *MockUserRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.User), args.Error(1)
}

func (m *MockUserRepository) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	args := m.Called(ctx, email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.User), args.Error(1)
}

func (m *MockUserRepository) GetByPhone(ctx context.Context, phone string) (*domain.User, error) {
	args := m.Called(ctx, phone)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.User), args.Error(1)
}

func (m *MockUserRepository) Update(ctx context.Context, user *domain.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *MockUserRepository) Delete(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockUserRepository) List(ctx context.Context, limit, offset int) ([]*domain.User, error) {
	args := m.Called(ctx, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.User), args.Error(1)
}

type MockRefreshTokenRepository struct {
	tmock.Mock
}

func (m *MockRefreshTokenRepository) Create(token *domain.RefreshToken) error {
	args := m.Called(token)
	return args.Error(0)
}

func (m *MockRefreshTokenRepository) GetByToken(token string) (*domain.RefreshToken, error) {
	args := m.Called(token)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.RefreshToken), args.Error(1)
}

func (m *MockRefreshTokenRepository) DeleteByToken(token string) error {
	args := m.Called(token)
	return args.Error(0)
}

func (m *MockRefreshTokenRepository) DeleteAllByUserID(userID uuid.UUID) error {
	args := m.Called(userID)
	return args.Error(0)
}

// Helper function
func setupAuthService() (*authService, *MockUserRepository, *MockRefreshTokenRepository) {
	mockUserRepo := new(MockUserRepository)
	mockRefreshRepo := new(MockRefreshTokenRepository)
	jwtManager := jwt.NewManager("test-secret", time.Hour, time.Hour*24)

	service := &authService{
		userRepo:         mockUserRepo,
		refreshTokenRepo: mockRefreshRepo,
		jwtManager:       jwtManager,
	}

	return service, mockUserRepo, mockRefreshRepo
}

// Test Register - Success
func TestRegister_Success(t *testing.T) {
	service, mockUserRepo, mockRefreshRepo := setupAuthService()

	// Arrange
	email := "test@example.com"
	password := "password123"
	name := "Test User"
	phone := "1234567890"
	role := domain.UserRoleCustomer

	mockUserRepo.On("GetByEmail", tmock.Anything, email).Return(nil, gorm.ErrRecordNotFound)
	mockUserRepo.On("Create", tmock.Anything, tmock.AnythingOfType("*domain.User")).Return(nil)
	mockRefreshRepo.On("Create", tmock.AnythingOfType("*domain.RefreshToken")).Return(nil)

	// Act
	user, accessToken, refreshToken, err := service.Register(email, password, name, &phone, role)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, user)
	assert.NotEmpty(t, accessToken)
	assert.NotEmpty(t, refreshToken)
	assert.Equal(t, email, user.Email)
	assert.NotEmpty(t, user.FirstName)
	assert.Equal(t, phone, user.Phone)
	assert.Equal(t, role, user.Role)

	mockUserRepo.AssertExpectations(t)
	mockRefreshRepo.AssertExpectations(t)
}

// Test Register - Invalid Email
func TestRegister_InvalidEmail(t *testing.T) {
	service, _, _ := setupAuthService()

	phone := "1234567890"
	// Act
	_, _, _, err := service.Register("invalid-email", "password123", "Test User", &phone, domain.UserRoleCustomer)

	// Assert
	assert.Error(t, err)
	assert.Equal(t, ErrInvalidEmail, err)
}

// Test Register - Short Password
func TestRegister_ShortPassword(t *testing.T) {
	service, _, _ := setupAuthService()

	phone := "1234567890"
	// Act
	_, _, _, err := service.Register("test@example.com", "short", "Test User", &phone, domain.UserRoleCustomer)

	// Assert
	assert.Error(t, err)
	assert.Equal(t, ErrInvalidPassword, err)
}

// Test Register - Email Already Exists
func TestRegister_EmailExists(t *testing.T) {
	service, mockUserRepo, _ := setupAuthService()

	existingUser := &domain.User{
		ID:    uuid.New(),
		Email: "test@example.com",
	}

	mockUserRepo.On("GetByEmail", tmock.Anything, "test@example.com").Return(existingUser, nil)

	phone := "1234567890"
	// Act
	_, _, _, err := service.Register("test@example.com", "password123", "Test User", &phone, domain.UserRoleCustomer)

	// Assert
	assert.Error(t, err)
	assert.Equal(t, ErrEmailExists, err)
	mockUserRepo.AssertExpectations(t)
}

// Test Login - Success
func TestLogin_Success(t *testing.T) {
	service, mockUserRepo, mockRefreshRepo := setupAuthService()

	// Arrange
	email := "test@example.com"
	password := "password123"
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)

	existingUser := &domain.User{
		ID:       uuid.New(),
		Email:    email,
		Password: string(hashedPassword),
		Role:     domain.UserRoleCustomer,
	}

	mockUserRepo.On("GetByEmail", tmock.Anything, email).Return(existingUser, nil)
	mockRefreshRepo.On("Create", tmock.AnythingOfType("*domain.RefreshToken")).Return(nil)

	// Act
	accessToken, refreshToken, user, err := service.Login(email, password)

	// Assert
	assert.NoError(t, err)
	assert.NotEmpty(t, accessToken)
	assert.NotEmpty(t, refreshToken)
	assert.NotNil(t, user)
	assert.Equal(t, email, user.Email)

	mockUserRepo.AssertExpectations(t)
	mockRefreshRepo.AssertExpectations(t)
}

// Test Login - User Not Found
func TestLogin_UserNotFound(t *testing.T) {
	service, mockUserRepo, _ := setupAuthService()

	mockUserRepo.On("GetByEmail", tmock.Anything, "nonexistent@example.com").Return(nil, gorm.ErrRecordNotFound)

	// Act
	_, _, _, err := service.Login("nonexistent@example.com", "password123")

	// Assert
	assert.Error(t, err)
	assert.Equal(t, ErrInvalidCredentials, err)
	mockUserRepo.AssertExpectations(t)
}

// Test Login - Wrong Password
func TestLogin_WrongPassword(t *testing.T) {
	service, mockUserRepo, _ := setupAuthService()

	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("correctpassword"), bcrypt.DefaultCost)
	existingUser := &domain.User{
		ID:       uuid.New(),
		Email:    "test@example.com",
		Password: string(hashedPassword),
		Role:     domain.UserRoleCustomer,
	}

	mockUserRepo.On("GetByEmail", tmock.Anything, "test@example.com").Return(existingUser, nil)

	// Act
	_, _, _, err := service.Login("test@example.com", "wrongpassword")

	// Assert
	assert.Error(t, err)
	assert.Equal(t, ErrInvalidCredentials, err)
	mockUserRepo.AssertExpectations(t)
}

// Test RefreshToken - Success
func TestRefreshToken_Success(t *testing.T) {
	service, mockUserRepo, mockRefreshRepo := setupAuthService()

	// Arrange
	userID := uuid.New()
	oldRefreshToken := "old-refresh-token"

	existingUser := &domain.User{
		ID:   userID,
		Role: domain.UserRoleCustomer,
	}

	existingRefreshToken := &domain.RefreshToken{
		ID:        uuid.New(),
		UserID:    userID,
		Token:     oldRefreshToken,
		ExpiresAt: time.Now().Add(time.Hour * 24),
	}

	mockRefreshRepo.On("GetByToken", oldRefreshToken).Return(existingRefreshToken, nil)
	mockUserRepo.On("GetByID", tmock.Anything, userID).Return(existingUser, nil)
	mockRefreshRepo.On("DeleteByToken", oldRefreshToken).Return(nil)
	mockRefreshRepo.On("Create", tmock.AnythingOfType("*domain.RefreshToken")).Return(nil)

	// Act
	newAccessToken, newRefreshToken, err := service.RefreshToken(oldRefreshToken)

	// Assert
	assert.NoError(t, err)
	assert.NotEmpty(t, newAccessToken)
	assert.NotEmpty(t, newRefreshToken)
	assert.NotEqual(t, oldRefreshToken, newRefreshToken)

	mockRefreshRepo.AssertExpectations(t)
	mockUserRepo.AssertExpectations(t)
}

// Test RefreshToken - Token Not Found
func TestRefreshToken_TokenNotFound(t *testing.T) {
	service, _, mockRefreshRepo := setupAuthService()

	mockRefreshRepo.On("GetByToken", "invalid-token").Return(nil, gorm.ErrRecordNotFound)

	// Act
	_, _, err := service.RefreshToken("invalid-token")

	// Assert
	assert.Error(t, err)
	assert.Equal(t, ErrInvalidRefreshToken, err)
	mockRefreshRepo.AssertExpectations(t)
}

// Test RefreshToken - Expired Token
func TestRefreshToken_ExpiredToken(t *testing.T) {
	service, _, mockRefreshRepo := setupAuthService()

	expiredToken := &domain.RefreshToken{
		ID:        uuid.New(),
		UserID:    uuid.New(),
		Token:     "expired-token",
		ExpiresAt: time.Now().Add(-time.Hour), // Expired 1 hour ago
	}

	mockRefreshRepo.On("GetByToken", "expired-token").Return(expiredToken, nil)
	mockRefreshRepo.On("DeleteByToken", "expired-token").Return(nil)

	// Act
	_, _, err := service.RefreshToken("expired-token")

	// Assert
	assert.Error(t, err)
	assert.Equal(t, ErrExpiredRefreshToken, err)
	mockRefreshRepo.AssertExpectations(t)
}

// Test Logout - Success
func TestLogout_Success(t *testing.T) {
	service, _, mockRefreshRepo := setupAuthService()

	refreshToken := "test-refresh-token"
	mockRefreshRepo.On("DeleteByToken", refreshToken).Return(nil)

	// Act
	err := service.Logout(refreshToken)

	// Assert
	assert.NoError(t, err)
	mockRefreshRepo.AssertExpectations(t)
}

// Test Logout - Token Not Found (should not fail)
func TestLogout_TokenNotFound(t *testing.T) {
	service, _, mockRefreshRepo := setupAuthService()

	mockRefreshRepo.On("DeleteByToken", "nonexistent-token").Return(gorm.ErrRecordNotFound)

	// Act
	err := service.Logout("nonexistent-token")

	// Assert
	assert.Error(t, err)
	assert.Equal(t, gorm.ErrRecordNotFound, err)
	mockRefreshRepo.AssertExpectations(t)
}

// Test isValidEmail
func TestIsValidEmail(t *testing.T) {
	tests := []struct {
		email    string
		expected bool
	}{
		{"test@example.com", true},
		{"user.name@example.co.uk", true},
		{"user+tag@example.com", true},
		{"invalid-email", false},
		{"@example.com", false},
		{"test@", false},
		{"test@.com", false},
		{"", false},
	}

	for _, tt := range tests {
		t.Run(tt.email, func(t *testing.T) {
			result := isValidEmail(tt.email)
			assert.Equal(t, tt.expected, result)
		})
	}
}
