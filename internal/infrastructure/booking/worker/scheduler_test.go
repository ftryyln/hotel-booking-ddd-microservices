package worker_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	bookingworker "github.com/ftryyln/hotel-booking-microservices/internal/infrastructure/booking/worker"
	bookinguc "github.com/ftryyln/hotel-booking-microservices/internal/usecase/booking"
	domain "github.com/ftryyln/hotel-booking-microservices/internal/domain/booking"
	hdomain "github.com/ftryyln/hotel-booking-microservices/internal/domain/hotel"
	"github.com/ftryyln/hotel-booking-microservices/pkg/query"
)

func TestAutoCheckoutSchedulerCreation(t *testing.T) {
	logger := zap.NewNop()
	repo := &bookingRepoStub{store: map[uuid.UUID]domain.Booking{}}
	hotelRepo := &hotelRepoStub{roomType: hdomain.RoomType{ID: uuid.New(), BasePrice: 500000}}
	payment := &paymentGatewayStub{}
	notifier := &notificationGatewayStub{}
	service := bookinguc.NewService(repo, hotelRepo, payment, notifier)

	scheduler := bookingworker.NewAutoCheckoutScheduler(service, logger)
	require.NotNil(t, scheduler)
}

func TestAutoCheckoutSchedulerStartStop(t *testing.T) {
	logger := zap.NewNop()
	repo := &bookingRepoStub{store: map[uuid.UUID]domain.Booking{}}
	hotelRepo := &hotelRepoStub{roomType: hdomain.RoomType{ID: uuid.New(), BasePrice: 500000}}
	payment := &paymentGatewayStub{}
	notifier := &notificationGatewayStub{}
	service := bookinguc.NewService(repo, hotelRepo, payment, notifier)

	scheduler := bookingworker.NewAutoCheckoutScheduler(service, logger)
	
	// Start scheduler
	err := scheduler.Start()
	require.NoError(t, err)
	
	// Give it a moment to start
	time.Sleep(100 * time.Millisecond)
	
	// Stop scheduler
	scheduler.Stop()
	
	// Should be able to stop gracefully
	time.Sleep(100 * time.Millisecond)
}

// Test stubs
type bookingRepoStub struct {
	store map[uuid.UUID]domain.Booking
}

func (b *bookingRepoStub) Create(ctx context.Context, bk domain.Booking) error {
	b.store[bk.ID] = bk
	return nil
}

func (b *bookingRepoStub) FindByID(ctx context.Context, id uuid.UUID) (domain.Booking, error) {
	bk, ok := b.store[id]
	if !ok {
		return domain.Booking{}, nil
	}
	return bk, nil
}

func (b *bookingRepoStub) FindByUserID(ctx context.Context, userID uuid.UUID) ([]domain.Booking, error) {
	var out []domain.Booking
	for _, v := range b.store {
		if v.UserID == userID {
			out = append(out, v)
		}
	}
	return out, nil
}

func (b *bookingRepoStub) UpdateStatus(ctx context.Context, id uuid.UUID, status string) error {
	if bk, ok := b.store[id]; ok {
		bk.Status = status
		b.store[id] = bk
		return nil
	}
	return nil
}

func (b *bookingRepoStub) List(ctx context.Context, _ query.Options) ([]domain.Booking, error) {
	var out []domain.Booking
	for _, v := range b.store {
		out = append(out, v)
	}
	return out, nil
}

func (b *bookingRepoStub) Save(ctx context.Context, bk domain.Booking) error {
	b.store[bk.ID] = bk
	return nil
}

type hotelRepoStub struct {
	roomType hdomain.RoomType
	err      error
}

func (h *hotelRepoStub) CreateHotel(context.Context, hdomain.Hotel) error { return nil }
func (h *hotelRepoStub) ListHotels(context.Context, query.Options) ([]hdomain.Hotel, error) {
	return nil, nil
}
func (h *hotelRepoStub) GetHotel(context.Context, uuid.UUID) (hdomain.Hotel, error) {
	return hdomain.Hotel{}, nil
}
func (h *hotelRepoStub) UpdateHotel(context.Context, uuid.UUID, hdomain.Hotel) error { return nil }
func (h *hotelRepoStub) DeleteHotel(context.Context, uuid.UUID) error                { return nil }
func (h *hotelRepoStub) CreateRoomType(context.Context, hdomain.RoomType) error      { return nil }
func (h *hotelRepoStub) ListRoomTypes(context.Context, uuid.UUID) ([]hdomain.RoomType, error) {
	return nil, nil
}
func (h *hotelRepoStub) ListAllRoomTypes(context.Context, query.Options) ([]hdomain.RoomType, error) {
	return []hdomain.RoomType{h.roomType}, h.err
}
func (h *hotelRepoStub) CreateRoom(context.Context, hdomain.Room) error { return nil }
func (h *hotelRepoStub) GetRoomType(ctx context.Context, id uuid.UUID) (hdomain.RoomType, error) {
	if h.err != nil {
		return hdomain.RoomType{}, h.err
	}
	return h.roomType, nil
}
func (h *hotelRepoStub) ListRooms(context.Context, query.Options) ([]hdomain.Room, error) {
	return nil, nil
}
func (h *hotelRepoStub) GetRoom(context.Context, uuid.UUID) (hdomain.Room, error) {
	return hdomain.Room{}, nil
}
func (h *hotelRepoStub) UpdateRoom(context.Context, uuid.UUID, hdomain.Room) error { return nil }
func (h *hotelRepoStub) DeleteRoom(context.Context, uuid.UUID) error               { return nil }

type paymentGatewayStub struct{}

func (p *paymentGatewayStub) Initiate(context.Context, uuid.UUID, float64) (domain.PaymentResult, error) {
	return domain.PaymentResult{
		ID:         uuid.New(),
		Status:     "pending",
		Provider:   "mock",
		PaymentURL: "http://mock",
	}, nil
}

type notificationGatewayStub struct{}

func (n *notificationGatewayStub) Notify(context.Context, string, any) error {
	return nil
}
