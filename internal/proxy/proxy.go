package proxy

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"net/http/httputil"
	"net/url"
	"sync"
	"time"

	"github.com/gatewayx/gatewayx/internal/auth"
	"github.com/gatewayx/gatewayx/internal/config"
	"github.com/gatewayx/gatewayx/pkg/loadbalancer"
)

type ReverseProxy struct {
	config    *config.Config
	logger    *slog.Logger
	server    *http.Server
	routes    map[string]*routeProxy
	mu        sync.RWMutex
}

type routeProxy struct {
	config   config.RouteConfig
	handler  http.Handler
	balancer loadbalancer.Balancer
	logger   *slog.Logger
}

func New(cfg *config.Config, logger *slog.Logger) (*ReverseProxy, error) {
	rp := &ReverseProxy{
		config: cfg,
		logger: logger,
		routes: make(map[string]*routeProxy),
	}

	for _, route := range cfg.Routes {
		if err := rp.addRoute(route); err != nil {
			return nil, fmt.Errorf("failed to add route %s: %w", route.Name, err)
		}
	}

	return rp, nil
}

func (rp *ReverseProxy) addRoute(routeCfg config.RouteConfig) error {
	backends := make([]*url.URL, 0, len(routeCfg.UpstreamURLs))
	for _, u := range routeCfg.UpstreamURLs {
		parsed, err := url.Parse(u)
		if err != nil {
			return fmt.Errorf("invalid upstream URL %s: %w", u, err)
		}
		backends = append(backends, parsed)
	}

	lb := loadbalancer.NewRoundRobin(backends)

	proxy := &httputil.ReverseProxy{
		Rewrite: func(pr *httputil.ProxyRequest) {
			backend := lb.Next()
			pr.Out.URL.Scheme = backend.Scheme
			pr.Out.URL.Host = backend.Host
			if routeCfg.StripPath {
				pr.Out.URL.Path = ""
			}
			if routeCfg.PreserveHost {
				pr.Out.Host = pr.In.Host
			}
		},
		ErrorHandler: func(w http.ResponseWriter, r *http.Request, err error) {
			rp.logger.Error("proxy error",
				"route", routeCfg.Name,
				"error", err,
				"method", r.Method,
				"path", r.URL.Path,
			)
			http.Error(w, "Bad Gateway", http.StatusBadGateway)
		},
		ModifyResponse: func(r *http.Response) error {
			for k, v := range routeCfg.Headers {
				r.Header.Set(k, v)
			}
			return nil
		},
	}

	var handler http.Handler = proxy
	handler = withTimeout(handler, routeCfg.Timeout)

	if routeCfg.Authentication != nil {
		authenticator, err := buildAuthenticator(routeCfg.Authentication)
		if err != nil {
			return fmt.Errorf("failed to build authenticator for route %s: %w", routeCfg.Name, err)
		}
		handler = auth.Middleware(authenticator)(handler)
	}

	handler = withLogger(rp.logger, routeCfg.Name)(handler)

	rp.mu.Lock()
	rp.routes[routeCfg.ListenPath] = &routeProxy{
		config:   routeCfg,
		handler:  handler,
		balancer: lb,
		logger:   rp.logger.With("route", routeCfg.Name),
	}
	rp.mu.Unlock()

	return nil
}

func (rp *ReverseProxy) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	rp.mu.RLock()
	route, ok := rp.routes[r.URL.Path]
	rp.mu.RUnlock()

	if !ok {
		rp.mu.RLock()
		for path, rp2 := range rp.routes {
			if len(r.URL.Path) >= len(path) && r.URL.Path[:len(path)] == path {
				route = rp2
				break
			}
		}
		rp.mu.RUnlock()
	}

	if route == nil {
		http.NotFound(w, r)
		return
	}

	route.handler.ServeHTTP(w, r)
}

func (rp *ReverseProxy) Start(ctx context.Context) error {
	mux := http.NewServeMux()
	mux.Handle("/", rp)

	addr := fmt.Sprintf("%s:%d", rp.config.Server.Host, rp.config.Server.Port)

	rp.server = &http.Server{
		Addr:           addr,
		Handler:        mux,
		ReadTimeout:    rp.config.Server.ReadTimeout,
		WriteTimeout:   rp.config.Server.WriteTimeout,
		IdleTimeout:    rp.config.Server.IdleTimeout,
		MaxHeaderBytes: rp.config.Server.MaxHeaderBytes,
	}

	rp.logger.Info("gateway starting", "addr", addr)

	go func() {
		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), rp.config.Server.ShutdownTimeout)
		defer cancel()
		if err := rp.server.Shutdown(shutdownCtx); err != nil {
			rp.logger.Error("graceful shutdown failed", "error", err)
		}
	}()

	if rp.config.TLS.Enabled {
		return rp.server.ListenAndServeTLS(rp.config.TLS.CertFile, rp.config.TLS.KeyFile)
	}
	return rp.server.ListenAndServe()
}

func (rp *ReverseProxy) GetServer() *http.Server {
	return rp.server
}

func withTimeout(next http.Handler, timeout time.Duration) http.Handler {
	if timeout <= 0 {
		return next
	}
	return http.TimeoutHandler(next, timeout, "Gateway Timeout")
}

func withLogger(logger *slog.Logger, name string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			sw := &statusWriter{ResponseWriter: w, status: http.StatusOK}
			next.ServeHTTP(sw, r)
			logger.Info("request",
				"route", name,
				"method", r.Method,
				"path", r.URL.Path,
				"status", sw.status,
				"duration", time.Since(start).String(),
				"remote_addr", r.RemoteAddr,
			)
		})
	}
}

type statusWriter struct {
	http.ResponseWriter
	status int
}

func (sw *statusWriter) WriteHeader(code int) {
	sw.status = code
	sw.ResponseWriter.WriteHeader(code)
}

func buildAuthenticator(ac *config.AuthConfig) (auth.Authenticator, error) {
	switch ac.Type {
	case auth.NameJWT:
		opts := ac.JWTOptions()
		algo := opts["algorithm"]
		secret := opts["secret"]
		secretFile := opts["secret_file"]
		publicKeyFile := opts["public_key_file"]

		return auth.NewJWT(auth.JWTOptions{
			Secret:        secret,
			SecretFile:    secretFile,
			PublicKeyFile: publicKeyFile,
			Algorithm:     algo,
		})

	case auth.NameAPIKey:
		opts := ac.APIKeyOptions()
		keysFile := opts["keys_file"]
		headerName := opts["header"]

		inlineKeys := make(map[string]string)
		for k, v := range opts {
			if k != "keys_file" && k != "header" && k != "type" {
				inlineKeys[k] = v
			}
		}

		return auth.NewAPIKey(auth.APIKeyOptions{
			Keys:       inlineKeys,
			KeysFile:   keysFile,
			HeaderName: headerName,
		})

	default:
		return nil, fmt.Errorf("unknown auth type: %s", ac.Type)
	}
}
