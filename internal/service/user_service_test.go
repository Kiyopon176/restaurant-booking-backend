package service

import (
	"restaurant-booking/internal/domain"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// Mock UserRepository for user service tests
type MockUserRepositoryForUserService struct {
	mock.Mock
}

func (m *MockUserRepositoryForUserService) Create(user *domain.User) error {
	args := m.Called(user)
	return args.Error(0)
}

func (m *MockUserRepositoryForUserService) GetByID(id uuid.UUID) (*domain.User, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.User), args.Error(1)
}

func (m *MockUserRepositoryForUserService) GetByEmail(email string) (*domain.User, error) {
	args := m.Called(email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.User), args.Error(1)
}

func (m *MockUserRepositoryForUserService) Update(user *domain.User) error {
	args := m.Called(user)
	return args.Error(0)
}

func (m *MockUserRepositoryForUserService) Delete(id uuid.UUID) error {
	args := m.Called(id)
	return args.Error(0)
}

func setupUserService() (UserService, *MockUserRepositoryForUserService) {
	mockUserRepo := new(MockUserRepositoryForUserService)
	service := NewUserService(mockUserRepo)
	return service, mockUserRepo
}

// Test GetUserByID - Success
func TestGetUserByID_Success(t *testing.T) {
	service, mockUserRepo := setupUserService()

	// Arrange
	userID := uuid.New()
	expectedUser := &domain.User{
		ID:        userID,
		Email:     "test@example.com",
		FirstName: "John",
		LastName:  "Doe",
		Phone:     "1234567890",
		Role:      domain.UserRoleCustomer,
	}

	mockUserRepo.On("GetByID", userID).Return(expectedUser, nil)

	// Act
	user, err := service.GetUserByID(userID)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, user)
	assert.Equal(t, userID, user.ID)
	assert.Equal(t, "test@example.com", user.Email)
	assert.Equal(t, "John", user.FirstName)
	assert.Equal(t, "Doe", user.LastName)

	mockUserRepo.AssertExpectations(t)
}

// Test GetUserByID - User Not Found
func TestGetUserByID_UserNotFound(t *testing.T) {
	service, mockUserRepo := setupUserService()

	// Arrange
	userID := uuid.New()
	mockUserRepo.On("GetByID", userID).Return(nil, gorm.ErrRecordNotFound)

	// Act
	user, err := service.GetUserByID(userID)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, user)
	assert.Equal(t, ErrUserNotFound, err)

	mockUserRepo.AssertExpectations(t)
}

// Test UpdateUser - Success
func TestUpdateUser_Success(t *testing.T) {
	service, mockUserRepo := setupUserService()

	// Arrange
	userID := uuid.New()
	existingUser := &domain.User{
		ID:        userID,
		Email:     "test@example.com",
		FirstName: "John",
		LastName:  "Doe",
		Phone:     "1234567890",
		Role:      domain.UserRoleCustomer,
	}

	updatedFirstName := "Jane"
	updatedLastName := "Smith"
	updatedPhone := "0987654321"

	mockUserRepo.On("GetByID", userID).Return(existingUser, nil)
	mockUserRepo.On("Update", mock.MatchedBy(func(u *domain.User) bool {
		return u.ID == userID &&
			u.FirstName == updatedFirstName &&
			u.LastName == updatedLastName &&
			u.Phone == updatedPhone
	})).Return(nil)

	// Act
	user, err := service.UpdateUser(userID, updatedFirstName, updatedLastName, updatedPhone)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, user)
	assert.Equal(t, updatedFirstName, user.FirstName)
	assert.Equal(t, updatedLastName, user.LastName)
	assert.Equal(t, updatedPhone, user.Phone)

	mockUserRepo.AssertExpectations(t)
}

// Test UpdateUser - User Not Found
func TestUpdateUser_UserNotFound(t *testing.T) {
	service, mockUserRepo := setupUserService()

	// Arrange
	userID := uuid.New()
	mockUserRepo.On("GetByID", userID).Return(nil, gorm.ErrRecordNotFound)

	// Act
	user, err := service.UpdateUser(userID, "Jane", "Smith", "0987654321")

	// Assert
	assert.Error(t, err)
	assert.Nil(t, user)
	assert.Equal(t, ErrUserNotFound, err)

	mockUserRepo.AssertExpectations(t)
}

// Test ChangePassword - Success
func TestChangePassword_Success(t *testing.T) {
	service, mockUserRepo := setupUserService()

	// Arrange
	userID := uuid.New()
	oldPassword := "oldpassword123"
	newPassword := "newpassword456"

	hashedOldPassword, _ := bcrypt.GenerateFromPassword([]byte(oldPassword), bcrypt.DefaultCost)
	existingUser := &domain.User{
		ID:       userID,
		Email:    "test@example.com",
		Password: string(hashedOldPassword),
		Role:     domain.UserRoleCustomer,
	}

	mockUserRepo.On("GetByID", userID).Return(existingUser, nil)
	mockUserRepo.On("Update", mock.MatchedBy(func(u *domain.User) bool {
		// Verify that password was hashed
		err := bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(newPassword))
		return err == nil && u.ID == userID
	})).Return(nil)

	// Act
	err := service.ChangePassword(userID, oldPassword, newPassword)

	// Assert
	assert.NoError(t, err)

	mockUserRepo.AssertExpectations(t)
}

// Test ChangePassword - User Not Found
func TestChangePassword_UserNotFound(t *testing.T) {
	service, mockUserRepo := setupUserService()

	// Arrange
	userID := uuid.New()
	mockUserRepo.On("GetByID", userID).Return(nil, gorm.ErrRecordNotFound)

	// Act
	err := service.ChangePassword(userID, "oldpassword", "newpassword123")

	// Assert
	assert.Error(t, err)
	assert.Equal(t, ErrUserNotFound, err)

	mockUserRepo.AssertExpectations(t)
}

// Test ChangePassword - Wrong Old Password
func TestChangePassword_WrongOldPassword(t *testing.T) {
	service, mockUserRepo := setupUserService()

	// Arrange
	userID := uuid.New()
	correctOldPassword := "correctpassword"
	wrongOldPassword := "wrongpassword"

	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(correctOldPassword), bcrypt.DefaultCost)
	existingUser := &domain.User{
		ID:       userID,
		Email:    "test@example.com",
		Password: string(hashedPassword),
		Role:     domain.UserRoleCustomer,
	}

	mockUserRepo.On("GetByID", userID).Return(existingUser, nil)

	// Act
	err := service.ChangePassword(userID, wrongOldPassword, "newpassword123")

	// Assert
	assert.Error(t, err)
	assert.Equal(t, ErrOldPasswordIncorrect, err)

	mockUserRepo.AssertExpectations(t)
}

// Test ChangePassword - Short New Password
func TestChangePassword_ShortNewPassword(t *testing.T) {
	service, mockUserRepo := setupUserService()

	// Arrange
	userID := uuid.New()
	oldPassword := "oldpassword123"

	hashedOldPassword, _ := bcrypt.GenerateFromPassword([]byte(oldPassword), bcrypt.DefaultCost)
	existingUser := &domain.User{
		ID:       userID,
		Email:    "test@example.com",
		Password: string(hashedOldPassword),
		Role:     domain.UserRoleCustomer,
	}

	mockUserRepo.On("GetByID", userID).Return(existingUser, nil)

	// Act
	err := service.ChangePassword(userID, oldPassword, "short")

	// Assert
	assert.Error(t, err)
	assert.Equal(t, ErrInvalidPassword, err)

	mockUserRepo.AssertExpectations(t)
}
