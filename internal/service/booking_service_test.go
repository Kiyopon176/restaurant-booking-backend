package service

import (
	"context"
	"restaurant-booking/internal/domain"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	tmock "github.com/stretchr/testify/mock"
)

type BookingMockBookingRepository struct {
	tmock.Mock
}

func (m *BookingMockBookingRepository) Create(ctx context.Context, booking *domain.Booking) error {
	args := m.Called(ctx, booking)
	return args.Error(0)
}

func (m *BookingMockBookingRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Booking, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Booking), args.Error(1)
}

func (m *BookingMockBookingRepository) GetByUserID(ctx context.Context, userID uuid.UUID) ([]*domain.Booking, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.Booking), args.Error(1)
}

func (m *BookingMockBookingRepository) GetByRestaurantID(ctx context.Context, restaurantID uuid.UUID, date time.Time) ([]*domain.Booking, error) {
	args := m.Called(ctx, restaurantID, date)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.Booking), args.Error(1)
}

func (m *BookingMockBookingRepository) Update(ctx context.Context, booking *domain.Booking) error {
	args := m.Called(ctx, booking)
	return args.Error(0)
}

func (m *BookingMockBookingRepository) Delete(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *BookingMockBookingRepository) CheckTableAvailability(ctx context.Context, tableID uuid.UUID, startTime, endTime time.Time) (bool, error) {
	args := m.Called(ctx, tableID, startTime, endTime)
	return args.Bool(0), args.Error(1)
}

type BookingMockTableRepository struct {
	tmock.Mock
}

func (m *BookingMockTableRepository) Create(ctx context.Context, table *domain.Table) error {
	args := m.Called(ctx, table)
	return args.Error(0)
}

func (m *BookingMockTableRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Table, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Table), args.Error(1)
}

func (m *BookingMockTableRepository) GetByRestaurantID(ctx context.Context, restaurantID uuid.UUID) ([]*domain.Table, error) {
	args := m.Called(ctx, restaurantID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.Table), args.Error(1)
}

func (m *BookingMockTableRepository) Update(ctx context.Context, table *domain.Table) error {
	args := m.Called(ctx, table)
	return args.Error(0)
}

func (m *BookingMockTableRepository) Delete(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *BookingMockTableRepository) GetAvailableTables(ctx context.Context, restaurantID uuid.UUID, minCapacity int) ([]*domain.Table, error) {
	args := m.Called(ctx, restaurantID, minCapacity)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.Table), args.Error(1)
}

func (m *BookingMockTableRepository) List(ctx context.Context, limit, offset int) ([]*domain.Table, error) {
	args := m.Called(ctx, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.Table), args.Error(1)
}

type BookingMockRestaurantRepository struct {
	tmock.Mock
}

func (m *BookingMockRestaurantRepository) Create(ctx context.Context, restaurant *domain.Restaurant) error {
	args := m.Called(ctx, restaurant)
	return args.Error(0)
}

func (m *BookingMockRestaurantRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Restaurant, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Restaurant), args.Error(1)
}

func (m *BookingMockRestaurantRepository) GetAll(ctx context.Context) ([]*domain.Restaurant, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.Restaurant), args.Error(1)
}

func (m *BookingMockRestaurantRepository) Update(ctx context.Context, restaurant *domain.Restaurant) error {
	args := m.Called(ctx, restaurant)
	return args.Error(0)
}

func (m *BookingMockRestaurantRepository) Delete(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *BookingMockRestaurantRepository) GetByManagerID(ctx context.Context, managerID uuid.UUID) ([]*domain.Restaurant, error) {
	args := m.Called(ctx, managerID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.Restaurant), args.Error(1)
}

func (m *BookingMockRestaurantRepository) GetByOwnerID(ctx context.Context, ownerID uuid.UUID) ([]*domain.Restaurant, error) {
	args := m.Called(ctx, ownerID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.Restaurant), args.Error(1)
}

func (m *BookingMockRestaurantRepository) List(ctx context.Context, limit, offset int) ([]*domain.Restaurant, error) {
	args := m.Called(ctx, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.Restaurant), args.Error(1)
}

func (m *BookingMockRestaurantRepository) Search(ctx context.Context, cuisineType *domain.CuisineType, minRating float64, limit, offset int) ([]*domain.Restaurant, error) {
	args := m.Called(ctx, cuisineType, minRating, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.Restaurant), args.Error(1)
}

func setupBookingService() (*BookingService, *BookingMockBookingRepository, *BookingMockTableRepository, *BookingMockRestaurantRepository, *NotificationService) {
	mockBookingRepo := new(BookingMockBookingRepository)
	mockTableRepo := new(BookingMockTableRepository)
	mockRestaurantRepo := new(BookingMockRestaurantRepository)
	notificationSvc := NewNotificationService(2, 10)

	service := NewBookingService(
		mockBookingRepo,
		mockTableRepo,
		mockRestaurantRepo,
		notificationSvc,
	)

	return service, mockBookingRepo, mockTableRepo, mockRestaurantRepo, notificationSvc
}

