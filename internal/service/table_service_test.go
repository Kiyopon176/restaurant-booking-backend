package service

import (
	"context"
	"errors"
	"restaurant-booking/internal/domain"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// MockTableRepository is a mock implementation of TableRepository
type MockTableRepository struct {
	mock.Mock
}

func (m *MockTableRepository) Create(ctx context.Context, table *domain.Table) error {
	args := m.Called(ctx, table)
	return args.Error(0)
}

func (m *MockTableRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Table, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Table), args.Error(1)
}

func (m *MockTableRepository) GetByRestaurantID(ctx context.Context, restaurantID uuid.UUID) ([]*domain.Table, error) {
	args := m.Called(ctx, restaurantID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.Table), args.Error(1)
}

func (m *MockTableRepository) GetAvailableTables(ctx context.Context, restaurantID uuid.UUID, minCapacity int) ([]*domain.Table, error) {
	args := m.Called(ctx, restaurantID, minCapacity)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.Table), args.Error(1)
}

func (m *MockTableRepository) Update(ctx context.Context, table *domain.Table) error {
	args := m.Called(ctx, table)
	return args.Error(0)
}

func (m *MockTableRepository) Delete(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockTableRepository) List(ctx context.Context, limit, offset int) ([]*domain.Table, error) {
	args := m.Called(ctx, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.Table), args.Error(1)
}

// setupTableService creates a table service instance with mock repositories
func setupTableService() (*tableService, *MockTableRepository, *MockRestaurantRepository, sqlmock.Sqlmock, *gorm.DB) {
	mockTableRepo := new(MockTableRepository)
	mockRestaurantRepo := new(MockRestaurantRepository)

	sqlDB, sqlMock, _ := sqlmock.New()
	dialector := postgres.New(postgres.Config{
		Conn:       sqlDB,
		DriverName: "postgres",
	})
	db, _ := gorm.Open(dialector, &gorm.Config{})

	service := &tableService{
		tableRepo:      mockTableRepo,
		restaurantRepo: mockRestaurantRepo,
		db:             db,
	}

	return service, mockTableRepo, mockRestaurantRepo, sqlMock, db
}

// TestNewTableService tests the service constructor
func TestNewTableService(t *testing.T) {
	mockTableRepo := new(MockTableRepository)
	mockRestaurantRepo := new(MockRestaurantRepository)
	sqlDB, _, _ := sqlmock.New()
	dialector := postgres.New(postgres.Config{
		Conn:       sqlDB,
		DriverName: "postgres",
	})
	db, _ := gorm.Open(dialector, &gorm.Config{})

	service := NewTableService(mockTableRepo, mockRestaurantRepo, db)

	assert.NotNil(t, service)
	assert.IsType(t, &tableService{}, service)
}

// TestCreateTable_Success tests successful table creation
func TestCreateTable_Success(t *testing.T) {
	service, mockTableRepo, mockRestaurantRepo, sqlMock, _ := setupTableService()
	ctx := context.Background()

	restaurantID := uuid.New()
	ownerID := uuid.New()

	restaurant := &domain.Restaurant{
		ID:      restaurantID,
		OwnerID: ownerID,
		Name:    "Test Restaurant",
	}

	xPos := 10
	yPos := 20

	req := CreateTableRequest{
		TableNumber:  "T1",
		MinCapacity:  2,
		MaxCapacity:  4,
		LocationType: domain.LocationWindow,
		XPosition:    &xPos,
		YPosition:    &yPos,
	}

	mockRestaurantRepo.On("GetByID", ctx, restaurantID).Return(restaurant, nil)
	sqlMock.ExpectQuery("SELECT (.+) FROM \"tables\"").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))
	mockTableRepo.On("Create", ctx, mock.AnythingOfType("*domain.Table")).Return(nil)

	result, err := service.CreateTable(ctx, restaurantID, ownerID, req)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "T1", result.TableNumber)
	assert.Equal(t, 2, result.MinCapacity)
	assert.Equal(t, 4, result.MaxCapacity)
	assert.Equal(t, domain.LocationWindow, result.LocationType)
	assert.True(t, result.IsActive)

	mockRestaurantRepo.AssertExpectations(t)
	mockTableRepo.AssertExpectations(t)
}

// TestCreateTable_RestaurantNotFound tests creating table when restaurant doesn't exist
func TestCreateTable_RestaurantNotFound(t *testing.T) {
	service, _, mockRestaurantRepo, _, _ := setupTableService()
	ctx := context.Background()

	restaurantID := uuid.New()
	ownerID := uuid.New()

	req := CreateTableRequest{
		TableNumber:  "T1",
		MinCapacity:  2,
		MaxCapacity:  4,
		LocationType: domain.LocationRegular,
	}

	mockRestaurantRepo.On("GetByID", ctx, restaurantID).Return(nil, gorm.ErrRecordNotFound)

	result, err := service.CreateTable(ctx, restaurantID, ownerID, req)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Equal(t, ErrRestaurantNotFound, err)

	mockRestaurantRepo.AssertExpectations(t)
}

// TestCreateTable_RestaurantGetError tests creating table when getting restaurant fails
func TestCreateTable_RestaurantGetError(t *testing.T) {
	service, _, mockRestaurantRepo, _, _ := setupTableService()
	ctx := context.Background()

	restaurantID := uuid.New()
	ownerID := uuid.New()

	req := CreateTableRequest{
		TableNumber:  "T1",
		MinCapacity:  2,
		MaxCapacity:  4,
		LocationType: domain.LocationRegular,
	}

	dbError := errors.New("database error")
	mockRestaurantRepo.On("GetByID", ctx, restaurantID).Return(nil, dbError)

	result, err := service.CreateTable(ctx, restaurantID, ownerID, req)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Equal(t, dbError, err)

	mockRestaurantRepo.AssertExpectations(t)
}

