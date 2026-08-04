package auth

import (
	"bufio"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"fmt"
	"net/http"
	"os"
	"strings"
	"sync"
)

const NameBasic = "basic"

type BasicAuthenticator struct {
	users  map[string]string
	realm  string
	mu     sync.RWMutex
}

type BasicOptions struct {
	Users    map[string]string
	Htpasswd string
	Realm    string
}

func NewBasic(opts BasicOptions) (*BasicAuthenticator, error) {
	ba := &BasicAuthenticator{
		users: make(map[string]string),
		realm: opts.Realm,
	}

	if ba.realm == "" {
		ba.realm = "GatewayX"
	}

	for k, v := range opts.Users {
		ba.users[k] = v
	}

	if opts.Htpasswd != "" {
		if err := ba.loadHtpasswd(opts.Htpasswd); err != nil {
			return nil, fmt.Errorf("failed to load htpasswd file: %w", err)
		}
	}

	if len(ba.users) == 0 {
		return nil, fmt.Errorf("basic: no users configured")
	}

	return ba, nil
}

func (ba *BasicAuthenticator) loadHtpasswd(path string) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.SplitN(line, ":", 2)
		if len(parts) == 2 {
			ba.users[parts[0]] = parts[1]
		}
	}
	return scanner.Err()
}

func (ba *BasicAuthenticator) Name() string {
	return NameBasic
}

func (ba *BasicAuthenticator) Authenticate(r *http.Request) (Claims, error) {
	user, pass, ok := r.BasicAuth()
	if !ok {
		w := &responseWriter{}
		w.Header().Set("WWW-Authenticate", fmt.Sprintf(`Basic realm="%s"`, ba.realm))
		return nil, fmt.Errorf("missing authorization header")
	}

	ba.mu.RLock()
	expectedPass, exists := ba.users[user]
	ba.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("invalid credentials")
	}

	if strings.HasPrefix(expectedPass, "{SHA}") {
		if !validateSHAPassword(pass, expectedPass) {
			return nil, fmt.Errorf("invalid credentials")
		}
	} else {
		if subtle.ConstantTimeCompare([]byte(pass), []byte(expectedPass)) != 1 {
			return nil, fmt.Errorf("invalid credentials")
		}
	}

	return Claims{
		"sub":  user,
		"type": "basic",
	}, nil
}

func validateSHAPassword(password, expected string) bool {
	hash := sha256.Sum256([]byte(password))
	encoded := base64.StdEncoding.EncodeToString(hash[:])
	prefix := "{SHA}"
	if len(expected) > len(prefix) {
		return subtle.ConstantTimeCompare([]byte(encoded), []byte(expected[len(prefix):])) == 1
	}
	return false
}

type responseWriter struct {
	headers http.Header
}

func (rw *responseWriter) Header() http.Header {
	if rw.headers == nil {
		rw.headers = make(http.Header)
	}
	return rw.headers
}

func (rw *responseWriter) Write([]byte) (int, error)     { return 0, nil }
func (rw *responseWriter) WriteHeader(int)                {}
