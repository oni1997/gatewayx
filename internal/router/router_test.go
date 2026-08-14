package router

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRouter_MethodMatch(t *testing.T) {
	r := New()
	r.Add(Route{
		Methods: []string{"GET"},
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}),
	})

	req := httptest.NewRequest("GET", "/", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200 for GET, got %d", rec.Code)
	}

	req = httptest.NewRequest("POST", "/", nil)
	rec = httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	if rec.Code != http.StatusNotFound {
		t.Errorf("expected 404 for POST, got %d", rec.Code)
	}
}

func TestRouter_HostMatch(t *testing.T) {
	r := New()
	r.Add(Route{
		Hosts:   []string{"api.example.com"},
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}),
	})

	req := httptest.NewRequest("GET", "/", nil)
	req.Host = "api.example.com"
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200 for matching host, got %d", rec.Code)
	}

	req = httptest.NewRequest("GET", "/", nil)
	req.Host = "other.example.com"
	rec = httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	if rec.Code != http.StatusNotFound {
		t.Errorf("expected 404 for non-matching host, got %d", rec.Code)
	}
}

func TestRouter_PathMatch(t *testing.T) {
	r := New()
	r.Add(Route{
		Paths:   []string{"/api"},
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}),
	})

	req := httptest.NewRequest("GET", "/api/users", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200 for /api/users, got %d", rec.Code)
	}

	req = httptest.NewRequest("GET", "/other", nil)
	rec = httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	if rec.Code != http.StatusNotFound {
		t.Errorf("expected 404 for /other, got %d", rec.Code)
	}
}

func TestRouter_StripPath(t *testing.T) {
	r := New()
	var gotPath string
	r.Add(Route{
		Paths:     []string{"/api"},
		StripPath: "/api",
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			gotPath = r.URL.Path
			w.WriteHeader(http.StatusOK)
		}),
	})

	req := httptest.NewRequest("GET", "/api/users/1", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if gotPath != "/users/1" {
		t.Errorf("expected /users/1 after strip, got %s", gotPath)
	}
}

func TestMatchMethods(t *testing.T) {
	if !matchMethods(nil, "GET") {
		t.Error("nil methods should match anything")
	}
	if !matchMethods([]string{"GET"}, "get") {
		t.Error("method match should be case-insensitive")
	}
	if matchMethods([]string{"POST"}, "GET") {
		t.Error("GET should not match POST-only")
	}
}

func TestMatchHosts(t *testing.T) {
	if !matchHosts(nil, "example.com") {
		t.Error("nil hosts should match anything")
	}
	if !matchHosts([]string{"example.com"}, "example.com:8080") {
		t.Error("host match should ignore port")
	}
	if matchHosts([]string{"example.com"}, "other.com") {
		t.Error("other.com should not match")
	}
}

func TestMatchPath(t *testing.T) {
	matched, _ := matchPath([]string{"/api"}, "/api/v1", "")
	if !matched {
		t.Error("expected /api to match /api/v1")
	}
	matched, _ = matchPath([]string{"/api"}, "/other", "")
	if matched {
		t.Error("/api should not match /other")
	}
	matched, strip := matchPath(nil, "/anything", "/prefix")
	if !matched || strip != "/prefix" {
		t.Error("nil paths should match and return strip path")
	}
}
