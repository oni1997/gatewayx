package cache

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestCache_GetSet(t *testing.T) {
	c := New(time.Minute, 100)

	req := httptest.NewRequest("GET", "/api/test", nil)
	c.Set(req, 200, http.Header{"Content-Type": []string{"application/json"}}, []byte(`{"ok":true}`))

	e, ok := c.Get(req)
	if !ok {
		t.Fatal("expected cache hit")
	}
	if e.status != 200 {
		t.Errorf("expected status 200, got %d", e.status)
	}
	if string(e.body) != `{"ok":true}` {
		t.Errorf("unexpected body: %s", e.body)
	}
}

func TestCache_Expiry(t *testing.T) {
	c := New(10*time.Millisecond, 100)

	req := httptest.NewRequest("GET", "/api/test", nil)
	c.Set(req, 200, nil, []byte("data"))

	time.Sleep(15 * time.Millisecond)

	if _, ok := c.Get(req); ok {
		t.Error("expected cache miss after expiry")
	}
}

func TestCache_MethodFilter(t *testing.T) {
	c := New(time.Minute, 100)

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		_, _ = w.Write([]byte("ok"))
	})
	wrapped := Middleware(c)(handler)

	getReq := httptest.NewRequest("GET", "/api/test", nil)
	rec := httptest.NewRecorder()
	wrapped.ServeHTTP(rec, getReq)
	if rec.Header().Get("X-Cache") != "MISS" {
		t.Errorf("first request should be MISS, got %s", rec.Header().Get("X-Cache"))
	}

	rec2 := httptest.NewRecorder()
	wrapped.ServeHTTP(rec2, getReq)
	if rec2.Header().Get("X-Cache") != "HIT" {
		t.Errorf("second request should be HIT, got %s", rec2.Header().Get("X-Cache"))
	}
}

func TestCache_SkipsPOST(t *testing.T) {
	c := New(time.Minute, 100)

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
	})
	wrapped := Middleware(c)(handler)

	postReq := httptest.NewRequest("POST", "/api/test", nil)
	rec := httptest.NewRecorder()
	wrapped.ServeHTTP(rec, postReq)

	if rec.Header().Get("X-Cache") != "" {
		t.Error("POST should bypass cache")
	}
}

func TestCache_Invalidate(t *testing.T) {
	c := New(time.Minute, 100)
	req := httptest.NewRequest("GET", "/api/test", nil)
	c.Set(req, 200, nil, []byte("data"))

	c.Invalidate()
	if c.Size() != 0 {
		t.Errorf("expected empty cache after invalidate, got %d", c.Size())
	}
}

func TestCache_MaxSize(t *testing.T) {
	c := New(time.Minute, 2)

	for i := 0; i < 5; i++ {
		req := httptest.NewRequest("GET", "/api/test?i="+string(rune('a'+i)), nil)
		c.Set(req, 200, nil, []byte("data"))
	}

	if c.Size() > 2 {
		t.Errorf("expected cache size capped at 2, got %d", c.Size())
	}
}
