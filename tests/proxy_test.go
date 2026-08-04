package main

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gatewayx/gatewayx/internal/config"
	"github.com/gatewayx/gatewayx/internal/proxy"
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
		w.Write([]byte("ok"))
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
