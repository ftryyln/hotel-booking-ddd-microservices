package hotel

import (
	"context"
	"database/sql"

	"github.com/google/uuid"

	domain "github.com/ftryyln/hotel-booking-microservices/internal/domain/hotel"
	"github.com/ftryyln/hotel-booking-microservices/pkg/dto"
	"github.com/ftryyln/hotel-booking-microservices/pkg/errors"
)

// Service exposes hotel catalog operations.
type Service struct {
	repo domain.Repository
}

func NewService(repo domain.Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) CreateHotel(ctx context.Context, req dto.HotelRequest) (uuid.UUID, error) {
	h := domain.Hotel{ID: uuid.New(), Name: req.Name, Description: req.Description, Address: req.Address}
	return h.ID, s.repo.CreateHotel(ctx, h)
}

func (s *Service) ListHotels(ctx context.Context) ([]dto.HotelResponse, error) {
	hotels, err := s.repo.ListHotels(ctx)
	if err != nil {
		return nil, err
	}
	var resp []dto.HotelResponse
	for _, h := range hotels {
		roomTypes, _ := s.repo.ListRoomTypes(ctx, h.ID)
		var summaries []dto.RoomTypeSummary
		for _, rt := range roomTypes {
			summaries = append(summaries, dto.RoomTypeSummary{
				ID:       rt.ID.String(),
				Name:     rt.Name,
				Capacity: rt.Capacity,
				Price:    rt.BasePrice,
			})
		}
		resp = append(resp, dto.HotelResponse{
			ID:          h.ID.String(),
			Name:        h.Name,
			Description: h.Description,
			Address:     h.Address,
			CreatedAt:   h.CreatedAt,
			RoomTypes:   summaries,
		})
	}
	return resp, nil
}

func (s *Service) CreateRoomType(ctx context.Context, req dto.RoomTypeRequest) (uuid.UUID, error) {
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
	room := domain.Room{ID: uuid.New(), RoomTypeID: uuid.MustParse(req.RoomTypeID), Number: req.Number, Status: req.Status}
	return room.ID, s.repo.CreateRoom(ctx, room)
}

func (s *Service) GetHotel(ctx context.Context, id uuid.UUID) (dto.HotelResponse, error) {
	h, err := s.repo.GetHotel(ctx, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return dto.HotelResponse{}, errors.New("not_found", "hotel not found")
		}
		return dto.HotelResponse{}, err
	}
	roomTypes, _ := s.repo.ListRoomTypes(ctx, h.ID)
	var summaries []dto.RoomTypeSummary
	for _, rt := range roomTypes {
		summaries = append(summaries, dto.RoomTypeSummary{
			ID:       rt.ID.String(),
			Name:     rt.Name,
			Capacity: rt.Capacity,
			Price:    rt.BasePrice,
		})
	}
	return dto.HotelResponse{
		ID:          h.ID.String(),
		Name:        h.Name,
		Description: h.Description,
		Address:     h.Address,
		CreatedAt:   h.CreatedAt,
		RoomTypes:   summaries,
	}, nil
}

func (s *Service) ListRoomTypes(ctx context.Context) ([]dto.RoomTypeResponse, error) {
	rts, err := s.repo.ListAllRoomTypes(ctx)
	if err != nil {
		return nil, err
	}
	var resp []dto.RoomTypeResponse
	for _, rt := range rts {
		resp = append(resp, dto.RoomTypeResponse{
			ID:        rt.ID.String(),
			HotelID:   rt.HotelID.String(),
			Name:      rt.Name,
			Capacity:  rt.Capacity,
			BasePrice: rt.BasePrice,
			Amenities: rt.Amenities,
		})
	}
	return resp, nil
}

func (s *Service) ListRooms(ctx context.Context) ([]dto.RoomResponse, error) {
	rooms, err := s.repo.ListRooms(ctx)
	if err != nil {
		return nil, err
	}
	var resp []dto.RoomResponse
	for _, r := range rooms {
		resp = append(resp, dto.RoomResponse{
			ID:         r.ID.String(),
			RoomTypeID: r.RoomTypeID.String(),
			Number:     r.Number,
			Status:     r.Status,
		})
	}
	return resp, nil
}
