package service

import (
	"context"
	"errors"
	"restaurant-booking/internal/domain"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gorm.io/gorm"
)

// MockRestaurantManagerRepository is a mock implementation of RestaurantManagerRepository
type MockRestaurantManagerRepository struct {
	mock.Mock
}

func (m *MockRestaurantManagerRepository) Create(ctx context.Context, manager *domain.RestaurantManager) error {
	args := m.Called(ctx, manager)
	return args.Error(0)
}

func (m *MockRestaurantManagerRepository) Delete(ctx context.Context, userID, restaurantID uuid.UUID) error {
	args := m.Called(ctx, userID, restaurantID)
	return args.Error(0)
}

func (m *MockRestaurantManagerRepository) GetManagersByRestaurant(ctx context.Context, restaurantID uuid.UUID) ([]*domain.RestaurantManager, error) {
	args := m.Called(ctx, restaurantID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.RestaurantManager), args.Error(1)
}

func (m *MockRestaurantManagerRepository) GetRestaurantsByManager(ctx context.Context, userID uuid.UUID) ([]*domain.RestaurantManager, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.RestaurantManager), args.Error(1)
}

func (m *MockRestaurantManagerRepository) IsManager(ctx context.Context, userID, restaurantID uuid.UUID) (bool, error) {
	args := m.Called(ctx, userID, restaurantID)
	return args.Bool(0), args.Error(1)
}

// setupManagerService creates a manager service instance with mock repositories
func setupManagerService() (*managerService, *MockRestaurantManagerRepository, *MockRestaurantRepository, *MockUserRepository) {
	mockManagerRepo := new(MockRestaurantManagerRepository)
	mockRestaurantRepo := new(MockRestaurantRepository)
	mockUserRepo := new(MockUserRepository)

	service := &managerService{
		managerRepo:    mockManagerRepo,
		restaurantRepo: mockRestaurantRepo,
		userRepo:       mockUserRepo,
	}

	return service, mockManagerRepo, mockRestaurantRepo, mockUserRepo
}

// TestAddManager_Success tests successful addition of a manager
func TestAddManager_Success(t *testing.T) {
	service, mockManagerRepo, mockRestaurantRepo, mockUserRepo := setupManagerService()
	ctx := context.Background()

	restaurantID := uuid.New()
	ownerID := uuid.New()
	userID := uuid.New()

	restaurant := &domain.Restaurant{
		ID:      restaurantID,
		OwnerID: ownerID,
		Name:    "Test Restaurant",
	}

	user := &domain.User{
		ID:        userID,
		Email:     "manager@test.com",
		FirstName: "Test",
		LastName:  "Manager",
	}

	req := AddManagerRequest{
		UserID: userID,
	}

	mockRestaurantRepo.On("GetByID", ctx, restaurantID).Return(restaurant, nil)
	mockUserRepo.On("GetByID", userID).Return(user, nil)
	mockManagerRepo.On("IsManager", ctx, userID, restaurantID).Return(false, nil)
	mockManagerRepo.On("Create", ctx, mock.AnythingOfType("*domain.RestaurantManager")).Return(nil)

	result, err := service.AddManager(ctx, restaurantID, ownerID, req)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, userID, result.UserID)
	assert.Equal(t, restaurantID, result.RestaurantID)
	assert.Equal(t, user, result.User)
	assert.Equal(t, restaurant, result.Restaurant)

	mockRestaurantRepo.AssertExpectations(t)
	mockUserRepo.AssertExpectations(t)
	mockManagerRepo.AssertExpectations(t)
}

// TestAddManager_RestaurantNotFound tests adding manager when restaurant doesn't exist
func TestAddManager_RestaurantNotFound(t *testing.T) {
	service, _, mockRestaurantRepo, _ := setupManagerService()
	ctx := context.Background()

	restaurantID := uuid.New()
	ownerID := uuid.New()
	userID := uuid.New()

	req := AddManagerRequest{
		UserID: userID,
	}

	mockRestaurantRepo.On("GetByID", ctx, restaurantID).Return(nil, gorm.ErrRecordNotFound)

	result, err := service.AddManager(ctx, restaurantID, ownerID, req)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Equal(t, ErrRestaurantNotFound, err)

	mockRestaurantRepo.AssertExpectations(t)
}

