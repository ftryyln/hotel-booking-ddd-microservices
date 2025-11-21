package gateway

import (
	"fmt"
	"strings"
	"time"
)

func (u *upstreamTarget) healthURL() string {
	healthPath := u.health
	if !strings.HasPrefix(healthPath, "/") {
		healthPath = "/" + healthPath
	}
	return fmt.Sprintf("%s://%s%s", u.url.Scheme, u.url.Host, healthPath)
}

func (u *upstreamTarget) markHealthy() {
	u.mu.Lock()
	defer u.mu.Unlock()
	u.status.Healthy = true
	u.status.LastChecked = time.Now()
	u.status.LastError = ""
	u.status.UnhealthySince = time.Time{}
	u.status.ConsecutiveErrors = 0
	if time.Now().After(u.status.CircuitOpenUntil) {
		u.status.CircuitOpenUntil = time.Time{}
	}
}

func (u *upstreamTarget) markUnhealthy(err error) {
	u.mu.Lock()
	defer u.mu.Unlock()
	u.status.Healthy = false
	u.status.LastChecked = time.Now()
	u.status.LastError = err.Error()
	if u.status.UnhealthySince.IsZero() {
		u.status.UnhealthySince = time.Now()
	}
	u.status.ConsecutiveErrors++
}

func (u *upstreamTarget) isAvailable(now time.Time) (bool, string) {
	u.mu.RLock()
	defer u.mu.RUnlock()
	if !u.status.CircuitOpenUntil.IsZero() && now.Before(u.status.CircuitOpenUntil) {
		return false, "circuit_open"
	}
	if !u.status.Healthy {
		return false, u.status.LastError
	}
	return true, ""
}

func (u *upstreamTarget) snapshot() upstreamStatus {
	u.mu.RLock()
	defer u.mu.RUnlock()
	return u.status
}

func (u *upstreamTarget) recordResult(success bool, now time.Time, window time.Duration, threshold float64, cooldown time.Duration) bool {
	if window <= 0 {
		window = 30 * time.Second
	}
	if threshold <= 0 {
		threshold = 0.5
	}

	u.mu.Lock()
	defer u.mu.Unlock()

	if now.Sub(u.status.WindowStartedAt) > window {
		u.status.WindowStartedAt = now
		u.status.RequestsInWindow = 0
		u.status.FailuresInWindow = 0
	}

	u.status.RequestsInWindow++
	if !success {
		u.status.FailuresInWindow++
		u.status.ConsecutiveErrors++
	} else {
		u.status.ConsecutiveErrors = 0
	}

	if !success && u.status.RequestsInWindow >= 3 {
		ratio := float64(u.status.FailuresInWindow) / float64(u.status.RequestsInWindow)
		if ratio >= threshold {
			u.status.CircuitOpenUntil = now.Add(cooldown)
			u.status.LastError = "circuit opened due to error ratio"
			return true
		}
	}
	return false
}
