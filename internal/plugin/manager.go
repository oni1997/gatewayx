package plugin

import (
	"context"
	"fmt"
	"net/http"

	sdkplugin "github.com/oni1997/gatewayx/sdk/plugin"
)

type Manager struct {
	host   *Host
	events *sdkplugin.Events
}

func NewManager(host *Host, events *sdkplugin.Events) *Manager {
	return &Manager{host: host, events: events}
}

func (m *Manager) ExecutePreRequest(ctx context.Context, req *http.Request) (context.Context, error) {
	plugins := m.host.List()
	for _, p := range plugins {
		var err error
		ctx, err = p.OnRequest(ctx, req)
		if err != nil {
			p.OnError(ctx, err)
			return ctx, fmt.Errorf("plugin %s OnRequest: %w", p.Name(), err)
		}
	}
	return ctx, nil
}

func (m *Manager) ExecutePostResponse(ctx context.Context, res *http.Response) error {
	plugins := m.host.List()
	for _, p := range plugins {
		if err := p.OnResponse(ctx, res); err != nil {
			p.OnError(ctx, err)
			return fmt.Errorf("plugin %s OnResponse: %w", p.Name(), err)
		}
	}
	return nil
}

func (m *Manager) ExecuteOnError(ctx context.Context, err error) {
	plugins := m.host.List()
	for _, p := range plugins {
		p.OnError(ctx, err)
	}
}

func (m *Manager) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx, err := m.ExecutePreRequest(r.Context(), r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		r = r.WithContext(ctx)

		sw := &responseCapture{ResponseWriter: w, status: 200}
		next.ServeHTTP(sw, r)

		if sw.status >= 400 {
			m.ExecuteOnError(ctx, fmt.Errorf("HTTP %d", sw.status))
		}
	})
}

type responseCapture struct {
	http.ResponseWriter
	status int
}

func (rc *responseCapture) WriteHeader(code int) {
	rc.status = code
	rc.ResponseWriter.WriteHeader(code)
}
