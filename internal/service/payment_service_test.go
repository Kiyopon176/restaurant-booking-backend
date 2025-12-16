package service

import (
	"context"
	_ "errors"
	"restaurant-booking/internal/domain"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	tmock "github.com/stretchr/testify/mock"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// Mock PaymentRepository
type MockPaymentRepository struct {
	tmock.Mock
}

func (m *MockPaymentRepository) Create(ctx context.Context, payment *domain.Payment) error {
	args := m.Called(ctx, payment)
	return args.Error(0)
}

func (m *MockPaymentRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Payment, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Payment), args.Error(1)
}

func (m *MockPaymentRepository) GetByExternalID(ctx context.Context, externalID string) (*domain.Payment, error) {
	args := m.Called(ctx, externalID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Payment), args.Error(1)
}

func (m *MockPaymentRepository) GetByUserID(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*domain.Payment, error) {
	args := m.Called(ctx, userID, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.Payment), args.Error(1)
}

func (m *MockPaymentRepository) Update(ctx context.Context, payment *domain.Payment) error {
	args := m.Called(ctx, payment)
	return args.Error(0)
}

// Mock WalletService
type MockWalletService struct {
	tmock.Mock
}

func (m *MockWalletService) GetOrCreateWallet(ctx context.Context, userID uuid.UUID) (*domain.Wallet, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Wallet), args.Error(1)
}

func (m *MockWalletService) GetBalance(ctx context.Context, userID uuid.UUID) (int, error) {
	args := m.Called(ctx, userID)
	return args.Int(0), args.Error(1)
}

func (m *MockWalletService) Deposit(ctx context.Context, userID uuid.UUID, amount int, description string) error {
	args := m.Called(ctx, userID, amount, description)
	return args.Error(0)
}

func (m *MockWalletService) Withdraw(ctx context.Context, userID uuid.UUID, amount int, description string) error {
	args := m.Called(ctx, userID, amount, description)
	return args.Error(0)
}

func (m *MockWalletService) ChargeForBooking(ctx context.Context, userID uuid.UUID, amount int, bookingID uuid.UUID) error {
	args := m.Called(ctx, userID, amount, bookingID)
	return args.Error(0)
}

func (m *MockWalletService) RefundBooking(ctx context.Context, userID uuid.UUID, amount int, bookingID uuid.UUID, reason string) error {
	args := m.Called(ctx, userID, amount, bookingID, reason)
	return args.Error(0)
}

func (m *MockWalletService) GetTransactions(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*domain.WalletTransaction, error) {
	args := m.Called(ctx, userID, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.WalletTransaction), args.Error(1)
}

// Helper function
func setupPaymentService() (*paymentService, *MockPaymentRepository, *MockWalletService, sqlmock.Sqlmock, *gorm.DB) {
	mockPaymentRepo := new(MockPaymentRepository)
	mockWalletService := new(MockWalletService)

	// Setup sqlmock for transaction testing
	sqlDB, mock, _ := sqlmock.New()
	dialector := postgres.New(postgres.Config{
		Conn:       sqlDB,
		DriverName: "postgres",
	})
	db, _ := gorm.Open(dialector, &gorm.Config{})

	service := &paymentService{
		paymentRepo:   mockPaymentRepo,
		walletService: mockWalletService,
		db:            db,
	}

	return service, mockPaymentRepo, mockWalletService, mock, db
}

// Test CreatePayment - Wallet Payment Success
func TestCreatePayment_WalletPayment_Success(t *testing.T) {
	service, mockPaymentRepo, mockWalletService, mock, _ := setupPaymentService()
	ctx := context.Background()

	// Arrange
	userID := uuid.New()
	bookingID := uuid.New()
	amount := 10000

	mock.ExpectBegin()
	mockPaymentRepo.On("Create", ctx, tmock.AnythingOfType("*domain.Payment")).Return(nil)
	mockPaymentRepo.On("GetByID", ctx, tmock.AnythingOfType("uuid.UUID")).Return(&domain.Payment{
		ID:            uuid.New(),
		UserID:        userID,
		Amount:        amount,
		PaymentMethod: domain.PaymentMethodWallet,
		PaymentStatus: domain.PaymentStatusCompleted,
	}, nil)
	mockWalletService.On("ChargeForBooking", ctx, userID, amount, bookingID).Return(nil)
	mockPaymentRepo.On("Update", ctx, tmock.AnythingOfType("*domain.Payment")).Return(nil)
	mock.ExpectCommit()

	// Act
	payment, err := service.CreatePayment(ctx, userID, amount, domain.PaymentMethodWallet, &bookingID)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, payment)
	assert.Equal(t, amount, payment.Amount)
	mockPaymentRepo.AssertExpectations(t)
}

// Test CreatePayment - Invalid Amount
func TestCreatePayment_InvalidAmount(t *testing.T) {
	service, _, _, _, _ := setupPaymentService()
	ctx := context.Background()

	// Act
	payment, err := service.CreatePayment(ctx, uuid.New(), 0, domain.PaymentMethodWallet, nil)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, payment)
	assert.Equal(t, ErrInvalidAmount, err)
}

