package gateway

import "time"

// DebugRouteResponse documents a route entry for swagger.
type DebugRouteResponse struct {
	Name        string            `json:"name"`
	Prefix      string            `json:"prefix"`
	Upstream    string            `json:"upstream"`
	RequireAuth bool              `json:"require_auth"`
	Auth        string            `json:"auth_strategy"`
	Status      UpstreamStatusDoc `json:"status"`
}

// HealthResponse documents upstream health for swagger.
type HealthResponse struct {
	Upstream string            `json:"upstream"`
	Status   UpstreamStatusDoc `json:"status"`
}

// UpstreamStatusDoc mirrors upstreamStatus but is exported for swagger docs.
type UpstreamStatusDoc struct {
	Healthy           bool   `json:"healthy"`
	LastChecked       string `json:"last_checked"`
	LastError         string `json:"last_error,omitempty"`
	CircuitOpenUntil  string `json:"circuit_open_until,omitempty"`
	RequestsInWindow  int    `json:"requests_in_window"`
	FailuresInWindow  int    `json:"failures_in_window"`
	WindowStartedAt   string `json:"window_started_at"`
	UnhealthySince    string `json:"unhealthy_since,omitempty"`
	ConsecutiveErrors int    `json:"consecutive_errors"`
}

func statusDocFrom(s upstreamStatus) UpstreamStatusDoc {
	return UpstreamStatusDoc{
		Healthy:           s.Healthy,
		LastChecked:       s.LastChecked.Format(time.RFC3339),
		LastError:         s.LastError,
		CircuitOpenUntil:  formatTime(s.CircuitOpenUntil),
		RequestsInWindow:  s.RequestsInWindow,
		FailuresInWindow:  s.FailuresInWindow,
		WindowStartedAt:   s.WindowStartedAt.Format(time.RFC3339),
		UnhealthySince:    formatTime(s.UnhealthySince),
		ConsecutiveErrors: s.ConsecutiveErrors,
	}
}

func formatTime(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return t.Format(time.RFC3339)
}