// TestCreateTable_Unauthorized tests creating table when user is not the owner
func TestCreateTable_Unauthorized(t *testing.T) {
	service, _, mockRestaurantRepo, _, _ := setupTableService()
	ctx := context.Background()

	restaurantID := uuid.New()
	ownerID := uuid.New()
	differentOwnerID := uuid.New()

	restaurant := &domain.Restaurant{
		ID:      restaurantID,
		OwnerID: differentOwnerID,
		Name:    "Test Restaurant",
	}

	req := CreateTableRequest{
		TableNumber:  "T1",
		MinCapacity:  2,
		MaxCapacity:  4,
		LocationType: domain.LocationRegular,
	}

	mockRestaurantRepo.On("GetByID", ctx, restaurantID).Return(restaurant, nil)

	result, err := service.CreateTable(ctx, restaurantID, ownerID, req)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Equal(t, ErrUnauthorized, err)

	mockRestaurantRepo.AssertExpectations(t)
}

// TestCreateTable_InvalidTableNumber tests creating table with empty table number
func TestCreateTable_InvalidTableNumber(t *testing.T) {
	service, _, mockRestaurantRepo, _, _ := setupTableService()
	ctx := context.Background()

	restaurantID := uuid.New()
	ownerID := uuid.New()

	restaurant := &domain.Restaurant{
		ID:      restaurantID,
		OwnerID: ownerID,
		Name:    "Test Restaurant",
	}

	req := CreateTableRequest{
		TableNumber:  "   ",
		MinCapacity:  2,
		MaxCapacity:  4,
		LocationType: domain.LocationRegular,
	}

	mockRestaurantRepo.On("GetByID", ctx, restaurantID).Return(restaurant, nil)

	result, err := service.CreateTable(ctx, restaurantID, ownerID, req)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Equal(t, ErrInvalidTableNumber, err)

	mockRestaurantRepo.AssertExpectations(t)
}

// TestCreateTable_InvalidCapacity tests creating table with invalid capacity
func TestCreateTable_InvalidCapacity(t *testing.T) {
	service, _, mockRestaurantRepo, _, _ := setupTableService()
	ctx := context.Background()

	restaurantID := uuid.New()
	ownerID := uuid.New()

	restaurant := &domain.Restaurant{
		ID:      restaurantID,
		OwnerID: ownerID,
		Name:    "Test Restaurant",
	}

	req := CreateTableRequest{
		TableNumber:  "T1",
		MinCapacity:  6,
		MaxCapacity:  4,
		LocationType: domain.LocationRegular,
	}

	mockRestaurantRepo.On("GetByID", ctx, restaurantID).Return(restaurant, nil)

	result, err := service.CreateTable(ctx, restaurantID, ownerID, req)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Equal(t, ErrInvalidCapacity, err)

	mockRestaurantRepo.AssertExpectations(t)
}

// TestCreateTable_DuplicateTableNumber tests creating table with existing table number
func TestCreateTable_DuplicateTableNumber(t *testing.T) {
	service, _, mockRestaurantRepo, sqlMock, _ := setupTableService()
	ctx := context.Background()

	restaurantID := uuid.New()
	ownerID := uuid.New()

	restaurant := &domain.Restaurant{
		ID:      restaurantID,
		OwnerID: ownerID,
		Name:    "Test Restaurant",
	}

	req := CreateTableRequest{
		TableNumber:  "T1",
		MinCapacity:  2,
		MaxCapacity:  4,
		LocationType: domain.LocationRegular,
	}

	mockRestaurantRepo.On("GetByID", ctx, restaurantID).Return(restaurant, nil)
	sqlMock.ExpectQuery("SELECT (.+) FROM \"tables\"").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))

	result, err := service.CreateTable(ctx, restaurantID, ownerID, req)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Equal(t, ErrDuplicateTableNumber, err)

	mockRestaurantRepo.AssertExpectations(t)
}

// TestCreateTable_DuplicateCheckError tests creating table when duplicate check fails
func TestCreateTable_DuplicateCheckError(t *testing.T) {
	service, _, mockRestaurantRepo, sqlMock, _ := setupTableService()
	ctx := context.Background()

	restaurantID := uuid.New()
	ownerID := uuid.New()

	restaurant := &domain.Restaurant{
		ID:      restaurantID,
		OwnerID: ownerID,
		Name:    "Test Restaurant",
	}

	req := CreateTableRequest{
		TableNumber:  "T1",
		MinCapacity:  2,
		MaxCapacity:  4,
		LocationType: domain.LocationRegular,
	}

	mockRestaurantRepo.On("GetByID", ctx, restaurantID).Return(restaurant, nil)
	sqlMock.ExpectQuery("SELECT (.+) FROM \"tables\"").
		WillReturnError(errors.New("database error"))

	result, err := service.CreateTable(ctx, restaurantID, ownerID, req)

	assert.Error(t, err)
	assert.Nil(t, result)

	mockRestaurantRepo.AssertExpectations(t)
}

// TestCreateTable_CreateError tests creating table when creation fails
func TestCreateTable_CreateError(t *testing.T) {
	service, mockTableRepo, mockRestaurantRepo, sqlMock, _ := setupTableService()
	ctx := context.Background()

	restaurantID := uuid.New()
	ownerID := uuid.New()

	restaurant := &domain.Restaurant{
		ID:      restaurantID,
		OwnerID: ownerID,
		Name:    "Test Restaurant",
	}

	req := CreateTableRequest{
		TableNumber:  "T1",
		MinCapacity:  2,
		MaxCapacity:  4,
		LocationType: domain.LocationRegular,
	}

	dbError := errors.New("database error")
	mockRestaurantRepo.On("GetByID", ctx, restaurantID).Return(restaurant, nil)
	sqlMock.ExpectQuery("SELECT (.+) FROM \"tables\"").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))
	mockTableRepo.On("Create", ctx, mock.AnythingOfType("*domain.Table")).Return(dbError)

	result, err := service.CreateTable(ctx, restaurantID, ownerID, req)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Equal(t, dbError, err)

	mockRestaurantRepo.AssertExpectations(t)
	mockTableRepo.AssertExpectations(t)
}

