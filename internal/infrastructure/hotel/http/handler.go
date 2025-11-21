package hotelhttp

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	hoteluc "github.com/ftryyln/hotel-booking-microservices/internal/usecase/hotel"
	"github.com/ftryyln/hotel-booking-microservices/internal/usecase/hotel/assembler"
	"github.com/ftryyln/hotel-booking-microservices/pkg/dto"
	pkgErrors "github.com/ftryyln/hotel-booking-microservices/pkg/errors"
	"github.com/ftryyln/hotel-booking-microservices/pkg/middleware"
	"github.com/ftryyln/hotel-booking-microservices/pkg/query"
	"github.com/ftryyln/hotel-booking-microservices/pkg/utils"
)

// Handler exposes hotel endpoints.
type Handler struct {
	service   *hoteluc.Service
	jwtSecret string
}

func NewHandler(service *hoteluc.Service, jwtSecret string) *Handler {
	return &Handler{service: service, jwtSecret: jwtSecret}
}

func (h *Handler) Routes() http.Handler {
	r := chi.NewRouter()
	r.Get("/hotels", h.listHotels)
	r.Get("/hotels/{id}", h.getHotel)
	r.Get("/room-types", h.listRoomTypes)
	r.Get("/rooms", h.listRooms)
	r.Group(func(r chi.Router) {
		r.Use(middleware.JWT(h.jwtSecret, "admin"))
		r.Post("/hotels", h.createHotel)
		r.Post("/room-types", h.createRoomType)
		r.Post("/rooms", h.createRoom)
	})
	return r
}

// @Summary Create hotel
// @Tags Hotels
// @Accept json
// @Produce json
// @Param request body dto.HotelRequest true "Hotel payload"
// @Success 201 {object} dto.CreatedHotelResponse
// @Failure 400 {object} dto.ErrorResponse
// @Security BearerAuth
// @Router /hotels [post]
func (h *Handler) createHotel(w http.ResponseWriter, r *http.Request) {
	var req dto.HotelRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, pkgErrors.New("bad_request", "invalid payload"))
		return
	}
	id, err := h.service.CreateHotel(r.Context(), req)
	if err != nil {
		writeError(w, pkgErrors.FromError(err))
		return
	}
	resource := utils.NewResource(id.String(), "hotel", "/api/v1/hotels/"+id.String(), dto.CreatedHotelResponse{
		ID:          id.String(),
		Name:        req.Name,
		Description: req.Description,
		Address:     req.Address,
		Message:     "hotel created",
	})
	utils.Respond(w, http.StatusCreated, "hotel created", resource)
}

// @Summary List hotels
// @Tags Hotels
// @Produce json
// @Param limit query int false "pagination limit (default 50)"
// @Param offset query int false "pagination offset"
// @Success 200 {array} dto.HotelResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /hotels [get]
func (h *Handler) listHotels(w http.ResponseWriter, r *http.Request) {
	opts := parseQueryOptions(r)
	resp, err := h.service.ListHotels(r.Context(), opts)
	if err != nil {
		writeError(w, pkgErrors.FromError(err))
		return
	}
	dtoList := assembler.ToHotelList(resp)
	var resources []utils.Resource
	for _, h := range dtoList {
		resources = append(resources, utils.NewResource(h.ID, "hotel", "/api/v1/hotels/"+h.ID, h))
	}
	utils.RespondWithCount(w, http.StatusOK, "hotels listed", resources, len(resources))
}

// @Summary Get hotel detail
// @Tags Hotels
// @Produce json
// @Param id path string true "Hotel ID"
// @Success 200 {object} dto.HotelResponse
// @Failure 404 {object} dto.ErrorResponse
// @Router /hotels/{id} [get]
func (h *Handler) getHotel(w http.ResponseWriter, r *http.Request) {
	idParam := chi.URLParam(r, "id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		writeError(w, pkgErrors.New("bad_request", "invalid id"))
		return
	}
	resp, err := h.service.GetHotel(r.Context(), id, query.Options{})
	if err != nil {
		writeError(w, pkgErrors.FromError(err))
		return
	}
	dtoResp := assembler.ToHotelResponse(resp)
	resource := utils.NewResource(dtoResp.ID, "hotel", "/api/v1/hotels/"+dtoResp.ID, dtoResp)
	utils.Respond(w, http.StatusOK, "hotel retrieved", resource)
}

