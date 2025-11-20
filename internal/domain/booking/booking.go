package booking

import (
	"context"
	"time"

	"github.com/google/uuid"
)

const (
	StatusPendingPayment = "pending_payment"
	StatusConfirmed      = "confirmed"
	StatusCancelled      = "cancelled"
	StatusCheckedIn      = "checked_in"
	StatusCompleted      = "completed"
)

// Booking aggregate.
type Booking struct {
	ID          uuid.UUID `db:"id"`
	UserID      uuid.UUID `db:"user_id"`
	RoomTypeID  uuid.UUID `db:"room_type_id"`
	CheckIn     time.Time `db:"check_in"`
	CheckOut    time.Time `db:"check_out"`
	Status      string    `db:"status"`
	Guests      int       `db:"guests"`
	TotalPrice  float64   `db:"total_price"`
	TotalNights int       `db:"total_nights"`
	CreatedAt   time.Time `db:"created_at"`
}

// Repository handles persistence.
type Repository interface {
	Create(ctx context.Context, b Booking) error
	FindByID(ctx context.Context, id uuid.UUID) (Booking, error)
	List(ctx context.Context) ([]Booking, error)
	UpdateStatus(ctx context.Context, id uuid.UUID, status string) error
}

// PaymentGateway used by booking service.
type PaymentGateway interface {
	Initiate(ctx context.Context, bookingID uuid.UUID, amount float64) (PaymentResult, error)
}

// NotificationGateway for events.
type NotificationGateway interface {
	Notify(ctx context.Context, event string, payload any) error
}

// PaymentResult carries minimal payment data after initiation.
type PaymentResult struct {
	ID         uuid.UUID
	Status     string
	Provider   string
	PaymentURL string
}
