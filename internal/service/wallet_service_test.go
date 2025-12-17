package service

import (
	"context"
	"restaurant-booking/internal/domain"
	"restaurant-booking/internal/repository"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// Mock WalletRepository
type MockWalletRepository struct {
	mock.Mock
}

func (m *MockWalletRepository) Create(ctx context.Context, w *domain.Wallet) error {
	return m.Called(ctx, w).Error(0)
}

func (m *MockWalletRepository) GetByUserID(ctx context.Context, userID uuid.UUID) (*domain.Wallet, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Wallet), args.Error(1)
}

func (m *MockWalletRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Wallet, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Wallet), args.Error(1)
}

func (m *MockWalletRepository) Update(ctx context.Context, w *domain.Wallet) error {
	return m.Called(ctx, w).Error(0)
}

func (m *MockWalletRepository) GetTransactions(ctx context.Context, walletID uuid.UUID, limit, offset int) ([]*domain.WalletTransaction, error) {
	args := m.Called(ctx, walletID, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.WalletTransaction), args.Error(1)
}

func (m *MockWalletRepository) CreateTransaction(ctx context.Context, tx *domain.WalletTransaction) error {
	return m.Called(ctx, tx).Error(0)
}

// Ensure mock implements interface
var _ repository.WalletRepository = (*MockWalletRepository)(nil)

func setupWalletService() (*walletService, *MockWalletRepository, sqlmock.Sqlmock) {
	repo := new(MockWalletRepository)
	sqlDB, dbMock, _ := sqlmock.New()
	dialector := postgres.New(postgres.Config{
		Conn:       sqlDB,
		DriverName: "postgres",
	})
	db, _ := gorm.Open(dialector, &gorm.Config{})

	service := &walletService{
		walletRepo: repo,
		db:         db,
	}

	return service, repo, dbMock
}

// --- Tests ---

func TestGetOrCreateWallet_Existing(t *testing.T) {
	service, repo, _ := setupWalletService()
	ctx := context.Background()
	userID := uuid.New()
	existingWallet := &domain.Wallet{UserID: userID, Balance: 100}

	repo.On("GetByUserID", ctx, userID).Return(existingWallet, nil)

	wallet, err := service.GetOrCreateWallet(ctx, userID)

	assert.NoError(t, err)
	assert.Equal(t, existingWallet, wallet)
	repo.AssertExpectations(t)
}

func TestGetOrCreateWallet_CreateNew(t *testing.T) {
	service, repo, _ := setupWalletService()
	ctx := context.Background()
	userID := uuid.New()

	repo.On("GetByUserID", ctx, userID).Return(nil, gorm.ErrRecordNotFound)
	repo.On("Create", ctx, mock.AnythingOfType("*domain.Wallet")).Return(nil)

	wallet, err := service.GetOrCreateWallet(ctx, userID)

	assert.NoError(t, err)
	assert.NotNil(t, wallet)
	assert.Equal(t, userID, wallet.UserID)
	repo.AssertExpectations(t)
}

func TestDeposit_Success(t *testing.T) {
	service, _, dbMock := setupWalletService()
	ctx := context.Background()
	userID := uuid.New()
	walletID := uuid.New()

	dbMock.ExpectBegin()
	// GORM adds LIMIT 1 to .First()
	dbMock.ExpectQuery(`SELECT .* FROM "wallets" WHERE user_id = .*`).
		WithArgs(userID, 1).
		WillReturnRows(sqlmock.NewRows([]string{"id", "user_id", "balance"}).
			AddRow(walletID, userID, 1000))

	dbMock.ExpectExec(`UPDATE "wallets"`).WillReturnResult(sqlmock.NewResult(1, 1))
	dbMock.ExpectQuery(`INSERT INTO "wallet_transactions"`).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(uuid.New()))
	dbMock.ExpectCommit()

	err := service.Deposit(ctx, userID, 500, "Deposit")

	assert.NoError(t, err)
	assert.NoError(t, dbMock.ExpectationsWereMet())
}

func TestWithdraw_InsufficientBalance(t *testing.T) {
	service, _, dbMock := setupWalletService()
	ctx := context.Background()
	userID := uuid.New()

	dbMock.ExpectBegin()
	dbMock.ExpectQuery(`SELECT .* FROM "wallets" WHERE user_id = .*`).
		WithArgs(userID, 1).
		WillReturnRows(sqlmock.NewRows([]string{"id", "user_id", "balance"}).
			AddRow(uuid.New(), userID, 100))

	dbMock.ExpectRollback()

	err := service.Withdraw(ctx, userID, 500, "Withdraw")

	assert.Error(t, err)
	assert.Equal(t, ErrInsufficientBalance, err)
	assert.NoError(t, dbMock.ExpectationsWereMet())
}

func TestChargeForBooking_Success(t *testing.T) {
	service, _, dbMock := setupWalletService()
	ctx := context.Background()
	userID := uuid.New()
	bookingID := uuid.New()

	dbMock.ExpectBegin()
	dbMock.ExpectQuery(`SELECT .* FROM "wallets" WHERE user_id = .*`).
		WithArgs(userID, 1).
		WillReturnRows(sqlmock.NewRows([]string{"id", "user_id", "balance"}).
			AddRow(uuid.New(), userID, 2000))

	dbMock.ExpectExec(`UPDATE "wallets"`).WillReturnResult(sqlmock.NewResult(1, 1))
	dbMock.ExpectQuery(`INSERT INTO "wallet_transactions"`).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(uuid.New()))
	dbMock.ExpectCommit()

	err := service.ChargeForBooking(ctx, userID, 500, bookingID)

	assert.NoError(t, err)
	assert.NoError(t, dbMock.ExpectationsWereMet())
}

func TestRefundBooking_Success(t *testing.T) {
	service, _, dbMock := setupWalletService()
	ctx := context.Background()
	userID := uuid.New()
	bookingID := uuid.New()

	dbMock.ExpectBegin()
	dbMock.ExpectQuery(`SELECT .* FROM "wallets" WHERE user_id = .*`).
		WithArgs(userID, 1).
		WillReturnRows(sqlmock.NewRows([]string{"id", "user_id", "balance"}).
			AddRow(uuid.New(), userID, 1000))

	dbMock.ExpectExec(`UPDATE "wallets"`).WillReturnResult(sqlmock.NewResult(1, 1))
	dbMock.ExpectQuery(`INSERT INTO "wallet_transactions"`).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(uuid.New()))
	dbMock.ExpectCommit()

	err := service.RefundBooking(ctx, userID, 300, bookingID, "Refund")

	assert.NoError(t, err)
	assert.NoError(t, dbMock.ExpectationsWereMet())
}

func TestGetTransactions_Success(t *testing.T) {
	service, repo, _ := setupWalletService()
	ctx := context.Background()
	userID := uuid.New()
	walletID := uuid.New()

	wallet := &domain.Wallet{ID: walletID, UserID: userID}
	expectedTxs := []*domain.WalletTransaction{{ID: uuid.New(), WalletID: walletID, Amount: 100}}

	repo.On("GetByUserID", ctx, userID).Return(wallet, nil)
	repo.On("GetTransactions", ctx, walletID, 10, 0).Return(expectedTxs, nil)

	txs, err := service.GetTransactions(ctx, userID, 10, 0)

	assert.NoError(t, err)
	assert.Equal(t, expectedTxs, txs)
	repo.AssertExpectations(t)
}
