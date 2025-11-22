package hotelhttp_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"

	domain "github.com/ftryyln/hotel-booking-microservices/internal/domain/hotel"
	hotelhttp "github.com/ftryyln/hotel-booking-microservices/internal/infrastructure/hotel/http"
	"github.com/ftryyln/hotel-booking-microservices/internal/usecase/hotel"
	"github.com/ftryyln/hotel-booking-microservices/pkg/query"
)

func TestHotelHandlerListHotelsPagination(t *testing.T) {
	repo := &hotelRepoStub{}
	svc := hotel.NewService(repo)
	h := hotelhttp.NewHandler(svc, "secret")
	r := chi.NewRouter()
	r.Mount("/", h.Routes())

	req := httptest.NewRequest(http.MethodGet, "/hotels?limit=1&offset=2", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	require.Equal(t, http.StatusOK, rec.Code)
}

// minimal repo stub for handler test
type hotelRepoStub struct{}

func (h *hotelRepoStub) CreateHotel(ctx context.Context, hotel domain.Hotel) error { return nil }
func (h *hotelRepoStub) ListHotels(ctx context.Context, opts query.Options) ([]domain.Hotel, error) {
	return []domain.Hotel{{ID: uuid.New(), Name: "H", Address: "Addr"}}, nil
}
func (h *hotelRepoStub) CreateRoomType(context.Context, domain.RoomType) error { return nil }
func (h *hotelRepoStub) ListRoomTypes(context.Context, uuid.UUID) ([]domain.RoomType, error) {
	return []domain.RoomType{}, nil
}
func (h *hotelRepoStub) ListAllRoomTypes(context.Context, query.Options) ([]domain.RoomType, error) {
	return []domain.RoomType{}, nil
}
func (h *hotelRepoStub) CreateRoom(context.Context, domain.Room) error { return nil }
func (h *hotelRepoStub) GetRoomType(context.Context, uuid.UUID) (domain.RoomType, error) {
	return domain.RoomType{}, nil
}
func (h *hotelRepoStub) ListRooms(context.Context, query.Options) ([]domain.Room, error) {
	return []domain.Room{}, nil
}
func (h *hotelRepoStub) GetHotel(context.Context, uuid.UUID) (domain.Hotel, error) {
	return domain.Hotel{ID: uuid.New(), Name: "H", Address: "A"}, nil
}
func (h *hotelRepoStub) UpdateHotel(ctx context.Context, id uuid.UUID, hotel domain.Hotel) error {
	return nil
}
func (h *hotelRepoStub) DeleteHotel(context.Context, uuid.UUID) error { return nil }
func (h *hotelRepoStub) GetRoom(ctx context.Context, id uuid.UUID) (domain.Room, error) {
	return domain.Room{ID: id, Number: "101", Status: "available"}, nil
}
func (h *hotelRepoStub) UpdateRoom(context.Context, uuid.UUID, domain.Room) error { return nil }
func (h *hotelRepoStub) DeleteRoom(context.Context, uuid.UUID) error              { return nil }

func TestHotelHandlerUpdateHotel(t *testing.T) {
	repo := &hotelRepoStub{}
	svc := hotel.NewService(repo)
	h := hotelhttp.NewHandler(svc, "secret")
	r := chi.NewRouter()
	r.Mount("/", h.Routes())

	hotelID := uuid.New()
	body := `{"name":"Updated Hotel","description":"Updated desc","address":"Updated addr"}`
	req := httptest.NewRequest(http.MethodPut, "/hotels/"+hotelID.String(),
		strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	// Update requires admin auth, so without JWT we expect 401
	require.Equal(t, http.StatusUnauthorized, rec.Code)
}

func TestHotelHandlerDeleteHotel(t *testing.T) {
	repo := &hotelRepoStub{}
	svc := hotel.NewService(repo)
	h := hotelhttp.NewHandler(svc, "secret")
	r := chi.NewRouter()
	r.Mount("/", h.Routes())

	hotelID := uuid.New()
	req := httptest.NewRequest(http.MethodDelete, "/hotels/"+hotelID.String(), nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	// Delete requires admin auth, so without JWT we expect 401
	require.Equal(t, http.StatusUnauthorized, rec.Code)
}

func TestHotelHandlerGetRoom(t *testing.T) {
	repo := &hotelRepoStub{}
	svc := hotel.NewService(repo)
	h := hotelhttp.NewHandler(svc, "secret")
	r := chi.NewRouter()
	r.Mount("/", h.Routes())

	roomID := uuid.New()
	req := httptest.NewRequest(http.MethodGet, "/rooms/"+roomID.String(), nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
}

func TestHotelHandlerUpdateRoom(t *testing.T) {
	repo := &hotelRepoStub{}
	svc := hotel.NewService(repo)
	h := hotelhttp.NewHandler(svc, "secret")
	r := chi.NewRouter()
	r.Mount("/", h.Routes())

	roomID := uuid.New()
	body := `{"number":"102","status":"maintenance"}`
	req := httptest.NewRequest(http.MethodPut, "/rooms/"+roomID.String(),
		strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	// Update requires admin auth, so without JWT we expect 401
	require.Equal(t, http.StatusUnauthorized, rec.Code)
}

func TestHotelHandlerDeleteRoom(t *testing.T) {
	repo := &hotelRepoStub{}
	svc := hotel.NewService(repo)
	h := hotelhttp.NewHandler(svc, "secret")
	r := chi.NewRouter()
	r.Mount("/", h.Routes())

	roomID := uuid.New()
	req := httptest.NewRequest(http.MethodDelete, "/rooms/"+roomID.String(), nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	// Delete requires admin auth, so without JWT we expect 401
	require.Equal(t, http.StatusUnauthorized, rec.Code)
}

func TestHotelHandlerUpdateHotelInvalidID(t *testing.T) {
	repo := &hotelRepoStub{}
	svc := hotel.NewService(repo)
	h := hotelhttp.NewHandler(svc, "secret")
	r := chi.NewRouter()
	r.Mount("/", h.Routes())

	body := `{"name":"Test","address":"Test"}`
	req := httptest.NewRequest(http.MethodPut, "/hotels/invalid-uuid",
		strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	// Update requires admin auth, so without JWT we expect 401 (not 400 for invalid ID)
	require.Equal(t, http.StatusUnauthorized, rec.Code)
}
