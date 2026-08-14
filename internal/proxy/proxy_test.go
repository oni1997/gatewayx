package proxy

import (
	"testing"
	"time"

	"github.com/oni1997/gatewayx/internal/config"
)

func TestBuildAuthenticator_UnknownType(t *testing.T) {
	_, err := buildAuthenticator(&config.AuthConfig{Type: "unknown"})
	if err == nil {
		t.Error("expected error for unknown auth type")
	}
}

func TestBuildAuthenticator_JWT(t *testing.T) {
	a, err := buildAuthenticator(&config.AuthConfig{
		Type:    "jwt",
		Options: map[string]string{"secret": "test"},
	})
	if err != nil {
		t.Fatalf("expected success, got %v", err)
	}
	if a.Name() != "jwt" {
		t.Errorf("expected jwt, got %s", a.Name())
	}
}

func TestBuildAuthenticator_JWTWithCache(t *testing.T) {
	a, err := buildAuthenticator(&config.AuthConfig{
		Type:    "jwt",
		Options: map[string]string{"secret": "test", "cache_ttl": "1m"},
	})
	if err != nil {
		t.Fatalf("expected success, got %v", err)
	}
	if a.Name() != "cached_jwt" {
		t.Errorf("expected cached_jwt, got %s", a.Name())
	}
}

func TestBuildAuthenticator_APIKey(t *testing.T) {
	a, err := buildAuthenticator(&config.AuthConfig{
		Type:    "api_key",
		Options: map[string]string{"key1": "owner1", "header": "X-Key"},
	})
	if err != nil {
		t.Fatalf("expected success, got %v", err)
	}
	if a.Name() != "api_key" {
		t.Errorf("expected api_key, got %s", a.Name())
	}
}

func TestBuildAuthenticator_Basic(t *testing.T) {
	a, err := buildAuthenticator(&config.AuthConfig{
		Type:    "basic",
		Options: map[string]string{"admin": "pass"},
	})
	if err != nil {
		t.Fatalf("expected success, got %v", err)
	}
	if a.Name() != "basic" {
		t.Errorf("expected basic, got %s", a.Name())
	}
}

func TestBuildAuthenticator_HMAC(t *testing.T) {
	a, err := buildAuthenticator(&config.AuthConfig{
		Type:    "hmac",
		Options: map[string]string{"secret": "test"},
	})
	if err != nil {
		t.Fatalf("expected success, got %v", err)
	}
	if a.Name() != "hmac" {
		t.Errorf("expected hmac, got %s", a.Name())
	}
}

func TestBuildAuthenticator_MTLS_MissingCA(t *testing.T) {
	_, err := buildAuthenticator(&config.AuthConfig{
		Type:    "mtls",
		Options: map[string]string{"ca_cert": "/nonexistent"},
	})
	if err == nil {
		t.Error("expected error for missing CA cert")
	}
}

func TestBuildAuthenticator_OAuth(t *testing.T) {
	a, err := buildAuthenticator(&config.AuthConfig{
		Type:    "oauth",
		Options: map[string]string{"provider": "github", "client_id": "id", "client_secret": "secret"},
	})
	if err != nil {
		t.Fatalf("expected success, got %v", err)
	}
	if a.Name() != "oauth" {
		t.Errorf("expected oauth, got %s", a.Name())
	}
}

func TestBuildAuthenticator_Session(t *testing.T) {
	a, err := buildAuthenticator(&config.AuthConfig{
		Type:    "session",
		Options: map[string]string{"ttl": "1h", "max_sessions": "100"},
	})
	if err != nil {
		t.Fatalf("expected success, got %v", err)
	}
	if a.Name() != "session" {
		t.Errorf("expected session, got %s", a.Name())
	}
}

func TestBuildAuthenticator_RBAC(t *testing.T) {
	a, err := buildAuthenticator(&config.AuthConfig{
		Type: "rbac",
		Options: map[string]string{
			"delegate": "jwt",
			"secret":   "test",
		},
	})
	if err != nil {
		t.Fatalf("expected success, got %v", err)
	}
	if a.Name() != "rbac" {
		t.Errorf("expected rbac, got %s", a.Name())
	}
}

func TestBuildRBACEngine(t *testing.T) {
	opts := map[string]string{
		"perm_1": "/admin/**:admin:GET,POST",
		"perm_2": "/api/**:admin,developer",
	}
	engine := buildRBACEngine(opts)

	if !engine.CheckPermission([]string{"admin"}, "GET", "/admin/users") {
		t.Error("admin should access /admin/users")
	}
	if engine.CheckPermission([]string{"developer"}, "GET", "/admin/users") {
		t.Error("developer should not access /admin")
	}
}

func TestParseDuration(t *testing.T) {
	if d := parseDuration("30s"); d != 30*time.Second {
		t.Errorf("expected 30s, got %v", d)
	}
	if d := parseDuration(""); d != 0 {
		t.Errorf("expected 0 for empty, got %v", d)
	}
	if d := parseDuration("invalid"); d != 0 {
		t.Errorf("expected 0 for invalid, got %v", d)
	}
}
