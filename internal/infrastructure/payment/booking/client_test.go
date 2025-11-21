package booking

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestHTTPGatewayUpdateSuccess(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	gw := NewHTTPStatusClient(srv.URL)
	err := gw.Update(context.Background(), uuid.New(), "confirmed")
	require.NoError(t, err)
}

func TestHTTPGatewayUpdateError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "fail", http.StatusBadRequest)
	}))
	defer srv.Close()

	gw := NewHTTPStatusClient(srv.URL)
	err := gw.Update(context.Background(), uuid.New(), "confirmed")
	require.Error(t, err)
}
