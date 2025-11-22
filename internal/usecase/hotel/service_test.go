package hotel_test

import (
	"context"
	stdErrors "errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"

	domain "github.com/ftryyln/hotel-booking-microservices/internal/domain/hotel"
	"github.com/ftryyln/hotel-booking-microservices/internal/usecase/hotel"
	"github.com/ftryyln/hotel-booking-microservices/pkg/dto"
	"github.com/ftryyln/hotel-booking-microservices/pkg/query"
)

func TestCreateHotelValidates(t *testing.T) {
	repo := &hotelRepoStub{}
	svc := hotel.NewService(repo)

	_, err := svc.CreateHotel(context.Background(), dto.HotelRequest{Name: "", Address: ""})
	require.Error(t, err)

	id, err := svc.CreateHotel(context.Background(), dto.HotelRequest{Name: "Hilton", Address: "Jakarta"})
	require.NoError(t, err)
	require.NotEqual(t, uuid.Nil, id)
}

func TestCreateRoomTypeAndRoom(t *testing.T) {
	repo := &hotelRepoStub{}
	svc := hotel.NewService(repo)
	hID, _ := uuid.Parse("11111111-1111-1111-1111-111111111111")

	rtID, err := svc.CreateRoomType(context.Background(), dto.RoomTypeRequest{
		HotelID:   hID.String(),
		Name:      "Deluxe",
		Capacity:  2,
		BasePrice: 1000,
	})
	require.NoError(t, err)
	require.NotEqual(t, uuid.Nil, rtID)

	_, err = svc.CreateRoom(context.Background(), dto.RoomRequest{
		RoomTypeID: rtID.String(),
		Number:     "101",
		Status:     "available",
	})
	require.NoError(t, err)

	_, err = svc.CreateRoom(context.Background(), dto.RoomRequest{
		RoomTypeID: rtID.String(),
		Number:     "102",
		Status:     "invalid",
	})
	require.Error(t, err)
}

func TestListHotelsRooms(t *testing.T) {
	repo := &hotelRepoStub{}
	hID := uuid.New()
	now := time.Now()
	repo.hotels = append(repo.hotels, domain.Hotel{ID: hID, Name: "H", Address: "Addr", CreatedAt: now})
	repo.roomTypes = append(repo.roomTypes, domain.RoomType{ID: uuid.New(), HotelID: hID, Name: "RT", Capacity: 2, BasePrice: 10})
	repo.rooms = append(repo.rooms, domain.Room{ID: uuid.New(), RoomTypeID: repo.roomTypes[0].ID, Number: "1", Status: "available"})
	svc := hotel.NewService(repo)

	hotels, err := svc.ListHotels(context.Background(), query.Options{Limit: 10})
	require.NoError(t, err)
	require.Len(t, hotels, 1)

	rt, err := svc.ListRoomTypes(context.Background(), query.Options{Limit: 10})
	require.NoError(t, err)
	require.Len(t, rt, 1)

	rooms, err := svc.ListRooms(context.Background(), query.Options{Limit: 10})
	require.NoError(t, err)
	require.Len(t, rooms, 1)
}

// stub repo
type hotelRepoStub struct {
	hotels    []domain.Hotel
	roomTypes []domain.RoomType
	rooms     []domain.Room
}

