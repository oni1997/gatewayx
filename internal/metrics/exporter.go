package metrics

import (
	"encoding/json"
	"fmt"
	"net/http"
	"runtime"
	"time"
)

func Exporter(collector *Collector) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		snapshot := collector.Snapshot()

		accept := r.Header.Get("Accept")
		if accept == "application/json" {
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(snapshot)
			return
		}

		w.Header().Set("Content-Type", "text/plain; version=0.0.4")

		_, _ = fmt.Fprintf(w, "# HELP gatewayx_requests_total Total number of HTTP requests processed\n")
		_, _ = fmt.Fprintf(w, "# TYPE gatewayx_requests_total counter\n")
		_, _ = fmt.Fprintf(w, "gatewayx_requests_total %d\n", snapshot.Requests)

		_, _ = fmt.Fprintf(w, "# HELP gatewayx_active_connections Current number of active connections\n")
		_, _ = fmt.Fprintf(w, "# TYPE gatewayx_active_connections gauge\n")
		_, _ = fmt.Fprintf(w, "gatewayx_active_connections %d\n", snapshot.ActiveConns)

		_, _ = fmt.Fprintf(w, "# HELP gatewayx_bytes_sent_total Total bytes sent\n")
		_, _ = fmt.Fprintf(w, "# TYPE gatewayx_bytes_sent_total counter\n")
		_, _ = fmt.Fprintf(w, "gatewayx_bytes_sent_total %d\n", snapshot.BytesSent)

		_, _ = fmt.Fprintf(w, "# HELP gatewayx_bytes_received_total Total bytes received\n")
		_, _ = fmt.Fprintf(w, "# TYPE gatewayx_bytes_received_total counter\n")
		_, _ = fmt.Fprintf(w, "gatewayx_bytes_received_total %d\n", snapshot.BytesReceived)

		_, _ = fmt.Fprintf(w, "# HELP gatewayx_uptime_seconds Gateway uptime in seconds\n")
		_, _ = fmt.Fprintf(w, "# TYPE gatewayx_uptime_seconds gauge\n")
		_, _ = fmt.Fprintf(w, "gatewayx_uptime_seconds %.0f\n", time.Since(snapshot.Uptime).Seconds())

		for _, route := range snapshot.Routes {
			label := fmt.Sprintf(`route="%s"`, route.Name)
			_, _ = fmt.Fprintf(w, "# HELP gatewayx_route_requests_total Route request count\n")
			_, _ = fmt.Fprintf(w, "# TYPE gatewayx_route_requests_total counter\n")
			_, _ = fmt.Fprintf(w, "gatewayx_route_requests_total{%s} %d\n", label, route.Count)

			_, _ = fmt.Fprintf(w, "# HELP gatewayx_route_latency_ms Route latency\n")
			_, _ = fmt.Fprintf(w, "# TYPE gatewayx_route_latency_ms gauge\n")
			_, _ = fmt.Fprintf(w, "gatewayx_route_latency_avg_ms{%s} %.2f\n", label, route.AvgMs)
			_, _ = fmt.Fprintf(w, "gatewayx_route_latency_min_ms{%s} %.2f\n", label, route.MinMs)
			_, _ = fmt.Fprintf(w, "gatewayx_route_latency_max_ms{%s} %.2f\n", label, route.MaxMs)
		}

		var mem runtime.MemStats
		runtime.ReadMemStats(&mem)
		_, _ = fmt.Fprintf(w, "# HELP gatewayx_memory_alloc_bytes Memory allocated\n")
		_, _ = fmt.Fprintf(w, "# TYPE gatewayx_memory_alloc_bytes gauge\n")
		_, _ = fmt.Fprintf(w, "gatewayx_memory_alloc_bytes %d\n", mem.Alloc)

		_, _ = fmt.Fprintf(w, "# HELP go_goroutines Number of goroutines\n")
		_, _ = fmt.Fprintf(w, "# TYPE go_goroutines gauge\n")
		_, _ = fmt.Fprintf(w, "go_goroutines %d\n", runtime.NumGoroutine())
	})
}
