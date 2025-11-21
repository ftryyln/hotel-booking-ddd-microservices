package notification

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestHTTPClientNotifySuccess(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusAccepted)
	}))
	defer srv.Close()

	client := NewHTTPGateway(srv.URL)
	err := client.Notify(context.Background(), "booking_created", "payload")
	require.NoError(t, err)
}

func TestHTTPClientNotifyError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "fail", http.StatusBadRequest)
	}))
	defer srv.Close()

	client := NewHTTPGateway(srv.URL)
	err := client.Notify(context.Background(), "booking_created", "payload")
	require.Error(t, err)
}
