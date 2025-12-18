package service

import (
	"context"
	"errors"
	"fmt"
	"restaurant-booking/internal/domain"
	"restaurant-booking/internal/repository"
	"restaurant-booking/pkg/logger"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

var (
	ErrPaymentNotFound         = errors.New("payment not found")
	ErrInvalidPaymentStatus    = errors.New("invalid payment status")
	ErrPaymentAlreadyProcessed = errors.New("payment already processed")
)

type PaymentService interface {
	CreatePayment(ctx context.Context, userID uuid.UUID, amount int, method domain.PaymentMethod, bookingID *uuid.UUID) (*domain.Payment, error)
	ProcessWalletPayment(ctx context.Context, paymentID uuid.UUID) error
	CreateHalykPayment(ctx context.Context, paymentID uuid.UUID) (string, error)
	CreateKaspiPayment(ctx context.Context, paymentID uuid.UUID) (string, error)
	ProcessExternalPaymentCallback(ctx context.Context, externalPaymentID string, success bool) error
	RefundPayment(ctx context.Context, paymentID uuid.UUID) error
	GetPaymentsByUser(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*domain.Payment, error)
}

type paymentService struct {
	paymentRepo   repository.PaymentRepository
	walletService WalletService
	db            *gorm.DB
	log           logger.Logger
}

func NewPaymentService(
	paymentRepo repository.PaymentRepository,
	walletService WalletService,
	db *gorm.DB,
	log logger.Logger,
) PaymentService {
	return &paymentService{
		paymentRepo:   paymentRepo,
		walletService: walletService,
		db:            db,
		log:           log,
	}
}

func (s *paymentService) CreatePayment(ctx context.Context, userID uuid.UUID, amount int, method domain.PaymentMethod, bookingID *uuid.UUID) (*domain.Payment, error) {
	if amount <= 0 {
		return nil, ErrInvalidAmount
	}

	payment := &domain.Payment{
		UserID:        userID,
		BookingID:     bookingID,
		Amount:        amount,
		PaymentMethod: method,
		PaymentStatus: domain.PaymentStatusPending,
	}

	if err := s.paymentRepo.Create(ctx, payment); err != nil {
		return nil, err
	}

	if method == domain.PaymentMethodWallet {
		if err := s.ProcessWalletPayment(ctx, payment.ID); err != nil {
			return nil, err
		}

		payment, _ = s.paymentRepo.GetByID(ctx, payment.ID)
	}

	return payment, nil
}

func (s *paymentService) ProcessWalletPayment(ctx context.Context, paymentID uuid.UUID) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		payment, err := s.paymentRepo.GetByID(ctx, paymentID)
		if err != nil {
			return err
		}

		if payment.PaymentStatus == domain.PaymentStatusCompleted {
			return nil
		}

		if payment.PaymentStatus != domain.PaymentStatusPending {
			return fmt.Errorf("%w: %s", ErrInvalidPaymentStatus, payment.PaymentStatus)
		}

		var bookingID uuid.UUID
		if payment.BookingID != nil {
			bookingID = *payment.BookingID
		}

		if err := s.walletService.ChargeForBooking(ctx, payment.UserID, payment.Amount, bookingID); err != nil {
			payment.PaymentStatus = domain.PaymentStatusFailed
			errMsg := err.Error()
			payment.ErrorMessage = &errMsg
			_ = s.paymentRepo.Update(ctx, payment)
			return err
		}

		payment.PaymentStatus = domain.PaymentStatusCompleted
		return s.paymentRepo.Update(ctx, payment)
	})
}

func (s *paymentService) CreateHalykPayment(ctx context.Context, paymentID uuid.UUID) (string, error) {
	payment, err := s.paymentRepo.GetByID(ctx, paymentID)
	if err != nil {
		return "", err
	}

	if payment.PaymentStatus != domain.PaymentStatusPending {
		return "", ErrInvalidPaymentStatus
	}

	externalID := uuid.New().String()
	url := fmt.Sprintf("https://halyk-mock.kz/pay/%s", externalID)

	payment.ExternalPaymentID = &externalID
	payment.ExternalPaymentURL = &url

	if err := s.paymentRepo.Update(ctx, payment); err != nil {
		return "", err
	}

	return url, nil
}

func (s *paymentService) CreateKaspiPayment(ctx context.Context, paymentID uuid.UUID) (string, error) {
	payment, err := s.paymentRepo.GetByID(ctx, paymentID)
	if err != nil {
		return "", err
	}

	if payment.PaymentStatus != domain.PaymentStatusPending {
		return "", ErrInvalidPaymentStatus
	}

	externalID := uuid.New().String()
	url := fmt.Sprintf("https://kaspi-mock.kz/pay/%s", externalID)

	payment.ExternalPaymentID = &externalID
	payment.ExternalPaymentURL = &url

	if err := s.paymentRepo.Update(ctx, payment); err != nil {
		return "", err
	}

	return url, nil
}

func (s *paymentService) ProcessExternalPaymentCallback(ctx context.Context, externalPaymentID string, success bool) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		payment, err := s.paymentRepo.GetByExternalID(ctx, externalPaymentID)
		if err != nil {
			return err
		}

		if payment == nil {
			return ErrPaymentNotFound
		}

		if payment.PaymentStatus == domain.PaymentStatusCompleted {
			return nil
		}

		if payment.PaymentStatus != domain.PaymentStatusPending {
			return fmt.Errorf("%w: %s", ErrInvalidPaymentStatus, payment.PaymentStatus)
		}

		if success {
			payment.PaymentStatus = domain.PaymentStatusCompleted
			if err := s.paymentRepo.Update(ctx, payment); err != nil {
				return err
			}

			if payment.PaymentMethod == domain.PaymentMethodHalyk || payment.PaymentMethod == domain.PaymentMethodKaspi {
				desc := fmt.Sprintf("Top-up via %s (Payment ID: %s)", payment.PaymentMethod, payment.ID)
				return s.walletService.Deposit(ctx, payment.UserID, payment.Amount, desc)
			}
		} else {
			payment.PaymentStatus = domain.PaymentStatusFailed
			failMsg := "External payment failed"
			payment.ErrorMessage = &failMsg
			return s.paymentRepo.Update(ctx, payment)
		}

		return nil
	})
}

func (s *paymentService) RefundPayment(ctx context.Context, paymentID uuid.UUID) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		payment, err := s.paymentRepo.GetByID(ctx, paymentID)
		if err != nil {
			return err
		}

		if payment.PaymentStatus == domain.PaymentStatusRefunded {
			return nil
		}

		if payment.PaymentStatus != domain.PaymentStatusCompleted {
			return errors.New("can only refund completed payments")
		}

		var bookingID uuid.UUID
		if payment.BookingID != nil {
			bookingID = *payment.BookingID
		}

		reason := fmt.Sprintf("Refund for payment %s", payment.ID)
		if err := s.walletService.RefundBooking(ctx, payment.UserID, payment.Amount, bookingID, reason); err != nil {
			return err
		}

		payment.PaymentStatus = domain.PaymentStatusRefunded
		return s.paymentRepo.Update(ctx, payment)
	})
}

func (s *paymentService) GetPaymentsByUser(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*domain.Payment, error) {
	return s.paymentRepo.GetByUserID(ctx, userID, limit, offset)
}
