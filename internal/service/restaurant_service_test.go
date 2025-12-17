package service

import (
	"context"
	"testing"

	"restaurant-booking/internal/domain"
	"restaurant-booking/internal/repository" // Важно для проверки интерфейса

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

//
// Mock RestaurantRepository
//

type MockRestaurantRepository struct {
	mock.Mock
}

func (m *MockRestaurantRepository) Create(ctx context.Context, r *domain.Restaurant) error {
	return m.Called(ctx, r).Error(0)
}

func (m *MockRestaurantRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Restaurant, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Restaurant), args.Error(1)
}

func (m *MockRestaurantRepository) GetByOwnerID(ctx context.Context, ownerID uuid.UUID) ([]*domain.Restaurant, error) {
	args := m.Called(ctx, ownerID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.Restaurant), args.Error(1)
}

func (m *MockRestaurantRepository) List(ctx context.Context, limit, offset int) ([]*domain.Restaurant, error) {
	args := m.Called(ctx, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.Restaurant), args.Error(1)
}

func (m *MockRestaurantRepository) Update(ctx context.Context, r *domain.Restaurant) error {
	return m.Called(ctx, r).Error(0)
}

// !!! ДОБАВЛЕН МЕТОД DELETE, КОТОРОГО НЕ ХВАТАЛО !!!
func (m *MockRestaurantRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return m.Called(ctx, id).Error(0)
}

// !!! ДОБАВЛЕН МЕТОД SEARCH (обычно он тоже нужен для полного соответствия интерфейсу) !!!
func (m *MockRestaurantRepository) Search(ctx context.Context, cuisineType *domain.CuisineType, minRating float64, limit, offset int) ([]*domain.Restaurant, error) {
	args := m.Called(ctx, cuisineType, minRating, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.Restaurant), args.Error(1)
}

// Проверка, что мок реализует интерфейс
var _ repository.RestaurantRepository = (*MockRestaurantRepository)(nil)

//
// Helper
//

func setupRestaurantService() (*restaurantService, *MockRestaurantRepository, sqlmock.Sqlmock) {
	repo := new(MockRestaurantRepository)

	// Переименовали переменную в dbMock, чтобы не конфликтовать с пакетом mock
	sqlDB, dbMock, _ := sqlmock.New()
	dialector := postgres.New(postgres.Config{
		Conn:       sqlDB,
		DriverName: "postgres",
	})
	db, _ := gorm.Open(dialector, &gorm.Config{})

	service := &restaurantService{
		restaurantRepo: repo,
		db:             db,
	}

	return service, repo, dbMock
}

//
// Tests
//

func TestCreateRestaurant_Success(t *testing.T) {
	service, repo, _ := setupRestaurantService()
	ctx := context.Background()

	ownerID := uuid.New()

	req := CreateRestaurantRequest{
		Name:         "Test Restaurant",
		Address:      "Test Address",
		AveragePrice: 5000,
	}

	repo.On("Create", ctx, mock.AnythingOfType("*domain.Restaurant")).Return(nil)

	restaurant, err := service.CreateRestaurant(ctx, ownerID, req)

	assert.NoError(t, err)
	assert.NotNil(t, restaurant)
	assert.Equal(t, ownerID, restaurant.OwnerID)
	assert.Equal(t, "Test Restaurant", restaurant.Name)
	assert.True(t, restaurant.IsActive)
	repo.AssertExpectations(t)
}

func TestCreateRestaurant_InvalidName(t *testing.T) {
	service, _, _ := setupRestaurantService()
	ctx := context.Background()

	_, err := service.CreateRestaurant(ctx, uuid.New(), CreateRestaurantRequest{
		Name: " ",
	})

	assert.Error(t, err)
	assert.Equal(t, ErrInvalidRestaurantName, err)
}

