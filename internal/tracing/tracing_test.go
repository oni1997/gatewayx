package tracing

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestTracer_GeneratesTraceID(t *testing.T) {
	tracer := New(true, nil)

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		traceID := GetTraceID(r.Context())
		if traceID == "" {
			t.Error("expected trace ID in context")
		}
		spanID := GetSpanID(r.Context())
		if spanID == "" {
			t.Error("expected span ID in context")
		}
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest("GET", "/api/test", nil)
	rec := httptest.NewRecorder()

	tracer.Middleware(handler).ServeHTTP(rec, req)

	if rec.Header().Get("X-Trace-ID") == "" {
		t.Error("expected X-Trace-ID header")
	}
	if rec.Header().Get("X-Span-ID") == "" {
		t.Error("expected X-Span-ID header")
	}
}

func TestTracer_PropagatesExistingTraceID(t *testing.T) {
	tracer := New(true, nil)

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		traceID := GetTraceID(r.Context())
		if traceID != "my-trace-id" {
			t.Errorf("expected my-trace-id, got %s", traceID)
		}
	})

	req := httptest.NewRequest("GET", "/api/test", nil)
	req.Header.Set("X-Trace-ID", "my-trace-id")
	req.Header.Set("X-Span-ID", "parent-span-id")
	rec := httptest.NewRecorder()

	tracer.Middleware(handler).ServeHTTP(rec, req)

	if rec.Header().Get("X-Parent-Span-ID") != "parent-span-id" {
		t.Error("expected X-Parent-Span-ID header")
	}
}

func TestTracer_Disabled(t *testing.T) {
	tracer := New(false, nil)

	var traceID string
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		traceID = GetTraceID(r.Context())
	})

	req := httptest.NewRequest("GET", "/api/test", nil)
	rec := httptest.NewRecorder()

	tracer.Middleware(handler).ServeHTTP(rec, req)

	if traceID != "" {
		t.Error("trace ID should be empty when tracing is disabled")
	}
}

func TestTracer_W3CTraceparent(t *testing.T) {
	tracer := New(true, nil)

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		traceID := GetTraceID(r.Context())
		if traceID == "" {
			t.Error("expected trace ID from traceparent")
		}
	})

	req := httptest.NewRequest("GET", "/api/test", nil)
	req.Header.Set("traceparent", "00-0af7651916cd43dd8448eb211c80319c-b7ad6b7169203331-01")
	rec := httptest.NewRecorder()

	tracer.Middleware(handler).ServeHTTP(rec, req)
}

func TestGetTraceID_NoContext(t *testing.T) {
	if GetTraceID(context.Background()) != "" {
		t.Error("expected empty trace ID in background context")
	}
}

func TestGetSpanID_NoContext(t *testing.T) {
	if GetSpanID(context.Background()) != "" {
		t.Error("expected empty span ID in background context")
	}
}