// TestGetTablesByRestaurant_Success tests successful retrieval of tables
func TestGetTablesByRestaurant_Success(t *testing.T) {
	service, mockTableRepo, _, _, _ := setupTableService()
	ctx := context.Background()

	restaurantID := uuid.New()

	tables := []*domain.Table{
		{
			ID:           uuid.New(),
			RestaurantID: restaurantID,
			TableNumber:  "T1",
			MinCapacity:  2,
			MaxCapacity:  4,
		},
		{
			ID:           uuid.New(),
			RestaurantID: restaurantID,
			TableNumber:  "T2",
			MinCapacity:  4,
			MaxCapacity:  6,
		},
	}

	mockTableRepo.On("GetByRestaurantID", ctx, restaurantID).Return(tables, nil)

	result, err := service.GetTablesByRestaurant(ctx, restaurantID)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Len(t, result, 2)
	assert.Equal(t, tables, result)

	mockTableRepo.AssertExpectations(t)
}

// TestGetTablesByRestaurant_Error tests getting tables when retrieval fails
func TestGetTablesByRestaurant_Error(t *testing.T) {
	service, mockTableRepo, _, _, _ := setupTableService()
	ctx := context.Background()

	restaurantID := uuid.New()
	dbError := errors.New("database error")

	mockTableRepo.On("GetByRestaurantID", ctx, restaurantID).Return(nil, dbError)

	result, err := service.GetTablesByRestaurant(ctx, restaurantID)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Equal(t, dbError, err)

	mockTableRepo.AssertExpectations(t)
}

// TestGetTablesByRestaurant_EmptyList tests getting tables when there are no tables
func TestGetTablesByRestaurant_EmptyList(t *testing.T) {
	service, mockTableRepo, _, _, _ := setupTableService()
	ctx := context.Background()

	restaurantID := uuid.New()
	emptyTables := []*domain.Table{}

	mockTableRepo.On("GetByRestaurantID", ctx, restaurantID).Return(emptyTables, nil)

	result, err := service.GetTablesByRestaurant(ctx, restaurantID)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Len(t, result, 0)

	mockTableRepo.AssertExpectations(t)
}

// TestUpdateTable_Success tests successful table update
func TestUpdateTable_Success(t *testing.T) {
	service, mockTableRepo, mockRestaurantRepo, sqlMock, _ := setupTableService()
	ctx := context.Background()

	tableID := uuid.New()
	restaurantID := uuid.New()
	ownerID := uuid.New()

	restaurant := &domain.Restaurant{
		ID:      restaurantID,
		OwnerID: ownerID,
		Name:    "Test Restaurant",
	}

	existingTable := &domain.Table{
		ID:           tableID,
		RestaurantID: restaurantID,
		TableNumber:  "T1",
		MinCapacity:  2,
		MaxCapacity:  4,
		LocationType: domain.LocationRegular,
		IsActive:     true,
	}

	newTableNumber := "T2"
	newMinCapacity := 4
	newMaxCapacity := 8
	newLocationType := domain.LocationVIP
	newXPos := 15
	newYPos := 25
	newIsActive := false

	req := UpdateTableRequest{
		TableNumber:  &newTableNumber,
		MinCapacity:  &newMinCapacity,
		MaxCapacity:  &newMaxCapacity,
		LocationType: &newLocationType,
		XPosition:    &newXPos,
		YPosition:    &newYPos,
		IsActive:     &newIsActive,
	}

	mockRestaurantRepo.On("GetByID", ctx, restaurantID).Return(restaurant, nil)
	mockTableRepo.On("GetByID", ctx, tableID).Return(existingTable, nil)
	sqlMock.ExpectQuery("SELECT (.+) FROM \"tables\"").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))
	mockTableRepo.On("Update", ctx, mock.AnythingOfType("*domain.Table")).Return(nil)

	result, err := service.UpdateTable(ctx, tableID, restaurantID, ownerID, req)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "T2", result.TableNumber)
	assert.Equal(t, 4, result.MinCapacity)
	assert.Equal(t, 8, result.MaxCapacity)
	assert.Equal(t, domain.LocationVIP, result.LocationType)
	assert.False(t, result.IsActive)

	mockRestaurantRepo.AssertExpectations(t)
	mockTableRepo.AssertExpectations(t)
}

// TestUpdateTable_RestaurantNotFound tests updating table when restaurant doesn't exist
func TestUpdateTable_RestaurantNotFound(t *testing.T) {
	service, _, mockRestaurantRepo, _, _ := setupTableService()
	ctx := context.Background()

	tableID := uuid.New()
	restaurantID := uuid.New()
	ownerID := uuid.New()

	req := UpdateTableRequest{}

	mockRestaurantRepo.On("GetByID", ctx, restaurantID).Return(nil, gorm.ErrRecordNotFound)

	result, err := service.UpdateTable(ctx, tableID, restaurantID, ownerID, req)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Equal(t, ErrRestaurantNotFound, err)

	mockRestaurantRepo.AssertExpectations(t)
}

// TestUpdateTable_RestaurantGetError tests updating table when getting restaurant fails
func TestUpdateTable_RestaurantGetError(t *testing.T) {
	service, _, mockRestaurantRepo, _, _ := setupTableService()
	ctx := context.Background()

	tableID := uuid.New()
	restaurantID := uuid.New()
	ownerID := uuid.New()

	req := UpdateTableRequest{}
	dbError := errors.New("database error")

	mockRestaurantRepo.On("GetByID", ctx, restaurantID).Return(nil, dbError)

	result, err := service.UpdateTable(ctx, tableID, restaurantID, ownerID, req)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Equal(t, dbError, err)

	mockRestaurantRepo.AssertExpectations(t)
}

