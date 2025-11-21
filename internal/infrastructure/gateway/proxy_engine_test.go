package gateway

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strings"
	"testing"
	"time"

	"go.uber.org/zap/zaptest"

	"github.com/ftryyln/hotel-booking-microservices/pkg/config"
)

func TestServeHTTP_RouteMatchAndAuthForward(t *testing.T) {
	// stub upstream that always OK
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer upstream.Close()

	cfg := config.Config{
		GatewayMode:     ModeProxyAll,
		RoutesFile:      "", // we will inject routes manually
		UpstreamRetries: 1,
		UpstreamTimeout: 2 * time.Second,
		HealthInterval:  time.Second,
	}
	engine := &proxyEngine{
		mode:             ModeProxyAll,
		log:              zaptest.NewLogger(t),
		timeout:          cfg.UpstreamTimeout,
		retries:          cfg.UpstreamRetries,
		metrics:          newGatewayMetrics(),
		transport:        http.DefaultTransport,
		healthInterval:   cfg.HealthInterval,
		healthClient:     &http.Client{Timeout: time.Second},
		readyCh:          make(chan struct{}),
		upstreams:        map[string]*upstreamTarget{},
		circuitWindow:    5 * time.Second,
		circuitThreshold: 0.5,
		circuitCooldown:  2 * time.Second,
	}

	upURL := upstream.URL
	rt := &route{
		name:        "test",
		prefix:      "/api",
		target:      &upstreamTarget{url: mustParseURL(upURL), health: "/healthz", status: upstreamStatus{Healthy: true, WindowStartedAt: time.Now()}},
		proxy:       engine.newReverseProxy(mustParseURL(upURL)),
		requireAuth: false,
	}
	engine.routes = []*route{rt}
	engine.upstreams[upURL] = rt.target

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/hello", nil)
	engine.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}

func TestServeHTTP_AuthValidateMissingToken(t *testing.T) {
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer upstream.Close()

	engine := newTestEngine()
	engine.jwtSecret = "secret"
	upURL := upstream.URL
	target := &upstreamTarget{url: mustParseURL(upURL), health: "/healthz", status: upstreamStatus{Healthy: true, WindowStartedAt: time.Now()}}
	engine.upstreams[upURL] = target
	engine.routes = []*route{{
		name:         "auth",
		prefix:       "/secure",
		target:       target,
		requireAuth:  true,
		authStrategy: "validate",
		proxy:        engine.newReverseProxy(target.url),
	}}

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/secure/foo", nil)
	engine.ServeHTTP(rec, req)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rec.Code)
	}
}

func TestDebugEndpoints(t *testing.T) {
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer upstream.Close()

	engine := &proxyEngine{
		mode:         ModeProxyAll,
		log:          zaptest.NewLogger(t),
		timeout:      2 * time.Second,
		metrics:      newGatewayMetrics(),
		healthClient: &http.Client{Timeout: time.Second},
		upstreams:    map[string]*upstreamTarget{},
	}
	u := &upstreamTarget{url: mustParseURL(upstream.URL), health: "/healthz", status: upstreamStatus{Healthy: true, WindowStartedAt: time.Now()}}
	engine.upstreams[u.url.String()] = u
	engine.routes = []*route{{
		name:        "debug",
		prefix:      "/debug",
		target:      u,
		requireAuth: false,
		proxy:       engine.newReverseProxy(u.url),
	}}

	// metrics endpoint
	recM := httptest.NewRecorder()
	engine.Metrics(recM, httptest.NewRequest(http.MethodGet, "/metrics", nil))
	if recM.Code != http.StatusOK {
		t.Fatalf("metrics status %d", recM.Code)
	}

	// debug routes
	recD := httptest.NewRecorder()
	engine.DebugRoutes(recD, httptest.NewRequest(http.MethodGet, "/debug/routes", nil))
	if recD.Code != http.StatusOK {
		t.Fatalf("debug routes status %d", recD.Code)
	}
	if !strings.Contains(recD.Body.String(), "\"prefix\":\"/debug\"") {
		t.Fatalf("debug routes payload missing route info")
	}

	// healthz
	recH := httptest.NewRecorder()
	engine.Healthz(recH, httptest.NewRequest(http.MethodGet, "/healthz", nil))
	if recH.Code != http.StatusOK {
		t.Fatalf("healthz status %d", recH.Code)
	}
	if !strings.Contains(recH.Body.String(), upstream.URL) {
		t.Fatalf("health payload missing upstream url")
	}
}