func TestCreateBookingWithNotification_Success(t *testing.T) {
	service, _, _, _, notificationSvc := setupBookingService()
	defer notificationSvc.Shutdown()

	ctx := context.Background()
	userID := uuid.New()
	tableID := uuid.New()
	startTime := time.Now().Add(24 * time.Hour)
	endTime := startTime.Add(2 * time.Hour)
	guestCount := 4
	userEmail := "test@example.com"

	booking, err := service.CreateBookingWithNotification(ctx, userID, tableID, startTime, endTime, guestCount, userEmail)

	assert.NoError(t, err)
	assert.NotNil(t, booking)
	assert.Equal(t, userID, booking.UserID)
	assert.Equal(t, tableID, booking.TableID)
	assert.Equal(t, domain.BookingStatusPending, booking.Status)

	time.Sleep(50 * time.Millisecond)
}

func TestCheckMultipleTablesAvailability_Success(t *testing.T) {
	service, _, _, _, notificationSvc := setupBookingService()
	defer notificationSvc.Shutdown()

	ctx := context.Background()
	tableIDs := []uuid.UUID{uuid.New(), uuid.New(), uuid.New()}
	startTime := time.Now().Add(24 * time.Hour)
	endTime := startTime.Add(2 * time.Hour)

	results := service.CheckMultipleTablesAvailability(ctx, tableIDs, startTime, endTime)

	assert.NotNil(t, results)
	assert.Equal(t, len(tableIDs), len(results))

	for _, tableID := range tableIDs {
		availability, exists := results[tableID]
		assert.True(t, exists)
		assert.True(t, availability)
	}
}

func TestCheckMultipleTablesAvailability_EmptyList(t *testing.T) {
	service, _, _, _, notificationSvc := setupBookingService()
	defer notificationSvc.Shutdown()

	ctx := context.Background()
	startTime := time.Now().Add(24 * time.Hour)
	endTime := startTime.Add(2 * time.Hour)

	results := service.CheckMultipleTablesAvailability(ctx, []uuid.UUID{}, startTime, endTime)

	assert.NotNil(t, results)
	assert.Equal(t, 0, len(results))
}

func TestProcessBulkBookings_Success(t *testing.T) {
	service, _, _, _, notificationSvc := setupBookingService()
	defer notificationSvc.Shutdown()

	ctx := context.Background()
	bookings := []domain.Booking{
		{
			ID:        uuid.New(),
			UserID:    uuid.New(),
			TableID:   uuid.New(),
			StartTime: time.Now().Add(24 * time.Hour),
			EndTime:   time.Now().Add(26 * time.Hour),
			Status:    domain.BookingStatusPending,
		},
		{
			ID:        uuid.New(),
			UserID:    uuid.New(),
			TableID:   uuid.New(),
			StartTime: time.Now().Add(48 * time.Hour),
			EndTime:   time.Now().Add(50 * time.Hour),
			Status:    domain.BookingStatusPending,
		},
	}

	results := service.ProcessBulkBookings(ctx, bookings, 2)

	assert.NotNil(t, results)
	assert.Equal(t, len(bookings), len(results))

	for _, result := range results {
		assert.NoError(t, result.Error)
		assert.NotNil(t, result.Booking)
	}
}

func TestProcessBulkBookings_EmptyList(t *testing.T) {
	service, _, _, _, notificationSvc := setupBookingService()
	defer notificationSvc.Shutdown()

	ctx := context.Background()

	results := service.ProcessBulkBookings(ctx, []domain.Booking{}, 2)

	assert.NotNil(t, results)
	assert.Equal(t, 0, len(results))
}

func TestProcessBulkBookings_WithConcurrencyLimit(t *testing.T) {
	service, _, _, _, notificationSvc := setupBookingService()
	defer notificationSvc.Shutdown()

	ctx := context.Background()
	bookings := make([]domain.Booking, 10)
	for i := 0; i < 10; i++ {
		bookings[i] = domain.Booking{
			ID:        uuid.New(),
			UserID:    uuid.New(),
			TableID:   uuid.New(),
			StartTime: time.Now().Add(time.Duration(i*24) * time.Hour),
			EndTime:   time.Now().Add(time.Duration(i*24+2) * time.Hour),
			Status:    domain.BookingStatusPending,
		}
	}

	results := service.ProcessBulkBookings(ctx, bookings, 3)

	assert.NotNil(t, results)
	assert.Equal(t, 10, len(results))
}

func TestSearchAvailableTablesParallel_Success(t *testing.T) {
	service, _, _, _, notificationSvc := setupBookingService()
	defer notificationSvc.Shutdown()

	ctx := context.Background()
	restaurantIDs := []uuid.UUID{uuid.New(), uuid.New()}
	startTime := time.Now().Add(24 * time.Hour)
	endTime := startTime.Add(2 * time.Hour)
	guestCount := 4

	results := service.SearchAvailableTablesParallel(ctx, restaurantIDs, startTime, endTime, guestCount)

	assert.NotNil(t, results)
	assert.Equal(t, len(restaurantIDs), len(results))

	for _, restaurantID := range restaurantIDs {
		tables, exists := results[restaurantID]
		assert.True(t, exists)
		assert.NotEmpty(t, tables)
	}
}

