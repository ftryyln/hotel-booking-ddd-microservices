package provider

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"

	domain "github.com/ftryyln/hotel-booking-microservices/internal/domain/payment"
)

func TestXenditMockProviderInitiate(t *testing.T) {
	p := &XenditMockProvider{}
	payment := domain.Payment{ID: uuid.New(), BookingID: uuid.New(), Currency: "IDR", Amount: 1000}
	res, err := p.Initiate(context.Background(), payment)
	require.NoError(t, err)
	require.Equal(t, payment.ID, res.ID)
	require.NotEmpty(t, res.PaymentURL)
}

func TestXenditMockProviderVerifySignature(t *testing.T) {
	p := NewXenditMockProvider("secret")
	payload := `{"payment_id":"id","status":"paid"}`
	// compute signature manually
	sig := signPayload("secret", payload)
	require.True(t, p.VerifySignature(context.Background(), payload, sig))
	require.False(t, p.VerifySignature(context.Background(), payload, "bad"))
}

func TestXenditMockProviderRefund(t *testing.T) {
	p := NewXenditMockProvider("secret")
	ref, err := p.Refund(context.Background(), domain.Payment{ID: uuid.New()}, "reason")
	require.NoError(t, err)
	require.NotEmpty(t, ref)
}

func signPayload(secret, payload string) string {
	h := hmac.New(sha256.New, []byte(secret))
	h.Write([]byte(payload))
	return hex.EncodeToString(h.Sum(nil))
}
