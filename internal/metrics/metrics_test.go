package metrics

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestCollector_RecordRequest(t *testing.T) {
	c := NewCollector()

	c.RecordRequest("/api/test", "GET", 200, 5*time.Millisecond, 100, 200)

	if c.Requests.Load() != 1 {
		t.Error("expected 1 request")
	}
	if c.BytesSent.Load() != 200 {
		t.Error("expected 200 bytes sent")
	}
	if c.BytesReceived.Load() != 100 {
		t.Error("expected 100 bytes received")
	}

	snapshot := c.Snapshot()
	if len(snapshot.Routes) != 1 {
		t.Fatalf("expected 1 route, got %d", len(snapshot.Routes))
	}
	if snapshot.Routes[0].Count != 1 {
		t.Error("expected count 1")
	}
}

func TestCollector_MultipleRequests(t *testing.T) {
	c := NewCollector()

	c.RecordRequest("/api/test", "GET", 200, 10*time.Millisecond, 0, 500)
	c.RecordRequest("/api/test", "GET", 200, 5*time.Millisecond, 0, 300)
	c.RecordRequest("/api/test", "POST", 400, 20*time.Millisecond, 0, 100)

	if c.Requests.Load() != 3 {
		t.Error("expected 3 requests")
	}

	snapshot := c.Snapshot()
	if len(snapshot.Routes) != 1 {
		t.Fatalf("expected 1 route, got %d", len(snapshot.Routes))
	}
	if snapshot.Routes[0].Count != 3 {
		t.Errorf("expected count 3, got %d", snapshot.Routes[0].Count)
	}
}

func TestCollector_ActiveConnections(t *testing.T) {
	c := NewCollector()
	c.ActiveConns.Add(1)
	c.ActiveConns.Add(1)
	c.ActiveConns.Add(-1)

	if c.ActiveConns.Load() != 1 {
		t.Errorf("expected 1 active conn, got %d", c.ActiveConns.Load())
	}
}

func TestMetricsMiddleware(t *testing.T) {
	c := NewCollector()
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("hello"))
	})
	wrapped := Middleware(c)(handler)

	req := httptest.NewRequest("GET", "/api/test", nil)
	rec := httptest.NewRecorder()

	wrapped.ServeHTTP(rec, req)

	if c.Requests.Load() != 1 {
		t.Error("expected 1 request recorded")
	}
	if c.BytesSent.Load() < 5 {
		t.Error("expected bytes sent to be recorded")
	}
}

func TestExporter_JSON(t *testing.T) {
	c := NewCollector()
	c.RecordRequest("/api/test", "GET", 200, time.Millisecond, 0, 100)

	req := httptest.NewRequest("GET", "/metrics", nil)
	req.Header.Set("Accept", "application/json")
	rec := httptest.NewRecorder()

	Exporter(c).ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
	if rec.Header().Get("Content-Type") != "application/json" {
		t.Error("expected application/json")
	}
}

func TestExporter_Prometheus(t *testing.T) {
	c := NewCollector()
	c.RecordRequest("/api/test", "GET", 200, time.Millisecond, 0, 100)

	req := httptest.NewRequest("GET", "/metrics", nil)
	rec := httptest.NewRecorder()

	Exporter(c).ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
	body := rec.Body.String()
	if body == "" {
		t.Error("expected body content")
	}
}
