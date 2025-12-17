package service

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"restaurant-booking/internal/domain"
	"restaurant-booking/internal/repository"

	"github.com/google/uuid"
)

type BookingService struct {
	bookingRepo     repository.BookingRepository
	tableRepo       repository.TableRepository
	restaurantRepo  repository.RestaurantRepository
	notificationSvc *NotificationService
	mu              sync.RWMutex
}

func NewBookingService(
	bookingRepo repository.BookingRepository,
	tableRepo repository.TableRepository,
	restaurantRepo repository.RestaurantRepository,
	notificationSvc *NotificationService,
) *BookingService {
	return &BookingService{
		bookingRepo:     bookingRepo,
		tableRepo:       tableRepo,
		restaurantRepo:  restaurantRepo,
		notificationSvc: notificationSvc,
	}
}

type BookingResult struct {
	Booking *domain.Booking
	Error   error
}

func (s *BookingService) CreateBookingWithNotification(
	ctx context.Context,
	userID uuid.UUID,
	tableID uuid.UUID,
	startTime, endTime time.Time,
	guestCount int,
	userEmail string,
) (*domain.Booking, error) {

	bookingChan := make(chan BookingResult, 1)
	notificationChan := make(chan error, 1)

	go func() {
		booking := &domain.Booking{
			ID:        uuid.New(),
			UserID:    userID,
			TableID:   tableID,
			StartTime: startTime,
			EndTime:   endTime,
			Status:    "pending",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		time.Sleep(100 * time.Millisecond)

		bookingChan <- BookingResult{Booking: booking, Error: nil}
	}()

	result := <-bookingChan
	if result.Error != nil {
		return nil, result.Error
	}

	go func() {
		err := s.notificationSvc.SendEmail(
			userEmail,
			"Booking Confirmation",
			fmt.Sprintf("Your booking for %s has been created. Booking ID: %s",
				startTime.Format("2006-01-02 15:04"), result.Booking.ID),
		)
		notificationChan <- err
	}()

	log.Printf("Booking created: %s, notification being sent asynchronously", result.Booking.ID)

	return result.Booking, nil
}

func (s *BookingService) CheckMultipleTablesAvailability(
	ctx context.Context,
	tableIDs []uuid.UUID,
	startTime, endTime time.Time,
) map[uuid.UUID]bool {
	results := make(map[uuid.UUID]bool)
	resultsChan := make(chan struct {
		TableID   uuid.UUID
		Available bool
	}, len(tableIDs))

	var wg sync.WaitGroup

	for _, tableID := range tableIDs {
		wg.Add(1)
		go func(tid uuid.UUID) {
			defer wg.Done()

			time.Sleep(50 * time.Millisecond)
			available := time.Now().UnixNano()%2 == 0

			resultsChan <- struct {
				TableID   uuid.UUID
				Available bool
			}{TableID: tid, Available: available}
		}(tableID)
	}

	go func() {
		wg.Wait()
		close(resultsChan)
	}()

	for result := range resultsChan {
		results[result.TableID] = result.Available
		log.Printf("Table %s availability: %v", result.TableID, result.Available)
	}

	return results
}

func (s *BookingService) ProcessBulkBookings(
	ctx context.Context,
	bookings []domain.Booking,
	maxConcurrent int,
) []BookingResult {
	results := make([]BookingResult, len(bookings))

	semaphore := make(chan struct{}, maxConcurrent)
	var wg sync.WaitGroup

	for i, booking := range bookings {
		wg.Add(1)

		go func(index int, b domain.Booking) {
			defer wg.Done()

			semaphore <- struct{}{}
			defer func() { <-semaphore }()

			if err := s.processBooking(ctx, &b); err != nil {
				results[index] = BookingResult{Booking: nil, Error: err}
			} else {
				results[index] = BookingResult{Booking: &b, Error: nil}
			}
		}(i, booking)
	}

	wg.Wait()
	log.Printf("Processed %d bookings", len(bookings))

	return results
}

func (s *BookingService) processBooking(ctx context.Context, booking *domain.Booking) error {

	time.Sleep(100 * time.Millisecond)

	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
		return nil
	}
}

