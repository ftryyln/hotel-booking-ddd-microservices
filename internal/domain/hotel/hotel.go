package hotel

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// Hotel entity.
type Hotel struct {
	ID          uuid.UUID `db:"id"`
	Name        string    `db:"name"`
	Description string    `db:"description"`
	Address     string    `db:"address"`
	CreatedAt   time.Time `db:"created_at"`
}

// RoomType entity.
type RoomType struct {
	ID        uuid.UUID `db:"id"`
	HotelID   uuid.UUID `db:"hotel_id"`
	Name      string    `db:"name"`
	Capacity  int       `db:"capacity"`
	BasePrice float64   `db:"base_price"`
	Amenities string    `db:"amenities"`
}

// Room entity.
type Room struct {
	ID         uuid.UUID `db:"id"`
	RoomTypeID uuid.UUID `db:"room_type_id"`
	Number     string    `db:"number"`
	Status     string    `db:"status"`
}

// Repository contract.
type Repository interface {
	CreateHotel(ctx context.Context, h Hotel) error
	ListHotels(ctx context.Context) ([]Hotel, error)
	CreateRoomType(ctx context.Context, rt RoomType) error
	ListRoomTypes(ctx context.Context, hotelID uuid.UUID) ([]RoomType, error)
	ListAllRoomTypes(ctx context.Context) ([]RoomType, error)
	CreateRoom(ctx context.Context, room Room) error
	GetRoomType(ctx context.Context, id uuid.UUID) (RoomType, error)
	ListRooms(ctx context.Context) ([]Room, error)
	GetHotel(ctx context.Context, id uuid.UUID) (Hotel, error)
}