func (h *hotelRepoStub) CreateHotel(ctx context.Context, v domain.Hotel) error {
	h.hotels = append(h.hotels, v)
	return nil
}
func (h *hotelRepoStub) ListHotels(ctx context.Context, opts query.Options) ([]domain.Hotel, error) {
	return h.hotels, nil
}
func (h *hotelRepoStub) CreateRoomType(ctx context.Context, rt domain.RoomType) error {
	h.roomTypes = append(h.roomTypes, rt)
	return nil
}
func (h *hotelRepoStub) ListRoomTypes(ctx context.Context, hotelID uuid.UUID) ([]domain.RoomType, error) {
	var out []domain.RoomType
	for _, rt := range h.roomTypes {
		if rt.HotelID == hotelID {
			out = append(out, rt)
		}
	}
	return out, nil
}
func (h *hotelRepoStub) ListAllRoomTypes(ctx context.Context, opts query.Options) ([]domain.RoomType, error) {
	return h.roomTypes, nil
}
func (h *hotelRepoStub) CreateRoom(ctx context.Context, r domain.Room) error {
	h.rooms = append(h.rooms, r)
	return nil
}
func (h *hotelRepoStub) GetRoomType(ctx context.Context, id uuid.UUID) (domain.RoomType, error) {
	for _, rt := range h.roomTypes {
		if rt.ID == id {
			return rt, nil
		}
	}
	return domain.RoomType{}, stdErrors.New("not found")
}
func (h *hotelRepoStub) ListRooms(ctx context.Context, opts query.Options) ([]domain.Room, error) {
	return h.rooms, nil
}
func (h *hotelRepoStub) GetHotel(ctx context.Context, id uuid.UUID) (domain.Hotel, error) {
	for _, ht := range h.hotels {
		if ht.ID == id {
			return ht, nil
		}
	}
	return domain.Hotel{}, stdErrors.New("not found")
}

func (h *hotelRepoStub) UpdateHotel(ctx context.Context, id uuid.UUID, hotel domain.Hotel) error {
	for i, ht := range h.hotels {
		if ht.ID == id {
			h.hotels[i].Name = hotel.Name
			h.hotels[i].Description = hotel.Description
			h.hotels[i].Address = hotel.Address
			return nil
		}
	}
	return stdErrors.New("not found")
}

func (h *hotelRepoStub) DeleteHotel(ctx context.Context, id uuid.UUID) error {
	for i, ht := range h.hotels {
		if ht.ID == id {
			h.hotels = append(h.hotels[:i], h.hotels[i+1:]...)
			return nil
		}
	}
	return stdErrors.New("not found")
}

func (h *hotelRepoStub) GetRoom(ctx context.Context, id uuid.UUID) (domain.Room, error) {
	for _, r := range h.rooms {
		if r.ID == id {
			return r, nil
		}
	}
	return domain.Room{}, stdErrors.New("not found")
}

func (h *hotelRepoStub) UpdateRoom(ctx context.Context, id uuid.UUID, room domain.Room) error {
	for i, r := range h.rooms {
		if r.ID == id {
			h.rooms[i].Number = room.Number
			h.rooms[i].Status = room.Status
			return nil
		}
	}
	return stdErrors.New("not found")
}

func (h *hotelRepoStub) DeleteRoom(ctx context.Context, id uuid.UUID) error {
	for i, r := range h.rooms {
		if r.ID == id {
			h.rooms = append(h.rooms[:i], h.rooms[i+1:]...)
			return nil
		}
	}
	return stdErrors.New("not found")
}

func TestUpdateHotel(t *testing.T) {
	repo := &hotelRepoStub{}
	svc := hotel.NewService(repo)
	
	// Create a hotel first
	hID, err := svc.CreateHotel(context.Background(), dto.HotelRequest{
		Name:        "Original Hotel",
		Description: "Original description",
		Address:     "Original address",
	})
	require.NoError(t, err)
	
	// Update the hotel
	err = svc.UpdateHotel(context.Background(), hID, dto.HotelUpdateRequest{
		Name:        "Updated Hotel",
		Description: "Updated description",
		Address:     "Updated address",
	})
	require.NoError(t, err)
	
	// Verify update
	h, err := svc.GetHotel(context.Background(), hID, query.Options{})
	require.NoError(t, err)
	require.Equal(t, "Updated Hotel", h.Hotel.Name)
	require.Equal(t, "Updated description", h.Hotel.Description)
	require.Equal(t, "Updated address", h.Hotel.Address)
}

func TestUpdateHotelNotFound(t *testing.T) {
	repo := &hotelRepoStub{}
	svc := hotel.NewService(repo)
	
	err := svc.UpdateHotel(context.Background(), uuid.New(), dto.HotelUpdateRequest{
		Name:    "Test",
		Address: "Test",
	})
	require.Error(t, err)
}

