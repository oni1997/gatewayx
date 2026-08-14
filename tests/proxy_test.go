package main

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/oni1997/gatewayx/internal/config"
	"github.com/oni1997/gatewayx/internal/proxy"
)

func TestProxyStartup(t *testing.T) {
	cfg := &config.Config{
		Server: config.ServerConfig{
			Host: "127.0.0.1",
			Port: 0,
		},
		Health: config.HealthConfig{
			Enabled: false,
		},
	}

	logger := stubLogger()

	_, err := proxy.New(cfg, logger)
	if err != nil {
		t.Fatalf("failed to create proxy: %v", err)
	}
}

func TestProxyNoRoutes(t *testing.T) {
	cfg := &config.Config{
		Server: config.ServerConfig{
			Host: "127.0.0.1",
			Port: 0,
		},
	}

	logger := stubLogger()

	rp, err := proxy.New(cfg, logger)
	if err != nil {
		t.Fatalf("failed to create proxy: %v", err)
	}

	req := httptest.NewRequest("GET", "/anything", nil)
	rec := httptest.NewRecorder()

	rp.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", rec.Code)
	}
}

func TestProxyWithRoute(t *testing.T) {
	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	}))
	defer backend.Close()

	cfg := &config.Config{
		Server: config.ServerConfig{
			Host: "127.0.0.1",
			Port: 0,
		},
		Routes: []config.RouteConfig{
			{
				Name:         "test",
				ListenPath:   "/api",
				UpstreamURLs: []string{backend.URL},
			},
		},
	}

	logger := stubLogger()

	rp, err := proxy.New(cfg, logger)
	if err != nil {
		t.Fatalf("failed to create proxy: %v", err)
	}

	req := httptest.NewRequest("GET", "/api/test", nil)
	rec := httptest.NewRecorder()

	rp.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestProxyWithCache(t *testing.T) {
	hits := 0
	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hits++
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("cached"))
	}))
	defer backend.Close()

	cfg := &config.Config{
		Server: config.ServerConfig{Host: "127.0.0.1", Port: 0},
		Routes: []config.RouteConfig{
			{
				Name:         "cached-route",
				ListenPath:   "/cached",
				UpstreamURLs: []string{backend.URL},
				Cache:        &config.CacheConfig{TTL: 30 * time.Second},
			},
		},
	}

	rp, err := proxy.New(cfg, stubLogger())
	if err != nil {
		t.Fatalf("failed to create proxy: %v", err)
	}

	for i := 0; i < 3; i++ {
		req := httptest.NewRequest("GET", "/cached/data", nil)
		rec := httptest.NewRecorder()
		rp.ServeHTTP(rec, req)
		if rec.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d", rec.Code)
		}
	}

	if hits != 1 {
		t.Errorf("expected backend hit once (cached), got %d hits", hits)
	}
}

func TestProxyWithCompression(t *testing.T) {
	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("compress me please"))
	}))
	defer backend.Close()

	cfg := &config.Config{
		Server: config.ServerConfig{Host: "127.0.0.1", Port: 0},
		Routes: []config.RouteConfig{
			{
				Name:         "compressed",
				ListenPath:   "/gzip",
				UpstreamURLs: []string{backend.URL},
				Compression:  true,
			},
		},
	}

	rp, err := proxy.New(cfg, stubLogger())
	if err != nil {
		t.Fatalf("failed to create proxy: %v", err)
	}

	req := httptest.NewRequest("GET", "/gzip/data", nil)
	req.Header.Set("Accept-Encoding", "gzip")
	rec := httptest.NewRecorder()

	rp.ServeHTTP(rec, req)

	if rec.Header().Get("Content-Encoding") != "gzip" {
		t.Errorf("expected gzip content encoding, got %s", rec.Header().Get("Content-Encoding"))
	}
}

func TestProxyWithWebsocket(t *testing.T) {
	cfg := &config.Config{
		Server: config.ServerConfig{Host: "127.0.0.1", Port: 0},
		Routes: []config.RouteConfig{
			{
				Name:         "ws",
				ListenPath:   "/ws",
				UpstreamURLs: []string{"http://localhost:9999"},
				Websocket:    true,
			},
		},
	}

	rp, err := proxy.New(cfg, stubLogger())
	if err != nil {
		t.Fatalf("failed to create proxy: %v", err)
	}

	req := httptest.NewRequest("GET", "/ws/socket", nil)
	rec := httptest.NewRecorder()

	rp.ServeHTTP(rec, req)

	// Will be 502 since backend is unreachable, but should not be a timeout
	if rec.Code != http.StatusBadGateway {
		t.Errorf("expected 502 for unreachable ws backend, got %d", rec.Code)
	}
}

func TestProxyWithHealthCheck(t *testing.T) {
	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer backend.Close()

	cfg := &config.Config{
		Server: config.ServerConfig{Host: "127.0.0.1", Port: 0},
		Routes: []config.RouteConfig{
			{
				Name:         "healthy",
				ListenPath:   "/hc",
				UpstreamURLs: []string{backend.URL},
				HealthCheck: &config.HealthCheckConfig{
					Path:      "/",
					Interval:  100 * time.Millisecond,
					Healthy:   1,
					Unhealthy: 1,
				},
			},
		},
	}

	rp, err := proxy.New(cfg, stubLogger())
	if err != nil {
		t.Fatalf("failed to create proxy: %v", err)
	}

	req := httptest.NewRequest("GET", "/hc/test", nil)
	rec := httptest.NewRecorder()

	rp.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestProxyReloadConfig(t *testing.T) {
	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer backend.Close()

	cfg := &config.Config{
		Server: config.ServerConfig{Host: "127.0.0.1", Port: 0},
		Routes: []config.RouteConfig{
			{Name: "r1", ListenPath: "/v1", UpstreamURLs: []string{backend.URL}},
		},
	}

	rp, err := proxy.New(cfg, stubLogger())
	if err != nil {
		t.Fatalf("failed to create proxy: %v", err)
	}

	newCfg := &config.Config{
		Server: config.ServerConfig{Host: "127.0.0.1", Port: 0},
		Routes: []config.RouteConfig{
			{Name: "r1", ListenPath: "/v1", UpstreamURLs: []string{backend.URL}},
			{Name: "r2", ListenPath: "/v2", UpstreamURLs: []string{backend.URL}},
		},
	}

	if err := rp.ReloadConfig(newCfg); err != nil {
		t.Fatalf("reload failed: %v", err)
	}

	req := httptest.NewRequest("GET", "/v2/test", nil)
	rec := httptest.NewRecorder()
	rp.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected 200 for newly added route, got %d", rec.Code)
	}
}

func TestConfigValidation(t *testing.T) {
	tests := []struct {
		name    string
		cfg     *config.Config
		wantErr bool
	}{
		{
			name: "valid config",
			cfg: &config.Config{
				Server: config.ServerConfig{Port: 8080},
			},
			wantErr: false,
		},
		{
			name: "invalid port",
			cfg: &config.Config{
				Server: config.ServerConfig{Port: 0},
			},
			wantErr: true,
		},
		{
			name: "route without upstreams",
			cfg: &config.Config{
				Server: config.ServerConfig{Port: 8080},
				Routes: []config.RouteConfig{
					{Name: "bad", ListenPath: "/api"},
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.cfg.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
