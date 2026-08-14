package alert

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestNewWebhook_EmptyURL(t *testing.T) {
	w := NewWebhook("")
	w.Send("test", "title", "message")
}

func TestWebhook_Send(t *testing.T) {
	received := make(chan string, 1)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		received <- "got-it"
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	w := NewWebhook(server.URL)
	w.Send("test_event", "Test Title", "Test Message")

	select {
	case <-received:
		// success
	case <-time.After(2 * time.Second):
		t.Error("expected webhook to be called")
	}
}

func TestWebhook_Cooldown(t *testing.T) {
	calls := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	w := NewWebhook(server.URL)

	w.Send("dup_event", "title", "msg")
	time.Sleep(100 * time.Millisecond)
	w.Send("dup_event", "title", "msg")

	time.Sleep(200 * time.Millisecond)

	if calls != 1 {
		t.Errorf("expected 1 call due to cooldown, got %d", calls)
	}
}

func TestAlertRateLimit_BelowThreshold(t *testing.T) {
	w := NewWebhook("")
	w.AlertRateLimit("/api", 50)
}

func TestAlertSecurity_NonCritical(t *testing.T) {
	w := NewWebhook("")
	w.AlertSecurity("path_traversal", "10.0.0.1", "detail")
}

func TestAlertBackendDown(t *testing.T) {
	w := NewWebhook("")
	w.AlertBackendDown("/api", "backend-1")
}