func TestDeleteHotel(t *testing.T) {
	repo := &hotelRepoStub{}
	svc := hotel.NewService(repo)
	
	// Create a hotel
	hID, err := svc.CreateHotel(context.Background(), dto.HotelRequest{
		Name:    "Test Hotel",
		Address: "Test address",
	})
	require.NoError(t, err)
	
	// Delete the hotel
	err = svc.DeleteHotel(context.Background(), hID)
	require.NoError(t, err)
	
	// Verify deletion
	_, err = svc.GetHotel(context.Background(), hID, query.Options{})
	require.Error(t, err)
}

func TestGetRoom(t *testing.T) {
	repo := &hotelRepoStub{}
	svc := hotel.NewService(repo)
	
	// Create room type and room
	rtID, _ := svc.CreateRoomType(context.Background(), dto.RoomTypeRequest{
		HotelID:   uuid.New().String(),
		Name:      "Deluxe",
		Capacity:  2,
		BasePrice: 1000,
	})
	
	roomID, err := svc.CreateRoom(context.Background(), dto.RoomRequest{
		RoomTypeID: rtID.String(),
		Number:     "101",
		Status:     "available",
	})
	require.NoError(t, err)
	
	// Get room
	room, err := svc.GetRoom(context.Background(), roomID)
	require.NoError(t, err)
	require.Equal(t, "101", room.Number)
	require.Equal(t, "available", room.Status)
}

func TestUpdateRoom(t *testing.T) {
	repo := &hotelRepoStub{}
	svc := hotel.NewService(repo)
	
	// Create room
	rtID, _ := svc.CreateRoomType(context.Background(), dto.RoomTypeRequest{
		HotelID:   uuid.New().String(),
		Name:      "Deluxe",
		Capacity:  2,
		BasePrice: 1000,
	})
	
	roomID, err := svc.CreateRoom(context.Background(), dto.RoomRequest{
		RoomTypeID: rtID.String(),
		Number:     "101",
		Status:     "available",
	})
	require.NoError(t, err)
	
	// Update room
	err = svc.UpdateRoom(context.Background(), roomID, dto.RoomUpdateRequest{
		Number: "102",
		Status: "maintenance",
	})
	require.NoError(t, err)
	
	// Verify update
	room, err := svc.GetRoom(context.Background(), roomID)
	require.NoError(t, err)
	require.Equal(t, "102", room.Number)
	require.Equal(t, "maintenance", room.Status)
}

func TestUpdateRoomPartial(t *testing.T) {
	repo := &hotelRepoStub{}
	svc := hotel.NewService(repo)
	
	// Create room
	rtID, _ := svc.CreateRoomType(context.Background(), dto.RoomTypeRequest{
		HotelID:   uuid.New().String(),
		Name:      "Deluxe",
		Capacity:  2,
		BasePrice: 1000,
	})
	
	roomID, err := svc.CreateRoom(context.Background(), dto.RoomRequest{
		RoomTypeID: rtID.String(),
		Number:     "101",
		Status:     "available",
	})
	require.NoError(t, err)
	
	// Update only status
	err = svc.UpdateRoom(context.Background(), roomID, dto.RoomUpdateRequest{
		Status: "maintenance",
	})
	require.NoError(t, err)
	
	// Verify partial update
	room, err := svc.GetRoom(context.Background(), roomID)
	require.NoError(t, err)
	require.Equal(t, "101", room.Number) // Number unchanged
	require.Equal(t, "maintenance", room.Status) // Status updated
}

func TestDeleteRoom(t *testing.T) {
	repo := &hotelRepoStub{}
	svc := hotel.NewService(repo)
	
	// Create room
	rtID, _ := svc.CreateRoomType(context.Background(), dto.RoomTypeRequest{
		HotelID:   uuid.New().String(),
		Name:      "Deluxe",
		Capacity:  2,
		BasePrice: 1000,
	})
	
	roomID, err := svc.CreateRoom(context.Background(), dto.RoomRequest{
		RoomTypeID: rtID.String(),
		Number:     "101",
		Status:     "available",
	})
	require.NoError(t, err)
	
	// Delete room
	err = svc.DeleteRoom(context.Background(), roomID)
	require.NoError(t, err)
	
	// Verify deletion
	_, err = svc.GetRoom(context.Background(), roomID)
	require.Error(t, err)
}