func (s *BookingService) SearchAvailableTablesParallel(
	ctx context.Context,
	restaurantIDs []uuid.UUID,
	startTime, endTime time.Time,
	guestCount int,
) map[uuid.UUID][]uuid.UUID {
	type RestaurantTables struct {
		RestaurantID uuid.UUID
		TableIDs     []uuid.UUID
	}

	resultsChan := make(chan RestaurantTables, len(restaurantIDs))
	var wg sync.WaitGroup

	for _, restaurantID := range restaurantIDs {
		wg.Add(1)

		go func(rid uuid.UUID) {
			defer wg.Done()

			time.Sleep(100 * time.Millisecond)

			availableTables := []uuid.UUID{uuid.New(), uuid.New()}

			resultsChan <- RestaurantTables{
				RestaurantID: rid,
				TableIDs:     availableTables,
			}

			log.Printf("Found %d available tables in restaurant %s",
				len(availableTables), rid)
		}(restaurantID)
	}

	go func() {
		wg.Wait()
		close(resultsChan)
	}()

	results := make(map[uuid.UUID][]uuid.UUID)
	for rt := range resultsChan {
		results[rt.RestaurantID] = rt.TableIDs
	}

	return results
}

func (s *BookingService) CancelBookingWithRefund(
	ctx context.Context,
	bookingID uuid.UUID,
	userEmail string,
) error {
	var wg sync.WaitGroup
	errChan := make(chan error, 3)

	wg.Add(1)
	go func() {
		defer wg.Done()
		time.Sleep(50 * time.Millisecond)
		log.Printf("Booking %s status updated to cancelled", bookingID)
		errChan <- nil
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		time.Sleep(100 * time.Millisecond)
		log.Printf("Refund processed for booking %s", bookingID)
		errChan <- nil
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		err := s.notificationSvc.SendEmail(
			userEmail,
			"Booking Cancelled",
			fmt.Sprintf("Your booking %s has been cancelled and refund processed", bookingID),
		)
		errChan <- err
	}()

	wg.Wait()
	close(errChan)

	for err := range errChan {
		if err != nil {
			return err
		}
	}

	return nil
}

func (s *BookingService) GetBookingStatistics(ctx context.Context, restaurantID uuid.UUID) (map[string]int, error) {
	type StatResult struct {
		Key   string
		Value int
		Error error
	}

	statsChan := make(chan StatResult, 4)
	var wg sync.WaitGroup

	stats := []struct {
		key  string
		calc func() (int, error)
	}{
		{"total_bookings", func() (int, error) {
			time.Sleep(50 * time.Millisecond)
			return 150, nil
		}},
		{"active_bookings", func() (int, error) {
			time.Sleep(30 * time.Millisecond)
			return 25, nil
		}},
		{"completed_bookings", func() (int, error) {
			time.Sleep(40 * time.Millisecond)
			return 100, nil
		}},
		{"cancelled_bookings", func() (int, error) {
			time.Sleep(35 * time.Millisecond)
			return 25, nil
		}},
	}

	for _, stat := range stats {
		wg.Add(1)
		go func(key string, calcFunc func() (int, error)) {
			defer wg.Done()
			value, err := calcFunc()
			statsChan <- StatResult{Key: key, Value: value, Error: err}
		}(stat.key, stat.calc)
	}

	go func() {
		wg.Wait()
		close(statsChan)
	}()

	results := make(map[string]int)
	for result := range statsChan {
		if result.Error != nil {
			return nil, result.Error
		}
		results[result.Key] = result.Value
	}

	log.Printf("Statistics calculated for restaurant %s: %+v", restaurantID, results)
	return results, nil
}