// TestUpdateTable_Unauthorized tests updating table when user is not the owner
func TestUpdateTable_Unauthorized(t *testing.T) {
	service, _, mockRestaurantRepo, _, _ := setupTableService()
	ctx := context.Background()

	tableID := uuid.New()
	restaurantID := uuid.New()
	ownerID := uuid.New()
	differentOwnerID := uuid.New()

	restaurant := &domain.Restaurant{
		ID:      restaurantID,
		OwnerID: differentOwnerID,
		Name:    "Test Restaurant",
	}

	req := UpdateTableRequest{}

	mockRestaurantRepo.On("GetByID", ctx, restaurantID).Return(restaurant, nil)

	result, err := service.UpdateTable(ctx, tableID, restaurantID, ownerID, req)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Equal(t, ErrUnauthorized, err)

	mockRestaurantRepo.AssertExpectations(t)
}

// TestUpdateTable_TableNotFound tests updating table when table doesn't exist
func TestUpdateTable_TableNotFound(t *testing.T) {
	service, mockTableRepo, mockRestaurantRepo, _, _ := setupTableService()
	ctx := context.Background()

	tableID := uuid.New()
	restaurantID := uuid.New()
	ownerID := uuid.New()

	restaurant := &domain.Restaurant{
		ID:      restaurantID,
		OwnerID: ownerID,
		Name:    "Test Restaurant",
	}

	req := UpdateTableRequest{}

	mockRestaurantRepo.On("GetByID", ctx, restaurantID).Return(restaurant, nil)
	mockTableRepo.On("GetByID", ctx, tableID).Return(nil, gorm.ErrRecordNotFound)

	result, err := service.UpdateTable(ctx, tableID, restaurantID, ownerID, req)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Equal(t, ErrTableNotFound, err)

	mockRestaurantRepo.AssertExpectations(t)
	mockTableRepo.AssertExpectations(t)
}

// TestUpdateTable_TableGetError tests updating table when getting table fails
func TestUpdateTable_TableGetError(t *testing.T) {
	service, mockTableRepo, mockRestaurantRepo, _, _ := setupTableService()
	ctx := context.Background()

	tableID := uuid.New()
	restaurantID := uuid.New()
	ownerID := uuid.New()

	restaurant := &domain.Restaurant{
		ID:      restaurantID,
		OwnerID: ownerID,
		Name:    "Test Restaurant",
	}

	req := UpdateTableRequest{}
	dbError := errors.New("database error")

	mockRestaurantRepo.On("GetByID", ctx, restaurantID).Return(restaurant, nil)
	mockTableRepo.On("GetByID", ctx, tableID).Return(nil, dbError)

	result, err := service.UpdateTable(ctx, tableID, restaurantID, ownerID, req)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Equal(t, dbError, err)

	mockRestaurantRepo.AssertExpectations(t)
	mockTableRepo.AssertExpectations(t)
}

// TestUpdateTable_TableRestaurantMismatch tests updating table from different restaurant
func TestUpdateTable_TableRestaurantMismatch(t *testing.T) {
	service, mockTableRepo, mockRestaurantRepo, _, _ := setupTableService()
	ctx := context.Background()

	tableID := uuid.New()
	restaurantID := uuid.New()
	differentRestaurantID := uuid.New()
	ownerID := uuid.New()

	restaurant := &domain.Restaurant{
		ID:      restaurantID,
		OwnerID: ownerID,
		Name:    "Test Restaurant",
	}

	existingTable := &domain.Table{
		ID:           tableID,
		RestaurantID: differentRestaurantID,
		TableNumber:  "T1",
	}

	req := UpdateTableRequest{}

	mockRestaurantRepo.On("GetByID", ctx, restaurantID).Return(restaurant, nil)
	mockTableRepo.On("GetByID", ctx, tableID).Return(existingTable, nil)

	result, err := service.UpdateTable(ctx, tableID, restaurantID, ownerID, req)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Equal(t, ErrUnauthorized, err)

	mockRestaurantRepo.AssertExpectations(t)
	mockTableRepo.AssertExpectations(t)
}

// TestUpdateTable_InvalidTableNumber tests updating with empty table number
func TestUpdateTable_InvalidTableNumber(t *testing.T) {
	service, mockTableRepo, mockRestaurantRepo, _, _ := setupTableService()
	ctx := context.Background()

	tableID := uuid.New()
	restaurantID := uuid.New()
	ownerID := uuid.New()

	restaurant := &domain.Restaurant{
		ID:      restaurantID,
		OwnerID: ownerID,
		Name:    "Test Restaurant",
	}

	existingTable := &domain.Table{
		ID:           tableID,
		RestaurantID: restaurantID,
		TableNumber:  "T1",
		MinCapacity:  2,
		MaxCapacity:  4,
	}

	emptyTableNumber := "   "
	req := UpdateTableRequest{
		TableNumber: &emptyTableNumber,
	}

	mockRestaurantRepo.On("GetByID", ctx, restaurantID).Return(restaurant, nil)
	mockTableRepo.On("GetByID", ctx, tableID).Return(existingTable, nil)

	result, err := service.UpdateTable(ctx, tableID, restaurantID, ownerID, req)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Equal(t, ErrInvalidTableNumber, err)

	mockRestaurantRepo.AssertExpectations(t)
	mockTableRepo.AssertExpectations(t)
}

