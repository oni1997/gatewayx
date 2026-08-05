package metrics

import (
	"net/http"
	"time"
)

func Middleware(collector *Collector) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			collector.ActiveConns.Add(1)
			defer collector.ActiveConns.Add(-1)

			start := time.Now()

			sw := &statusWriter{ResponseWriter: w, status: http.StatusOK, bytesWritten: 0}
			next.ServeHTTP(sw, r)

			route := r.URL.Path
			if route == "" {
				route = "/"
			}

			bytesRead := r.ContentLength
			if bytesRead < 0 {
				bytesRead = 0
			}

			collector.RecordRequest(route, r.Method, sw.status, time.Since(start), bytesRead, sw.bytesWritten)
		})
	}
}

type statusWriter struct {
	http.ResponseWriter
	status       int
	bytesWritten int64
}

func (sw *statusWriter) WriteHeader(code int) {
	sw.status = code
	sw.ResponseWriter.WriteHeader(code)
}

func (sw *statusWriter) Write(b []byte) (int, error) {
	n, err := sw.ResponseWriter.Write(b)
	sw.bytesWritten += int64(n)
	return n, err
}
