package proxy

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/oni1997/gatewayx/internal/auth"
	"github.com/oni1997/gatewayx/internal/cache"
	"github.com/oni1997/gatewayx/internal/config"
	"github.com/oni1997/gatewayx/internal/healthcheck"
	"github.com/oni1997/gatewayx/internal/ratelimit"
	"github.com/oni1997/gatewayx/pkg/circuitbreaker"
	"github.com/oni1997/gatewayx/pkg/compression"
	"github.com/oni1997/gatewayx/pkg/loadbalancer"
)

type ReverseProxy struct {
	config         *config.Config
	logger         *slog.Logger
	server         *http.Server
	routes         map[string]*routeProxy
	breakers       map[string]*circuitbreaker.Breaker
	healthMonitors []*healthcheck.Monitor
	keyResolver    func(rawKey string) string
	mu             sync.RWMutex
}

func (rp *ReverseProxy) SetKeyResolver(resolver func(rawKey string) string) {
	rp.keyResolver = resolver
}

type routeProxy struct {
	config   config.RouteConfig
	handler  http.Handler
	balancer loadbalancer.Balancer
	logger   *slog.Logger
}

func New(cfg *config.Config, logger *slog.Logger) (*ReverseProxy, error) {
	rp := &ReverseProxy{
		config:   cfg,
		logger:   logger,
		routes:   make(map[string]*routeProxy),
		breakers: make(map[string]*circuitbreaker.Breaker),
	}

	for _, route := range cfg.Routes {
		if err := rp.addRoute(route); err != nil {
			return nil, fmt.Errorf("failed to add route %s: %w", route.Name, err)
		}
	}

	return rp, nil
}

func (rp *ReverseProxy) ReloadConfig(cfg *config.Config) error {
	rp.mu.Lock()
	defer rp.mu.Unlock()

	for _, m := range rp.healthMonitors {
		m.Stop()
	}

	rp.routes = make(map[string]*routeProxy)
	rp.breakers = make(map[string]*circuitbreaker.Breaker)
	rp.healthMonitors = nil
	rp.config = cfg

	for _, route := range cfg.Routes {
		if err := rp.addRouteLocked(route); err != nil {
			return fmt.Errorf("failed to add route %s: %w", route.Name, err)
		}
	}

	rp.logger.Info("configuration reloaded", "routes", len(rp.routes))
	return nil
}

func (rp *ReverseProxy) addRoute(routeCfg config.RouteConfig) error {
	rp.mu.Lock()
	defer rp.mu.Unlock()
	return rp.addRouteLocked(routeCfg)
}