// TestUpdateTable_DuplicateTableNumber tests updating with existing table number
func TestUpdateTable_DuplicateTableNumber(t *testing.T) {
	service, mockTableRepo, mockRestaurantRepo, sqlMock, _ := setupTableService()
	ctx := context.Background()

	tableID := uuid.New()
	restaurantID := uuid.New()
	ownerID := uuid.New()

	restaurant := &domain.Restaurant{
		ID:      restaurantID,
		OwnerID: ownerID,
		Name:    "Test Restaurant",
	}

	existingTable := &domain.Table{
		ID:           tableID,
		RestaurantID: restaurantID,
		TableNumber:  "T1",
		MinCapacity:  2,
		MaxCapacity:  4,
	}

	newTableNumber := "T2"
	req := UpdateTableRequest{
		TableNumber: &newTableNumber,
	}

	mockRestaurantRepo.On("GetByID", ctx, restaurantID).Return(restaurant, nil)
	mockTableRepo.On("GetByID", ctx, tableID).Return(existingTable, nil)
	sqlMock.ExpectQuery("SELECT (.+) FROM \"tables\"").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))

	result, err := service.UpdateTable(ctx, tableID, restaurantID, ownerID, req)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Equal(t, ErrDuplicateTableNumber, err)

	mockRestaurantRepo.AssertExpectations(t)
	mockTableRepo.AssertExpectations(t)
}

// TestUpdateTable_InvalidCapacity tests updating with invalid capacity
func TestUpdateTable_InvalidCapacity(t *testing.T) {
	service, mockTableRepo, mockRestaurantRepo, _, _ := setupTableService()
	ctx := context.Background()

	tableID := uuid.New()
	restaurantID := uuid.New()
	ownerID := uuid.New()

	restaurant := &domain.Restaurant{
		ID:      restaurantID,
		OwnerID: ownerID,
		Name:    "Test Restaurant",
	}

	existingTable := &domain.Table{
		ID:           tableID,
		RestaurantID: restaurantID,
		TableNumber:  "T1",
		MinCapacity:  2,
		MaxCapacity:  4,
	}

	newMinCapacity := 8
	newMaxCapacity := 4
	req := UpdateTableRequest{
		MinCapacity: &newMinCapacity,
		MaxCapacity: &newMaxCapacity,
	}

	mockRestaurantRepo.On("GetByID", ctx, restaurantID).Return(restaurant, nil)
	mockTableRepo.On("GetByID", ctx, tableID).Return(existingTable, nil)

	result, err := service.UpdateTable(ctx, tableID, restaurantID, ownerID, req)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Equal(t, ErrInvalidCapacity, err)

	mockRestaurantRepo.AssertExpectations(t)
	mockTableRepo.AssertExpectations(t)
}

// TestUpdateTable_UpdateError tests updating table when update fails
func TestUpdateTable_UpdateError(t *testing.T) {
	service, mockTableRepo, mockRestaurantRepo, _, _ := setupTableService()
	ctx := context.Background()

	tableID := uuid.New()
	restaurantID := uuid.New()
	ownerID := uuid.New()

	restaurant := &domain.Restaurant{
		ID:      restaurantID,
		OwnerID: ownerID,
		Name:    "Test Restaurant",
	}

	existingTable := &domain.Table{
		ID:           tableID,
		RestaurantID: restaurantID,
		TableNumber:  "T1",
		MinCapacity:  2,
		MaxCapacity:  4,
	}

	newIsActive := false
	req := UpdateTableRequest{
		IsActive: &newIsActive,
	}

	dbError := errors.New("database error")
	mockRestaurantRepo.On("GetByID", ctx, restaurantID).Return(restaurant, nil)
	mockTableRepo.On("GetByID", ctx, tableID).Return(existingTable, nil)
	mockTableRepo.On("Update", ctx, mock.AnythingOfType("*domain.Table")).Return(dbError)

	result, err := service.UpdateTable(ctx, tableID, restaurantID, ownerID, req)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Equal(t, dbError, err)

	mockRestaurantRepo.AssertExpectations(t)
	mockTableRepo.AssertExpectations(t)
}

// TestDeleteTable_Success tests successful table deletion
func TestDeleteTable_Success(t *testing.T) {
	service, mockTableRepo, mockRestaurantRepo, _, _ := setupTableService()
	ctx := context.Background()

	tableID := uuid.New()
	restaurantID := uuid.New()
	ownerID := uuid.New()

	restaurant := &domain.Restaurant{
		ID:      restaurantID,
		OwnerID: ownerID,
		Name:    "Test Restaurant",
	}

	existingTable := &domain.Table{
		ID:           tableID,
		RestaurantID: restaurantID,
		TableNumber:  "T1",
		IsActive:     true,
	}

	mockRestaurantRepo.On("GetByID", ctx, restaurantID).Return(restaurant, nil)
	mockTableRepo.On("GetByID", ctx, tableID).Return(existingTable, nil)
	mockTableRepo.On("Update", ctx, mock.AnythingOfType("*domain.Table")).Return(nil)

	err := service.DeleteTable(ctx, tableID, restaurantID, ownerID)

	assert.NoError(t, err)

	mockRestaurantRepo.AssertExpectations(t)
	mockTableRepo.AssertExpectations(t)
}

// TestDeleteTable_RestaurantNotFound tests deleting table when restaurant doesn't exist
func TestDeleteTable_RestaurantNotFound(t *testing.T) {
	service, _, mockRestaurantRepo, _, _ := setupTableService()
	ctx := context.Background()

	tableID := uuid.New()
	restaurantID := uuid.New()
	ownerID := uuid.New()

	mockRestaurantRepo.On("GetByID", ctx, restaurantID).Return(nil, gorm.ErrRecordNotFound)

	err := service.DeleteTable(ctx, tableID, restaurantID, ownerID)

	assert.Error(t, err)
	assert.Equal(t, ErrRestaurantNotFound, err)

	mockRestaurantRepo.AssertExpectations(t)
}

