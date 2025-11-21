package gateway

import (
	"context"
	"encoding/json"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"go.uber.org/zap"

	pkgErrors "github.com/ftryyln/hotel-booking-microservices/pkg/errors"
	"github.com/ftryyln/hotel-booking-microservices/pkg/middleware"
)

// proxy serve
func (p *proxyEngine) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if p.mode != ModeProxyAll {
		writeProxyError(w, pkgErrors.New("not_found", "proxy mode disabled (whitelist)"))
		return
	}

	route := p.matchRoute(r.URL.Path)
	if route == nil {
		writeProxyError(w, pkgErrors.New("not_found", "no upstream mapping"))
		return
	}

	if len(route.methods) > 0 {
		if _, ok := route.methods[r.Method]; !ok {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
	}

	now := time.Now()
	if ok, reason := route.target.isAvailable(now); !ok {
		writeProxyError(w, pkgErrors.New("service_unavailable", reason))
		p.metrics.Observe(route.name, http.StatusServiceUnavailable, 0)
		return
	}

	if err := p.ensureAuth(r, route); err.Code != "" {
		writeProxyError(w, err)
		p.metrics.Observe(route.name, pkgErrors.StatusCode(err), 0)
		return
	}

	p.forward(route, w, r)
}

func (p *proxyEngine) ensureAuth(r *http.Request, route *route) pkgErrors.APIError {
	if !route.requireAuth {
		return pkgErrors.APIError{}
	}

	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		return pkgErrors.New("unauthorized", "missing Authorization header")
	}

	if route.authStrategy != "validate" {
		return pkgErrors.APIError{}
	}
	tokenString := extractBearer(authHeader)
	if tokenString == "" {
		return pkgErrors.New("unauthorized", "missing bearer token")
	}

	if p.jwtSecret == "" {
		return pkgErrors.New("unauthorized", "jwt secret not configured")
	}

	claims := &middleware.Claims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		return []byte(p.jwtSecret), nil
	})
	if err != nil || !token.Valid {
		return pkgErrors.New("unauthorized", "invalid token")
	}
	return pkgErrors.APIError{}
}

func extractBearer(header string) string {
	parts := strings.Split(header, " ")
	if len(parts) != 2 {
		return ""
	}
	if !strings.EqualFold(parts[0], "bearer") {
		return ""
	}
	return parts[1]
}

func (p *proxyEngine) matchRoute(path string) *route {
	for _, route := range p.routes {
		if route.matches(path) {
			return route
		}
	}
	return nil
}

func (r *route) matches(path string) bool {
	if r.prefix == "/" {
		return true
	}
	if strings.HasPrefix(path, r.prefix) {
		if len(path) == len(r.prefix) {
			return true
		}
		if r.prefix == "/" {
			return true
		}
		if strings.HasPrefix(path, r.prefix+"/") {
			return true
		}
	}
	return false
}

func (p *proxyEngine) forward(route *route, w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	ctx, cancel := context.WithTimeout(r.Context(), p.timeout)
	defer cancel()

	req := r.Clone(ctx)
	req.URL.Path = route.rewritePath(r.URL.Path)
	req.URL.RawPath = req.URL.Path

	rec := &proxyResponseWriter{ResponseWriter: w, status: http.StatusOK}

	route.proxy.ServeHTTP(rec, req)
	elapsed := time.Since(start)

	p.metrics.Observe(route.name, rec.status, elapsed)
	if route.target.recordResult(rec.status < http.StatusInternalServerError, time.Now(), p.circuitWindow, p.circuitThreshold, p.circuitCooldown) {
		go p.checkUpstream(route.target)
	}

	p.log.Info("proxy request",
		zap.String("method", r.Method),
		zap.String("path", r.URL.Path),
		zap.String("route", route.prefix),
		zap.String("upstream", route.target.url.Host),
		zap.Int("status", rec.status),
		zap.Float64("latency_ms", float64(elapsed.Milliseconds())),
		zap.String("remote_ip", remoteIP(r)),
		zap.String("user_agent", r.UserAgent()),
	)
}

func (r *route) rewritePath(path string) string {
	if r.prefix == "/" {
		return path
	}

	if r.rewrite != "" {
		suffix := strings.TrimPrefix(path, r.prefix)
		return cleanPath(r.rewrite + suffix)
	}

	if r.stripPrefix && strings.HasPrefix(path, r.prefix) {
		newPath := strings.TrimPrefix(path, r.prefix)
		if newPath == "" {
			return "/"
		}
		if !strings.HasPrefix(newPath, "/") {
			newPath = "/" + newPath
		}
		return newPath
	}
	return path
}

func remoteIP(r *http.Request) string {
	if ip := r.Header.Get("X-Forwarded-For"); ip != "" {
		parts := strings.Split(ip, ",")
		return strings.TrimSpace(parts[0])
	}
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}

func writeProxyError(w http.ResponseWriter, err pkgErrors.APIError) {
	status := pkgErrors.StatusCode(err)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(err)
}

type proxyResponseWriter struct {
	http.ResponseWriter
	status int
}

func (w *proxyResponseWriter) WriteHeader(code int) {
	w.status = code
	w.ResponseWriter.WriteHeader(code)
}
