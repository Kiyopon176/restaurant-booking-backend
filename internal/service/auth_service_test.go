package service

import (
	_ "errors"
	"restaurant-booking/internal/domain"
	"restaurant-booking/pkg/jwt"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type MockUserRepository struct {
	mock.Mock
}

func (m *MockUserRepository) Create(user *domain.User) error {
	args := m.Called(user)
	return args.Error(0)
}

func (m *MockUserRepository) GetByID(id uuid.UUID) (*domain.User, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.User), args.Error(1)
}

func (m *MockUserRepository) GetByEmail(email string) (*domain.User, error) {
	args := m.Called(email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.User), args.Error(1)
}

func (m *MockUserRepository) Update(user *domain.User) error {
	args := m.Called(user)
	return args.Error(0)
}

func (m *MockUserRepository) Delete(id uuid.UUID) error {
	args := m.Called(id)
	return args.Error(0)
}

type MockRefreshTokenRepository struct {
	mock.Mock
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

func TestRegister_Success(t *testing.T) {
	service, mockUserRepo, mockRefreshRepo := setupAuthService()

	email := "test@example.com"
	password := "password123"
	firstName := "Test"
	lastName := "User"
	phone := "1234567890"
	role := domain.UserRoleCustomer

	mockUserRepo.On("GetByEmail", email).Return(nil, gorm.ErrRecordNotFound)
	mockUserRepo.On("Create", mock.AnythingOfType("*domain.User")).Return(nil)
	mockRefreshRepo.On("Create", mock.AnythingOfType("*domain.RefreshToken")).Return(nil)

	user, accessToken, refreshToken, err := service.Register(email, password, firstName, lastName, phone, role)

	assert.NoError(t, err)
	assert.NotNil(t, user)
	assert.NotEmpty(t, accessToken)
	assert.NotEmpty(t, refreshToken)
	assert.Equal(t, email, user.Email)
	assert.Equal(t, firstName, user.FirstName)
	assert.Equal(t, lastName, user.LastName)
	assert.Equal(t, phone, user.Phone)
	assert.Equal(t, role, user.Role)

	mockUserRepo.AssertExpectations(t)
	mockRefreshRepo.AssertExpectations(t)
}

func TestRegister_InvalidEmail(t *testing.T) {
	service, _, _ := setupAuthService()

	_, _, _, err := service.Register("invalid-email", "password123", "Test", "User", "1234567890", domain.UserRoleCustomer)

	assert.Error(t, err)
	assert.Equal(t, ErrInvalidEmail, err)
}

func TestRegister_ShortPassword(t *testing.T) {
	service, _, _ := setupAuthService()

	_, _, _, err := service.Register("test@example.com", "short", "Test", "User", "1234567890", domain.UserRoleCustomer)

	assert.Error(t, err)
	assert.Equal(t, ErrInvalidPassword, err)
}

func TestRegister_EmailExists(t *testing.T) {
	service, mockUserRepo, _ := setupAuthService()

	existingUser := &domain.User{
		ID:    uuid.New(),
		Email: "test@example.com",
	}

	mockUserRepo.On("GetByEmail", "test@example.com").Return(existingUser, nil)

	_, _, _, err := service.Register("test@example.com", "password123", "Test", "User", "1234567890", domain.UserRoleCustomer)

	assert.Error(t, err)
	assert.Equal(t, ErrEmailExists, err)
	mockUserRepo.AssertExpectations(t)
}

func TestLogin_Success(t *testing.T) {
	service, mockUserRepo, mockRefreshRepo := setupAuthService()

	email := "test@example.com"
	password := "password123"
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)

	existingUser := &domain.User{
		ID:       uuid.New(),
		Email:    email,
		Password: string(hashedPassword),
		Role:     domain.UserRoleCustomer,
	}

	mockUserRepo.On("GetByEmail", email).Return(existingUser, nil)
	mockRefreshRepo.On("Create", mock.AnythingOfType("*domain.RefreshToken")).Return(nil)

	accessToken, refreshToken, user, err := service.Login(email, password)

	assert.NoError(t, err)
	assert.NotEmpty(t, accessToken)
	assert.NotEmpty(t, refreshToken)
	assert.NotNil(t, user)
	assert.Equal(t, email, user.Email)

	mockUserRepo.AssertExpectations(t)
	mockRefreshRepo.AssertExpectations(t)
}

