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

func setupPaymentService() (*paymentService, *MockPaymentRepository, *MockWalletService, sqlmock.Sqlmock, *gorm.DB) {
	mockPaymentRepo := new(MockPaymentRepository)
	mockWalletService := new(MockWalletService)

	sqlDB, sqlMock, _ := sqlmock.New()
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

	return service, mockPaymentRepo, mockWalletService, sqlMock, db
}

func TestCreatePayment_WalletPayment_Success(t *testing.T) {
	service, mockPaymentRepo, mockWalletService, sqlMock, _ := setupPaymentService()
	ctx := context.Background()

	userID := uuid.New()
	bookingID := uuid.New()
	amount := 10000

	sqlMock.ExpectBegin()
	mockPaymentRepo.On("Create", ctx, mock.AnythingOfType("*domain.Payment")).Return(nil)
	mockPaymentRepo.On("GetByID", ctx, mock.AnythingOfType("uuid.UUID")).Return(&domain.Payment{
		ID:            uuid.New(),
		UserID:        userID,
		Amount:        amount,
		PaymentMethod: domain.PaymentMethodWallet,
		PaymentStatus: domain.PaymentStatusCompleted,
	}, nil)
	mockWalletService.On("ChargeForBooking", ctx, userID, amount, bookingID).Return(nil)
	mockPaymentRepo.On("Update", ctx, mock.AnythingOfType("*domain.Payment")).Return(nil)
	sqlMock.ExpectCommit()

	payment, err := service.CreatePayment(ctx, userID, amount, domain.PaymentMethodWallet, &bookingID)

	assert.NoError(t, err)
	assert.NotNil(t, payment)
	assert.Equal(t, amount, payment.Amount)
	mockPaymentRepo.AssertExpectations(t)
}

func TestCreatePayment_InvalidAmount(t *testing.T) {
	service, _, _, _, _ := setupPaymentService()
	ctx := context.Background()

	payment, err := service.CreatePayment(ctx, uuid.New(), 0, domain.PaymentMethodWallet, nil)

	assert.Error(t, err)
	assert.Nil(t, payment)
	assert.Equal(t, ErrInvalidAmount, err)
}

func TestProcessWalletPayment_Success(t *testing.T) {
	service, mockPaymentRepo, mockWalletService, sqlMock, _ := setupPaymentService()
	ctx := context.Background()

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

	sqlMock.ExpectBegin()
	mockPaymentRepo.On("GetByID", ctx, paymentID).Return(payment, nil)
	mockWalletService.On("ChargeForBooking", ctx, userID, amount, bookingID).Return(nil)
	mockPaymentRepo.On("Update", ctx, payment).Return(nil)
	sqlMock.ExpectCommit()

	err := service.ProcessWalletPayment(ctx, paymentID)

	assert.NoError(t, err)
	assert.Equal(t, domain.PaymentStatusCompleted, payment.PaymentStatus)
	mockPaymentRepo.AssertExpectations(t)
	mockWalletService.AssertExpectations(t)
}

func TestProcessWalletPayment_InsufficientBalance(t *testing.T) {
	service, mockPaymentRepo, mockWalletService, sqlMock, _ := setupPaymentService()
	ctx := context.Background()

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

	sqlMock.ExpectBegin()
	mockPaymentRepo.On("GetByID", ctx, paymentID).Return(payment, nil)
	mockWalletService.On("ChargeForBooking", ctx, userID, 10000, bookingID).Return(ErrInsufficientBalance)
	mockPaymentRepo.On("Update", ctx, payment).Return(nil)
	sqlMock.ExpectRollback()

	err := service.ProcessWalletPayment(ctx, paymentID)

	assert.Error(t, err)
	assert.Equal(t, ErrInsufficientBalance, err)
	assert.Equal(t, domain.PaymentStatusFailed, payment.PaymentStatus)
	mockPaymentRepo.AssertExpectations(t)
	mockWalletService.AssertExpectations(t)
}

func TestCreateHalykPayment_Success(t *testing.T) {
	service, mockPaymentRepo, _, _, _ := setupPaymentService()
	ctx := context.Background()

	paymentID := uuid.New()
	payment := &domain.Payment{
		ID:            paymentID,
		PaymentMethod: domain.PaymentMethodHalyk,
		PaymentStatus: domain.PaymentStatusPending,
	}

	mockPaymentRepo.On("GetByID", ctx, paymentID).Return(payment, nil)
	mockPaymentRepo.On("Update", ctx, payment).Return(nil)

	url, err := service.CreateHalykPayment(ctx, paymentID)

	assert.NoError(t, err)
	assert.NotEmpty(t, url)
	assert.Contains(t, url, "halyk-mock.kz")
	assert.NotNil(t, payment.ExternalPaymentID)
	assert.NotNil(t, payment.ExternalPaymentURL)
	mockPaymentRepo.AssertExpectations(t)
}

