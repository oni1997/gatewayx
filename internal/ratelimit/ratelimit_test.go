package ratelimit

import (
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/oni1997/gatewayx/internal/auth"
)

func TestTokenBucket_Allow(t *testing.T) {
	tb := NewTokenBucket(10, 20)

	for i := 0; i < 20; i++ {
		if !tb.Allow() {
			t.Fatalf("token %d should be allowed", i)
		}
	}

	if tb.Allow() {
		t.Error("should be rate limited after burst exhausted")
	}
}

func TestTokenBucket_Refill(t *testing.T) {
	tb := NewTokenBucket(100, 10)

	for i := 0; i < 10; i++ {
		tb.Allow()
	}

	if tb.Allow() {
		t.Error("should be denied immediately after exhausting burst")
	}

	time.Sleep(50 * time.Millisecond)

	if !tb.Allow() {
		t.Error("should be allowed after refill")
	}
}

func TestTokenBucket_AllowN(t *testing.T) {
	tb := NewTokenBucket(10, 50)

	if !tb.AllowN(30) {
		t.Error("should allow 30 tokens")
	}

	if !tb.AllowN(20) {
		t.Error("should allow remaining 20 tokens")
	}

	if tb.AllowN(1) {
		t.Error("should deny when empty")
	}
}

func TestTokenBucket_Concurrent(t *testing.T) {
	tb := NewTokenBucket(1000, 100)
	var wg sync.WaitGroup
	allowed := make(chan bool, 100)

	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			allowed <- tb.Allow()
		}()
	}
	wg.Wait()
	close(allowed)

	count := 0
	for a := range allowed {
		if a {
			count++
		}
	}
	if count != 100 {
		t.Errorf("expected 100 allowed, got %d", count)
	}
}

func TestSlidingWindow_Allow(t *testing.T) {
	sw := NewSlidingWindow(5, 0)

	for i := 0; i < 5; i++ {
		if !sw.Allow() {
			t.Fatalf("request %d should be allowed", i)
		}
	}

	if sw.Allow() {
		t.Error("should deny when limit exceeded")
	}
}

func TestSlidingWindow_AllowN(t *testing.T) {
	sw := NewSlidingWindow(10, 0)

	if !sw.AllowN(7) {
		t.Error("should allow 7")
	}

	if !sw.AllowN(3) {
		t.Error("should allow 3")
	}

	if sw.AllowN(1) {
		t.Error("should deny when full")
	}
}

func TestMemoryStore_Basic(t *testing.T) {
	cfg := Config{
		Rate:     10,
		Burst:    20,
		Strategy: "token_bucket",
	}
	store := NewMemoryStore(cfg)

	key := "test-user"
	for i := 0; i < 20; i++ {
		if !store.Allow(key) {
			t.Fatalf("request %d should be allowed", i)
		}
	}

	if store.Allow(key) {
		t.Error("should be rate limited")
	}
}

func TestMemoryStore_SlidingWindow(t *testing.T) {
	cfg := Config{
		Rate:     5,
		Burst:    0,
		Strategy: "sliding_window",
	}
	store := NewMemoryStore(cfg)

	key := "test-ip"
	for i := 0; i < 5; i++ {
		if !store.Allow(key) {
			t.Fatalf("request %d should be allowed", i)
		}
	}

	if store.Allow(key) {
		t.Error("should be rate limited")
	}
}

func TestMemoryStore_SeparateKeys(t *testing.T) {
	cfg := Config{
		Rate:  5,
		Burst: 5,
	}
	store := NewMemoryStore(cfg)

	for i := 0; i < 5; i++ {
		store.Allow("user-a")
	}

	if !store.Allow("user-b") {
		t.Error("separate keys should have separate limits")
	}
}

func TestMemoryStore_AllowN(t *testing.T) {
	cfg := Config{
		Rate:  20,
		Burst: 50,
	}
	store := NewMemoryStore(cfg)

	if !store.AllowN("key", 40) {
		t.Error("should allow 40")
	}

	if !store.AllowN("key", 10) {
		t.Error("should allow remaining 10")
	}

	if store.AllowN("key", 1) {
		t.Error("should deny when empty")
	}
}

func TestMiddleware_BlocksExceeded(t *testing.T) {
	cfg := Config{
		Rate:      100,
		Burst:     3,
		Strategy:  "token_bucket",
		RouteName: "test-route",
	}
	store := NewMemoryStore(cfg)
	mw := NewMiddleware(store, cfg)

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	for i := 0; i < 3; i++ {
		req := httptest.NewRequest("GET", "/api/test", nil)
		rec := httptest.NewRecorder()
		mw.Handler(handler).ServeHTTP(rec, req)
		if rec.Code != http.StatusOK {
			t.Fatalf("request %d should pass, got %d", i, rec.Code)
		}
	}

	req := httptest.NewRequest("GET", "/api/test", nil)
	rec := httptest.NewRecorder()
	mw.Handler(handler).ServeHTTP(rec, req)

	if rec.Code != http.StatusTooManyRequests {
		t.Errorf("expected 429, got %d", rec.Code)
	}
}

