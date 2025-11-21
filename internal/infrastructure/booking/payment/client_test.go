package payment

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestHTTPGatewayInitiateSuccess(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"data":{"id":"` + uuid.New().String() + `","type":"payment","attributes":{"id":"` + uuid.New().String() + `","status":"pending","provider":"mock","payment_url":"http://pay"}},"meta":{"message":"ok","requestId":"test"}}`))
	}))
	defer srv.Close()

	gw := NewHTTPGateway(srv.URL)
	res, err := gw.Initiate(context.Background(), uuid.New(), 1000)
	require.NoError(t, err)
	require.Equal(t, "pending", res.Status)
}

func TestHTTPGatewayInitiateError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "bad", http.StatusBadRequest)
	}))
	defer srv.Close()

	gw := NewHTTPGateway(srv.URL)
	_, err := gw.Initiate(context.Background(), uuid.New(), 1000)
	require.Error(t, err)
}

func TestHTTPGatewayInitiateBadJSON(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{"data":{bad json`))
	}))
	defer srv.Close()

	gw := NewHTTPGateway(srv.URL)
	_, err := gw.Initiate(context.Background(), uuid.New(), 1000)
	require.Error(t, err)
}
