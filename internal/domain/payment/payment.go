package payment

import (
	"context"
	"time"

	"github.com/google/uuid"
)

const (
	StatusPending = "pending"
	StatusPaid    = "paid"
	StatusFailed  = "failed"
)

// Payment aggregates payment state.
type Payment struct {
	ID         uuid.UUID `db:"id"`
	BookingID  uuid.UUID `db:"booking_id"`
	Amount     float64   `db:"amount"`
	Currency   string    `db:"currency"`
	Status     string    `db:"status"`
	Provider   string    `db:"provider"`
	PaymentURL string    `db:"payment_url"`
	CreatedAt  time.Time `db:"created_at"`
}

// Provider integrates external gateway.
type Provider interface {
	Initiate(ctx context.Context, payment Payment) (Payment, error)
	VerifySignature(ctx context.Context, payload, signature string) bool
	Refund(ctx context.Context, payment Payment, reason string) (string, error)
}

// Repository persists payments.
type Repository interface {
	Create(ctx context.Context, p Payment) error
	FindByID(ctx context.Context, id uuid.UUID) (Payment, error)
	FindByBookingID(ctx context.Context, bookingID uuid.UUID) (Payment, error)
	UpdateStatus(ctx context.Context, id uuid.UUID, status, paymentURL string) error
}

// BookingStatusUpdater notifies booking service.
type BookingStatusUpdater interface {
	Update(ctx context.Context, bookingID uuid.UUID, status string) error
}