func TestHealthCheckMarksUnhealthy(t *testing.T) {
	badHealth := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer badHealth.Close()

	engine := newTestEngine()
	up := &upstreamTarget{url: mustParseURL(badHealth.URL), health: "/", status: upstreamStatus{Healthy: true, WindowStartedAt: time.Now()}}
	engine.healthClient = &http.Client{Timeout: time.Second}
	engine.checkUpstream(up)
	if up.status.Healthy {
		t.Fatalf("expected unhealthy status")
	}
}

func TestRetryTransportSucceedsAfterFailures(t *testing.T) {
	failures := 0
	rt := &retryTransport{
		base: roundTripFunc(func(req *http.Request) (*http.Response, error) {
			if failures < 2 {
				failures++
				return nil, fmt.Errorf("temp fail")
			}
			return &http.Response{
				StatusCode: http.StatusOK,
				Body:       http.NoBody,
				Request:    req,
			}, nil
		}),
		retries: 2,
	}
	req := httptest.NewRequest(http.MethodGet, "http://example.com", nil)
	resp, err := rt.RoundTrip(req)
	if err != nil {
		t.Fatalf("expected success after retries, got %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("unexpected status %d", resp.StatusCode)
	}
	if failures != 2 {
		t.Fatalf("expected 2 failures before success, got %d", failures)
	}
}

func TestLoadRouteDefinitionsWithFallback(t *testing.T) {
	file := `
routes:
  - name: auth
    prefix: /api/v1/auth
    upstream: http://auth:8080
fallback:
  base_path: /api
  health_path: /healthz
  mapping:
    hotels:
      upstream: http://hotel:8081
`
	tmp := t.TempDir() + "/routes.yml"
	if err := os.WriteFile(tmp, []byte(file), 0o644); err != nil {
		t.Fatal(err)
	}
	defs, err := loadRouteDefinitions(tmp)
	if err != nil {
		t.Fatalf("loadRouteDefinitions error: %v", err)
	}
	if len(defs) != 2 {
		t.Fatalf("expected 2 routes (1 explicit + 1 fallback), got %d", len(defs))
	}
	foundFallback := false
	for _, d := range defs {
		if strings.HasPrefix(d.Name, "fallback-") {
			foundFallback = true
		}
	}
	if !foundFallback {
		t.Fatalf("fallback route not generated")
	}
}

func TestCircuitBreakerOpens(t *testing.T) {
	up := &upstreamTarget{
		status: upstreamStatus{
			Healthy:         true,
			WindowStartedAt: time.Now(),
		},
	}
	now := time.Now()
	window := 30 * time.Second
	threshold := 0.5
	cooldown := 2 * time.Second

	// 3 failures out of 3 should open circuit
	for i := 0; i < 3; i++ {
		up.recordResult(false, now, window, threshold, cooldown)
	}
	up.mu.RLock()
	open := !up.status.CircuitOpenUntil.IsZero()
	up.mu.RUnlock()
	if !open {
		t.Fatalf("circuit should be open after consecutive failures")
	}
}

// helper
func mustParseURL(raw string) *url.URL {
	u, _ := url.Parse(raw)
	return u
}

func newTestEngine() *proxyEngine {
	return &proxyEngine{
		mode:             ModeProxyAll,
		log:              zaptest.NewLogger(nil),
		timeout:          2 * time.Second,
		retries:          1,
		metrics:          newGatewayMetrics(),
		transport:        http.DefaultTransport,
		upstreams:        map[string]*upstreamTarget{},
		readyCh:          make(chan struct{}),
		healthClient:     &http.Client{Timeout: time.Second},
		circuitWindow:    5 * time.Second,
		circuitThreshold: 0.5,
		circuitCooldown:  2 * time.Second,
	}
}

type roundTripFunc func(req *http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}