// TestDeleteTable_RestaurantGetError tests deleting table when getting restaurant fails
func TestDeleteTable_RestaurantGetError(t *testing.T) {
	service, _, mockRestaurantRepo, _, _ := setupTableService()
	ctx := context.Background()

	tableID := uuid.New()
	restaurantID := uuid.New()
	ownerID := uuid.New()

	dbError := errors.New("database error")
	mockRestaurantRepo.On("GetByID", ctx, restaurantID).Return(nil, dbError)

	err := service.DeleteTable(ctx, tableID, restaurantID, ownerID)

	assert.Error(t, err)
	assert.Equal(t, dbError, err)

	mockRestaurantRepo.AssertExpectations(t)
}

// TestDeleteTable_Unauthorized tests deleting table when user is not the owner
func TestDeleteTable_Unauthorized(t *testing.T) {
	service, _, mockRestaurantRepo, _, _ := setupTableService()
	ctx := context.Background()

	tableID := uuid.New()
	restaurantID := uuid.New()
	ownerID := uuid.New()
	differentOwnerID := uuid.New()

	restaurant := &domain.Restaurant{
		ID:      restaurantID,
		OwnerID: differentOwnerID,
		Name:    "Test Restaurant",
	}

	mockRestaurantRepo.On("GetByID", ctx, restaurantID).Return(restaurant, nil)

	err := service.DeleteTable(ctx, tableID, restaurantID, ownerID)

	assert.Error(t, err)
	assert.Equal(t, ErrUnauthorized, err)

	mockRestaurantRepo.AssertExpectations(t)
}

// TestDeleteTable_TableNotFound tests deleting table when table doesn't exist
func TestDeleteTable_TableNotFound(t *testing.T) {
	service, mockTableRepo, mockRestaurantRepo, _, _ := setupTableService()
	ctx := context.Background()

	tableID := uuid.New()
	restaurantID := uuid.New()
	ownerID := uuid.New()

	restaurant := &domain.Restaurant{
		ID:      restaurantID,
		OwnerID: ownerID,
		Name:    "Test Restaurant",
	}

	mockRestaurantRepo.On("GetByID", ctx, restaurantID).Return(restaurant, nil)
	mockTableRepo.On("GetByID", ctx, tableID).Return(nil, gorm.ErrRecordNotFound)

	err := service.DeleteTable(ctx, tableID, restaurantID, ownerID)

	assert.Error(t, err)
	assert.Equal(t, ErrTableNotFound, err)

	mockRestaurantRepo.AssertExpectations(t)
	mockTableRepo.AssertExpectations(t)
}

// TestDeleteTable_TableGetError tests deleting table when getting table fails
func TestDeleteTable_TableGetError(t *testing.T) {
	service, mockTableRepo, mockRestaurantRepo, _, _ := setupTableService()
	ctx := context.Background()

	tableID := uuid.New()
	restaurantID := uuid.New()
	ownerID := uuid.New()

	restaurant := &domain.Restaurant{
		ID:      restaurantID,
		OwnerID: ownerID,
		Name:    "Test Restaurant",
	}

	dbError := errors.New("database error")
	mockRestaurantRepo.On("GetByID", ctx, restaurantID).Return(restaurant, nil)
	mockTableRepo.On("GetByID", ctx, tableID).Return(nil, dbError)

	err := service.DeleteTable(ctx, tableID, restaurantID, ownerID)

	assert.Error(t, err)
	assert.Equal(t, dbError, err)

	mockRestaurantRepo.AssertExpectations(t)
	mockTableRepo.AssertExpectations(t)
}

// TestDeleteTable_TableRestaurantMismatch tests deleting table from different restaurant
func TestDeleteTable_TableRestaurantMismatch(t *testing.T) {
	service, mockTableRepo, mockRestaurantRepo, _, _ := setupTableService()
	ctx := context.Background()

	tableID := uuid.New()
	restaurantID := uuid.New()
	differentRestaurantID := uuid.New()
	ownerID := uuid.New()

	restaurant := &domain.Restaurant{
		ID:      restaurantID,
		OwnerID: ownerID,
		Name:    "Test Restaurant",
	}

	existingTable := &domain.Table{
		ID:           tableID,
		RestaurantID: differentRestaurantID,
		TableNumber:  "T1",
	}

	mockRestaurantRepo.On("GetByID", ctx, restaurantID).Return(restaurant, nil)
	mockTableRepo.On("GetByID", ctx, tableID).Return(existingTable, nil)

	err := service.DeleteTable(ctx, tableID, restaurantID, ownerID)

	assert.Error(t, err)
	assert.Equal(t, ErrUnauthorized, err)

	mockRestaurantRepo.AssertExpectations(t)
	mockTableRepo.AssertExpectations(t)
}

// TestDeleteTable_UpdateError tests deleting table when update fails
func TestDeleteTable_UpdateError(t *testing.T) {
	service, mockTableRepo, mockRestaurantRepo, _, _ := setupTableService()
	ctx := context.Background()

	tableID := uuid.New()
	restaurantID := uuid.New()
	ownerID := uuid.New()

	restaurant := &domain.Restaurant{
		ID:      restaurantID,
		OwnerID: ownerID,
		Name:    "Test Restaurant",
	}

	existingTable := &domain.Table{
		ID:           tableID,
		RestaurantID: restaurantID,
		TableNumber:  "T1",
		IsActive:     true,
	}

	dbError := errors.New("database error")
	mockRestaurantRepo.On("GetByID", ctx, restaurantID).Return(restaurant, nil)
	mockTableRepo.On("GetByID", ctx, tableID).Return(existingTable, nil)
	mockTableRepo.On("Update", ctx, mock.AnythingOfType("*domain.Table")).Return(dbError)

	err := service.DeleteTable(ctx, tableID, restaurantID, ownerID)

	assert.Error(t, err)
	assert.Equal(t, dbError, err)

	mockRestaurantRepo.AssertExpectations(t)
	mockTableRepo.AssertExpectations(t)
}

