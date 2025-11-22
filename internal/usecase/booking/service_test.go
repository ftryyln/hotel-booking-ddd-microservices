package booking_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"

	domain "github.com/ftryyln/hotel-booking-microservices/internal/domain/booking"
	hdomain "github.com/ftryyln/hotel-booking-microservices/internal/domain/hotel"
	"github.com/ftryyln/hotel-booking-microservices/internal/usecase/booking"
	"github.com/ftryyln/hotel-booking-microservices/internal/usecase/booking/assembler"
	"github.com/ftryyln/hotel-booking-microservices/pkg/dto"
	"github.com/ftryyln/hotel-booking-microservices/pkg/query"
	"github.com/ftryyln/hotel-booking-microservices/pkg/valueobject"
)

func TestCreateBooking(t *testing.T) {
	roomTypeID := uuid.New()
	repo := &bookingRepoStub{store: map[uuid.UUID]domain.Booking{}}
	hotelRepo := &hotelRepoStub{roomType: hdomain.RoomType{ID: roomTypeID, BasePrice: 500000}}
	payment := &paymentGatewayStub{}
	notifier := &notificationGatewayStub{}
	service := booking.NewService(repo, hotelRepo, payment, notifier)

	tests := []struct {
		name    string
		req     dto.BookingRequest
		wantErr bool
	}{
		{
			name: "happy path",
			req: dto.BookingRequest{
				UserID:     uuid.New().String(),
				RoomTypeID: roomTypeID.String(),
				CheckIn:    dto.Date{Time: time.Now().Add(24 * time.Hour)},
				CheckOut:   dto.Date{Time: time.Now().Add(48 * time.Hour)},
			},
		},
		{
			name: "invalid dates",
			req: dto.BookingRequest{
				UserID:     uuid.New().String(),
				RoomTypeID: roomTypeID.String(),
				CheckIn:    dto.Date{Time: time.Now().Add(48 * time.Hour)},
				CheckOut:   dto.Date{Time: time.Now().Add(24 * time.Hour)},
			},
			wantErr: true,
		},
		{
			name: "missing room type",
			req: dto.BookingRequest{
				UserID:     uuid.New().String(),
				RoomTypeID: uuid.New().String(),
				CheckIn:    dto.Date{Time: time.Now().Add(24 * time.Hour)},
				CheckOut:   dto.Date{Time: time.Now().Add(48 * time.Hour)},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.name == "missing room type" {
				hotelRepo.err = errors.New("not found")
			} else {
				hotelRepo.err = nil
			}

			cmd, err := assembler.FromRequest(tt.req)
			if err == nil {
				_, _, err = service.CreateBooking(context.Background(), cmd)
			}
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestApplyStatusInvalidTransition(t *testing.T) {
	repo := &bookingRepoStub{store: map[uuid.UUID]domain.Booking{}}
	hotelRepo := &hotelRepoStub{roomType: hdomain.RoomType{ID: uuid.New(), BasePrice: 500000}}
	payment := &paymentGatewayStub{}
	notifier := &notificationGatewayStub{}
	service := booking.NewService(repo, hotelRepo, payment, notifier)

	id := uuid.New()
	repo.store[id] = domain.Booking{ID: id, Status: string(valueobject.StatusCancelled)}

	err := service.ApplyStatus(context.Background(), id, string(valueobject.StatusConfirmed))
	require.Error(t, err)
}

// stubs

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
		return domain.Booking{}, errors.New("not found")
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
	return errors.New("not found")
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
func (h *hotelRepoStub) CreateRoomType(context.Context, hdomain.RoomType) error { return nil }
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
func (h *hotelRepoStub) GetHotel(context.Context, uuid.UUID) (hdomain.Hotel, error) {
	return hdomain.Hotel{}, nil
}
func (h *hotelRepoStub) UpdateHotel(context.Context, uuid.UUID, hdomain.Hotel) error { return nil }
func (h *hotelRepoStub) DeleteHotel(context.Context, uuid.UUID) error                { return nil }
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

func TestAutoCheckout(t *testing.T) {
	repo := &bookingRepoStub{store: map[uuid.UUID]domain.Booking{}}
	hotelRepo := &hotelRepoStub{roomType: hdomain.RoomType{ID: uuid.New(), BasePrice: 500000}}
	payment := &paymentGatewayStub{}
	notifier := &notificationGatewayStub{}
	service := booking.NewService(repo, hotelRepo, payment, notifier)

	today := time.Now().Truncate(24 * time.Hour)
	tomorrow := today.Add(24 * time.Hour)
	yesterday := today.Add(-24 * time.Hour)

	// Create bookings with different scenarios
	// 1. Booking that should be auto-checked-out (checkout today, status checked_in)
	booking1 := domain.Booking{
		ID:       uuid.New(),
		UserID:   uuid.New(),
		CheckIn:  yesterday,
		CheckOut: today,
		Status:   string(valueobject.StatusCheckedIn),
	}
	repo.store[booking1.ID] = booking1

	// 2. Booking with checkout today but not checked in (should be ignored)
	booking2 := domain.Booking{
		ID:       uuid.New(),
		UserID:   uuid.New(),
		CheckIn:  yesterday,
		CheckOut: today,
		Status:   string(valueobject.StatusConfirmed),
	}
	repo.store[booking2.ID] = booking2

	// 3. Booking checked in but checkout tomorrow (should be ignored)
	booking3 := domain.Booking{
		ID:       uuid.New(),
		UserID:   uuid.New(),
		CheckIn:  today,
		CheckOut: tomorrow,
		Status:   string(valueobject.StatusCheckedIn),
	}
	repo.store[booking3.ID] = booking3

	// Run auto-checkout
	count, err := service.AutoCheckout(context.Background())
	require.NoError(t, err)
	require.Equal(t, 1, count) // Only booking1 should be checked out

	// Verify booking1 is completed
	updated, err := repo.FindByID(context.Background(), booking1.ID)
	require.NoError(t, err)
	require.Equal(t, string(valueobject.StatusCompleted), updated.Status)

	// Verify booking2 is still confirmed
	unchanged2, err := repo.FindByID(context.Background(), booking2.ID)
	require.NoError(t, err)
	require.Equal(t, string(valueobject.StatusConfirmed), unchanged2.Status)

	// Verify booking3 is still checked in
	unchanged3, err := repo.FindByID(context.Background(), booking3.ID)
	require.NoError(t, err)
	require.Equal(t, string(valueobject.StatusCheckedIn), unchanged3.Status)
}

func TestAutoCheckoutNoBookings(t *testing.T) {
	repo := &bookingRepoStub{store: map[uuid.UUID]domain.Booking{}}
	hotelRepo := &hotelRepoStub{roomType: hdomain.RoomType{ID: uuid.New(), BasePrice: 500000}}
	payment := &paymentGatewayStub{}
	notifier := &notificationGatewayStub{}
	service := booking.NewService(repo, hotelRepo, payment, notifier)

	// Run auto-checkout with no bookings
	count, err := service.AutoCheckout(context.Background())
	require.NoError(t, err)
	require.Equal(t, 0, count)
}
