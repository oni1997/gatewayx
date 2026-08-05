package auth

import (
	"context"
	"net/http"
)

type contextKey string

const claimsKey contextKey = "auth-claims"

type Claims map[string]any

type Authenticator interface {
	Authenticate(r *http.Request) (Claims, error)
	Name() string
}

type Registry struct {
	authenticators map[string]Authenticator
}

func NewRegistry() *Registry {
	return &Registry{
		authenticators: make(map[string]Authenticator),
	}
}

func (r *Registry) Register(a Authenticator) {
	r.authenticators[a.Name()] = a
}

func (r *Registry) Get(name string) (Authenticator, bool) {
	a, ok := r.authenticators[name]
	return a, ok
}

func Middleware(authenticator Authenticator) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			claims, err := authenticator.Authenticate(r)
			if err != nil {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusUnauthorized)
				_, _ = w.Write([]byte(`{"error":"unauthorized","message":"` + err.Error() + `"}`))
				return
			}
			ctx := WithClaims(r.Context(), claims)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func WithClaims(ctx context.Context, claims Claims) context.Context {
	return context.WithValue(ctx, claimsKey, claims)
}

func GetClaims(ctx context.Context) (Claims, bool) {
	claims, ok := ctx.Value(claimsKey).(Claims)
	return claims, ok
}