func (rp *ReverseProxy) addRouteLocked(routeCfg config.RouteConfig) error {
	backends := make([]*url.URL, 0, len(routeCfg.UpstreamURLs))
	for _, u := range routeCfg.UpstreamURLs {
		parsed, err := url.Parse(u)
		if err != nil {
			return fmt.Errorf("invalid upstream URL %s: %w", u, err)
		}
		backends = append(backends, parsed)
	}

	lb := loadbalancer.NewRoundRobin(backends)
	cb := circuitbreaker.New(5, 3, 30*time.Second)
	rp.breakers[routeCfg.Name] = cb

	var healthBalancer *loadbalancer.HealthAwareRoundRobin
	var healthMonitor *healthcheck.Monitor
	if routeCfg.HealthCheck != nil {
		healthBalancer = loadbalancer.NewHealthAwareRoundRobin(backends)
		healthMonitor = healthcheck.NewMonitor(healthBalancer, healthcheck.Config{
			Path:      routeCfg.HealthCheck.Path,
			Interval:  routeCfg.HealthCheck.Interval,
			Timeout:   routeCfg.HealthCheck.Timeout,
			Healthy:   routeCfg.HealthCheck.Healthy,
			Unhealthy: routeCfg.HealthCheck.Unhealthy,
		})
		healthMonitor.Start()
		rp.healthMonitors = append(rp.healthMonitors, healthMonitor)
	}

	proxy := &httputil.ReverseProxy{
		Rewrite: func(pr *httputil.ProxyRequest) {
			if !cb.Allow() {
				pr.Out.URL.Scheme = ""
				pr.Out.URL.Host = ""
				return
			}
			var backend *url.URL
			if healthBalancer != nil {
				backend = healthBalancer.Next()
			} else {
				backend = lb.Next()
			}
			if backend == nil {
				pr.Out.URL.Scheme = ""
				pr.Out.URL.Host = ""
				return
			}
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
			cb.Failure()
			http.Error(w, "Bad Gateway", http.StatusBadGateway)
		},
		ModifyResponse: func(r *http.Response) error {
			cb.Success()
			if r.StatusCode >= 500 {
				cb.Failure()
			}
			for k, v := range routeCfg.Headers {
				r.Header.Set(k, v)
			}
			return nil
		},
	}

	var handler http.Handler = proxy

	if routeCfg.Websocket {
		handler = withWebSocketUpgrade(handler)
	} else {
		handler = withTimeout(handler, routeCfg.Timeout)
	}

	if routeCfg.Authentication != nil {
		authenticator, err := buildAuthenticator(routeCfg.Authentication)
		if err != nil {
			return fmt.Errorf("failed to build authenticator for route %s: %w", routeCfg.Name, err)
		}
		handler = auth.Middleware(authenticator)(handler)
	}

	if routeCfg.RateLimit != nil {
		rlCfg := ratelimit.Config{
			Rate:        routeCfg.RateLimit.Rate,
			Burst:       routeCfg.RateLimit.Burst,
			Strategy:    routeCfg.RateLimit.Strategy,
			PerUser:     routeCfg.RateLimit.PerUser,
			PerIP:       routeCfg.RateLimit.PerIP,
			PerKey:      routeCfg.RateLimit.PerKey,
			RedisAddr:   routeCfg.RateLimit.RedisAddr,
			RedisPrefix: routeCfg.RateLimit.RedisPrefix,
			RouteName:   routeCfg.Name,
		}
		rlStore := buildRateLimitStore(rlCfg)
		rlMw := ratelimit.NewMiddleware(rlStore, rlCfg)
		rlMw.SetKeyResolver(rp.keyResolver)
		handler = rlMw.Handler(handler)
	}

	if routeCfg.Cache != nil && routeCfg.Cache.TTL > 0 {
		routeCache := cache.New(routeCfg.Cache.TTL, routeCfg.Cache.MaxSize)
		handler = cache.Middleware(routeCache)(handler)
	}

	if routeCfg.Compression {
		handler = compression.GzipMiddleware(handler)
	}

	handler = withLogger(rp.logger, routeCfg.Name)(handler)

	rp.routes[routeCfg.ListenPath] = &routeProxy{
		config:   routeCfg,
		handler:  handler,
		balancer: lb,
		logger:   rp.logger.With("route", routeCfg.Name),
	}

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

func withWebSocketUpgrade(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.EqualFold(r.Header.Get("Upgrade"), "websocket") {
			r.Header.Set("Connection", "Upgrade")
		}
		next.ServeHTTP(w, r)
	})
}

