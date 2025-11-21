package assembler

import (
	"fmt"

	"github.com/google/uuid"

	domain "github.com/ftryyln/hotel-booking-microservices/internal/domain/payment"
	"github.com/ftryyln/hotel-booking-microservices/pkg/dto"
	"github.com/ftryyln/hotel-booking-microservices/pkg/errors"
	"github.com/ftryyln/hotel-booking-microservices/pkg/valueobject"
)

// InitiateCommand represents inbound payment initiation intent.
type InitiateCommand struct {
	BookingID uuid.UUID
	Money     valueobject.Money
}

// WebhookCommand represents inbound webhook update.
type WebhookCommand struct {
	PaymentID  uuid.UUID
	Status     string
	Signature  string
	RawPayload string
}

// RefundCommand represents refund intent.
type RefundCommand struct {
	PaymentID uuid.UUID
	Reason    string
}

// RefundResult represents refund outcome for handler mapping.
type RefundResult struct {
	PaymentID uuid.UUID
	Status    string
	Reference string
}

// FromPaymentRequest validates and builds an initiate command.
func FromPaymentRequest(req dto.PaymentRequest) (InitiateCommand, error) {
	bookingID, err := uuid.Parse(req.BookingID)
	if err != nil {
		return InitiateCommand{}, errors.New("bad_request", "invalid booking id")
	}
	money, err := valueobject.NewMoney(req.Amount, req.Currency)
	if err != nil {
		return InitiateCommand{}, err
	}
	return InitiateCommand{BookingID: bookingID, Money: money}, nil
}

// FromWebhook builds webhook command.
func FromWebhook(req dto.WebhookRequest, raw string) (WebhookCommand, error) {
	paymentID, err := uuid.Parse(req.PaymentID)
	if err != nil {
		return WebhookCommand{}, errors.New("bad_request", "invalid payment id")
	}
	if req.Status == "" || req.Signature == "" {
		return WebhookCommand{}, errors.New("bad_request", "missing webhook fields")
	}
	return WebhookCommand{
		PaymentID:  paymentID,
		Status:     req.Status,
		Signature:  req.Signature,
		RawPayload: raw,
	}, nil
}

// FromRefundRequest builds refund command.
func FromRefundRequest(req dto.RefundRequest) (RefundCommand, error) {
	paymentID, err := uuid.Parse(req.PaymentID)
	if err != nil {
		return RefundCommand{}, errors.New("bad_request", "invalid payment id")
	}
	return RefundCommand{PaymentID: paymentID, Reason: req.Reason}, nil
}

// ToRefundResult maps provider response to result.
func ToRefundResult(paymentID uuid.UUID, ref string) RefundResult {
	return RefundResult{PaymentID: paymentID, Status: "refunded", Reference: ref}
}

// ToResponse maps domain Payment to DTO.
func ToResponse(p domain.Payment) dto.PaymentResponse {
	return dto.PaymentResponse{
		ID:         p.ID.String(),
		Status:     p.Status,
		Provider:   p.Provider,
		PaymentURL: p.PaymentURL,
	}
}

// ToRefundResponse maps refund result to DTO.
func ToRefundResponse(res RefundResult) dto.RefundResponse {
	return dto.RefundResponse{
		ID:        res.PaymentID.String(),
		Status:    res.Status,
        Reference: res.Reference,
	}
}

// CanonicalPayload constructs canonical payload for signature verify.
func CanonicalPayload(cmd WebhookCommand) string {
	return fmt.Sprintf("{\"payment_id\":\"%s\",\"status\":\"%s\"}", cmd.PaymentID.String(), cmd.Status)
}
