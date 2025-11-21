package dto

import (
	"fmt"
	"strings"
	"time"
)

// Date allows YYYY-MM-DD or RFC3339 in JSON and stores as time.Time.
type Date struct{ time.Time }

// UnmarshalJSON accepts date-only (YYYY-MM-DD) or RFC3339 timestamps.
func (d *Date) UnmarshalJSON(b []byte) error {
	s := strings.Trim(string(b), `"`)
	if s == "" || s == "null" {
		d.Time = time.Time{}
		return nil
	}
	if t, err := time.Parse(time.RFC3339, s); err == nil {
		d.Time = t
		return nil
	}
	if t, err := time.Parse("2006-01-02", s); err == nil {
		d.Time = t
		return nil
	}
	return fmt.Errorf("invalid date format")
}

// MarshalJSON outputs YYYY-MM-DD for consistency.
func (d Date) MarshalJSON() ([]byte, error) {
	if d.Time.IsZero() {
		return []byte("null"), nil
	}
	return []byte(fmt.Sprintf(`"%s"`, d.Time.Format("2006-01-02"))), nil
}

// BookingRequest is used for booking creation.
type BookingRequest struct {
	UserID     string `json:"user_id"`
	RoomTypeID string `json:"room_type_id"`
	CheckIn    Date   `json:"check_in"`
	CheckOut   Date   `json:"check_out"`
	Guests     int    `json:"guests"`
}

// BookingResponse returns booking info.
type BookingResponse struct {
	ID          string           `json:"id"`
	Status      string           `json:"status"`
	Guests      int              `json:"guests"`
	TotalNights int              `json:"total_nights"`
	TotalPrice  float64          `json:"total_price"`
	CheckIn     time.Time        `json:"check_in"`
	CheckOut    time.Time        `json:"check_out"`
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
