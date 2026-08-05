package main

import (
	"net/http"
	"os"
	"path/filepath"
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
			w.Write([]byte(`<html><body style="background:#0f172a;color:#fff;font-family:sans-serif;display:flex;align-items:center;justify-content:center;height:100vh"><div style="text-align:center"><h1>GatewayX Dashboard</h1><p>Dashboard not built. Run: <code>cd apps/dashboard && npm run build</code></p></div></body></html>`))
		})
	}
	return http.FileServer(http.Dir(distPath))
}