func TestMiddleware_PerIP(t *testing.T) {
	cfg := Config{
		Rate:      100,
		Burst:     2,
		Strategy:  "token_bucket",
		PerIP:     true,
		RouteName: "ip-test",
	}
	store := NewMemoryStore(cfg)
	mw := NewMiddleware(store, cfg)

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	req1 := httptest.NewRequest("GET", "/api/test", nil)
	req1.RemoteAddr = "10.0.0.1:12345"

	req2 := httptest.NewRequest("GET", "/api/test", nil)
	req2.RemoteAddr = "10.0.0.2:12345"

	for i := 0; i < 2; i++ {
		rec := httptest.NewRecorder()
		mw.Handler(handler).ServeHTTP(rec, req1)
		if rec.Code != http.StatusOK {
			t.Fatalf("ip1 req %d should pass", i)
		}
	}

	rec := httptest.NewRecorder()
	mw.Handler(handler).ServeHTTP(rec, req2)
	if rec.Code != http.StatusOK {
		t.Error("different IP should not be blocked")
	}
}

func TestMiddleware_PerUser(t *testing.T) {
	cfg := Config{
		Rate:      100,
		Burst:     2,
		Strategy:  "token_bucket",
		PerUser:   true,
		RouteName: "user-test",
	}
	store := NewMemoryStore(cfg)
	mw := NewMiddleware(store, cfg)

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest("GET", "/api/test", nil)
	req = req.WithContext(auth.WithClaims(req.Context(), auth.Claims{"sub": "user-42"}))

	for i := 0; i < 2; i++ {
		rec := httptest.NewRecorder()
		mw.Handler(handler).ServeHTTP(rec, req)
		if rec.Code != http.StatusOK {
			t.Fatalf("user req %d should pass", i)
		}
	}

	rec := httptest.NewRecorder()
	mw.Handler(handler).ServeHTTP(rec, req)
	if rec.Code != http.StatusTooManyRequests {
		t.Errorf("expected 429, got %d", rec.Code)
	}
}

func TestMiddleware_RetryAfterHeader(t *testing.T) {
	cfg := Config{
		Rate:      100,
		Burst:     0,
		Strategy:  "token_bucket",
		RouteName: "retry-test",
	}
	store := NewMemoryStore(cfg)
	mw := NewMiddleware(store, cfg)

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest("GET", "/api/test", nil)
	rec := httptest.NewRecorder()
	mw.Handler(handler).ServeHTTP(rec, req)

	if retryAfter := rec.Header().Get("Retry-After"); retryAfter != "1" {
		t.Errorf("expected Retry-After=1, got %s", retryAfter)
	}

	if rateLimit := rec.Header().Get("X-RateLimit-Limit"); rateLimit != "100" {
		t.Errorf("expected X-RateLimit-Limit=100, got %s", rateLimit)
	}
}

func TestExtractIP(t *testing.T) {
	tests := []struct {
		name       string
		remoteAddr string
		headers    map[string]string
		want       string
	}{
		{
			name:       "direct connection",
			remoteAddr: "192.168.1.1:54321",
			want:       "192.168.1.1",
		},
		{
			name:       "forwarded for",
			remoteAddr: "10.0.0.1:12345",
			headers:    map[string]string{"X-Forwarded-For": "203.0.113.1, 10.0.0.1"},
			want:       "203.0.113.1",
		},
		{
			name:       "real ip",
			remoteAddr: "10.0.0.1:12345",
			headers:    map[string]string{"X-Real-IP": "198.51.100.1"},
			want:       "198.51.100.1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/", nil)
			req.RemoteAddr = tt.remoteAddr
			for k, v := range tt.headers {
				req.Header.Set(k, v)
			}
			got := extractIP(req)
			if got != tt.want {
				t.Errorf("extractIP() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestConfigDefaults(t *testing.T) {
	cfg := DefaultConfig()
	if cfg.Strategy != "token_bucket" {
		t.Errorf("expected token_bucket, got %s", cfg.Strategy)
	}
	if cfg.RedisPrefix != "gatewayx:ratelimit" {
		t.Errorf("expected gatewayx:ratelimit prefix, got %s", cfg.RedisPrefix)
	}
}
