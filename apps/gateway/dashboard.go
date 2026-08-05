package main

import (
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

var dashboardPath = "/opt/gatewayx/dashboard"

func init() {
	if p := os.Getenv("GATEWAYX_DASHBOARD_PATH"); p != "" {
		dashboardPath = p
	}
}

func dashboardHandler() http.Handler {
	distPath := filepath.Join(dashboardPath, "dist")
	if _, err := os.Stat(distPath); os.IsNotExist(err) {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "text/html")
			_, _ = w.Write([]byte(`<html><body style="background:#0f172a;color:#fff;font-family:sans-serif;display:flex;align-items:center;justify-content:center;height:100vh"><div style="text-align:center"><h1>GatewayX Dashboard</h1><p>Dashboard not built. Run: <code>cd apps/dashboard && npm run build</code></p></div></body></html>`))
		})
	}

	fs := http.FileServer(http.Dir(distPath))
	indexPath := filepath.Join(distPath, "index.html")

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := filepath.Join(distPath, filepath.Clean(r.URL.Path))
		if _, err := os.Stat(path); os.IsNotExist(err) || strings.HasSuffix(path, "/") && !strings.HasSuffix(r.URL.Path, ".html") {
			if strings.HasPrefix(r.URL.Path, "/assets/") || strings.Contains(r.URL.Path, ".") {
				fs.ServeHTTP(w, r)
				return
			}
			http.ServeFile(w, r, indexPath)
			return
		}
		fs.ServeHTTP(w, r)
	})
}
