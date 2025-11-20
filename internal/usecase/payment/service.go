package payment

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/lib/pq"

	domain "github.com/ftryyln/hotel-booking-microservices/internal/domain/payment"
	"github.com/ftryyln/hotel-booking-microservices/pkg/dto"
	pkgErrors "github.com/ftryyln/hotel-booking-microservices/pkg/errors"
)

// Service orchestrates payments.
type Service struct {
	repo           domain.Repository
	provider       domain.Provider
	bookingUpdater domain.BookingStatusUpdater
}

func NewService(repo domain.Repository, provider domain.Provider, updater domain.BookingStatusUpdater) *Service {
	return &Service{repo: repo, provider: provider, bookingUpdater: updater}
}

func (s *Service) Initiate(ctx context.Context, req dto.PaymentRequest) (dto.PaymentResponse, error) {
	bookingID, err := uuid.Parse(req.BookingID)
	if err != nil {
		return dto.PaymentResponse{}, pkgErrors.New("bad_request", "invalid booking id")
	}

	if existing, err := s.repo.FindByBookingID(ctx, bookingID); err == nil {
		return toDTO(existing), pkgErrors.New("conflict", "payment already exists for booking")
	}

	payment := domain.Payment{
		ID:        uuid.New(),
		BookingID: bookingID,
		Amount:    req.Amount,
		Currency:  req.Currency,
		Status:    domain.StatusPending,
		Provider:  "xendit-mock",
	}

	initiated, err := s.provider.Initiate(ctx, payment)
	if err != nil {
		return dto.PaymentResponse{}, err
	}
	if err := s.repo.Create(ctx, initiated); err != nil {
		if isUniqueViolation(err) {
			if existing, errLookup := s.repo.FindByBookingID(ctx, bookingID); errLookup == nil {
				return toDTO(existing), pkgErrors.New("conflict", "payment already exists for booking")
			}
			return dto.PaymentResponse{}, pkgErrors.New("conflict", "payment already exists for booking")
		}
		return dto.PaymentResponse{}, err
	}

	return toDTO(initiated), nil
}

func (s *Service) HandleWebhook(ctx context.Context, req dto.WebhookRequest, payload string) error {
	paymentID, err := uuid.Parse(req.PaymentID)
	if err != nil {
		return pkgErrors.New("bad_request", "invalid payment id")
	}

	payment, err := s.repo.FindByID(ctx, paymentID)
	if err != nil {
		return pkgErrors.New("not_found", "payment not found")
	}

	canonical := fmt.Sprintf("{\"payment_id\":\"%s\",\"status\":\"%s\"}", req.PaymentID, req.Status)
	if !s.provider.VerifySignature(ctx, canonical, req.Signature) {
		return pkgErrors.New("forbidden", "invalid signature")
	}

	if err := s.repo.UpdateStatus(ctx, payment.ID, req.Status, payment.PaymentURL); err != nil {
		return err
	}

	if s.bookingUpdater != nil {
		var bookingStatus string
		switch req.Status {
		case domain.StatusPaid:
			bookingStatus = "confirmed"
		case domain.StatusFailed:
			bookingStatus = "cancelled"
		}
		if bookingStatus != "" {
			_ = s.bookingUpdater.Update(ctx, payment.BookingID, bookingStatus)
		}
	}

	return nil
}

func (s *Service) Refund(ctx context.Context, req dto.RefundRequest) (dto.RefundResponse, error) {
	paymentID, err := uuid.Parse(req.PaymentID)
	if err != nil {
		return dto.RefundResponse{}, pkgErrors.New("bad_request", "invalid payment id")
	}

	payment, err := s.repo.FindByID(ctx, paymentID)
	if err != nil {
		return dto.RefundResponse{}, err
	}

	ref, err := s.provider.Refund(ctx, payment, req.Reason)
	if err != nil {
		return dto.RefundResponse{}, err
	}

	return dto.RefundResponse{ID: payment.ID.String(), Status: "refunded", Reference: ref}, nil
}

func (s *Service) GetPayment(ctx context.Context, id uuid.UUID) (dto.PaymentResponse, error) {
	p, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return dto.PaymentResponse{}, err
	}
	return toDTO(p), nil
}

// GetByBooking returns payment for a given booking.
func (s *Service) GetByBooking(ctx context.Context, bookingID uuid.UUID) (dto.PaymentResponse, error) {
	p, err := s.repo.FindByBookingID(ctx, bookingID)
	if err != nil {
		return dto.PaymentResponse{}, err
	}
	return toDTO(p), nil
}

func toDTO(p domain.Payment) dto.PaymentResponse {
	return dto.PaymentResponse{
		ID:         p.ID.String(),
		Status:     p.Status,
		Provider:   p.Provider,
		PaymentURL: p.PaymentURL,
	}
}

func isUniqueViolation(err error) bool {
	var pqErr *pq.Error
	if errors.As(err, &pqErr) {
		return pqErr.Code == "23505"
	}
	return false
}
