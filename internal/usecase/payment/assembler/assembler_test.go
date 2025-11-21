package assembler

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"

	"github.com/ftryyln/hotel-booking-microservices/pkg/dto"
)

func TestFromPaymentRequest(t *testing.T) {
	req := dto.PaymentRequest{
		BookingID: uuid.New().String(),
		Amount:    1000,
		Currency:  "IDR",
	}
	cmd, err := FromPaymentRequest(req)
	require.NoError(t, err)
	require.Equal(t, req.Currency, cmd.Money.Currency)

	_, err = FromPaymentRequest(dto.PaymentRequest{BookingID: "bad", Amount: 1000, Currency: "IDR"})
	require.Error(t, err)

	_, err = FromPaymentRequest(dto.PaymentRequest{BookingID: uuid.New().String(), Amount: -1, Currency: "IDR"})
	require.Error(t, err)
}

func TestFromWebhook(t *testing.T) {
	req := dto.WebhookRequest{PaymentID: uuid.New().String(), Status: "paid", Signature: "sig"}
	cmd, err := FromWebhook(req, `{"payment_id":"x","status":"paid"}`)
	require.NoError(t, err)
	require.Equal(t, req.Signature, cmd.Signature)

	_, err = FromWebhook(dto.WebhookRequest{PaymentID: "bad"}, "{}")
	require.Error(t, err)
}

func TestFromRefundRequest(t *testing.T) {
	req := dto.RefundRequest{PaymentID: uuid.New().String(), Reason: "test"}
	cmd, err := FromRefundRequest(req)
	require.NoError(t, err)
	require.Equal(t, req.Reason, cmd.Reason)

	_, err = FromRefundRequest(dto.RefundRequest{PaymentID: "bad"})
	require.Error(t, err)
}