func withLogger(logger *slog.Logger, name string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if !logger.Enabled(r.Context(), slog.LevelInfo) {
				next.ServeHTTP(w, r)
				return
			}
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
	opts := ac.AllOptions()

	switch ac.Type {
	case auth.NameJWT:
		cacheTTL := parseDuration(opts["cache_ttl"])
		ja, err := auth.NewJWT(auth.JWTOptions{
			Secret:        opts["secret"],
			SecretFile:    opts["secret_file"],
			PublicKeyFile: opts["public_key_file"],
			Algorithm:     opts["algorithm"],
		})
		if err != nil {
			return nil, err
		}
		if cacheTTL > 0 {
			return auth.NewCachedJWT(ja, cacheTTL), nil
		}
		return ja, nil

	case auth.NameAPIKey:
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

	case auth.NameBasic:
		htpasswdFile := opts["htpasswd_file"]
		realm := opts["realm"]
		inlineUsers := make(map[string]string)
		for k, v := range opts {
			if k != "htpasswd_file" && k != "realm" && k != "type" {
				inlineUsers[k] = v
			}
		}
		return auth.NewBasic(auth.BasicOptions{
			Users:    inlineUsers,
			Htpasswd: htpasswdFile,
			Realm:    realm,
		})

	case auth.NameHMAC:
		algo := opts["algorithm"]
		headerName := opts["header"]
		skew := parseDuration(opts["clock_skew"])
		return auth.NewHMAC(auth.HMACOptions{
			Secret:     opts["secret"],
			Algorithm:  algo,
			HeaderName: headerName,
			ClockSkew:  skew,
		})

	case "rbac":
		rolesClaim := opts["roles_claim"]
		delegateType := opts["delegate"]
		delegateAc := &config.AuthConfig{
			Type:    delegateType,
			Options: ac.Options,
		}
		delegate, err := buildAuthenticator(delegateAc)
		if err != nil {
			return nil, fmt.Errorf("rbac delegate auth: %w", err)
		}
		engine := buildRBACEngine(opts)
		return auth.NewRBAC(delegate, engine, rolesClaim), nil

	case auth.NameSession:
		ttl := parseDuration(opts["ttl"])
		if ttl == 0 {
			ttl = 30 * time.Minute
		}
		maxSessions := int64(10000)
		if msRaw := opts["max_sessions"]; msRaw != "" {
			_, _ = fmt.Sscanf(msRaw, "%d", &maxSessions)
		}
		return auth.NewSession(auth.SessionOptions{
			TTL:         ttl,
			MaxSessions: maxSessions,
			CookieName:  opts["cookie_name"],
			HeaderName:  opts["header_name"],
		})

	case auth.NameOAuth:
		return auth.NewOAuth(auth.OAuthOptions{
			Provider:     opts["provider"],
			ClientID:     opts["client_id"],
			ClientSecret: opts["client_secret"],
			RedirectURL:  opts["redirect_url"],
		})

	case auth.NameMTLS:
		verifyDepth := 1
		if depthStr := opts["verify_depth"]; depthStr != "" {
			_, _ = fmt.Sscanf(depthStr, "%d", &verifyDepth)
		}
		return auth.NewMTLS(auth.MTLS{
			CACertFile:  opts["ca_cert"],
			VerifyDepth: verifyDepth,
		})

	default:
		return nil, fmt.Errorf("unknown auth type: %s", ac.Type)
	}
}

func buildRBACEngine(opts map[string]string) *auth.RBACEngine {
	engine := auth.NewRBACEngine()

	for k, v := range opts {
		if !strings.HasPrefix(k, "perm_") {
			continue
		}
		permParts := strings.SplitN(v, ":", 3)
		perm := auth.Permission{}
		switch len(permParts) {
		case 3:
			perm.Path = permParts[0]
			perm.Roles = strings.Split(permParts[1], ",")
			perm.Methods = strings.Split(permParts[2], ",")
		case 2:
			perm.Path = permParts[0]
			perm.Roles = strings.Split(permParts[1], ",")
		case 1:
			perm.Path = permParts[0]
		}
		if perm.Path != "" {
			engine.AddPermission(perm)
		}
	}

	return engine
}

func parseDuration(s string) time.Duration {
	if s == "" {
		return 0
	}
	d, err := time.ParseDuration(s)
	if err != nil {
		return 0
	}
	return d
}

func buildRateLimitStore(cfg ratelimit.Config) ratelimit.Store {
	if cfg.RedisAddr != "" {
		return ratelimit.NewMemoryStore(cfg)
	}
	return ratelimit.NewMemoryStore(cfg)
}