func TestCreateKaspiPayment_Success(t *testing.T) {
	service, mockPaymentRepo, _, _, _ := setupPaymentService()
	ctx := context.Background()

	paymentID := uuid.New()
	payment := &domain.Payment{
		ID:            paymentID,
		PaymentMethod: domain.PaymentMethodKaspi,
		PaymentStatus: domain.PaymentStatusPending,
	}

	mockPaymentRepo.On("GetByID", ctx, paymentID).Return(payment, nil)
	mockPaymentRepo.On("Update", ctx, payment).Return(nil)

	url, err := service.CreateKaspiPayment(ctx, paymentID)

	assert.NoError(t, err)
	assert.NotEmpty(t, url)
	assert.Contains(t, url, "kaspi-mock.kz")
	assert.NotNil(t, payment.ExternalPaymentID)
	assert.NotNil(t, payment.ExternalPaymentURL)
	mockPaymentRepo.AssertExpectations(t)
}

func TestProcessExternalPaymentCallback_Success(t *testing.T) {
	service, mockPaymentRepo, mockWalletService, sqlMock, _ := setupPaymentService()
	ctx := context.Background()

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

	sqlMock.ExpectBegin()
	mockPaymentRepo.On("GetByExternalID", ctx, externalID).Return(payment, nil)
	mockPaymentRepo.On("Update", ctx, payment).Return(nil)
	mockWalletService.On("Deposit", ctx, userID, amount, mock.AnythingOfType("string")).Return(nil)
	sqlMock.ExpectCommit()

	err := service.ProcessExternalPaymentCallback(ctx, externalID, true)

	assert.NoError(t, err)
	assert.Equal(t, domain.PaymentStatusCompleted, payment.PaymentStatus)
	mockPaymentRepo.AssertExpectations(t)
	mockWalletService.AssertExpectations(t)
}

func TestProcessExternalPaymentCallback_Failed(t *testing.T) {
	service, mockPaymentRepo, _, sqlMock, _ := setupPaymentService()
	ctx := context.Background()

	externalID := "external-123"
	payment := &domain.Payment{
		ID:                uuid.New(),
		PaymentMethod:     domain.PaymentMethodHalyk,
		PaymentStatus:     domain.PaymentStatusPending,
		ExternalPaymentID: &externalID,
	}

	sqlMock.ExpectBegin()
	mockPaymentRepo.On("GetByExternalID", ctx, externalID).Return(payment, nil)
	mockPaymentRepo.On("Update", ctx, payment).Return(nil)
	sqlMock.ExpectCommit()

	err := service.ProcessExternalPaymentCallback(ctx, externalID, false)

	assert.NoError(t, err)
	assert.Equal(t, domain.PaymentStatusFailed, payment.PaymentStatus)
	assert.NotNil(t, payment.ErrorMessage)
	mockPaymentRepo.AssertExpectations(t)
}

func TestRefundPayment_Success(t *testing.T) {
	service, mockPaymentRepo, mockWalletService, sqlMock, _ := setupPaymentService()
	ctx := context.Background()

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

	sqlMock.ExpectBegin()
	mockPaymentRepo.On("GetByID", ctx, paymentID).Return(payment, nil)
	mockWalletService.On("RefundBooking", ctx, userID, amount, bookingID, tmock.AnythingOfType("string")).Return(nil)
	mockPaymentRepo.On("Update", ctx, payment).Return(nil)
	sqlMock.ExpectCommit()

	err := service.RefundPayment(ctx, paymentID)

	assert.NoError(t, err)
	assert.Equal(t, domain.PaymentStatusRefunded, payment.PaymentStatus)
	mockPaymentRepo.AssertExpectations(t)
	mockWalletService.AssertExpectations(t)
}

func TestRefundPayment_AlreadyRefunded(t *testing.T) {
	service, mockPaymentRepo, _, sqlMock, _ := setupPaymentService()
	ctx := context.Background()

	paymentID := uuid.New()
	payment := &domain.Payment{
		ID:            paymentID,
		PaymentStatus: domain.PaymentStatusRefunded,
	}

	sqlMock.ExpectBegin()
	mockPaymentRepo.On("GetByID", ctx, paymentID).Return(payment, nil)
	sqlMock.ExpectCommit()

	err := service.RefundPayment(ctx, paymentID)

	assert.NoError(t, err)
	mockPaymentRepo.AssertExpectations(t)
}

func TestRefundPayment_InvalidStatus(t *testing.T) {
	service, mockPaymentRepo, _, sqlMock, _ := setupPaymentService()
	ctx := context.Background()

	paymentID := uuid.New()
	payment := &domain.Payment{
		ID:            paymentID,
		PaymentStatus: domain.PaymentStatusPending,
	}

	sqlMock.ExpectBegin()
	mockPaymentRepo.On("GetByID", ctx, paymentID).Return(payment, nil)
	sqlMock.ExpectRollback()

	err := service.RefundPayment(ctx, paymentID)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "can only refund completed payments")
	mockPaymentRepo.AssertExpectations(t)
}

func TestGetPaymentsByUser(t *testing.T) {
	service, mockPaymentRepo, _, _, _ := setupPaymentService()
	ctx := context.Background()

	userID := uuid.New()
	payments := []*domain.Payment{
		{ID: uuid.New(), UserID: userID},
		{ID: uuid.New(), UserID: userID},
	}

	mockPaymentRepo.On("GetByUserID", ctx, userID, 10, 0).Return(payments, nil)

	result, err := service.GetPaymentsByUser(ctx, userID, 10, 0)

	assert.NoError(t, err)
	assert.Len(t, result, 2)
	mockPaymentRepo.AssertExpectations(t)
}