// TestBulkCreateTables_Success tests successful bulk table creation
func TestBulkCreateTables_Success(t *testing.T) {
	service, mockTableRepo, mockRestaurantRepo, sqlMock, _ := setupTableService()
	ctx := context.Background()

	restaurantID := uuid.New()
	ownerID := uuid.New()

	restaurant := &domain.Restaurant{
		ID:      restaurantID,
		OwnerID: ownerID,
		Name:    "Test Restaurant",
	}

	req := BulkCreateTablesRequest{
		Tables: []CreateTableRequest{
			{
				TableNumber:  "T1",
				MinCapacity:  2,
				MaxCapacity:  4,
				LocationType: domain.LocationWindow,
			},
			{
				TableNumber:  "T2",
				MinCapacity:  4,
				MaxCapacity:  6,
				LocationType: domain.LocationRegular,
			},
		},
	}

	mockRestaurantRepo.On("GetByID", ctx, restaurantID).Return(restaurant, nil)

	// Expect duplicate check for T1
	sqlMock.ExpectQuery("SELECT (.+) FROM \"tables\"").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))
	// Expect duplicate check for T2
	sqlMock.ExpectQuery("SELECT (.+) FROM \"tables\"").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))

	mockTableRepo.On("Create", ctx, mock.AnythingOfType("*domain.Table")).Return(nil).Times(2)

	result, err := service.BulkCreateTables(ctx, restaurantID, ownerID, req)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Len(t, result, 2)

	mockRestaurantRepo.AssertExpectations(t)
	mockTableRepo.AssertExpectations(t)
}

// TestBulkCreateTables_RestaurantNotFound tests bulk creation when restaurant doesn't exist
func TestBulkCreateTables_RestaurantNotFound(t *testing.T) {
	service, _, mockRestaurantRepo, _, _ := setupTableService()
	ctx := context.Background()

	restaurantID := uuid.New()
	ownerID := uuid.New()

	req := BulkCreateTablesRequest{
		Tables: []CreateTableRequest{
			{
				TableNumber:  "T1",
				MinCapacity:  2,
				MaxCapacity:  4,
				LocationType: domain.LocationRegular,
			},
		},
	}

	mockRestaurantRepo.On("GetByID", ctx, restaurantID).Return(nil, gorm.ErrRecordNotFound)

	result, err := service.BulkCreateTables(ctx, restaurantID, ownerID, req)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Equal(t, ErrRestaurantNotFound, err)

	mockRestaurantRepo.AssertExpectations(t)
}

// TestBulkCreateTables_RestaurantGetError tests bulk creation when getting restaurant fails
func TestBulkCreateTables_RestaurantGetError(t *testing.T) {
	service, _, mockRestaurantRepo, _, _ := setupTableService()
	ctx := context.Background()

	restaurantID := uuid.New()
	ownerID := uuid.New()

	req := BulkCreateTablesRequest{
		Tables: []CreateTableRequest{
			{
				TableNumber:  "T1",
				MinCapacity:  2,
				MaxCapacity:  4,
				LocationType: domain.LocationRegular,
			},
		},
	}

	dbError := errors.New("database error")
	mockRestaurantRepo.On("GetByID", ctx, restaurantID).Return(nil, dbError)

	result, err := service.BulkCreateTables(ctx, restaurantID, ownerID, req)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Equal(t, dbError, err)

	mockRestaurantRepo.AssertExpectations(t)
}

// TestBulkCreateTables_Unauthorized tests bulk creation when user is not the owner
func TestBulkCreateTables_Unauthorized(t *testing.T) {
	service, _, mockRestaurantRepo, _, _ := setupTableService()
	ctx := context.Background()

	restaurantID := uuid.New()
	ownerID := uuid.New()
	differentOwnerID := uuid.New()

	restaurant := &domain.Restaurant{
		ID:      restaurantID,
		OwnerID: differentOwnerID,
		Name:    "Test Restaurant",
	}

	req := BulkCreateTablesRequest{
		Tables: []CreateTableRequest{
			{
				TableNumber:  "T1",
				MinCapacity:  2,
				MaxCapacity:  4,
				LocationType: domain.LocationRegular,
			},
		},
	}

	mockRestaurantRepo.On("GetByID", ctx, restaurantID).Return(restaurant, nil)

	result, err := service.BulkCreateTables(ctx, restaurantID, ownerID, req)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Equal(t, ErrUnauthorized, err)

	mockRestaurantRepo.AssertExpectations(t)
}

// TestBulkCreateTables_InvalidTableNumber tests bulk creation with empty table number
func TestBulkCreateTables_InvalidTableNumber(t *testing.T) {
	service, _, mockRestaurantRepo, _, _ := setupTableService()
	ctx := context.Background()

	restaurantID := uuid.New()
	ownerID := uuid.New()

	restaurant := &domain.Restaurant{
		ID:      restaurantID,
		OwnerID: ownerID,
		Name:    "Test Restaurant",
	}

	req := BulkCreateTablesRequest{
		Tables: []CreateTableRequest{
			{
				TableNumber:  "   ",
				MinCapacity:  2,
				MaxCapacity:  4,
				LocationType: domain.LocationRegular,
			},
		},
	}

	mockRestaurantRepo.On("GetByID", ctx, restaurantID).Return(restaurant, nil)

	result, err := service.BulkCreateTables(ctx, restaurantID, ownerID, req)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "table at index 0")
	assert.Contains(t, err.Error(), ErrInvalidTableNumber.Error())

	mockRestaurantRepo.AssertExpectations(t)
}

