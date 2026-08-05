package tracing

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"net/http"
	"time"

	"github.com/oni1997/gatewayx/internal/history"
)

type contextKey string

const (
	traceIDKey contextKey = "trace-id"
	spanIDKey  contextKey = "span-id"
)

type Tracer struct {
	enabled bool
	history *history.Buffer
}

func New(enabled bool, history *history.Buffer) *Tracer {
	return &Tracer{enabled: enabled, history: history}
}

func (t *Tracer) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !t.enabled {
			next.ServeHTTP(w, r)
			return
		}

		traceID := r.Header.Get("X-Trace-ID")
		if traceID == "" {
			traceID = t.extractOrGenerate(r, "traceparent")
		}
		if traceID == "" {
			traceID = generateID(16)
		}
		spanID := generateID(8)

		parentSpanID := r.Header.Get("X-Span-ID")

		ctx := context.WithValue(r.Context(), traceIDKey, traceID)
		ctx = context.WithValue(ctx, spanIDKey, spanID)
		r = r.WithContext(ctx)

		w.Header().Set("X-Trace-ID", traceID)
		w.Header().Set("X-Span-ID", spanID)
		if parentSpanID != "" {
			w.Header().Set("X-Parent-Span-ID", parentSpanID)
		}

		start := time.Now()
		next.ServeHTTP(w, r)
		duration := time.Since(start)

		if t.history != nil {
			t.history.Push(history.Entry{
				Timestamp: time.Now(),
				TraceID:   traceID,
				SpanID:    spanID,
				Method:    r.Method,
				Path:      r.URL.Path,
				Host:      r.Host,
				Duration:  duration,
			})
		}
	})
}

func (t *Tracer) extractOrGenerate(r *http.Request, header string) string {
	val := r.Header.Get(header)
	if len(val) >= 55 {
		return val[3:35]
	}
	return ""
}

func generateID(length int) string {
	b := make([]byte, length)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}

func GetTraceID(ctx context.Context) string {
	if v, ok := ctx.Value(traceIDKey).(string); ok {
		return v
	}
	return ""
}

func GetSpanID(ctx context.Context) string {
	if v, ok := ctx.Value(spanIDKey).(string); ok {
		return v
	}
	return ""
}