func TestLogin_UserNotFound(t *testing.T) {
	service, mockUserRepo, _ := setupAuthService()

	mockUserRepo.On("GetByEmail", "nonexistent@example.com").Return(nil, gorm.ErrRecordNotFound)

	_, _, _, err := service.Login("nonexistent@example.com", "password123")

	assert.Error(t, err)
	assert.Equal(t, ErrInvalidCredentials, err)
	mockUserRepo.AssertExpectations(t)
}

func TestLogin_WrongPassword(t *testing.T) {
	service, mockUserRepo, _ := setupAuthService()

	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("correctpassword"), bcrypt.DefaultCost)
	existingUser := &domain.User{
		ID:       uuid.New(),
		Email:    "test@example.com",
		Password: string(hashedPassword),
		Role:     domain.UserRoleCustomer,
	}

	mockUserRepo.On("GetByEmail", "test@example.com").Return(existingUser, nil)

	_, _, _, err := service.Login("test@example.com", "wrongpassword")

	assert.Error(t, err)
	assert.Equal(t, ErrInvalidCredentials, err)
	mockUserRepo.AssertExpectations(t)
}

func TestRefreshToken_Success(t *testing.T) {
	service, mockUserRepo, mockRefreshRepo := setupAuthService()

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
	mockUserRepo.On("GetByID", userID).Return(existingUser, nil)
	mockRefreshRepo.On("DeleteByToken", oldRefreshToken).Return(nil)
	mockRefreshRepo.On("Create", mock.AnythingOfType("*domain.RefreshToken")).Return(nil)

	newAccessToken, newRefreshToken, err := service.RefreshToken(oldRefreshToken)

	assert.NoError(t, err)
	assert.NotEmpty(t, newAccessToken)
	assert.NotEmpty(t, newRefreshToken)
	assert.NotEqual(t, oldRefreshToken, newRefreshToken)

	mockRefreshRepo.AssertExpectations(t)
	mockUserRepo.AssertExpectations(t)
}

func TestRefreshToken_TokenNotFound(t *testing.T) {
	service, _, mockRefreshRepo := setupAuthService()

	mockRefreshRepo.On("GetByToken", "invalid-token").Return(nil, gorm.ErrRecordNotFound)

	_, _, err := service.RefreshToken("invalid-token")

	assert.Error(t, err)
	assert.Equal(t, ErrInvalidRefreshToken, err)
	mockRefreshRepo.AssertExpectations(t)
}

func TestRefreshToken_ExpiredToken(t *testing.T) {
	service, _, mockRefreshRepo := setupAuthService()

	expiredToken := &domain.RefreshToken{
		ID:        uuid.New(),
		UserID:    uuid.New(),
		Token:     "expired-token",
		ExpiresAt: time.Now().Add(-time.Hour),
	}

	mockRefreshRepo.On("GetByToken", "expired-token").Return(expiredToken, nil)
	mockRefreshRepo.On("DeleteByToken", "expired-token").Return(nil)

	_, _, err := service.RefreshToken("expired-token")

	assert.Error(t, err)
	assert.Equal(t, ErrExpiredRefreshToken, err)
	mockRefreshRepo.AssertExpectations(t)
}

func TestLogout_Success(t *testing.T) {
	service, _, mockRefreshRepo := setupAuthService()

	refreshToken := "test-refresh-token"
	mockRefreshRepo.On("DeleteByToken", refreshToken).Return(nil)

	err := service.Logout(refreshToken)

	assert.NoError(t, err)
	mockRefreshRepo.AssertExpectations(t)
}

func TestLogout_TokenNotFound(t *testing.T) {
	service, _, mockRefreshRepo := setupAuthService()

	mockRefreshRepo.On("DeleteByToken", "nonexistent-token").Return(gorm.ErrRecordNotFound)

	err := service.Logout("nonexistent-token")

	assert.Error(t, err)
	assert.Equal(t, gorm.ErrRecordNotFound, err)
	mockRefreshRepo.AssertExpectations(t)
}

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
