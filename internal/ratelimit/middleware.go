package ratelimit

import (
	"fmt"
	"net"
	"net/http"
	"strings"

	"github.com/oni1997/gatewayx/internal/auth"
)

type Store interface {
	Allow(key string) bool
	AllowN(key string, n int) bool
}

type Middleware struct {
	store       Store
	config      Config
	extractors  []KeyExtractor
	keyResolver func(rawKey string) string
}

func NewMiddleware(store Store, cfg Config) *Middleware {
	mw := &Middleware{
		store:  store,
		config: cfg,
	}
	mw.buildExtractors()
	return mw
}

func (rm *Middleware) SetKeyResolver(resolver func(rawKey string) string) {
	rm.keyResolver = resolver
}

func (rm *Middleware) buildExtractors() {
	useSpecific := rm.config.PerIP || rm.config.PerUser || rm.config.PerKey

	if rm.config.PerIP {
		rm.extractors = append(rm.extractors, func(r *http.Request) string {
			return rm.config.RouteName + ":ip:" + extractIP(r)
		})
	}

	if rm.config.PerUser {
		rm.extractors = append(rm.extractors, func(r *http.Request) string {
			claims, ok := auth.GetClaims(r.Context())
			if !ok {
				return rm.config.RouteName + ":user:anonymous"
			}
			sub, _ := claims["sub"].(string)
			if sub == "" {
				sub = "unknown"
			}
			return rm.config.RouteName + ":user:" + sub
		})
	}

	if rm.config.PerKey {
		rm.extractors = append(rm.extractors, func(r *http.Request) string {
			key := r.Header.Get("X-API-Key")
			if key == "" {
				key = "anonymous"
			}
			if rm.keyResolver != nil {
				if resolved := rm.keyResolver(key); resolved != "" {
					return rm.config.RouteName + ":key:" + resolved
				}
			}
			return rm.config.RouteName + ":key:" + key
		})
	}

	if !useSpecific {
		rm.extractors = append(rm.extractors, func(r *http.Request) string {
			return rm.config.RouteName + ":global"
		})
	}
}

func (rm *Middleware) Handler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		for _, extract := range rm.extractors {
			key := extract(r)
			if !rm.store.Allow(key) {
				w.Header().Set("Content-Type", "application/json")
				w.Header().Set("Retry-After", "1")
				w.Header().Set("X-RateLimit-Limit", fmt.Sprintf("%.0f", rm.config.Rate))
				w.WriteHeader(http.StatusTooManyRequests)
				_, _ = w.Write([]byte(`{"error":"rate_limit_exceeded","message":"too many requests"}`))
				return
			}
		}
		next.ServeHTTP(w, r)
	})
}

func extractIP(r *http.Request) string {
	if fwd := r.Header.Get("X-Forwarded-For"); fwd != "" {
		parts := strings.Split(fwd, ",")
		return strings.TrimSpace(parts[0])
	}

	if realIP := r.Header.Get("X-Real-IP"); realIP != "" {
		return strings.TrimSpace(realIP)
	}

	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}