// TestAddManager_RestaurantGetError tests adding manager when getting restaurant fails
func TestAddManager_RestaurantGetError(t *testing.T) {
	service, _, mockRestaurantRepo, _ := setupManagerService()
	ctx := context.Background()

	restaurantID := uuid.New()
	ownerID := uuid.New()
	userID := uuid.New()

	req := AddManagerRequest{
		UserID: userID,
	}

	dbError := errors.New("database error")
	mockRestaurantRepo.On("GetByID", ctx, restaurantID).Return(nil, dbError)

	result, err := service.AddManager(ctx, restaurantID, ownerID, req)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Equal(t, dbError, err)

	mockRestaurantRepo.AssertExpectations(t)
}

// TestAddManager_Unauthorized tests adding manager when user is not the owner
func TestAddManager_Unauthorized(t *testing.T) {
	service, _, mockRestaurantRepo, _ := setupManagerService()
	ctx := context.Background()

	restaurantID := uuid.New()
	ownerID := uuid.New()
	differentOwnerID := uuid.New()
	userID := uuid.New()

	restaurant := &domain.Restaurant{
		ID:      restaurantID,
		OwnerID: differentOwnerID,
		Name:    "Test Restaurant",
	}

	req := AddManagerRequest{
		UserID: userID,
	}

	mockRestaurantRepo.On("GetByID", ctx, restaurantID).Return(restaurant, nil)

	result, err := service.AddManager(ctx, restaurantID, ownerID, req)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Equal(t, ErrUnauthorized, err)

	mockRestaurantRepo.AssertExpectations(t)
}

// TestAddManager_UserNotFound tests adding manager when user doesn't exist
func TestAddManager_UserNotFound(t *testing.T) {
	service, _, mockRestaurantRepo, mockUserRepo := setupManagerService()
	ctx := context.Background()

	restaurantID := uuid.New()
	ownerID := uuid.New()
	userID := uuid.New()

	restaurant := &domain.Restaurant{
		ID:      restaurantID,
		OwnerID: ownerID,
		Name:    "Test Restaurant",
	}

	req := AddManagerRequest{
		UserID: userID,
	}

	mockRestaurantRepo.On("GetByID", ctx, restaurantID).Return(restaurant, nil)
	mockUserRepo.On("GetByID", userID).Return(nil, gorm.ErrRecordNotFound)

	result, err := service.AddManager(ctx, restaurantID, ownerID, req)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Equal(t, ErrUserNotFound, err)

	mockRestaurantRepo.AssertExpectations(t)
	mockUserRepo.AssertExpectations(t)
}

// TestAddManager_UserGetError tests adding manager when getting user fails
func TestAddManager_UserGetError(t *testing.T) {
	service, _, mockRestaurantRepo, mockUserRepo := setupManagerService()
	ctx := context.Background()

	restaurantID := uuid.New()
	ownerID := uuid.New()
	userID := uuid.New()

	restaurant := &domain.Restaurant{
		ID:      restaurantID,
		OwnerID: ownerID,
		Name:    "Test Restaurant",
	}

	req := AddManagerRequest{
		UserID: userID,
	}

	dbError := errors.New("database error")
	mockRestaurantRepo.On("GetByID", ctx, restaurantID).Return(restaurant, nil)
	mockUserRepo.On("GetByID", userID).Return(nil, dbError)

	result, err := service.AddManager(ctx, restaurantID, ownerID, req)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Equal(t, dbError, err)

	mockRestaurantRepo.AssertExpectations(t)
	mockUserRepo.AssertExpectations(t)
}

// TestAddManager_AlreadyManager tests adding a user who is already a manager
func TestAddManager_AlreadyManager(t *testing.T) {
	service, mockManagerRepo, mockRestaurantRepo, mockUserRepo := setupManagerService()
	ctx := context.Background()

	restaurantID := uuid.New()
	ownerID := uuid.New()
	userID := uuid.New()

	restaurant := &domain.Restaurant{
		ID:      restaurantID,
		OwnerID: ownerID,
		Name:    "Test Restaurant",
	}

	user := &domain.User{
		ID:        userID,
		Email:     "manager@test.com",
		FirstName: "Test",
		LastName:  "Manager",
	}

	req := AddManagerRequest{
		UserID: userID,
	}

	mockRestaurantRepo.On("GetByID", ctx, restaurantID).Return(restaurant, nil)
	mockUserRepo.On("GetByID", userID).Return(user, nil)
	mockManagerRepo.On("IsManager", ctx, userID, restaurantID).Return(true, nil)

	result, err := service.AddManager(ctx, restaurantID, ownerID, req)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Equal(t, ErrManagerAlreadyExists, err)

	mockRestaurantRepo.AssertExpectations(t)
	mockUserRepo.AssertExpectations(t)
	mockManagerRepo.AssertExpectations(t)
}