// Test ProcessWalletPayment - Success
func TestProcessWalletPayment_Success(t *testing.T) {
	service, mockPaymentRepo, mockWalletService, mock, _ := setupPaymentService()
	ctx := context.Background()

	// Arrange
	paymentID := uuid.New()
	userID := uuid.New()
	bookingID := uuid.New()
	amount := 10000

	payment := &domain.Payment{
		ID:            paymentID,
		UserID:        userID,
		BookingID:     &bookingID,
		Amount:        amount,
		PaymentMethod: domain.PaymentMethodWallet,
		PaymentStatus: domain.PaymentStatusPending,
	}

	mock.ExpectBegin()
	mockPaymentRepo.On("GetByID", ctx, paymentID).Return(payment, nil)
	mockWalletService.On("ChargeForBooking", ctx, userID, amount, bookingID).Return(nil)
	mockPaymentRepo.On("Update", ctx, payment).Return(nil)
	mock.ExpectCommit()

	// Act
	err := service.ProcessWalletPayment(ctx, paymentID)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, domain.PaymentStatusCompleted, payment.PaymentStatus)
	mockPaymentRepo.AssertExpectations(t)
	mockWalletService.AssertExpectations(t)
}

// Test ProcessWalletPayment - Insufficient Balance
func TestProcessWalletPayment_InsufficientBalance(t *testing.T) {
	service, mockPaymentRepo, mockWalletService, mock, _ := setupPaymentService()
	ctx := context.Background()

	// Arrange
	paymentID := uuid.New()
	userID := uuid.New()
	bookingID := uuid.New()

	payment := &domain.Payment{
		ID:            paymentID,
		UserID:        userID,
		BookingID:     &bookingID,
		Amount:        10000,
		PaymentMethod: domain.PaymentMethodWallet,
		PaymentStatus: domain.PaymentStatusPending,
	}

	mock.ExpectBegin()
	mockPaymentRepo.On("GetByID", ctx, paymentID).Return(payment, nil)
	mockWalletService.On("ChargeForBooking", ctx, userID, 10000, bookingID).Return(ErrInsufficientBalance)
	mockPaymentRepo.On("Update", ctx, payment).Return(nil)
	mock.ExpectRollback()

	// Act
	err := service.ProcessWalletPayment(ctx, paymentID)

	// Assert
	assert.Error(t, err)
	assert.Equal(t, ErrInsufficientBalance, err)
	assert.Equal(t, domain.PaymentStatusFailed, payment.PaymentStatus)
	mockPaymentRepo.AssertExpectations(t)
	mockWalletService.AssertExpectations(t)
}

// Test CreateHalykPayment - Success
func TestCreateHalykPayment_Success(t *testing.T) {
	service, mockPaymentRepo, _, _, _ := setupPaymentService()
	ctx := context.Background()

	// Arrange
	paymentID := uuid.New()
	payment := &domain.Payment{
		ID:            paymentID,
		PaymentMethod: domain.PaymentMethodHalyk,
		PaymentStatus: domain.PaymentStatusPending,
	}

	mockPaymentRepo.On("GetByID", ctx, paymentID).Return(payment, nil)
	mockPaymentRepo.On("Update", ctx, payment).Return(nil)

	// Act
	url, err := service.CreateHalykPayment(ctx, paymentID)

	// Assert
	assert.NoError(t, err)
	assert.NotEmpty(t, url)
	assert.Contains(t, url, "halyk-mock.kz")
	assert.NotNil(t, payment.ExternalPaymentID)
	assert.NotNil(t, payment.ExternalPaymentURL)
	mockPaymentRepo.AssertExpectations(t)
}

// Test CreateKaspiPayment - Success
func TestCreateKaspiPayment_Success(t *testing.T) {
	service, mockPaymentRepo, _, _, _ := setupPaymentService()
	ctx := context.Background()

	// Arrange
	paymentID := uuid.New()
	payment := &domain.Payment{
		ID:            paymentID,
		PaymentMethod: domain.PaymentMethodKaspi,
		PaymentStatus: domain.PaymentStatusPending,
	}

	mockPaymentRepo.On("GetByID", ctx, paymentID).Return(payment, nil)
	mockPaymentRepo.On("Update", ctx, payment).Return(nil)

	// Act
	url, err := service.CreateKaspiPayment(ctx, paymentID)

	// Assert
	assert.NoError(t, err)
	assert.NotEmpty(t, url)
	assert.Contains(t, url, "kaspi-mock.kz")
	assert.NotNil(t, payment.ExternalPaymentID)
	assert.NotNil(t, payment.ExternalPaymentURL)
	mockPaymentRepo.AssertExpectations(t)
}

