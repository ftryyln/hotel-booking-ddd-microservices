package payment

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/lib/pq"

	domain "github.com/ftryyln/hotel-booking-microservices/internal/domain/payment"
	"github.com/ftryyln/hotel-booking-microservices/internal/usecase/payment/assembler"
	pkgErrors "github.com/ftryyln/hotel-booking-microservices/pkg/errors"
	"github.com/ftryyln/hotel-booking-microservices/pkg/valueobject"
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

// Initiate creates a new payment from a validated command.
func (s *Service) Initiate(ctx context.Context, cmd assembler.InitiateCommand) (domain.Payment, error) {
	if existing, err := s.repo.FindByBookingID(ctx, cmd.BookingID); err == nil {
		return existing, pkgErrors.New("conflict", "payment already exists for booking")
	}

	payment := domain.Payment{
		ID:        uuid.New(),
		BookingID: cmd.BookingID,
		Amount:    cmd.Money.Amount,
		Currency:  cmd.Money.Currency,
		Status:    string(valueobject.PaymentPending),
		Provider:  "xendit-mock",
	}

	initiated, err := s.provider.Initiate(ctx, payment)
	if err != nil {
		return domain.Payment{}, err
	}
	if err := s.repo.Create(ctx, initiated); err != nil {
		if isUniqueViolation(err) {
			if existing, errLookup := s.repo.FindByBookingID(ctx, cmd.BookingID); errLookup == nil {
				return existing, pkgErrors.New("conflict", "payment already exists for booking")
			}
			return domain.Payment{}, pkgErrors.New("conflict", "payment already exists for booking")
		}
		return domain.Payment{}, err
	}

	return initiated, nil
}

// HandleWebhook applies status update from provider webhook.
func (s *Service) HandleWebhook(ctx context.Context, cmd assembler.WebhookCommand) error {
	payment, err := s.repo.FindByID(ctx, cmd.PaymentID)
	if err != nil {
		return pkgErrors.New("not_found", "payment not found")
	}

	targetStatus, err := mapProviderStatus(cmd.Status)
	if err != nil {
		return err
	}

	currentStatus, err := valueobject.ValidatePaymentStatus(payment.Status)
	if err != nil {
		return err
	}
	if err := currentStatus.CanTransition(targetStatus); err != nil {
		return err
	}

	canonical := assembler.CanonicalPayload(cmd)
	if !s.provider.VerifySignature(ctx, canonical, cmd.Signature) {
		return pkgErrors.New("forbidden", "invalid signature")
	}

	if err := s.repo.UpdateStatus(ctx, payment.ID, string(targetStatus), payment.PaymentURL, cmd.RawPayload, cmd.Signature); err != nil {
		return err
	}

	if s.bookingUpdater != nil {
		var bookingStatus string
		switch cmd.Status {
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

func mapProviderStatus(status string) (valueobject.PaymentStatus, error) {
	switch status {
	case "PENDING", "pending":
		return valueobject.PaymentPending, nil
	case "PAID", "paid":
		return valueobject.PaymentPaid, nil
	case "EXPIRED", "FAILED", "expired", "failed", "REFUNDED", "refunded":
		return valueobject.PaymentFailed, nil
	default:
		return valueobject.ValidatePaymentStatus(status)
	}
}

// Refund requests refund via provider.
func (s *Service) Refund(ctx context.Context, cmd assembler.RefundCommand) (assembler.RefundResult, error) {
	payment, err := s.repo.FindByID(ctx, cmd.PaymentID)
	if err != nil {
		return assembler.RefundResult{}, err
	}

	ref, err := s.provider.Refund(ctx, payment, cmd.Reason)
	if err != nil {
		return assembler.RefundResult{}, err
	}

	return assembler.ToRefundResult(payment.ID, ref), nil
}

// GetPayment fetches payment by ID.
func (s *Service) GetPayment(ctx context.Context, id uuid.UUID) (domain.Payment, error) {
	pay, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return domain.Payment{}, pkgErrors.New("not_found", "payment not found")
	}
	return pay, nil
}

// GetByBooking returns payment for a given booking.
func (s *Service) GetByBooking(ctx context.Context, bookingID uuid.UUID) (domain.Payment, error) {
	return s.repo.FindByBookingID(ctx, bookingID)
}

func isUniqueViolation(err error) bool {
	var pqErr *pq.Error
	if errors.As(err, &pqErr) {
		return pqErr.Code == "23505"
	}
	return false
}
