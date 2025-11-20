package dto

import "time"

// BookingRequest is used for booking creation.
type BookingRequest struct {
	UserID     string    `json:"user_id"`
	RoomTypeID string    `json:"room_type_id"`
	CheckIn    time.Time `json:"check_in"`
	CheckOut   time.Time `json:"check_out"`
	Guests     int       `json:"guests"`
}

// BookingResponse returns booking info.
type BookingResponse struct {
	ID          string    `json:"id"`
	Status      string    `json:"status"`
	Guests      int       `json:"guests"`
	TotalNights int       `json:"total_nights"`
	TotalPrice  float64   `json:"total_price"`
	CheckIn     time.Time `json:"check_in"`
	CheckOut    time.Time `json:"check_out"`
	Payment     *PaymentResponse `json:"payment,omitempty"`
}

// BookingAggregateResponse merges booking+payment.
type BookingAggregateResponse struct {
	Booking BookingResponse `json:"booking"`
	Payment PaymentResponse `json:"payment"`
}

// CheckpointRequest handles lifecycle updates.
type CheckpointRequest struct {
	Action string `json:"action"`
}
