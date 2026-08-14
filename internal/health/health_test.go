package health

import (
	"encoding/json"
	"errors"
	"net/http/httptest"
	"testing"
)

func TestChecker_Healthy(t *testing.T) {
	c := New()
	c.Register("gateway", func() error { return nil })

	req := httptest.NewRequest("GET", "/health", nil)
	rec := httptest.NewRecorder()

	c.Handler().ServeHTTP(rec, req)

	if rec.Code != 200 {
		t.Errorf("expected 200, got %d", rec.Code)
	}

	var report Report
	if err := json.Unmarshal(rec.Body.Bytes(), &report); err != nil {
		t.Fatalf("failed to parse: %v", err)
	}
	if report.Status != StatusHealthy {
		t.Errorf("expected healthy, got %s", report.Status)
	}
	if report.Checks["gateway"] != "healthy" {
		t.Errorf("expected gateway healthy, got %s", report.Checks["gateway"])
	}
}

func TestChecker_Unhealthy(t *testing.T) {
	c := New()
	c.Register("db", func() error { return errors.New("connection refused") })

	req := httptest.NewRequest("GET", "/health", nil)
	rec := httptest.NewRecorder()

	c.Handler().ServeHTTP(rec, req)

	if rec.Code != 503 {
		t.Errorf("expected 503, got %d", rec.Code)
	}

	var report Report
	_ = json.Unmarshal(rec.Body.Bytes(), &report)
	if report.Status != StatusUnhealthy {
		t.Errorf("expected unhealthy, got %s", report.Status)
	}
}

func TestChecker_MultipleChecks(t *testing.T) {
	c := New()
	c.Register("a", func() error { return nil })
	c.Register("b", func() error { return errors.New("down") })

	req := httptest.NewRequest("GET", "/health", nil)
	rec := httptest.NewRecorder()

	c.Handler().ServeHTTP(rec, req)

	var report Report
	_ = json.Unmarshal(rec.Body.Bytes(), &report)
	if len(report.Checks) != 2 {
		t.Errorf("expected 2 checks, got %d", len(report.Checks))
	}
}
