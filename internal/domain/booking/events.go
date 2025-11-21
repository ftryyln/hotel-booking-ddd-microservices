package booking

import (
	"github.com/google/uuid"

	"github.com/ftryyln/hotel-booking-microservices/pkg/domain"
)

// Event type constants
const (
	EventTypeBookingCreated   = "booking.created"
	EventTypeBookingConfirmed = "booking.confirmed"
	EventTypeBookingCancelled = "booking.cancelled"
	EventTypeBookingCheckedIn = "booking.checked_in"
	EventTypeBookingCompleted = "booking.completed"
)

// BookingCreated event is raised when a new booking is created.
type BookingCreated struct {
	domain.BaseEvent
	BookingID  uuid.UUID
	UserID     uuid.UUID
	RoomTypeID uuid.UUID
	TotalPrice float64
	Guests     int
}

// NewBookingCreated creates a new BookingCreated event.
func NewBookingCreated(bookingID, userID, roomTypeID uuid.UUID, totalPrice float64, guests int) BookingCreated {
	return BookingCreated{
		BaseEvent:  domain.NewBaseEvent(bookingID, EventTypeBookingCreated),
		BookingID:  bookingID,
		UserID:     userID,
		RoomTypeID: roomTypeID,
		TotalPrice: totalPrice,
		Guests:     guests,
	}
}

// BookingConfirmed event is raised when a booking is confirmed after payment.
type BookingConfirmed struct {
	domain.BaseEvent
	BookingID uuid.UUID
}

// NewBookingConfirmed creates a new BookingConfirmed event.
func NewBookingConfirmed(bookingID uuid.UUID) BookingConfirmed {
	return BookingConfirmed{
		BaseEvent: domain.NewBaseEvent(bookingID, EventTypeBookingConfirmed),
		BookingID: bookingID,
	}
}

// BookingCancelled event is raised when a booking is cancelled.
type BookingCancelled struct {
	domain.BaseEvent
	BookingID uuid.UUID
	Reason    string
}

// NewBookingCancelled creates a new BookingCancelled event.
func NewBookingCancelled(bookingID uuid.UUID, reason string) BookingCancelled {
	return BookingCancelled{
		BaseEvent: domain.NewBaseEvent(bookingID, EventTypeBookingCancelled),
		BookingID: bookingID,
		Reason:    reason,
	}
}

// BookingCheckedIn event is raised when a guest checks in.
type BookingCheckedIn struct {
	domain.BaseEvent
	BookingID uuid.UUID
}

// NewBookingCheckedIn creates a new BookingCheckedIn event.
func NewBookingCheckedIn(bookingID uuid.UUID) BookingCheckedIn {
	return BookingCheckedIn{
		BaseEvent: domain.NewBaseEvent(bookingID, EventTypeBookingCheckedIn),
		BookingID: bookingID,
	}
}

// BookingCompleted event is raised when a booking is completed after checkout.
type BookingCompleted struct {
	domain.BaseEvent
	BookingID uuid.UUID
}

// NewBookingCompleted creates a new BookingCompleted event.
func NewBookingCompleted(bookingID uuid.UUID) BookingCompleted {
	return BookingCompleted{
		BaseEvent: domain.NewBaseEvent(bookingID, EventTypeBookingCompleted),
		BookingID: bookingID,
	}
}
