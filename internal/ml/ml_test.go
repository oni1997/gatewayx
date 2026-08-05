package ml

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/oni1997/gatewayx/internal/history"
	"github.com/oni1997/gatewayx/internal/metrics"
)

func TestSecurityScanner_CleanTraffic(t *testing.T) {
	s := NewSecurityScanner()
	entries := []history.Entry{
		{Method: "GET", Path: "/api/users", Status: 200, RemoteAddr: "10.0.0.1"},
		{Method: "GET", Path: "/api/products", Status: 200, RemoteAddr: "10.0.0.2"},
	}

	report := s.Scan(entries)
	if report.TotalAlerts != 0 {
		t.Errorf("expected 0 alerts for clean traffic, got %d", report.TotalAlerts)
	}
	if report.Summary != "No threats detected. Traffic appears normal." {
		t.Errorf("unexpected summary: %s", report.Summary)
	}
}

func TestSecurityScanner_SQLInjection(t *testing.T) {
	s := NewSecurityScanner()
	entries := []history.Entry{
		{Method: "GET", Path: "/api/users?id=1 OR 1=1", Status: 200, RemoteAddr: "10.0.0.5"},
	}

	report := s.Scan(entries)
	if report.TotalAlerts == 0 {
		t.Fatal("expected alerts for SQL injection")
	}

	found := false
	for _, threat := range report.Threats {
		if threat.Type == "sql_injection" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected sql_injection threat in report")
	}
}

func TestSecurityScanner_XSS(t *testing.T) {
	s := NewSecurityScanner()
	entries := []history.Entry{
		{Method: "GET", Path: "/search?q=<script>alert(1)</script>", Status: 200, RemoteAddr: "10.0.0.6"},
	}

	report := s.Scan(entries)
	found := false
	for _, threat := range report.Threats {
		if threat.Type == "xss" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected xss threat in report")
	}
}

func TestSecurityScanner_BruteForce(t *testing.T) {
	s := NewSecurityScanner()
	entries := make([]history.Entry, 15)
	for i := range entries {
		entries[i] = history.Entry{
			Method: "POST", Path: "/login",
			Status: 401, RemoteAddr: "10.0.0.99",
		}
	}

	report := s.Scan(entries)
	found := false
	for _, threat := range report.Threats {
		if threat.Type == "auth_failure" {
			found = true
			if threat.Severity != "medium" {
				t.Error("expected medium severity")
			}
			break
		}
	}
	if !found {
		t.Error("expected auth_failure threat")
	}
}

func TestSecurityScanner_HighVolume(t *testing.T) {
	s := NewSecurityScanner()
	entries := make([]history.Entry, 150)
	for i := range entries {
		entries[i] = history.Entry{
			Method: "GET", Path: "/api/data",
			Status: 200, RemoteAddr: "10.0.0.200",
		}
	}

	report := s.Scan(entries)
	found := false
	for _, threat := range report.Threats {
		if threat.Type == "high_volume" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected high_volume threat")
	}
}

func TestBottleneckAnalysis_SlowRoutes(t *testing.T) {
	c := metrics.NewCollector()

	c.RecordRequest("/api/fast", "GET", 200, 5*time.Millisecond, 0, 100)
	c.RecordRequest("/api/slow", "GET", 200, 600*time.Millisecond, 0, 200)
	c.RecordRequest("/api/slow", "GET", 200, 550*time.Millisecond, 0, 200)

	report := AnalyzeBottlenecks(c)
	if len(report.SlowRoutes) != 1 {
		t.Fatalf("expected 1 slow route, got %d", len(report.SlowRoutes))
	}
	if report.SlowRoutes[0].Name != "/api/slow" {
		t.Errorf("expected /api/slow, got %s", report.SlowRoutes[0].Name)
	}
	if report.SlowRoutes[0].Severity != "critical" {
		t.Errorf("expected critical severity, got %s", report.SlowRoutes[0].Severity)
	}
}

func TestBottleneckAnalysis_NoIssues(t *testing.T) {
	c := metrics.NewCollector()
	c.RecordRequest("/api/fast", "GET", 200, 5*time.Millisecond, 0, 100)

	report := AnalyzeBottlenecks(c)
	if len(report.SlowRoutes) != 0 {
		t.Errorf("expected 0 slow routes, got %d", len(report.SlowRoutes))
	}
	if report.Summary != "No bottlenecks detected. All routes performing normally." {
		t.Errorf("unexpected summary: %s", report.Summary)
	}
}

func TestRecommendations(t *testing.T) {
	c := metrics.NewCollector()

	for i := 0; i < 50; i++ {
		c.RecordRequest("/api/busy", "GET", 200, 80*time.Millisecond, 0, 100)
	}

	report := GenerateRecommendations(c, 2*time.Second)

	if len(report.RateLimits) == 0 {
		t.Error("expected rate limit suggestions for busy route")
	}
	if len(report.CacheHints) == 0 {
		t.Error("expected cache hint for slow route")
	}
}

func TestRecommendations_LowTraffic(t *testing.T) {
	c := metrics.NewCollector()
	c.RecordRequest("/api/quiet", "GET", 200, 10*time.Millisecond, 0, 100)

	report := GenerateRecommendations(c, time.Hour)
	if report.Summary != "No recommendations at this time. Traffic patterns are within normal range." {
		t.Errorf("unexpected summary: %s", report.Summary)
	}
}

func TestAPIHandlers(t *testing.T) {
	c := metrics.NewCollector()
	h := history.NewBuffer(10)

	c.RecordRequest("/api/test", "GET", 200, 10*time.Millisecond, 0, 100)
	h.Push(history.Entry{
		Method: "GET", Path: "/api/test", Status: 200,
		RemoteAddr: "10.0.0.1", Timestamp: time.Now(),
		Duration: 10 * time.Millisecond,
	})

	svc := NewAnalysisService(c, h)

	tests := []struct {
		path string
	}{
		{"/security"},
		{"/bottlenecks"},
		{"/recommendations"},
		{"/analysis"},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			var handler http.Handler
			switch tt.path {
			case "/security":
				handler = svc.SecurityHandler()
			case "/bottlenecks":
				handler = svc.BottlenecksHandler()
			case "/recommendations":
				handler = svc.RecommendationsHandler()
			case "/analysis":
				handler = svc.FullReportHandler()
			}

			req := httptest.NewRequest("GET", tt.path, nil)
			rec := httptest.NewRecorder()
			handler.ServeHTTP(rec, req)

			if rec.Code != 200 {
				t.Errorf("expected 200, got %d", rec.Code)
			}
			if rec.Header().Get("Content-Type") != "application/json" {
				t.Error("expected application/json content type")
			}

			var data map[string]any
			if err := json.Unmarshal(rec.Body.Bytes(), &data); err != nil {
				t.Errorf("failed to parse JSON from %s: %v", tt.path, err)
			}
		})
	}
}
