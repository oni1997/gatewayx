package alert

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"
)

type Webhook struct {
	url     string
	client  *http.Client
	mu      sync.Mutex
	lastSent map[string]time.Time
	cooldown time.Duration
}

func NewWebhook(url string) *Webhook {
	return &Webhook{
		url:      url,
		client:   &http.Client{Timeout: 5 * time.Second},
		lastSent: make(map[string]time.Time),
		cooldown: 5 * time.Minute,
	}
}

func (w *Webhook) Send(eventType, title, message string) {
	if w.url == "" {
		return
	}

	w.mu.Lock()
	if last, ok := w.lastSent[eventType]; ok {
		if time.Since(last) < w.cooldown {
			w.mu.Unlock()
			return
		}
	}
	w.lastSent[eventType] = time.Now()
	w.mu.Unlock()

	go func() {
		payload := map[string]any{
			"text": fmt.Sprintf("[%s]\n\n%s\n\n_%s_", title, message, time.Now().Format(time.RFC3339)),
		}

		body, _ := json.Marshal(payload)
		req, _ := http.NewRequest("POST", w.url, bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")

		resp, err := w.client.Do(req)
		if err != nil {
			return
		}
		_ = resp.Body.Close()
	}()
}

func (w *Webhook) AlertRateLimit(route string, count int) {
	if count > 100 {
		w.Send("rate_limit_spike", fmt.Sprintf("Rate Limit Spike: %s", route),
			fmt.Sprintf("Route `%s` has triggered rate limiting %d times in the last window. Check for potential abuse or adjust limits.", route, count))
	}
}

func (w *Webhook) AlertSecurity(threatType, source, detail string) {
	if strings.HasPrefix(threatType, "sql_") || threatType == "xss" || threatType == "shell_injection" {
		w.Send("security_critical", fmt.Sprintf("Security Alert: %s", threatType),
			fmt.Sprintf("Source: `%s`\n%s", source, detail))
	}
}

func (w *Webhook) AlertBackendDown(route, backend string) {
	w.Send("backend_down", fmt.Sprintf("Backend Down: %s", route),
		fmt.Sprintf("Backend `%s` for route `%s` is unreachable. Circuit breaker may be open.", backend, route))
}
