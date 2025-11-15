package repository

import (
	"errors"

	"github.com/Kiyopon176/restaurant-booking-backend/internal/domain"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

type PaymentRepository interface {
	Create(tx *sqlx.Tx, p *domain.Payment) error
	GetByID(tx *sqlx.Tx, id uuid.UUID) (*domain.Payment, error)
	GetByExternalID(tx *sqlx.Tx, externalID string) (*domain.Payment, error)
	Update(tx *sqlx.Tx, p *domain.Payment) error
	GetByUserID(tx *sqlx.Tx, userID uuid.UUID, pagination Pagination) ([]domain.Payment, int64, error)
}

type paymentRepo struct {
	db *sqlx.DB
}

func NewPaymentRepository(db *sqlx.DB) PaymentRepository {
	return &paymentRepo{db: db}
}

func (r *paymentRepo) exec(tx *sqlx.Tx) sqlx.Ext {
	if tx != nil {
		return tx
	}
	return r.db
}

func (r *paymentRepo) Create(tx *sqlx.Tx, p *domain.Payment) error {
	q := `
		INSERT INTO payments (
			id, user_id, booking_id, amount, payment_method, payment_status,
			external_payment_id, external_payment_url, error_message, created_at, updated_at
		)
		VALUES (
			:id, :user_id, :booking_id, :amount, :payment_method, :payment_status,
			:external_payment_id, :external_payment_url, :error_message, NOW(), NOW()
		)
	`
	_, err := sqlx.NamedExec(r.exec(tx), q, p)
	return err
}

func (r *paymentRepo) GetByID(tx *sqlx.Tx, id uuid.UUID) (*domain.Payment, error) {
	var p domain.Payment
	q := `SELECT * FROM payments WHERE id = $1 LIMIT 1`
	err := sqlx.Get(r.exec(tx), &p, q, id)
	if err != nil {
		return nil, err
	}
	return &p, nil
}

func (r *paymentRepo) GetByExternalID(tx *sqlx.Tx, externalID string) (*domain.Payment, error) {
	var p domain.Payment
	q := `SELECT * FROM payments WHERE external_payment_id = $1 LIMIT 1`
	err := sqlx.Get(r.exec(tx), &p, q, externalID)
	if err != nil {
		if errors.Is(err, sqlx.ErrNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &p, nil
}

func (r *paymentRepo) Update(tx *sqlx.Tx, p *domain.Payment) error {
	q := `
		UPDATE payments SET
			user_id = :user_id,
			booking_id = :booking_id,
			amount = :amount,
			payment_method = :payment_method,
			payment_status = :payment_status,
			external_payment_id = :external_payment_id,
			external_payment_url = :external_payment_url,
			error_message = :error_message,
			updated_at = NOW()
		WHERE id = :id
	`
	_, err := sqlx.NamedExec(r.exec(tx), q, p)
	return err
}

func (r *paymentRepo) GetByUserID(tx *sqlx.Tx, userID uuid.UUID, pagination Pagination) ([]domain.Payment, int64, error) {
	var total int64
	r.exec(tx).QueryRow(`SELECT COUNT(*) FROM payments WHERE user_id = $1`, userID).Scan(&total)

	var list []domain.Payment
	q := `
		SELECT * FROM payments
		WHERE user_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`
	err := sqlx.Select(r.exec(tx), &list, q, userID, pagination.Limit(), pagination.Offset())
	if err != nil {
		return nil, 0, err
	}

	return list, total, nil
}