func TestGetRestaurant_Success(t *testing.T) {
	service, repo, _ := setupRestaurantService()
	ctx := context.Background()

	id := uuid.New()
	restaurant := &domain.Restaurant{ID: id}

	repo.On("GetByID", ctx, id).Return(restaurant, nil)

	result, err := service.GetRestaurant(ctx, id)

	assert.NoError(t, err)
	assert.Equal(t, restaurant, result)
	repo.AssertExpectations(t)
}

func TestGetRestaurant_NotFound(t *testing.T) {
	service, repo, _ := setupRestaurantService()
	ctx := context.Background()

	id := uuid.New()

	repo.On("GetByID", ctx, id).Return(nil, gorm.ErrRecordNotFound)

	result, err := service.GetRestaurant(ctx, id)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Equal(t, ErrRestaurantNotFound, err)
}

func TestGetRestaurants_Success(t *testing.T) {
	service, repo, _ := setupRestaurantService()
	ctx := context.Background()

	list := []*domain.Restaurant{
		{ID: uuid.New()},
		{ID: uuid.New()},
	}

	repo.On("List", ctx, 10, 0).Return(list, nil)

	result, err := service.GetRestaurants(ctx, 10, 0)

	assert.NoError(t, err)
	assert.Len(t, result, 2)
	repo.AssertExpectations(t)
}

func TestUpdateRestaurant_Unauthorized(t *testing.T) {
	service, repo, _ := setupRestaurantService()
	ctx := context.Background()

	id := uuid.New()
	ownerID := uuid.New()

	restaurant := &domain.Restaurant{
		ID:      id,
		OwnerID: uuid.New(),
	}

	repo.On("GetByID", ctx, id).Return(restaurant, nil)

	_, err := service.UpdateRestaurant(ctx, id, ownerID, UpdateRestaurantRequest{})

	assert.Error(t, err)
	assert.Equal(t, ErrUnauthorized, err)
}

func TestDeleteRestaurant_Success(t *testing.T) {
	service, repo, _ := setupRestaurantService()
	ctx := context.Background()

	id := uuid.New()
	ownerID := uuid.New()

	restaurant := &domain.Restaurant{
		ID:       id,
		OwnerID:  ownerID,
		IsActive: true,
	}

	repo.On("GetByID", ctx, id).Return(restaurant, nil)
	repo.On("Update", ctx, restaurant).Return(nil)

	err := service.DeleteRestaurant(ctx, id, ownerID)

	assert.NoError(t, err)
	assert.False(t, restaurant.IsActive)
	repo.AssertExpectations(t)
}

func TestAddImage_Success(t *testing.T) {
	service, repo, dbMock := setupRestaurantService()
	ctx := context.Background()

	restaurantID := uuid.New()
	ownerID := uuid.New()

	restaurant := &domain.Restaurant{
		ID:      restaurantID,
		OwnerID: ownerID,
	}

	repo.On("GetByID", ctx, restaurantID).Return(restaurant, nil)

	dbMock.ExpectBegin()
	dbMock.ExpectQuery(`INSERT`).WillReturnRows(
		sqlmock.NewRows([]string{"id"}).AddRow(uuid.New()),
	)
	dbMock.ExpectCommit()

	image, err := service.AddImage(ctx, restaurantID, ownerID, AddImageRequest{
		CloudinaryURL:      "http://image",
		CloudinaryPublicID: "public-id",
		IsMain:             true,
	})

	assert.NoError(t, err)
	assert.NotNil(t, image)
}

func TestDeleteImage_NotFound(t *testing.T) {
	service, repo, dbMock := setupRestaurantService()
	ctx := context.Background()

	restaurantID := uuid.New()
	imageID := uuid.New()
	ownerID := uuid.New()

	restaurant := &domain.Restaurant{
		ID:      restaurantID,
		OwnerID: ownerID,
	}

	repo.On("GetByID", ctx, restaurantID).Return(restaurant, nil)

	dbMock.ExpectQuery(`SELECT`).WillReturnError(gorm.ErrRecordNotFound)

	err := service.DeleteImage(ctx, imageID, restaurantID, ownerID)

	assert.Error(t, err)
	assert.Equal(t, ErrImageNotFound, err)
}
