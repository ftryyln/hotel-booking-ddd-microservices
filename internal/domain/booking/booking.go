package booking

import (
	"context"
	"time"

	"github.com/google/uuid"

	"github.com/ftryyln/hotel-booking-microservices/pkg/domain"
	pkgErrors "github.com/ftryyln/hotel-booking-microservices/pkg/errors"
	"github.com/ftryyln/hotel-booking-microservices/pkg/query"
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
	ID          uuid.UUID
	UserID      uuid.UUID
	RoomTypeID  uuid.UUID
	CheckIn     time.Time
	CheckOut    time.Time
	Status      string
	Guests      int
	TotalPrice  float64
	TotalNights int
	CreatedAt   time.Time

	// events stores domain events raised by this aggregate
	events []domain.DomainEvent
}

// Confirm transitions booking to confirmed state.
func (b *Booking) Confirm() error {
	if b.Status != StatusPendingPayment {
		return pkgErrors.New("bad_request", "cannot confirm booking that is not pending payment")
	}
	b.Status = StatusConfirmed
	b.RecordEvent(NewBookingConfirmed(b.ID))
	return nil
}

// Cancel transitions booking to cancelled state.
func (b *Booking) Cancel(reason string) error {
	if b.Status == StatusCompleted {
		return pkgErrors.New("bad_request", "cannot cancel completed booking")
	}
	b.Status = StatusCancelled
	b.RecordEvent(NewBookingCancelled(b.ID, reason))
	return nil
}

// GuestCheckIn transitions booking to checked_in state.
func (b *Booking) GuestCheckIn() error {
	if b.Status != StatusConfirmed {
		return pkgErrors.New("bad_request", "booking must be confirmed before check-in")
	}
	b.Status = StatusCheckedIn
	b.RecordEvent(NewBookingCheckedIn(b.ID))
	return nil
}


// Complete transitions booking to completed state.
func (b *Booking) Complete() error {
	if b.Status != StatusCheckedIn {
		return pkgErrors.New("bad_request", "booking must be checked-in before completion")
	}
	b.Status = StatusCompleted
	b.RecordEvent(NewBookingCompleted(b.ID))
	return nil
}

// Events returns the domain events raised by this aggregate.
func (b *Booking) Events() []domain.DomainEvent {
	return b.events
}

// ClearEvents clears the domain events.
func (b *Booking) ClearEvents() {
	b.events = nil
}

func (b *Booking) RecordEvent(event domain.DomainEvent) {
	b.events = append(b.events, event)
}


// BookingReader handles queries (CQRS Read Side).
type BookingReader interface {
	FindByID(ctx context.Context, id uuid.UUID) (Booking, error)
	List(ctx context.Context, opts query.Options) ([]Booking, error)
	FindByUserID(ctx context.Context, userID uuid.UUID) ([]Booking, error)
}

// BookingWriter handles commands (CQRS Write Side).
type BookingWriter interface {
	Create(ctx context.Context, b Booking) error
	UpdateStatus(ctx context.Context, id uuid.UUID, status string) error
	Save(ctx context.Context, b Booking) error
}

// Repository handles persistence (combines Read and Write).
type Repository interface {
	BookingReader
	BookingWriter
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
