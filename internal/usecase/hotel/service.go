package hotel

import (
	"context"
	"database/sql"

	"github.com/google/uuid"

	"github.com/ftryyln/hotel-booking-microservices/internal/usecase/hotel/assembler"
	domain "github.com/ftryyln/hotel-booking-microservices/internal/domain/hotel"
	"github.com/ftryyln/hotel-booking-microservices/pkg/dto"
	"github.com/ftryyln/hotel-booking-microservices/pkg/errors"
	"github.com/ftryyln/hotel-booking-microservices/pkg/query"
	"github.com/ftryyln/hotel-booking-microservices/pkg/valueobject"
)

// Service exposes hotel catalog operations.
type Service struct {
	repo domain.Repository
}

func NewService(repo domain.Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) CreateHotel(ctx context.Context, req dto.HotelRequest) (uuid.UUID, error) {
	name, addr, err := valueobject.ValidateHotel(req.Name, req.Address)
	if err != nil {
		return uuid.Nil, err
	}
	h := domain.Hotel{ID: uuid.New(), Name: name, Description: req.Description, Address: addr}
	return h.ID, s.repo.CreateHotel(ctx, h)
}

func (s *Service) ListHotels(ctx context.Context, opts query.Options) ([]assembler.HotelAggregate, error) {
	hotels, err := s.repo.ListHotels(ctx, opts.Normalize(50))
	if err != nil {
		return nil, err
	}
	var aggs []assembler.HotelAggregate
	for _, h := range hotels {
		roomTypes, _ := s.repo.ListRoomTypes(ctx, h.ID)
		aggs = append(aggs, assembler.HotelAggregate{
			Hotel:     h,
			RoomTypes: roomTypes,
		})
	}
	return aggs, nil
}

func (s *Service) CreateRoomType(ctx context.Context, req dto.RoomTypeRequest) (uuid.UUID, error) {
	if err := valueobject.RoomTypeSpec(req.Capacity, req.BasePrice); err != nil {
		return uuid.Nil, err
	}
	rt := domain.RoomType{
		ID:        uuid.New(),
		HotelID:   uuid.MustParse(req.HotelID),
		Name:      req.Name,
		Capacity:  req.Capacity,
		BasePrice: req.BasePrice,
		Amenities: req.Amenities,
	}
	return rt.ID, s.repo.CreateRoomType(ctx, rt)
}

func (s *Service) CreateRoom(ctx context.Context, req dto.RoomRequest) (uuid.UUID, error) {
	status, err := valueobject.NormalizeRoomStatus(req.Status)
	if err != nil {
		return uuid.Nil, err
	}
	room := domain.Room{
		ID:         uuid.New(),
		RoomTypeID: uuid.MustParse(req.RoomTypeID),
		Number:     req.Number,
		Status:     string(status),
	}
	return room.ID, s.repo.CreateRoom(ctx, room)
}

func (s *Service) GetHotel(ctx context.Context, id uuid.UUID, opts query.Options) (assembler.HotelAggregate, error) {
	h, err := s.repo.GetHotel(ctx, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return assembler.HotelAggregate{}, errors.New("not_found", "hotel not found")
		}
		return assembler.HotelAggregate{}, err
	}
	roomTypes, _ := s.repo.ListRoomTypes(ctx, h.ID)
	return assembler.HotelAggregate{Hotel: h, RoomTypes: roomTypes}, nil
}

func (s *Service) ListRoomTypes(ctx context.Context, opts query.Options) ([]domain.RoomType, error) {
	rts, err := s.repo.ListAllRoomTypes(ctx, opts.Normalize(50))
	if err != nil {
		return nil, err
	}
	return rts, nil
}

func (s *Service) ListRooms(ctx context.Context, opts query.Options) ([]domain.Room, error) {
	rooms, err := s.repo.ListRooms(ctx, opts.Normalize(50))
	if err != nil {
		return nil, err
	}
	return rooms, nil
}

func (s *Service) UpdateHotel(ctx context.Context, id uuid.UUID, req dto.HotelUpdateRequest) error {
	name, addr, err := valueobject.ValidateHotel(req.Name, req.Address)
	if err != nil {
		return err
	}
	h := domain.Hotel{
		Name:        name,
		Description: req.Description,
		Address:     addr,
	}
	return s.repo.UpdateHotel(ctx, id, h)
}

func (s *Service) DeleteHotel(ctx context.Context, id uuid.UUID) error {
	return s.repo.DeleteHotel(ctx, id)
}

func (s *Service) GetRoom(ctx context.Context, id uuid.UUID) (domain.Room, error) {
	room, err := s.repo.GetRoom(ctx, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return domain.Room{}, errors.New("not_found", "room not found")
		}
		return domain.Room{}, err
	}
	return room, nil
}

func (s *Service) UpdateRoom(ctx context.Context, id uuid.UUID, req dto.RoomUpdateRequest) error {
	// Get existing room first
	existing, err := s.repo.GetRoom(ctx, id)
	if err != nil {
		return err
	}

	// Update only provided fields
	if req.Number != "" {
		existing.Number = req.Number
	}
	if req.Status != "" {
		status, err := valueobject.NormalizeRoomStatus(req.Status)
		if err != nil {
			return err
		}
		existing.Status = string(status)
	}

	return s.repo.UpdateRoom(ctx, id, existing)
}

func (s *Service) DeleteRoom(ctx context.Context, id uuid.UUID) error {
	return s.repo.DeleteRoom(ctx, id)
}
