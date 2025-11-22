package assembler

import (
	domain "github.com/ftryyln/hotel-booking-microservices/internal/domain/hotel"
	"github.com/ftryyln/hotel-booking-microservices/pkg/dto"
)

// HotelAggregate represents hotel with its room types.
type HotelAggregate struct {
	Hotel     domain.Hotel
	RoomTypes []domain.RoomType
}

// ToHotelResponse maps aggregate to DTO response.
func ToHotelResponse(agg HotelAggregate) dto.HotelResponse {
	var summaries []dto.RoomTypeSummary
	for _, rt := range agg.RoomTypes {
		summaries = append(summaries, dto.RoomTypeSummary{
			ID:       rt.ID.String(),
			Name:     rt.Name,
			Capacity: rt.Capacity,
			Price:    rt.BasePrice,
		})
	}
	return dto.HotelResponse{
		ID:          agg.Hotel.ID.String(),
		Name:        agg.Hotel.Name,
		Description: agg.Hotel.Description,
		Address:     agg.Hotel.Address,
		CreatedAt:   agg.Hotel.CreatedAt,
		RoomTypes:   summaries,
	}
}

// ToHotelList maps aggregates to DTO list.
func ToHotelList(aggs []HotelAggregate) []dto.HotelResponse {
	resp := make([]dto.HotelResponse, 0, len(aggs))
	for _, agg := range aggs {
		resp = append(resp, ToHotelResponse(agg))
	}
	return resp
}

// RoomTypesToDTO maps domain room types to DTOs.
func RoomTypesToDTO(rts []domain.RoomType) []dto.RoomTypeResponse {
	out := make([]dto.RoomTypeResponse, 0, len(rts))
	for _, rt := range rts {
		out = append(out, dto.RoomTypeResponse{
			ID:        rt.ID.String(),
			HotelID:   rt.HotelID.String(),
			Name:      rt.Name,
			Capacity:  rt.Capacity,
			BasePrice: rt.BasePrice,
			Amenities: rt.Amenities,
		})
	}
	return out
}

// RoomResponses maps rooms to DTOs.
func RoomResponses(rooms []domain.Room) []dto.RoomResponse {
	out := make([]dto.RoomResponse, 0, len(rooms))
	for _, r := range rooms {
		out = append(out, RoomResponse(r))
	}
	return out
}

// RoomResponse maps single room to DTO.
func RoomResponse(r domain.Room) dto.RoomResponse {
	return dto.RoomResponse{
		ID:         r.ID.String(),
		RoomTypeID: r.RoomTypeID.String(),
		Number:     r.Number,
		Status:     r.Status,
	}
}
