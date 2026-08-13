package oauth

import (
	"crypto/rand"
	"encoding/hex"
	"net/http"
	"time"

	"github.com/oni1997/gatewayx/internal/auth"
)

type Flow struct {
	authenticator *auth.OAuthAuthenticator
	cookieName    string
}

func NewFlow(authenticator *auth.OAuthAuthenticator) *Flow {
	return &Flow{
		authenticator: authenticator,
		cookieName:    "gatewayx_oauth",
	}
}

func (f *Flow) LoginHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		state := randomState()
		http.SetCookie(w, &http.Cookie{
			Name:     "gatewayx_oauth_state",
			Value:    state,
			Path:     "/",
			HttpOnly: true,
			MaxAge:   600,
		})
		http.Redirect(w, r, f.authenticator.AuthURL(state), http.StatusFound)
	})
}

func (f *Flow) CallbackHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		state := r.URL.Query().Get("state")
		code := r.URL.Query().Get("code")

		if state == "" || code == "" {
			http.Error(w, "missing state or code", http.StatusBadRequest)
			return
		}

		stateCookie, err := r.Cookie("gatewayx_oauth_state")
		if err != nil || stateCookie.Value != state {
			http.Error(w, "invalid state", http.StatusBadRequest)
			return
		}

		token, err := f.authenticator.ExchangeCode(code)
		if err != nil {
			http.Error(w, "token exchange failed: "+err.Error(), http.StatusBadRequest)
			return
		}

		http.SetCookie(w, &http.Cookie{
			Name:     f.cookieName,
			Value:    token,
			Path:     "/",
			HttpOnly: true,
			Secure:   r.TLS != nil,
			MaxAge:   3600,
		})

		http.Redirect(w, r, "/", http.StatusFound)
	})
}

func (f *Flow) LogoutHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.SetCookie(w, &http.Cookie{
			Name:     f.cookieName,
			Value:    "",
			Path:     "/",
			HttpOnly: true,
			MaxAge:   -1,
			Expires:  time.Unix(0, 0),
		})
		http.Redirect(w, r, "/", http.StatusFound)
	})
}

func randomState() string {
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}