func TestSearchAvailableTablesParallel_EmptyRestaurantList(t *testing.T) {
	service, _, _, _, notificationSvc := setupBookingService()
	defer notificationSvc.Shutdown()

	ctx := context.Background()
	startTime := time.Now().Add(24 * time.Hour)
	endTime := startTime.Add(2 * time.Hour)
	guestCount := 4

	results := service.SearchAvailableTablesParallel(ctx, []uuid.UUID{}, startTime, endTime, guestCount)

	assert.NotNil(t, results)
	assert.Equal(t, 0, len(results))
}

func TestCancelBookingWithRefund_Success(t *testing.T) {
	service, _, _, _, notificationSvc := setupBookingService()
	defer notificationSvc.Shutdown()

	ctx := context.Background()
	bookingID := uuid.New()
	userEmail := "test@example.com"

	err := service.CancelBookingWithRefund(ctx, bookingID, userEmail)

	assert.NoError(t, err)
}

func TestGetBookingStatistics_Success(t *testing.T) {
	service, _, _, _, notificationSvc := setupBookingService()
	defer notificationSvc.Shutdown()

	ctx := context.Background()
	restaurantID := uuid.New()

	stats, err := service.GetBookingStatistics(ctx, restaurantID)

	assert.NoError(t, err)
	assert.NotNil(t, stats)
	assert.Equal(t, 150, stats["total_bookings"])
	assert.Equal(t, 25, stats["active_bookings"])
	assert.Equal(t, 100, stats["completed_bookings"])
	assert.Equal(t, 25, stats["cancelled_bookings"])
}

func TestProcessBooking_Success(t *testing.T) {
	service, _, _, _, notificationSvc := setupBookingService()
	defer notificationSvc.Shutdown()

	ctx := context.Background()
	booking := &domain.Booking{
		ID:        uuid.New(),
		UserID:    uuid.New(),
		TableID:   uuid.New(),
		StartTime: time.Now().Add(24 * time.Hour),
		EndTime:   time.Now().Add(26 * time.Hour),
		Status:    domain.BookingStatusPending,
	}

	err := service.processBooking(ctx, booking)

	assert.NoError(t, err)
}

func TestProcessBooking_ContextCancelled(t *testing.T) {
	service, _, _, _, notificationSvc := setupBookingService()
	defer notificationSvc.Shutdown()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	booking := &domain.Booking{
		ID:        uuid.New(),
		UserID:    uuid.New(),
		TableID:   uuid.New(),
		StartTime: time.Now().Add(24 * time.Hour),
		EndTime:   time.Now().Add(26 * time.Hour),
		Status:    domain.BookingStatusPending,
	}

	err := service.processBooking(ctx, booking)

	assert.Error(t, err)
	assert.Equal(t, context.Canceled, err)
}

func TestConcurrentBookingOperations(t *testing.T) {
	service, _, _, _, notificationSvc := setupBookingService()
	defer notificationSvc.Shutdown()

	ctx := context.Background()
	done := make(chan bool, 5)

	for i := 0; i < 5; i++ {
		go func() {
			userID := uuid.New()
			tableID := uuid.New()
			startTime := time.Now().Add(24 * time.Hour)
			endTime := startTime.Add(2 * time.Hour)
			userEmail := "test@example.com"

			booking, err := service.CreateBookingWithNotification(ctx, userID, tableID, startTime, endTime, 4, userEmail)
			assert.NoError(t, err)
			assert.NotNil(t, booking)
			done <- true
		}()
	}

	for i := 0; i < 5; i++ {
		<-done
	}
}

func TestLargeScaleTableAvailabilityCheck(t *testing.T) {
	service, _, _, _, notificationSvc := setupBookingService()
	defer notificationSvc.Shutdown()

	ctx := context.Background()
	tableIDs := make([]uuid.UUID, 50)
	for i := 0; i < 50; i++ {
		tableIDs[i] = uuid.New()
	}

	startTime := time.Now().Add(24 * time.Hour)
	endTime := startTime.Add(2 * time.Hour)

	results := service.CheckMultipleTablesAvailability(ctx, tableIDs, startTime, endTime)

	assert.NotNil(t, results)
	assert.Equal(t, 50, len(results))
}

func TestMultipleRestaurantSearch(t *testing.T) {
	service, _, _, _, notificationSvc := setupBookingService()
	defer notificationSvc.Shutdown()

	ctx := context.Background()
	restaurantIDs := make([]uuid.UUID, 10)
	for i := 0; i < 10; i++ {
		restaurantIDs[i] = uuid.New()
	}

	startTime := time.Now().Add(24 * time.Hour)
	endTime := startTime.Add(2 * time.Hour)

	results := service.SearchAvailableTablesParallel(ctx, restaurantIDs, startTime, endTime, 4)

	assert.NotNil(t, results)
	assert.Equal(t, 10, len(results))
}
