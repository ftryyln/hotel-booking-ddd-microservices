package gateway

import (
	"net/http"
	"strings"
	"time"
)

// transports and retry helpers
type retryTransport struct {
	base    http.RoundTripper
	retries int
}

func (t *retryTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	attempts := 1
	if strings.EqualFold(req.Method, http.MethodGet) && t.retries > 0 {
		attempts += t.retries
	}
	var lastErr error
	for i := 0; i < attempts; i++ {
		cloned := cloneRequest(req)
		resp, err := t.base.RoundTrip(cloned)
		if err == nil {
			return resp, nil
		}
		lastErr = err
		time.Sleep(backoffDelay(i))
	}
	return nil, lastErr
}

func cloneRequest(r *http.Request) *http.Request {
	return r.Clone(r.Context())
}

func backoffDelay(attempt int) time.Duration {
	if attempt <= 0 {
		return 50 * time.Millisecond
	}
	delay := time.Duration(1<<uint(attempt-1)) * 100 * time.Millisecond
	if delay > time.Second {
		return time.Second
	}
	return delay
}
