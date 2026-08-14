package compression

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestGzipMiddleware_Compresses(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(strings.Repeat("hello world ", 100)))
	})

	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("Accept-Encoding", "gzip")
	rec := httptest.NewRecorder()

	GzipMiddleware(handler).ServeHTTP(rec, req)

	if rec.Header().Get("Content-Encoding") != "gzip" {
		t.Errorf("expected gzip encoding, got %s", rec.Header().Get("Content-Encoding"))
	}
}

func TestGzipMiddleware_SkipsWithoutHeader(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("plain"))
	})

	req := httptest.NewRequest("GET", "/", nil)
	rec := httptest.NewRecorder()

	GzipMiddleware(handler).ServeHTTP(rec, req)

	if rec.Header().Get("Content-Encoding") == "gzip" {
		t.Error("should not compress without Accept-Encoding header")
	}
	if rec.Body.String() != "plain" {
		t.Errorf("expected plain body, got %s", rec.Body.String())
	}
}