// TestAddManager_IsManagerError tests adding manager when checking existing manager fails
func TestAddManager_IsManagerError(t *testing.T) {
	service, mockManagerRepo, mockRestaurantRepo, mockUserRepo := setupManagerService()
	ctx := context.Background()

	restaurantID := uuid.New()
	ownerID := uuid.New()
	userID := uuid.New()

	restaurant := &domain.Restaurant{
		ID:      restaurantID,
		OwnerID: ownerID,
		Name:    "Test Restaurant",
	}

	user := &domain.User{
		ID:        userID,
		Email:     "manager@test.com",
		FirstName: "Test",
		LastName:  "Manager",
	}

	req := AddManagerRequest{
		UserID: userID,
	}

	dbError := errors.New("database error")
	mockRestaurantRepo.On("GetByID", ctx, restaurantID).Return(restaurant, nil)
	mockUserRepo.On("GetByID", userID).Return(user, nil)
	mockManagerRepo.On("IsManager", ctx, userID, restaurantID).Return(false, dbError)

	result, err := service.AddManager(ctx, restaurantID, ownerID, req)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Equal(t, dbError, err)

	mockRestaurantRepo.AssertExpectations(t)
	mockUserRepo.AssertExpectations(t)
	mockManagerRepo.AssertExpectations(t)
}

// TestAddManager_CreateError tests adding manager when creation fails
func TestAddManager_CreateError(t *testing.T) {
	service, mockManagerRepo, mockRestaurantRepo, mockUserRepo := setupManagerService()
	ctx := context.Background()

	restaurantID := uuid.New()
	ownerID := uuid.New()
	userID := uuid.New()

	restaurant := &domain.Restaurant{
		ID:      restaurantID,
		OwnerID: ownerID,
		Name:    "Test Restaurant",
	}

	user := &domain.User{
		ID:        userID,
		Email:     "manager@test.com",
		FirstName: "Test",
		LastName:  "Manager",
	}

	req := AddManagerRequest{
		UserID: userID,
	}

	dbError := errors.New("database error")
	mockRestaurantRepo.On("GetByID", ctx, restaurantID).Return(restaurant, nil)
	mockUserRepo.On("GetByID", userID).Return(user, nil)
	mockManagerRepo.On("IsManager", ctx, userID, restaurantID).Return(false, nil)
	mockManagerRepo.On("Create", ctx, mock.AnythingOfType("*domain.RestaurantManager")).Return(dbError)

	result, err := service.AddManager(ctx, restaurantID, ownerID, req)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Equal(t, dbError, err)

	mockRestaurantRepo.AssertExpectations(t)
	mockUserRepo.AssertExpectations(t)
	mockManagerRepo.AssertExpectations(t)
}

// TestRemoveManager_Success tests successful removal of a manager
func TestRemoveManager_Success(t *testing.T) {
	service, mockManagerRepo, mockRestaurantRepo, _ := setupManagerService()
	ctx := context.Background()

	restaurantID := uuid.New()
	ownerID := uuid.New()
	userID := uuid.New()

	restaurant := &domain.Restaurant{
		ID:      restaurantID,
		OwnerID: ownerID,
		Name:    "Test Restaurant",
	}

	mockRestaurantRepo.On("GetByID", ctx, restaurantID).Return(restaurant, nil)
	mockManagerRepo.On("IsManager", ctx, userID, restaurantID).Return(true, nil)
	mockManagerRepo.On("Delete", ctx, userID, restaurantID).Return(nil)

	err := service.RemoveManager(ctx, restaurantID, ownerID, userID)

	assert.NoError(t, err)

	mockRestaurantRepo.AssertExpectations(t)
	mockManagerRepo.AssertExpectations(t)
}

// TestRemoveManager_RestaurantNotFound tests removing manager when restaurant doesn't exist
func TestRemoveManager_RestaurantNotFound(t *testing.T) {
	service, _, mockRestaurantRepo, _ := setupManagerService()
	ctx := context.Background()

	restaurantID := uuid.New()
	ownerID := uuid.New()
	userID := uuid.New()

	mockRestaurantRepo.On("GetByID", ctx, restaurantID).Return(nil, gorm.ErrRecordNotFound)

	err := service.RemoveManager(ctx, restaurantID, ownerID, userID)

	assert.Error(t, err)
	assert.Equal(t, ErrRestaurantNotFound, err)

	mockRestaurantRepo.AssertExpectations(t)
}