// @Summary Create room type
// @Tags Hotels
// @Accept json
// @Produce json
// @Param request body dto.RoomTypeRequest true "Room type payload"
// @Success 201 {object} dto.CreatedRoomTypeResponse
// @Failure 400 {object} dto.ErrorResponse
// @Security BearerAuth
// @Router /room-types [post]
func (h *Handler) createRoomType(w http.ResponseWriter, r *http.Request) {
	var req dto.RoomTypeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, pkgErrors.New("bad_request", "invalid payload"))
		return
	}
	id, err := h.service.CreateRoomType(r.Context(), req)
	if err != nil {
		writeError(w, pkgErrors.FromError(err))
		return
	}
	resource := utils.NewResource(id.String(), "room_type", "/api/v1/room-types/"+id.String(), dto.CreatedRoomTypeResponse{
		ID:        id.String(),
		HotelID:   req.HotelID,
		Name:      req.Name,
		Capacity:  req.Capacity,
		BasePrice: req.BasePrice,
		Amenities: req.Amenities,
		Message:   "room type created",
	})
	utils.Respond(w, http.StatusCreated, "room type created", resource)
}

// @Summary List room types
// @Tags Hotels
// @Produce json
// @Param limit query int false "pagination limit (default 50)"
// @Param offset query int false "pagination offset"
// @Success 200 {array} dto.RoomTypeResponse
// @Router /room-types [get]
func (h *Handler) listRoomTypes(w http.ResponseWriter, r *http.Request) {
	opts := parseQueryOptions(r)
	resp, err := h.service.ListRoomTypes(r.Context(), opts)
	if err != nil {
		writeError(w, pkgErrors.FromError(err))
		return
	}
	dtoList := assembler.RoomTypesToDTO(resp)
	var resources []utils.Resource
	for _, rt := range dtoList {
		resources = append(resources, utils.NewResource(rt.ID, "room_type", "/api/v1/room-types/"+rt.ID, rt))
	}
	utils.RespondWithCount(w, http.StatusOK, "room types listed", resources, len(resources))
}

// @Summary Create room
// @Tags Hotels
// @Accept json
// @Produce json
// @Param request body dto.RoomRequest true "Room payload"
// @Success 201 {object} dto.CreatedRoomResponse
// @Failure 400 {object} dto.ErrorResponse
// @Security BearerAuth
// @Router /rooms [post]
func (h *Handler) createRoom(w http.ResponseWriter, r *http.Request) {
	var req dto.RoomRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, pkgErrors.New("bad_request", "invalid payload"))
		return
	}
	id, err := h.service.CreateRoom(r.Context(), req)
	if err != nil {
		writeError(w, pkgErrors.FromError(err))
		return
	}
	resource := utils.NewResource(id.String(), "room", "/api/v1/rooms/"+id.String(), dto.CreatedRoomResponse{
		ID:         id.String(),
		RoomTypeID: req.RoomTypeID,
		Number:     req.Number,
		Status:     req.Status,
		Message:    "room created",
	})
	utils.Respond(w, http.StatusCreated, "room created", resource)
}

// @Summary List rooms
// @Tags Hotels
// @Produce json
// @Param limit query int false "pagination limit (default 50)"
// @Param offset query int false "pagination offset"
// @Success 200 {array} dto.RoomResponse
// @Router /rooms [get]
func (h *Handler) listRooms(w http.ResponseWriter, r *http.Request) {
	opts := parseQueryOptions(r)
	resp, err := h.service.ListRooms(r.Context(), opts)
	if err != nil {
		writeError(w, pkgErrors.FromError(err))
		return
	}
	dtoRooms := assembler.RoomResponses(resp)
	var resources []utils.Resource
	for _, room := range dtoRooms {
		resources = append(resources, utils.NewResource(room.ID, "room", "/api/v1/rooms/"+room.ID, room))
	}
	utils.RespondWithCount(w, http.StatusOK, "rooms listed", resources, len(resources))
}

func writeJSON(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(body)
}

func writeError(w http.ResponseWriter, err pkgErrors.APIError) {
	utils.Respond(w, pkgErrors.StatusCode(err), err.Message, err)
}

func parseQueryOptions(r *http.Request) query.Options {
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
	return query.Options{Limit: limit, Offset: offset}
}
