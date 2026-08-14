package admin

import (
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/oni1997/gatewayx/internal/metrics"
)

func testHandler() http.Handler {
	store := NewStore()
	store.SetAuditLog(NewAuditLog(slog.New(slog.NewTextHandler(io.Discard, nil))))
	collector := metrics.NewCollector()
	return NewHandler(store, collector)
}

func TestListKeys_Empty(t *testing.T) {
	handler := testHandler()

	req := httptest.NewRequest("GET", "/api/keys", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
	if rec.Body.String() != "[]\n" {
		t.Errorf("expected empty array, got %s", rec.Body.String())
	}
}

func TestCreateKey(t *testing.T) {
	handler := testHandler()

	req := httptest.NewRequest("POST", "/api/keys", strings.NewReader(`{"name":"test","owner":"me"}`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Errorf("expected 201, got %d", rec.Code)
	}

	var result map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &result); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}
	if result["name"] != "test" {
		t.Errorf("expected name=test, got %v", result["name"])
	}
	key, ok := result["key"].(string)
	if !ok || !strings.HasPrefix(key, "sk-") {
		t.Errorf("expected key starting with sk-, got %v", result["key"])
	}
}

func TestCreateKey_InvalidBody(t *testing.T) {
	handler := testHandler()

	req := httptest.NewRequest("POST", "/api/keys", strings.NewReader(`{invalid`))
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rec.Code)
	}
}

func TestRevokeKey(t *testing.T) {
	store := NewStore()
	handler := NewHandler(store, metrics.NewCollector())

	key, _ := store.CreateKey("temp", "owner")

	req := httptest.NewRequest("DELETE", "/api/keys/"+key.ID, nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Errorf("expected 204, got %d", rec.Code)
	}
}

func TestRevokeKey_NotFound(t *testing.T) {
	handler := testHandler()

	req := httptest.NewRequest("DELETE", "/api/keys/nonexistent", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", rec.Code)
	}
}

func TestListCerts_Empty(t *testing.T) {
	handler := testHandler()

	req := httptest.NewRequest("GET", "/api/certs", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestCreateCert(t *testing.T) {
	handler := testHandler()

	req := httptest.NewRequest("POST", "/api/certs", strings.NewReader(`{"domain":"api.example.com","issuer":"LE"}`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Errorf("expected 201, got %d", rec.Code)
	}

	var cert Certificate
	if err := json.Unmarshal(rec.Body.Bytes(), &cert); err != nil {
		t.Fatalf("failed to parse cert: %v", err)
	}
	if cert.Domain != "api.example.com" {
		t.Errorf("expected domain api.example.com, got %s", cert.Domain)
	}
	if cert.Status != "active" {
		t.Errorf("expected active, got %s", cert.Status)
	}
}

func TestConfigEndpoint(t *testing.T) {
	handler := testHandler()

	req := httptest.NewRequest("GET", "/api/config", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}

	var result map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &result); err != nil {
		t.Fatalf("failed to parse config: %v", err)
	}
	if _, ok := result["uptime"]; !ok {
		t.Error("expected uptime field")
	}
}

func TestAuditEndpoint(t *testing.T) {
	store := NewStore()
	store.SetAuditLog(NewAuditLog(slog.New(slog.NewTextHandler(io.Discard, nil))))
	handler := NewHandler(store, metrics.NewCollector())

	store.CreateKey("audited", "owner")

	req := httptest.NewRequest("GET", "/api/audit", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}

	var entries []AuditEntry
	if err := json.Unmarshal(rec.Body.Bytes(), &entries); err != nil {
		t.Fatalf("failed to parse audit: %v", err)
	}
	if len(entries) != 1 {
		t.Fatalf("expected 1 audit entry, got %d", len(entries))
	}
	if entries[0].Action != "key_created" {
		t.Errorf("expected key_created, got %s", entries[0].Action)
	}
}

func TestAuditEndpoint_NilAudit(t *testing.T) {
	store := NewStore()
	handler := NewHandler(store, metrics.NewCollector())

	req := httptest.NewRequest("GET", "/api/audit", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestRequireAuth_NoTokenConfigured(t *testing.T) {
	store := NewStore()
	handler := RequireAuth(NewHandler(store, metrics.NewCollector()), "")

	req := httptest.NewRequest("GET", "/api/keys", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected 200 with no token configured, got %d", rec.Code)
	}
}

func TestRequireAuth_ValidBearerToken(t *testing.T) {
	store := NewStore()
	handler := RequireAuth(NewHandler(store, metrics.NewCollector()), "secret-token")

	req := httptest.NewRequest("GET", "/api/keys", nil)
	req.Header.Set("Authorization", "Bearer secret-token")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected 200 with valid token, got %d", rec.Code)
	}
}

func TestRequireAuth_MissingToken(t *testing.T) {
	store := NewStore()
	handler := RequireAuth(NewHandler(store, metrics.NewCollector()), "secret-token")

	req := httptest.NewRequest("GET", "/api/keys", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("expected 401 without token, got %d", rec.Code)
	}
}

func TestRequireAuth_InvalidToken(t *testing.T) {
	store := NewStore()
	handler := RequireAuth(NewHandler(store, metrics.NewCollector()), "secret-token")

	req := httptest.NewRequest("GET", "/api/keys", nil)
	req.Header.Set("Authorization", "Bearer wrong-token")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("expected 401 with invalid token, got %d", rec.Code)
	}
}

func TestRequireAuth_QueryParamToken(t *testing.T) {
	store := NewStore()
	handler := RequireAuth(NewHandler(store, metrics.NewCollector()), "secret-token")

	req := httptest.NewRequest("GET", "/api/keys?token=secret-token", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected 200 with query token, got %d", rec.Code)
	}
}
