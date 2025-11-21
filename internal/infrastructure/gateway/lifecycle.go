package gateway

import (
	"context"
	"fmt"
	"net/http"
	"time"
)

// lifecycle & health loop
func (p *proxyEngine) Start(ctx context.Context) {
	if len(p.routes) == 0 {
		p.readyOnce.Do(func() {
			close(p.readyCh)
		})
		return
	}

	go func() {
		p.runHealthChecks()
		p.readyOnce.Do(func() { close(p.readyCh) })

		if p.healthInterval <= 0 {
			p.healthInterval = 10 * time.Second
		}
		ticker := time.NewTicker(p.healthInterval)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				p.runHealthChecks()
			}
		}
	}()
}

func (p *proxyEngine) WaitUntilReady(ctx context.Context) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-p.readyCh:
		return nil
	}
}

func (p *proxyEngine) runHealthChecks() {
	for _, upstream := range p.upstreams {
		p.checkUpstream(upstream)
	}
}

func (p *proxyEngine) checkUpstream(target *upstreamTarget) {
	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, target.healthURL(), nil)
	if err != nil {
		return
	}

	resp, err := p.healthClient.Do(req)
	if err != nil {
		target.markUnhealthy(err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		target.markUnhealthy(fmt.Errorf("health check status %d", resp.StatusCode))
		return
	}
	target.markHealthy()
}
