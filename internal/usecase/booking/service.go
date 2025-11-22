package booking

import (
	"context"
	"database/sql"
	"time" // Added time import

	"github.com/google/uuid"

	domain "github.com/ftryyln/hotel-booking-microservices/internal/domain/booking"
	hdomain "github.com/ftryyln/hotel-booking-microservices/internal/domain/hotel"
	"github.com/ftryyln/hotel-booking-microservices/internal/usecase/booking/assembler"
	pkgDomain "github.com/ftryyln/hotel-booking-microservices/pkg/domain"
	"github.com/ftryyln/hotel-booking-microservices/pkg/errors"


	"github.com/ftryyln/hotel-booking-microservices/pkg/query"
	"github.com/ftryyln/hotel-booking-microservices/pkg/valueobject"
)

// Service handles booking lifecycle.
type Service struct {
	repo     domain.Repository
	hotels   hdomain.Repository
	payments domain.PaymentGateway
	notifier domain.NotificationGateway
}

func NewService(repo domain.Repository, hotels hdomain.Repository, payments domain.PaymentGateway, notifier domain.NotificationGateway) *Service {
	return &Service{repo: repo, hotels: hotels, payments: payments, notifier: notifier}
}

func (s *Service) CreateBooking(ctx context.Context, cmd assembler.CreateCommand) (domain.Booking, domain.PaymentResult, error) {
	// Use value objects
	dateRange, err := valueobject.NewDateRange(cmd.CheckIn, cmd.CheckOut)
	if err != nil {
		return domain.Booking{}, domain.PaymentResult{}, err
	}

	rt, err := s.hotels.GetRoomType(ctx, cmd.RoomTypeID)
	if err != nil {
		return domain.Booking{}, domain.PaymentResult{}, errors.New("not_found", "room type not found")
	}

	// Use domain service for pricing
	pricingService := domain.NewPricingService()
	baseTotal := pricingService.CalculateTotalPrice(rt.BasePrice, dateRange.Nights(), cmd.Guests)
	totalPrice := pricingService.ApplyDiscount(baseTotal, dateRange.Nights())

	booking := domain.Booking{
		ID:          uuid.New(),
		UserID:      cmd.UserID,
		RoomTypeID:  cmd.RoomTypeID,
		CheckIn:     cmd.CheckIn,
		CheckOut:    cmd.CheckOut,
		Status:      string(valueobject.StatusPendingPayment),
		Guests:      cmd.Guests,
		TotalPrice:  totalPrice,
		TotalNights: dateRange.Nights(),
		CreatedAt:   time.Now(),
	}

	// Record creation event
	booking.RecordEvent(domain.NewBookingCreated(booking.ID, booking.UserID, booking.RoomTypeID, booking.TotalPrice, booking.Guests))

	if err := s.repo.Create(ctx, booking); err != nil {
		return domain.Booking{}, domain.PaymentResult{}, err
	}

	// Publish domain events
	s.publishEvents(ctx, booking.Events())
	booking.ClearEvents()

	paymentResult, err := s.payments.Initiate(ctx, booking.ID, booking.TotalPrice)
	if err != nil {
		return domain.Booking{}, domain.PaymentResult{}, err
	}

	return booking, paymentResult, nil
}

func (s *Service) CancelBooking(ctx context.Context, id uuid.UUID) error {
	booking, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return err
	}

	// Use domain method
	if err := booking.Cancel("user_requested"); err != nil {
		return err
	}

	if err := s.repo.Save(ctx, booking); err != nil {
		return err
	}

	s.publishEvents(ctx, booking.Events())
	return nil
}

func (s *Service) ApplyStatus(ctx context.Context, id uuid.UUID, status string) error {
	// Deprecated: Use specific domain methods instead (Confirm, CheckIn, Complete)
	// Keeping for backward compatibility if needed, but redirecting to domain methods where possible
	booking, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return err
	}

	var updateErr error
	switch status {
	case domain.StatusConfirmed:
		updateErr = booking.Confirm()
	case domain.StatusCancelled:
		updateErr = booking.Cancel("admin_requested")
	case domain.StatusCheckedIn:
		updateErr = booking.GuestCheckIn()

	case domain.StatusCompleted:
		updateErr = booking.Complete()
	default:
		// Fallback for direct status update (legacy)
		return s.repo.UpdateStatus(ctx, id, status)
	}

	if updateErr != nil {
		return updateErr
	}

	if err := s.repo.Save(ctx, booking); err != nil {
		return err
	}

	s.publishEvents(ctx, booking.Events())
	return nil
}

func (s *Service) Checkpoint(ctx context.Context, id uuid.UUID, action string) error {
	bk, err := s.repo.FindByID(ctx, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return errors.New("not_found", "booking not found")
		}
		return err
	}

	var updateErr error
	switch action {
	case "check_in":
		updateErr = bk.GuestCheckIn()

	case "complete":
		updateErr = bk.Complete()
	default:
		return errors.New("bad_request", "unknown checkpoint action")
	}

	if updateErr != nil {
		return updateErr
	}

	if err := s.repo.Save(ctx, bk); err != nil {
		return err
	}

	s.publishEvents(ctx, bk.Events())
	return nil
}

func (s *Service) GetBooking(ctx context.Context, id uuid.UUID) (domain.Booking, error) {
	b, err := s.repo.FindByID(ctx, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return domain.Booking{}, errors.New("not_found", "booking not found")
		}
		return domain.Booking{}, err
	}
	return b, nil
}

func (s *Service) ListBookings(ctx context.Context, opts query.Options) ([]domain.Booking, error) {
	bks, err := s.repo.List(ctx, opts.Normalize(50))
	if err != nil {
		return nil, err
	}
	return bks, nil
}

func (s *Service) publishEvents(ctx context.Context, events []pkgDomain.DomainEvent) {
	for _, event := range events {
		_ = s.notifier.Notify(ctx, event.EventType(), event)
	}
}

// AutoCheckout processes bookings that should be automatically checked out.
// It finds all bookings with checkout_date = today and status = checked_in,
// then completes them automatically.
func (s *Service) AutoCheckout(ctx context.Context) (int, error) {
	today := time.Now().Truncate(24 * time.Hour)
	
	// Find bookings that need auto-checkout
	// We'll use the repository's List method with a filter
	// For simplicity, we'll get all bookings and filter in memory
	// In production, you'd want to add a specific repository method
	allBookings, err := s.repo.List(ctx, query.Options{Limit: 1000})
	if err != nil {
		return 0, err
	}

	count := 0
	for _, booking := range allBookings {
		// Check if booking should be auto-checked-out
		checkoutDate := booking.CheckOut.Truncate(24 * time.Hour)
		if checkoutDate.Equal(today) && booking.Status == string(valueobject.StatusCheckedIn) {
			// Complete the booking
			if err := booking.Complete(); err != nil {
				// Log error but continue with other bookings
				continue
			}

			if err := s.repo.Save(ctx, booking); err != nil {
				// Log error but continue
				continue
			}

			s.publishEvents(ctx, booking.Events())
			booking.ClearEvents()
			count++
		}
	}

	return count, nil
}
