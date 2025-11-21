package gateway

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"go.uber.org/zap"

	"github.com/ftryyln/hotel-booking-microservices/pkg/config"
)

const (
	// ModeWhitelist keeps existing explicit routes only.
	ModeWhitelist = "whitelist"
	// ModeProxyAll enables fallback routing for every service.
	ModeProxyAll = "proxy_all"
)

type proxyEngine struct {
	mode             string
	routes           []*route
	log              *zap.Logger
	timeout          time.Duration
	retries          int
	metrics          *gatewayMetrics
	transport        http.RoundTripper
	healthInterval   time.Duration
	healthClient     *http.Client
	readyCh          chan struct{}
	readyOnce        sync.Once
	upstreams        map[string]*upstreamTarget
	jwtSecret        string
	circuitWindow    time.Duration
	circuitThreshold float64
	circuitCooldown  time.Duration
}

func NewProxyEngine(cfg config.Config, log *zap.Logger) (*proxyEngine, error) {
	engine := &proxyEngine{
		mode:           cfg.GatewayMode,
		log:            log,
		timeout:        cfg.UpstreamTimeout,
		retries:        cfg.UpstreamRetries,
		metrics:        newGatewayMetrics(),
		healthInterval: cfg.HealthInterval,
		healthClient: &http.Client{
			Timeout: 3 * time.Second,
		},
		readyCh:          make(chan struct{}),
		upstreams:        map[string]*upstreamTarget{},
		jwtSecret:        cfg.JWTSecret,
		circuitWindow:    cfg.CircuitWindow,
		circuitThreshold: cfg.CircuitThreshold,
		circuitCooldown:  cfg.CircuitCooldown,
	}

	engine.transport = &retryTransport{
		base: &http.Transport{
			Proxy:                 http.ProxyFromEnvironment,
			MaxIdleConns:          128,
			MaxIdleConnsPerHost:   32,
			IdleConnTimeout:       90 * time.Second,
			TLSHandshakeTimeout:   5 * time.Second,
			ResponseHeaderTimeout: cfg.UpstreamTimeout,
		},
		retries: cfg.UpstreamRetries,
	}

	if engine.mode == "" {
		engine.mode = ModeWhitelist
	}

	definitions, err := loadRouteDefinitions(cfg.RoutesFile)
	if err != nil {
		if engine.mode == ModeProxyAll {
			return nil, fmt.Errorf("proxy mode requires routes configuration: %w", err)
		}
		log.Warn("unable to load routes file, proxy_all disabled", zap.Error(err))
		definitions = nil
	}

	if err := engine.buildRoutes(definitions); err != nil {
		return nil, err
	}

	return engine, nil
}

// Metrics
// @Summary Gateway Prometheus metrics
// @Tags Diagnostics
// @Produce plain
// @Success 200 {string} string "prometheus metrics"
// @Router /metrics [get]
func (p *proxyEngine) Metrics(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "text/plain; version=0.0.4")
	for _, line := range p.metrics.Format() {
		_, _ = w.Write([]byte(line + "\n"))
	}
}

// DebugRoutes
// @Summary List active proxy routes and upstream health
// @Tags Diagnostics
// @Produce json
// @Success 200 {array} gateway.DebugRouteResponse
// @Router /debug/routes [get]
func (p *proxyEngine) DebugRoutes(w http.ResponseWriter, _ *http.Request) {
	var payload []DebugRouteResponse
	for _, route := range p.routes {
		payload = append(payload, DebugRouteResponse{
			Name:        route.name,
			Prefix:      route.prefix,
			Upstream:    route.target.url.String(),
			RequireAuth: route.requireAuth,
			Auth:        route.authStrategy,
			Status:      statusDocFrom(route.target.snapshot()),
		})
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(payload)
}

// Healthz
// @Summary Aggregated upstream health
// @Tags Diagnostics
// @Produce json
// @Success 200 {array} gateway.HealthResponse
// @Failure 503 {array} gateway.HealthResponse
// @Router /healthz [get]
func (p *proxyEngine) Healthz(w http.ResponseWriter, _ *http.Request) {
	payload := make([]HealthResponse, 0, len(p.upstreams))
	healthy := true
	for _, up := range p.upstreams {
		state := up.snapshot()
		if !state.Healthy {
			healthy = false
		}
		payload = append(payload, HealthResponse{
			Upstream: up.url.String(),
			Status:   statusDocFrom(state),
		})
	}
	status := http.StatusOK
	if !healthy {
		status = http.StatusServiceUnavailable
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

// utility helpers used across files
func cleanPath(path string) string {
	if path == "" {
		return "/"
	}
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}
	for strings.Contains(path, "//") {
		path = strings.ReplaceAll(path, "//", "/")
	}
	return path
}
