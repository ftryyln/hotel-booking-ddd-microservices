package payment

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"

	domain "github.com/ftryyln/hotel-booking-microservices/internal/domain/booking"
	"github.com/ftryyln/hotel-booking-microservices/pkg/dto"
	"github.com/ftryyln/hotel-booking-microservices/pkg/middleware"
)

// HTTPGateway calls payment service over HTTP.
type HTTPGateway struct {
	baseURL string
	client  *http.Client
}

func NewHTTPGateway(baseURL string) domain.PaymentGateway {
	return &HTTPGateway{baseURL: baseURL, client: &http.Client{Timeout: 5 * time.Second}}
}

func (g *HTTPGateway) Initiate(ctx context.Context, bookingID uuid.UUID, amount float64) (domain.PaymentResult, error) {
	payload := map[string]any{"booking_id": bookingID.String(), "amount": amount, "currency": "IDR"}
	body, _ := json.Marshal(payload)
	req, _ := http.NewRequestWithContext(ctx, http.MethodPost, fmt.Sprintf("%s/payments", g.baseURL), bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	if token, ok := ctx.Value(middleware.AuthTokenKey).(string); ok && token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	resp, err := g.client.Do(req)
	if err != nil {
		return domain.PaymentResult{}, err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		return domain.PaymentResult{}, fmt.Errorf("payment initiation failed: %d", resp.StatusCode)
	}
	var result dto.PaymentResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return domain.PaymentResult{}, err
	}
	paymentID, _ := uuid.Parse(result.ID)
	return domain.PaymentResult{
		ID:         paymentID,
		Status:     result.Status,
		Provider:   result.Provider,
		PaymentURL: result.PaymentURL,
	}, nil
}