// TestBulkCreateTables_InvalidCapacity tests bulk creation with invalid capacity
func TestBulkCreateTables_InvalidCapacity(t *testing.T) {
	service, _, mockRestaurantRepo, _, _ := setupTableService()
	ctx := context.Background()

	restaurantID := uuid.New()
	ownerID := uuid.New()

	restaurant := &domain.Restaurant{
		ID:      restaurantID,
		OwnerID: ownerID,
		Name:    "Test Restaurant",
	}

	req := BulkCreateTablesRequest{
		Tables: []CreateTableRequest{
			{
				TableNumber:  "T1",
				MinCapacity:  6,
				MaxCapacity:  4,
				LocationType: domain.LocationRegular,
			},
		},
	}

	mockRestaurantRepo.On("GetByID", ctx, restaurantID).Return(restaurant, nil)

	result, err := service.BulkCreateTables(ctx, restaurantID, ownerID, req)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "table at index 0")
	assert.Contains(t, err.Error(), ErrInvalidCapacity.Error())

	mockRestaurantRepo.AssertExpectations(t)
}

// TestBulkCreateTables_DuplicateInRequest tests bulk creation with duplicate table numbers in request
func TestBulkCreateTables_DuplicateInRequest(t *testing.T) {
	service, _, mockRestaurantRepo, sqlMock, _ := setupTableService()
	ctx := context.Background()

	restaurantID := uuid.New()
	ownerID := uuid.New()

	restaurant := &domain.Restaurant{
		ID:      restaurantID,
		OwnerID: ownerID,
		Name:    "Test Restaurant",
	}

	req := BulkCreateTablesRequest{
		Tables: []CreateTableRequest{
			{
				TableNumber:  "T1",
				MinCapacity:  2,
				MaxCapacity:  4,
				LocationType: domain.LocationRegular,
			},
			{
				TableNumber:  "T1",
				MinCapacity:  4,
				MaxCapacity:  6,
				LocationType: domain.LocationRegular,
			},
		},
	}

	mockRestaurantRepo.On("GetByID", ctx, restaurantID).Return(restaurant, nil)
	// First table passes duplicate check
	sqlMock.ExpectQuery("SELECT (.+) FROM \"tables\"").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))

	result, err := service.BulkCreateTables(ctx, restaurantID, ownerID, req)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "table at index 1")
	assert.Contains(t, err.Error(), "duplicate table number 'T1' in request")

	mockRestaurantRepo.AssertExpectations(t)
}

// TestBulkCreateTables_DuplicateTableNumber tests bulk creation with existing table number
func TestBulkCreateTables_DuplicateTableNumber(t *testing.T) {
	service, _, mockRestaurantRepo, sqlMock, _ := setupTableService()
	ctx := context.Background()

	restaurantID := uuid.New()
	ownerID := uuid.New()

	restaurant := &domain.Restaurant{
		ID:      restaurantID,
		OwnerID: ownerID,
		Name:    "Test Restaurant",
	}

	req := BulkCreateTablesRequest{
		Tables: []CreateTableRequest{
			{
				TableNumber:  "T1",
				MinCapacity:  2,
				MaxCapacity:  4,
				LocationType: domain.LocationRegular,
			},
		},
	}

	mockRestaurantRepo.On("GetByID", ctx, restaurantID).Return(restaurant, nil)
	sqlMock.ExpectQuery("SELECT (.+) FROM \"tables\"").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))

	result, err := service.BulkCreateTables(ctx, restaurantID, ownerID, req)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "table at index 0")

	mockRestaurantRepo.AssertExpectations(t)
}

// TestBulkCreateTables_DuplicateCheckError tests bulk creation when duplicate check fails
func TestBulkCreateTables_DuplicateCheckError(t *testing.T) {
	service, _, mockRestaurantRepo, sqlMock, _ := setupTableService()
	ctx := context.Background()

	restaurantID := uuid.New()
	ownerID := uuid.New()

	restaurant := &domain.Restaurant{
		ID:      restaurantID,
		OwnerID: ownerID,
		Name:    "Test Restaurant",
	}

	req := BulkCreateTablesRequest{
		Tables: []CreateTableRequest{
			{
				TableNumber:  "T1",
				MinCapacity:  2,
				MaxCapacity:  4,
				LocationType: domain.LocationRegular,
			},
		},
	}

	mockRestaurantRepo.On("GetByID", ctx, restaurantID).Return(restaurant, nil)
	sqlMock.ExpectQuery("SELECT (.+) FROM \"tables\"").
		WillReturnError(errors.New("database error"))

	result, err := service.BulkCreateTables(ctx, restaurantID, ownerID, req)

	assert.Error(t, err)
	assert.Nil(t, result)

	mockRestaurantRepo.AssertExpectations(t)
}

// TestBulkCreateTables_CreateError tests bulk creation when creation fails
func TestBulkCreateTables_CreateError(t *testing.T) {
	service, mockTableRepo, mockRestaurantRepo, sqlMock, _ := setupTableService()
	ctx := context.Background()

	restaurantID := uuid.New()
	ownerID := uuid.New()

	restaurant := &domain.Restaurant{
		ID:      restaurantID,
		OwnerID: ownerID,
		Name:    "Test Restaurant",
	}

	req := BulkCreateTablesRequest{
		Tables: []CreateTableRequest{
			{
				TableNumber:  "T1",
				MinCapacity:  2,
				MaxCapacity:  4,
				LocationType: domain.LocationRegular,
			},
		},
	}

	dbError := errors.New("database error")
	mockRestaurantRepo.On("GetByID", ctx, restaurantID).Return(restaurant, nil)
	sqlMock.ExpectQuery("SELECT (.+) FROM \"tables\"").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))
	mockTableRepo.On("Create", ctx, mock.AnythingOfType("*domain.Table")).Return(dbError)

	result, err := service.BulkCreateTables(ctx, restaurantID, ownerID, req)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "failed to create table at index 0")

	mockRestaurantRepo.AssertExpectations(t)
	mockTableRepo.AssertExpectations(t)
}
