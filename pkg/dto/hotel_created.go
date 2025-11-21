package dto

// CreatedHotelResponse represents payload after creating a hotel.
type CreatedHotelResponse struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Address     string `json:"address"`
	Message     string `json:"message"`
}

// CreatedRoomTypeResponse represents payload after creating a room type.
type CreatedRoomTypeResponse struct {
	ID        string  `json:"id"`
	HotelID   string  `json:"hotel_id"`
	Name      string  `json:"name"`
	Capacity  int     `json:"capacity"`
	BasePrice float64 `json:"base_price"`
	Amenities string  `json:"amenities"`
	Message   string  `json:"message"`
}

// CreatedRoomResponse represents payload after creating a room.
type CreatedRoomResponse struct {
	ID         string `json:"id"`
	RoomTypeID string `json:"room_type_id"`
	Number     string `json:"number"`
	Status     string `json:"status"`
	Message    string `json:"message"`
}