// TestRemoveManager_RestaurantGetError tests removing manager when getting restaurant fails
func TestRemoveManager_RestaurantGetError(t *testing.T) {
	service, _, mockRestaurantRepo, _ := setupManagerService()
	ctx := context.Background()

	restaurantID := uuid.New()
	ownerID := uuid.New()
	userID := uuid.New()

	dbError := errors.New("database error")
	mockRestaurantRepo.On("GetByID", ctx, restaurantID).Return(nil, dbError)

	err := service.RemoveManager(ctx, restaurantID, ownerID, userID)

	assert.Error(t, err)
	assert.Equal(t, dbError, err)

	mockRestaurantRepo.AssertExpectations(t)
}

// TestRemoveManager_Unauthorized tests removing manager when user is not the owner
func TestRemoveManager_Unauthorized(t *testing.T) {
	service, _, mockRestaurantRepo, _ := setupManagerService()
	ctx := context.Background()

	restaurantID := uuid.New()
	ownerID := uuid.New()
	differentOwnerID := uuid.New()
	userID := uuid.New()

	restaurant := &domain.Restaurant{
		ID:      restaurantID,
		OwnerID: differentOwnerID,
		Name:    "Test Restaurant",
	}

	mockRestaurantRepo.On("GetByID", ctx, restaurantID).Return(restaurant, nil)

	err := service.RemoveManager(ctx, restaurantID, ownerID, userID)

	assert.Error(t, err)
	assert.Equal(t, ErrUnauthorized, err)

	mockRestaurantRepo.AssertExpectations(t)
}

// TestRemoveManager_NotFound tests removing a user who is not a manager
func TestRemoveManager_NotFound(t *testing.T) {
	service, mockManagerRepo, mockRestaurantRepo, _ := setupManagerService()
	ctx := context.Background()

	restaurantID := uuid.New()
	ownerID := uuid.New()
	userID := uuid.New()

	restaurant := &domain.Restaurant{
		ID:      restaurantID,
		OwnerID: ownerID,
		Name:    "Test Restaurant",
	}

	mockRestaurantRepo.On("GetByID", ctx, restaurantID).Return(restaurant, nil)
	mockManagerRepo.On("IsManager", ctx, userID, restaurantID).Return(false, nil)

	err := service.RemoveManager(ctx, restaurantID, ownerID, userID)

	assert.Error(t, err)
	assert.Equal(t, ErrManagerNotFound, err)

	mockRestaurantRepo.AssertExpectations(t)
	mockManagerRepo.AssertExpectations(t)
}

// TestRemoveManager_IsManagerError tests removing manager when checking existing manager fails
func TestRemoveManager_IsManagerError(t *testing.T) {
	service, mockManagerRepo, mockRestaurantRepo, _ := setupManagerService()
	ctx := context.Background()

	restaurantID := uuid.New()
	ownerID := uuid.New()
	userID := uuid.New()

	restaurant := &domain.Restaurant{
		ID:      restaurantID,
		OwnerID: ownerID,
		Name:    "Test Restaurant",
	}

	dbError := errors.New("database error")
	mockRestaurantRepo.On("GetByID", ctx, restaurantID).Return(restaurant, nil)
	mockManagerRepo.On("IsManager", ctx, userID, restaurantID).Return(false, dbError)

	err := service.RemoveManager(ctx, restaurantID, ownerID, userID)

	assert.Error(t, err)
	assert.Equal(t, dbError, err)

	mockRestaurantRepo.AssertExpectations(t)
	mockManagerRepo.AssertExpectations(t)
}

// TestRemoveManager_DeleteError tests removing manager when deletion fails
func TestRemoveManager_DeleteError(t *testing.T) {
	service, mockManagerRepo, mockRestaurantRepo, _ := setupManagerService()
	ctx := context.Background()

	restaurantID := uuid.New()
	ownerID := uuid.New()
	userID := uuid.New()

	restaurant := &domain.Restaurant{
		ID:      restaurantID,
		OwnerID: ownerID,
		Name:    "Test Restaurant",
	}

	dbError := errors.New("database error")
	mockRestaurantRepo.On("GetByID", ctx, restaurantID).Return(restaurant, nil)
	mockManagerRepo.On("IsManager", ctx, userID, restaurantID).Return(true, nil)
	mockManagerRepo.On("Delete", ctx, userID, restaurantID).Return(dbError)

	err := service.RemoveManager(ctx, restaurantID, ownerID, userID)

	assert.Error(t, err)
	assert.Equal(t, dbError, err)

	mockRestaurantRepo.AssertExpectations(t)
	mockManagerRepo.AssertExpectations(t)
}

