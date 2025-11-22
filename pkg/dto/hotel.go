package dto

import "time"

// HotelRequest defines admin input.
type HotelRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Address     string `json:"address"`
}

// RoomTypeRequest configures hotel room types.
type RoomTypeRequest struct {
	HotelID   string  `json:"hotel_id"`
	Name      string  `json:"name"`
	Capacity  int     `json:"capacity"`
	BasePrice float64 `json:"base_price"`
	Amenities string  `json:"amenities"`
}

// RoomTypeResponse exposes room type details.
type RoomTypeResponse struct {
	ID        string  `json:"id"`
	HotelID   string  `json:"hotel_id"`
	Name      string  `json:"name"`
	Capacity  int     `json:"capacity"`
	BasePrice float64 `json:"base_price"`
	Amenities string  `json:"amenities"`
}

// RoomRequest describes a physical room.
type RoomRequest struct {
	RoomTypeID string `json:"room_type_id"`
	Number     string `json:"number"`
	Status     string `json:"status"`
}

// RoomResponse shows room detail.
type RoomResponse struct {
	ID         string `json:"id"`
	RoomTypeID string `json:"room_type_id"`
	Number     string `json:"number"`
	Status     string `json:"status"`
}

// HotelResponse surfaces public data.
type HotelResponse struct {
	ID          string            `json:"id"`
	Name        string            `json:"name"`
	Description string            `json:"description"`
	Address     string            `json:"address"`
	CreatedAt   time.Time         `json:"created_at"`
	RoomTypes   []RoomTypeSummary `json:"room_types"`
}

// RoomTypeSummary short view.
type RoomTypeSummary struct {
	ID       string  `json:"id"`
	Name     string  `json:"name"`
	Capacity int     `json:"capacity"`
	Price    float64 `json:"price"`
}

// HotelUpdateRequest for updating hotel details.
type HotelUpdateRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Address     string `json:"address"`
}

// RoomUpdateRequest for updating room details.
type RoomUpdateRequest struct {
	Number string `json:"number,omitempty"`
	Status string `json:"status,omitempty"`
}