// Test ProcessExternalPaymentCallback - Success
func TestProcessExternalPaymentCallback_Success(t *testing.T) {
	service, mockPaymentRepo, mockWalletService, mock, _ := setupPaymentService()
	ctx := context.Background()

	// Arrange
	externalID := "external-123"
	userID := uuid.New()
	amount := 10000

	payment := &domain.Payment{
		ID:                uuid.New(),
		UserID:            userID,
		Amount:            amount,
		PaymentMethod:     domain.PaymentMethodHalyk,
		PaymentStatus:     domain.PaymentStatusPending,
		ExternalPaymentID: &externalID,
	}

	mock.ExpectBegin()
	mockPaymentRepo.On("GetByExternalID", ctx, externalID).Return(payment, nil)
	mockPaymentRepo.On("Update", ctx, payment).Return(nil)
	mockWalletService.On("Deposit", ctx, userID, amount, tmock.AnythingOfType("string")).Return(nil)
	mock.ExpectCommit()

	// Act
	err := service.ProcessExternalPaymentCallback(ctx, externalID, true)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, domain.PaymentStatusCompleted, payment.PaymentStatus)
	mockPaymentRepo.AssertExpectations(t)
	mockWalletService.AssertExpectations(t)
}

// Test ProcessExternalPaymentCallback - Failed Payment
func TestProcessExternalPaymentCallback_Failed(t *testing.T) {
	service, mockPaymentRepo, _, mock, _ := setupPaymentService()
	ctx := context.Background()

	// Arrange
	externalID := "external-123"
	payment := &domain.Payment{
		ID:                uuid.New(),
		PaymentMethod:     domain.PaymentMethodHalyk,
		PaymentStatus:     domain.PaymentStatusPending,
		ExternalPaymentID: &externalID,
	}

	mock.ExpectBegin()
	mockPaymentRepo.On("GetByExternalID", ctx, externalID).Return(payment, nil)
	mockPaymentRepo.On("Update", ctx, payment).Return(nil)
	mock.ExpectCommit()

	// Act
	err := service.ProcessExternalPaymentCallback(ctx, externalID, false)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, domain.PaymentStatusFailed, payment.PaymentStatus)
	assert.NotNil(t, payment.ErrorMessage)
	mockPaymentRepo.AssertExpectations(t)
}

// Test RefundPayment - Success
func TestRefundPayment_Success(t *testing.T) {
	service, mockPaymentRepo, mockWalletService, mock, _ := setupPaymentService()
	ctx := context.Background()

	// Arrange
	paymentID := uuid.New()
	userID := uuid.New()
	bookingID := uuid.New()
	amount := 10000

	payment := &domain.Payment{
		ID:            paymentID,
		UserID:        userID,
		BookingID:     &bookingID,
		Amount:        amount,
		PaymentStatus: domain.PaymentStatusCompleted,
	}

	mock.ExpectBegin()
	mockPaymentRepo.On("GetByID", ctx, paymentID).Return(payment, nil)
	mockWalletService.On("RefundBooking", ctx, userID, amount, bookingID, tmock.AnythingOfType("string")).Return(nil)
	mockPaymentRepo.On("Update", ctx, payment).Return(nil)
	mock.ExpectCommit()

	// Act
	err := service.RefundPayment(ctx, paymentID)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, domain.PaymentStatusRefunded, payment.PaymentStatus)
	mockPaymentRepo.AssertExpectations(t)
	mockWalletService.AssertExpectations(t)
}

// Test RefundPayment - Already Refunded (Idempotent)
func TestRefundPayment_AlreadyRefunded(t *testing.T) {
	service, mockPaymentRepo, _, mock, _ := setupPaymentService()
	ctx := context.Background()

	// Arrange
	paymentID := uuid.New()
	payment := &domain.Payment{
		ID:            paymentID,
		PaymentStatus: domain.PaymentStatusRefunded,
	}

	mock.ExpectBegin()
	mockPaymentRepo.On("GetByID", ctx, paymentID).Return(payment, nil)
	mock.ExpectCommit()

	// Act
	err := service.RefundPayment(ctx, paymentID)

	// Assert
	assert.NoError(t, err)
	mockPaymentRepo.AssertExpectations(t)
}

// Test RefundPayment - Invalid Status
func TestRefundPayment_InvalidStatus(t *testing.T) {
	service, mockPaymentRepo, _, mock, _ := setupPaymentService()
	ctx := context.Background()

	// Arrange
	paymentID := uuid.New()
	payment := &domain.Payment{
		ID:            paymentID,
		PaymentStatus: domain.PaymentStatusPending,
	}

	mock.ExpectBegin()
	mockPaymentRepo.On("GetByID", ctx, paymentID).Return(payment, nil)
	mock.ExpectRollback()

	// Act
	err := service.RefundPayment(ctx, paymentID)

	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "can only refund completed payments")
	mockPaymentRepo.AssertExpectations(t)
}

// Test GetPaymentsByUser
func TestGetPaymentsByUser(t *testing.T) {
	service, mockPaymentRepo, _, _, _ := setupPaymentService()
	ctx := context.Background()

	// Arrange
	userID := uuid.New()
	payments := []*domain.Payment{
		{ID: uuid.New(), UserID: userID},
		{ID: uuid.New(), UserID: userID},
	}

	mockPaymentRepo.On("GetByUserID", ctx, userID, 10, 0).Return(payments, nil)

	// Act
	result, err := service.GetPaymentsByUser(ctx, userID, 10, 0)

	// Assert
	assert.NoError(t, err)
	assert.Len(t, result, 2)
	mockPaymentRepo.AssertExpectations(t)
}
