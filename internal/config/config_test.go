package config

import (
	"os"
	"testing"
)

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()
	if cfg.Server.Port != 8080 {
		t.Errorf("expected port 8080, got %d", cfg.Server.Port)
	}
	if cfg.Logging.Level != "info" {
		t.Errorf("expected level info, got %s", cfg.Logging.Level)
	}
}

func TestLoadConfigFromFile(t *testing.T) {
	yaml := `
server:
  host: "127.0.0.1"
  port: 9090
logging:
  level: "debug"
routes:
  - name: "test"
    listen_path: "/api"
    upstream_urls:
      - "http://localhost:3000"
`

	tmpfile, err := os.CreateTemp("", "gatewayx-test-*.yaml")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.Remove(tmpfile.Name()) }()

	if _, err := tmpfile.Write([]byte(yaml)); err != nil {
		t.Fatal(err)
	}
	_ = tmpfile.Close()

	cfg, err := LoadConfig(tmpfile.Name())
	if err != nil {
		t.Fatalf("failed to load config: %v", err)
	}

	if cfg.Server.Port != 9090 {
		t.Errorf("expected port 9090, got %d", cfg.Server.Port)
	}
	if len(cfg.Routes) != 1 {
		t.Errorf("expected 1 route, got %d", len(cfg.Routes))
	}
}

func TestLoadConfig_DefaultsApplied(t *testing.T) {
	yaml := `
server:
  port: 7000
routes:
  - name: "r1"
    listen_path: "/api"
    upstream_urls: ["http://localhost:3000"]
`

	tmpfile, _ := os.CreateTemp("", "gatewayx-*.yaml")
	defer func() { _ = os.Remove(tmpfile.Name()) }()
	_, _ = tmpfile.Write([]byte(yaml))
	_ = tmpfile.Close()

	cfg, err := LoadConfig(tmpfile.Name())
	if err != nil {
		t.Fatalf("failed: %v", err)
	}

	if cfg.Routes[0].LoadBalancing != "round_robin" {
		t.Errorf("expected default round_robin, got %s", cfg.Routes[0].LoadBalancing)
	}
	if cfg.Routes[0].Timeout == 0 {
		t.Error("expected default timeout to be set")
	}
}

func TestLoadConfig_NonexistentFile(t *testing.T) {
	_, err := LoadConfig("/nonexistent/path.yaml")
	if err == nil {
		t.Error("expected error for explicitly missing config file")
	}
}

func TestLoadConfig_InvalidYAML(t *testing.T) {
	tmpfile, _ := os.CreateTemp("", "gatewayx-*.yaml")
	defer func() { _ = os.Remove(tmpfile.Name()) }()
	_, _ = tmpfile.Write([]byte("not: [valid: yaml"))
	_ = tmpfile.Close()

	_, err := LoadConfig(tmpfile.Name())
	if err == nil {
		t.Error("expected error for invalid YAML")
	}
}

func TestValidate_TLSRequiresCert(t *testing.T) {
	cfg := DefaultConfig()
	cfg.TLS.Enabled = true
	cfg.TLS.CertFile = ""
	cfg.TLS.KeyFile = ""

	err := cfg.Validate()
	if err == nil {
		t.Error("expected error when TLS enabled without cert")
	}
}

func TestValidate_RouteWithoutPath(t *testing.T) {
	cfg := DefaultConfig()
	cfg.Routes = []RouteConfig{
		{Name: "bad", UpstreamURLs: []string{"http://localhost:3000"}},
	}

	err := cfg.Validate()
	if err == nil {
		t.Error("expected error for route without listen_path")
	}
}

func TestJWTOptions_EnvExpansion(t *testing.T) {
	t.Setenv("TEST_SECRET", "mysecret")

	ac := &AuthConfig{
		Type: "jwt",
		Options: map[string]string{
			"secret": "${TEST_SECRET}",
		},
	}

	opts := ac.JWTOptions()
	if opts["secret"] != "mysecret" {
		t.Errorf("expected env expansion, got %s", opts["secret"])
	}
}

func TestAllOptions(t *testing.T) {
	ac := &AuthConfig{
		Options: map[string]string{"a": "1", "b": "2"},
	}
	opts := ac.AllOptions()
	if len(opts) != 2 {
		t.Errorf("expected 2 options, got %d", len(opts))
	}
}
