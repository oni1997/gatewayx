package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/gatewayx/gatewayx/internal/config"
	"github.com/gatewayx/gatewayx/internal/health"
	"github.com/gatewayx/gatewayx/internal/logger"
	"github.com/gatewayx/gatewayx/internal/middleware"
	"github.com/gatewayx/gatewayx/internal/proxy"
)

var (
	version   = "dev"
	commit    = "none"
	buildDate = "unknown"
)

func main() {
	cfgFile := os.Getenv("GATEWAYX_CONFIG")
	if cfgFile == "" {
		cfgFile = "gatewayx.yaml"
	}

	cfg, err := config.LoadConfig(cfgFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to load config: %v\n", err)
		os.Exit(1)
	}

	log := logger.New(cfg.Logging)
	log.Info("gatewayx starting", "version", version, "commit", commit, "build_date", buildDate)

	rp, err := proxy.New(cfg, log)
	if err != nil {
		log.Error("failed to create proxy", "error", err)
		os.Exit(1)
	}

	checker := health.New()
	checker.Register("gateway", func() error { return nil })

	mux := http.NewServeMux()
	mux.Handle(cfg.Health.Path, checker.Handler())
	mux.Handle("/", rp)

	var handler http.Handler = mux
	handler = middleware.Recovery(handler)
	handler = middleware.MaxBodySize(cfg.Security.MaxBodySize)(handler)

	if cfg.Metrics.Enabled {
		log.Info("metrics enabled", "port", cfg.Metrics.Port)
		go func() {
			metricsMux := http.NewServeMux()
			metricsMux.Handle(cfg.Metrics.Path, metricsHandler())
			metricsAddr := fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Metrics.Port)
			if err := http.ListenAndServe(metricsAddr, metricsMux); err != nil {
				log.Error("metrics server error", "error", err)
			}
		}()
	}

	server := &http.Server{
		Addr:         fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port),
		Handler:      handler,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
		IdleTimeout:  cfg.Server.IdleTimeout,
	}

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigCh
		log.Info("shutting down...")
		shutdownCtx, cancel := context.WithTimeout(context.Background(), cfg.Server.ShutdownTimeout)
		defer cancel()
		if err := server.Shutdown(shutdownCtx); err != nil {
			log.Error("graceful shutdown failed", "error", err)
		}
		}()

	log.Info("gateway listening", "addr", server.Addr)

	var serveErr error
	if cfg.TLS.Enabled {
		serveErr = server.ListenAndServeTLS(cfg.TLS.CertFile, cfg.TLS.KeyFile)
	} else {
		serveErr = server.ListenAndServe()
	}

	if serveErr != nil && serveErr != http.ErrServerClosed {
		log.Error("server error", "error", serveErr)
		os.Exit(1)
	}

	log.Info("gateway stopped")
}

func metricsHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain; version=0.0.4")
		w.Write([]byte("# GatewayX Metrics\n"))
	})
}
