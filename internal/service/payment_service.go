package service

import (
	"errors"
	"fmt"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/Kiyopon176/restaurant-booking-backend/internal/domain"
	"github.com/Kiyopon176/restaurant-booking-backend/internal/repository"
)

type PaymentService struct {
	DB            *gorm.DB
	PaymentRepo   repository.PaymentRepository
	WalletService *WalletService
}

func NewPaymentService(db *gorm.DB, pr repository.PaymentRepository, ws *WalletService) *PaymentService {
	return &PaymentService{
		DB:            db,
		PaymentRepo:   pr,
		WalletService: ws,
	}
}

func (s *PaymentService) CreatePayment(userID uuid.UUID, amount int, method domain.PaymentMethod, bookingID *uuid.UUID) (*domain.Payment, error) {
	if amount <= 0 {
		return nil, errors.New("invalid_amount")
	}
	p := &domain.Payment{
		UserID:        userID,
		BookingID:     bookingID,
		Amount:        amount,
		PaymentMethod: method,
		PaymentStatus: domain.PaymentStatusPending,
	}
	if err := s.PaymentRepo.Create(nil, p); err != nil {
		return nil, err
	}
	// If wallet payment, process immediately
	if method == domain.PaymentMethodWallet {
		if err := s.ProcessWalletPayment(p.ID); err != nil {
			return nil, err
		}
	}
	return p, nil
}

func (s *PaymentService) ProcessWalletPayment(paymentID uuid.UUID) error {
	return s.DB.Transaction(func(tx *gorm.DB) error {
		p, err := s.PaymentRepo.GetByID(tx, paymentID)
		if err != nil {
			return err
		}
		if p.PaymentStatus == domain.PaymentStatusCompleted {
			return nil
		}
		if p.PaymentStatus != domain.PaymentStatusPending {
			return fmt.Errorf("payment in invalid state: %s", p.PaymentStatus)
		}
		// attempt to charge wallet
		if err := s.WalletService.ChargeForBooking(p.UserID, p.Amount, uuid.Nil); err != nil {
			return err
		}
		p.PaymentStatus = domain.PaymentStatusCompleted
		if err := s.PaymentRepo.Update(tx, p); err != nil {
			return err
		}
		return nil
	})
}

func randomExternalURL(id string) string {
	return fmt.Sprintf("https://payment.mock/pay/%s", id)
}

func (s *PaymentService) CreateHalykPayment(paymentID uuid.UUID) (string, error) {
	p, err := s.PaymentRepo.GetByID(nil, paymentID)
	if err != nil {
		return "", err
	}
	if p.PaymentStatus != domain.PaymentStatusPending {
		return "", errors.New("payment not pending")
	}
	externalID := uuid.New().String()
	url := randomExternalURL(externalID)
	p.ExternalPaymentID = &externalID
	p.ExternalPaymentURL = &url
	if err := s.PaymentRepo.Update(nil, p); err != nil {
		return "", err
	}
	return url, nil
}

func (s *PaymentService) CreateKaspiPayment(paymentID uuid.UUID) (string, error) {
	return s.CreateHalykPayment(paymentID) // mock identical behavior
}

func (s *PaymentService) ProcessExternalPaymentCallback(externalPaymentID string, success bool) error {
	return s.DB.Transaction(func(tx *gorm.DB) error {
		p, err := s.PaymentRepo.GetByExternalID(tx, externalPaymentID)
		if err != nil {
			return err
		}
		if p == nil {
			return errors.New("payment_not_found")
		}
		if p.PaymentStatus == domain.PaymentStatusCompleted {
			// already processed -> idempotent
			return nil
		}
		if p.PaymentStatus != domain.PaymentStatusPending {
			return fmt.Errorf("payment in invalid state: %s", p.PaymentStatus)
		}
		if success {
			p.PaymentStatus = domain.PaymentStatusCompleted
			if err := s.PaymentRepo.Update(tx, p); err != nil {
				return err
			}
			// top-up wallet for external payments
			if p.PaymentMethod == domain.PaymentMethodHalyk || p.PaymentMethod == domain.PaymentMethodKaspi {
				// deposit into wallet
				if err := s.WalletService.Deposit(p.UserID, p.Amount, "Top-up via external:"+externalPaymentID); err != nil {
					return err
				}
			}
		} else {
			p.PaymentStatus = domain.PaymentStatusFailed
			if err := s.PaymentRepo.Update(tx, p); err != nil {
				return err
			}
		}
		return nil
	})
}

func (s *PaymentService) RefundPayment(paymentID uuid.UUID) error {
	return s.DB.Transaction(func(tx *gorm.DB) error {
		p, err := s.PaymentRepo.GetByID(tx, paymentID)
		if err != nil {
			return err
		}
		if p.PaymentStatus == domain.PaymentStatusRefunded {
			return nil
		}
		if p.PaymentStatus != domain.PaymentStatusCompleted {
			return errors.New("cannot_refund_non_completed_payment")
		}
		// if payment was via wallet -> refund to wallet
		if p.PaymentMethod == domain.PaymentMethodWallet {
			if err := s.WalletService.Deposit(p.UserID, p.Amount, "Refund for payment:"+p.ID.String()); err != nil {
				return err
			}
		}
		p.PaymentStatus = domain.PaymentStatusRefunded
		if err := s.PaymentRepo.Update(tx, p); err != nil {
			return err
		}
		return nil
	})
}
