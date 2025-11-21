package gateway

import (
	"errors"
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"sort"
	"strings"
	"time"

	"go.uber.org/zap"
	"gopkg.in/yaml.v3"

	pkgErrors "github.com/ftryyln/hotel-booking-microservices/pkg/errors"
)

// route building & config parsing
func (p *proxyEngine) buildRoutes(defs []routeDefinition) error {
	if len(defs) == 0 {
		p.routes = nil
		return nil
	}

	for _, def := range defs {
		if def.Upstream == "" || def.Prefix == "" {
			continue
		}

		targetURL, err := url.Parse(def.Upstream)
		if err != nil {
			return fmt.Errorf("invalid upstream for prefix %s: %w", def.Prefix, err)
		}

		key := targetURL.String()
		up, ok := p.upstreams[key]
		if !ok {
			healthPath := def.HealthPath
			if healthPath == "" {
				healthPath = "/healthz"
			}
			up = &upstreamTarget{
				name:   targetURL.Host,
				url:    targetURL,
				health: healthPath,
				status: upstreamStatus{
					Healthy:         true,
					WindowStartedAt: time.Now(),
				},
			}
			p.upstreams[key] = up
		}

		route := &route{
			name:         def.Name,
			prefix:       normalizePrefix(def.Prefix),
			stripPrefix:  def.StripPrefix,
			rewrite:      def.Rewrite,
			requireAuth:  def.RequireAuth,
			authStrategy: normalizeAuthStrategy(def.AuthStrategy),
			target:       up,
		}

		if len(def.Methods) > 0 {
			route.methods = map[string]struct{}{}
			for _, m := range def.Methods {
				route.methods[strings.ToUpper(m)] = struct{}{}
			}
		}

		if route.name == "" {
			route.name = route.prefix
		}

		route.proxy = p.newReverseProxy(up.url)
		route.proxy.ErrorHandler = p.makeErrorHandler(route)

		p.routes = append(p.routes, route)
	}

	sort.SliceStable(p.routes, func(i, j int) bool {
		return len(p.routes[i].prefix) > len(p.routes[j].prefix)
	})
	return nil
}

func (p *proxyEngine) newReverseProxy(target *url.URL) *httputil.ReverseProxy {
	proxy := httputil.NewSingleHostReverseProxy(target)
	proxy.Transport = p.transport
	return proxy
}

func (p *proxyEngine) makeErrorHandler(route *route) func(http.ResponseWriter, *http.Request, error) {
	return func(w http.ResponseWriter, r *http.Request, err error) {
		writeProxyError(w, pkgErrors.New("bad_gateway", err.Error()))
		p.metrics.Observe(route.name, http.StatusBadGateway, 0)
		if route.target.recordResult(false, time.Now(), p.circuitWindow, p.circuitThreshold, p.circuitCooldown) {
			go p.checkUpstream(route.target)
		}
		p.log.Warn("proxy upstream error",
			zap.String("route", route.prefix),
			zap.String("path", r.URL.Path),
			zap.Error(err),
		)
	}
}

func normalizeAuthStrategy(strategy string) string {
	switch strings.ToLower(strategy) {
	case "validate":
		return "validate"
	default:
		return "forward"
	}
}

func normalizePrefix(prefix string) string {
	if prefix == "" {
		return "/"
	}
	if !strings.HasPrefix(prefix, "/") {
		prefix = "/" + prefix
	}
	if len(prefix) > 1 {
		prefix = strings.TrimRight(prefix, "/")
	}
	return prefix
}

func loadRouteDefinitions(path string) ([]routeDefinition, error) {
	if path == "" {
		return nil, errors.New("routes file path is empty")
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var file routeFile
	if err := yaml.Unmarshal(data, &file); err != nil {
		return nil, err
	}
	defs := append([]routeDefinition{}, file.Routes...)
	if !nullFallback(file.Fallback) {
		base := strings.TrimRight(file.Fallback.BasePath, "/")
		if base == "" {
			base = "/api"
		}
		for key, mapping := range file.Fallback.Mapping {
			prefix := fmt.Sprintf("%s/%s", base, key)
			strip := mapping.StripPrefix
			if file.Fallback.StripBase {
				strip = true
			}
			defs = append(defs, routeDefinition{
				Name:         fmt.Sprintf("fallback-%s", key),
				Prefix:       prefix,
				Upstream:     mapping.Upstream,
				StripPrefix:  strip,
				RequireAuth:  mapping.RequireAuth,
				AuthStrategy: mapping.AuthStrategy,
				HealthPath:   firstNonEmpty(mapping.HealthPath, file.Fallback.HealthPath, "/healthz"),
			})
		}
	}
	return defs, nil
}

func nullFallback(fb *fallbackDefinition) bool {
	return fb == nil || fb.BasePath == "" || len(fb.Mapping) == 0
}

func firstNonEmpty(values ...string) string {
	for _, v := range values {
		if v != "" {
			return v
		}
	}
	return ""
}