// TestGetManagers_Success tests successful retrieval of managers
func TestGetManagers_Success(t *testing.T) {
	service, mockManagerRepo, mockRestaurantRepo, _ := setupManagerService()
	ctx := context.Background()

	restaurantID := uuid.New()

	restaurant := &domain.Restaurant{
		ID:   restaurantID,
		Name: "Test Restaurant",
	}

	managers := []*domain.RestaurantManager{
		{
			ID:           uuid.New(),
			UserID:       uuid.New(),
			RestaurantID: restaurantID,
			AssignedAt:   time.Now(),
		},
		{
			ID:           uuid.New(),
			UserID:       uuid.New(),
			RestaurantID: restaurantID,
			AssignedAt:   time.Now(),
		},
	}

	mockRestaurantRepo.On("GetByID", ctx, restaurantID).Return(restaurant, nil)
	mockManagerRepo.On("GetManagersByRestaurant", ctx, restaurantID).Return(managers, nil)

	result, err := service.GetManagers(ctx, restaurantID)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Len(t, result, 2)
	assert.Equal(t, managers, result)

	mockRestaurantRepo.AssertExpectations(t)
	mockManagerRepo.AssertExpectations(t)
}

// TestGetManagers_RestaurantNotFound tests getting managers when restaurant doesn't exist
func TestGetManagers_RestaurantNotFound(t *testing.T) {
	service, _, mockRestaurantRepo, _ := setupManagerService()
	ctx := context.Background()

	restaurantID := uuid.New()

	mockRestaurantRepo.On("GetByID", ctx, restaurantID).Return(nil, gorm.ErrRecordNotFound)

	result, err := service.GetManagers(ctx, restaurantID)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Equal(t, ErrRestaurantNotFound, err)

	mockRestaurantRepo.AssertExpectations(t)
}

// TestGetManagers_RestaurantGetError tests getting managers when getting restaurant fails
func TestGetManagers_RestaurantGetError(t *testing.T) {
	service, _, mockRestaurantRepo, _ := setupManagerService()
	ctx := context.Background()

	restaurantID := uuid.New()

	dbError := errors.New("database error")
	mockRestaurantRepo.On("GetByID", ctx, restaurantID).Return(nil, dbError)

	result, err := service.GetManagers(ctx, restaurantID)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Equal(t, dbError, err)

	mockRestaurantRepo.AssertExpectations(t)
}

// TestGetManagers_GetManagersError tests getting managers when retrieval fails
func TestGetManagers_GetManagersError(t *testing.T) {
	service, mockManagerRepo, mockRestaurantRepo, _ := setupManagerService()
	ctx := context.Background()

	restaurantID := uuid.New()

	restaurant := &domain.Restaurant{
		ID:   restaurantID,
		Name: "Test Restaurant",
	}

	dbError := errors.New("database error")
	mockRestaurantRepo.On("GetByID", ctx, restaurantID).Return(restaurant, nil)
	mockManagerRepo.On("GetManagersByRestaurant", ctx, restaurantID).Return(nil, dbError)

	result, err := service.GetManagers(ctx, restaurantID)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Equal(t, dbError, err)

	mockRestaurantRepo.AssertExpectations(t)
	mockManagerRepo.AssertExpectations(t)
}

// TestGetManagers_EmptyList tests getting managers when there are no managers
func TestGetManagers_EmptyList(t *testing.T) {
	service, mockManagerRepo, mockRestaurantRepo, _ := setupManagerService()
	ctx := context.Background()

	restaurantID := uuid.New()

	restaurant := &domain.Restaurant{
		ID:   restaurantID,
		Name: "Test Restaurant",
	}

	emptyManagers := []*domain.RestaurantManager{}

	mockRestaurantRepo.On("GetByID", ctx, restaurantID).Return(restaurant, nil)
	mockManagerRepo.On("GetManagersByRestaurant", ctx, restaurantID).Return(emptyManagers, nil)

	result, err := service.GetManagers(ctx, restaurantID)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Len(t, result, 0)

	mockRestaurantRepo.AssertExpectations(t)
	mockManagerRepo.AssertExpectations(t)
}

// TestNewManagerService tests the service constructor
func TestNewManagerService(t *testing.T) {
	mockManagerRepo := new(MockRestaurantManagerRepository)
	mockRestaurantRepo := new(MockRestaurantRepository)
	mockUserRepo := new(MockUserRepository)

	service := NewManagerService(mockManagerRepo, mockRestaurantRepo, mockUserRepo)

	assert.NotNil(t, service)
	assert.IsType(t, &managerService{}, service)
}
