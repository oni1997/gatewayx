package server

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/oni1997/gatewayx/internal/admin"
	"github.com/oni1997/gatewayx/internal/alert"
	"github.com/oni1997/gatewayx/internal/auth"
	"github.com/oni1997/gatewayx/internal/config"
	"github.com/oni1997/gatewayx/internal/health"
	"github.com/oni1997/gatewayx/internal/history"
	"github.com/oni1997/gatewayx/internal/logger"
	"github.com/oni1997/gatewayx/internal/metrics"
	"github.com/oni1997/gatewayx/internal/middleware"
	"github.com/oni1997/gatewayx/internal/ml"
	"github.com/oni1997/gatewayx/internal/oauth"
	"github.com/oni1997/gatewayx/internal/proxy"
	"github.com/oni1997/gatewayx/internal/tracing"
)

type Version struct {
	Version   string
	Commit    string
	BuildDate string
}

func Run(cfgFile string, v Version) {
	if cfgFile == "" {
		cfgFile = os.Getenv("GATEWAYX_CONFIG")
	}
	if cfgFile == "" {
		cfgFile = "gatewayx.yaml"
	}

	cfg, err := config.LoadConfig(cfgFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to load config: %v\n", err)
		os.Exit(1)
	}

	log := logger.New(cfg.Logging)
	log.Info("gatewayx starting", "version", v.Version, "commit", v.Commit, "build_date", v.BuildDate)

	collector := metrics.NewCollector()
	histBuf := history.NewBuffer(max(cfg.Metrics.History, 1000))
	tracer := tracing.New(cfg.Metrics.Tracing, histBuf)

	adminStore := admin.NewStore()
	if dbPath := os.Getenv("GATEWAYX_DB_PATH"); dbPath != "" {
		persistent, err := admin.NewStoreWithPersistence(dbPath)
		if err != nil {
			log.Error("failed to initialize persistence", "error", err, "path", dbPath)
		} else {
			adminStore = persistent
			log.Info("persistence enabled", "path", dbPath)
		}
	}
	adminStore.SetAuditLog(admin.NewAuditLog(log))
	wh := alert.NewWebhook(os.Getenv("GATEWAYX_WEBHOOK_URL"))

	rp, err := proxy.New(cfg, log)
	if err != nil {
		log.Error("failed to create proxy", "error", err)
		os.Exit(1)
	}

	rp.SetKeyResolver(func(rawKey string) string {
		if key, ok := adminStore.ValidateKey(rawKey); ok {
			return key.Owner
		}
		return ""
	})

	checker := health.New()
	checker.Register("gateway", func() error { return nil })

	mux := http.NewServeMux()
	mux.Handle(cfg.Health.Path, checker.Handler())
	mux.Handle("/", rp)

	var handler http.Handler = mux
	handler = middleware.Recovery(handler)
	handler = middleware.MaxBodySize(cfg.Security.MaxBodySize)(handler)
	if cfg.Metrics.Enabled {
		handler = metrics.Middleware(collector)(handler)
		handler = tracer.Middleware(handler)
	}

	if cfg.Metrics.Enabled {
		log.Info("metrics enabled", "port", cfg.Metrics.Port)
		go func() {
			analysisSvc := ml.NewAnalysisService(collector, histBuf)

			metricsMux := http.NewServeMux()
			metricsMux.Handle(cfg.Metrics.Path, metrics.Exporter(collector))
			metricsMux.Handle("/history", histBuf.Handler())
			metricsMux.Handle("/health", checker.Handler())
			metricsMux.Handle("/version", versionHandler(v))
			metricsMux.Handle("/security", analysisSvc.SecurityHandler())
			metricsMux.Handle("/bottlenecks", analysisSvc.BottlenecksHandler())
			metricsMux.Handle("/recommendations", analysisSvc.RecommendationsHandler())
			metricsMux.Handle("/analysis", analysisSvc.FullReportHandler())
			metricsMux.Handle("/api/", admin.RequireAuth(admin.NewHandler(adminStore, collector), cfg.Admin.Token))

			if cfg.OAuth.Provider != "" && cfg.OAuth.ClientID != "" && cfg.OAuth.ClientSecret != "" {
				oauthAuth, err := auth.NewOAuth(auth.OAuthOptions{
					Provider:     cfg.OAuth.Provider,
					ClientID:     cfg.OAuth.ClientID,
					ClientSecret: cfg.OAuth.ClientSecret,
					RedirectURL:  cfg.OAuth.RedirectURL,
				})
				if err == nil {
					flow := oauth.NewFlow(oauthAuth)
					metricsMux.Handle("/oauth/login", flow.LoginHandler())
					metricsMux.Handle("/oauth/callback", flow.CallbackHandler())
					metricsMux.Handle("/oauth/logout", flow.LogoutHandler())
					log.Info("OAuth flow enabled", "provider", cfg.OAuth.Provider)
				}
			}

			metricsMux.Handle("/", dashboardHandler())
			metricsAddr := fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Metrics.Port)
			log.Info("dashboard available", "url", "http://"+metricsAddr)
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
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)

	reload := func() {
		newCfg, err := config.LoadConfig(cfgFile)
		if err != nil {
			log.Error("failed to reload config", "error", err)
			wh.Send("config_error", "Config Reload Failed", fmt.Sprintf("Failed to reload configuration: %v", err))
			return
		}
		if err := rp.ReloadConfig(newCfg); err != nil {
			log.Error("failed to apply new config", "error", err)
			wh.Send("config_error", "Config Reload Failed", fmt.Sprintf("Failed to apply new configuration: %v", err))
			return
		}
		log.Info("configuration reloaded successfully")
	}

	stopWatch := watchConfigFile(cfgFile, reload)

	go func() {
		for sig := range sigCh {
			switch sig {
			case syscall.SIGHUP:
				log.Info("received SIGHUP, reloading configuration...")
				reload()
			default:
				log.Info("shutting down...")
				stopWatch()
				shutdownCtx, cancel := context.WithTimeout(context.Background(), cfg.Server.ShutdownTimeout)
				defer cancel()
				if err := server.Shutdown(shutdownCtx); err != nil {
					log.Error("graceful shutdown failed", "error", err)
				}
				return
			}
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

func versionHandler(v Version) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = fmt.Fprintf(w, `{"version":%q,"commit":%q,"build_date":%q}`, v.Version, v.Commit, v.BuildDate)
	})
}
