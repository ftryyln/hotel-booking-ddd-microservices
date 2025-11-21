package assembler

import (
	"time"

	"github.com/google/uuid"

	domain "github.com/ftryyln/hotel-booking-microservices/internal/domain/booking"
	"github.com/ftryyln/hotel-booking-microservices/pkg/dto"
	pkgErrors "github.com/ftryyln/hotel-booking-microservices/pkg/errors"
)

// CreateCommand represents inbound booking creation intent.
type CreateCommand struct {
	UserID     uuid.UUID
	RoomTypeID uuid.UUID
	CheckIn    time.Time
	CheckOut   time.Time
	Guests     int
}

// ToResponse maps domain booking plus optional payment info to response DTO.
func ToResponse(b domain.Booking, payment domain.PaymentResult) dto.BookingResponse {
	resp := dto.BookingResponse{
		ID:          b.ID.String(),
		Status:      b.Status,
		Guests:      b.Guests,
		TotalNights: b.TotalNights,
		TotalPrice:  b.TotalPrice,
		CheckIn:     b.CheckIn,
		CheckOut:    b.CheckOut,
	}
	if payment.ID != uuid.Nil {
		resp.Payment = &dto.PaymentResponse{
			ID:         payment.ID.String(),
			Status:     payment.Status,
			Provider:   payment.Provider,
			PaymentURL: payment.PaymentURL,
		}
	}
	return resp
}

// FromRequest validates incoming DTO to command.
func FromRequest(req dto.BookingRequest) (CreateCommand, error) {
	userID, err := uuid.Parse(req.UserID)
	if err != nil {
		return CreateCommand{}, pkgErrors.New("bad_request", "invalid user id")
	}
	roomTypeID, err := uuid.Parse(req.RoomTypeID)
	if err != nil {
		return CreateCommand{}, pkgErrors.New("bad_request", "invalid room type id")
	}
	if req.CheckIn.IsZero() || req.CheckOut.IsZero() {
		return CreateCommand{}, pkgErrors.New("bad_request", "date required")
	}
	if !req.CheckIn.Time.Before(req.CheckOut.Time) {
		return CreateCommand{}, pkgErrors.New("bad_request", "check_in must be before check_out")
	}
	guests := req.Guests
	if guests <= 0 {
		guests = 1
	}
	return CreateCommand{
		UserID:     userID,
		RoomTypeID: roomTypeID,
		CheckIn:    req.CheckIn.Time,
		CheckOut:   req.CheckOut.Time,
		Guests:     guests,
	}, nil
}
