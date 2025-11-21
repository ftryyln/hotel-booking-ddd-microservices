package gateway

import (
	"net/http/httputil"
	"net/url"
	"sync"
	"time"

	"gopkg.in/yaml.v3"
)

type route struct {
	name         string
	prefix       string
	stripPrefix  bool
	rewrite      string
	requireAuth  bool
	authStrategy string
	methods      map[string]struct{}
	target       *upstreamTarget
	proxy        *httputil.ReverseProxy
}

type upstreamTarget struct {
	name   string
	url    *url.URL
	health string

	mu     sync.RWMutex
	status upstreamStatus
}

type upstreamStatus struct {
	Healthy           bool      `json:"healthy"`
	LastChecked       time.Time `json:"last_checked"`
	LastError         string    `json:"last_error,omitempty"`
	CircuitOpenUntil  time.Time `json:"circuit_open_until,omitempty"`
	RequestsInWindow  int       `json:"requests_in_window"`
	FailuresInWindow  int       `json:"failures_in_window"`
	WindowStartedAt   time.Time `json:"window_started_at"`
	UnhealthySince    time.Time `json:"unhealthy_since,omitempty"`
	ConsecutiveErrors int       `json:"consecutive_errors"`
}

type routeFile struct {
	Routes   []routeDefinition   `yaml:"routes"`
	Fallback *fallbackDefinition `yaml:"fallback"`
}

type routeDefinition struct {
	Name         string   `yaml:"name"`
	Prefix       string   `yaml:"prefix"`
	Upstream     string   `yaml:"upstream"`
	StripPrefix  bool     `yaml:"strip_prefix"`
	Rewrite      string   `yaml:"rewrite"`
	RequireAuth  bool     `yaml:"require_auth"`
	AuthStrategy string   `yaml:"auth_strategy"`
	HealthPath   string   `yaml:"health_path"`
	Methods      []string `yaml:"methods"`
}

type fallbackDefinition struct {
	BasePath   string                   `yaml:"base_path"`
	StripBase  bool                     `yaml:"strip_base"`
	HealthPath string                   `yaml:"health_path"`
	Mapping    map[string]fallbackRoute `yaml:"mapping"`
}

type fallbackRoute struct {
	Upstream     string `yaml:"upstream"`
	StripPrefix  bool   `yaml:"strip_prefix"`
	RequireAuth  bool   `yaml:"require_auth"`
	AuthStrategy string `yaml:"auth_strategy"`
	HealthPath   string `yaml:"health_path"`
}

// ensure yaml is imported for route file parsing
var _ = yaml.Node{}
