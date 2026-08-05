package auth

import (
	"bufio"
	"fmt"
	"net/http"
	"os"
	"strings"
	"sync"
)

const NameAPIKey = "api_key"

type APIKeyAuthenticator struct {
	keys       map[string]string
	headerName string
	mu         sync.RWMutex
}

type APIKeyOptions struct {
	Keys       map[string]string
	KeysFile   string
	HeaderName string
}

func NewAPIKey(opts APIKeyOptions) (*APIKeyAuthenticator, error) {
	ak := &APIKeyAuthenticator{
		keys:       make(map[string]string),
		headerName: opts.HeaderName,
	}

	if ak.headerName == "" {
		ak.headerName = "X-API-Key"
	}

	for k, v := range opts.Keys {
		ak.keys[k] = v
	}

	if opts.KeysFile != "" {
		if err := ak.loadKeysFile(opts.KeysFile); err != nil {
			return nil, fmt.Errorf("failed to load api keys file: %w", err)
		}
	}

	if len(ak.keys) == 0 {
		return nil, fmt.Errorf("api_key: no keys configured (provide keys or keys_file)")
	}

	return ak, nil
}

func (ak *APIKeyAuthenticator) loadKeysFile(path string) error {
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
			ak.keys[strings.TrimSpace(parts[0])] = strings.TrimSpace(parts[1])
		} else {
			ak.keys[parts[0]] = parts[0]
		}
	}
	return scanner.Err()
}

func (ak *APIKeyAuthenticator) Name() string {
	return NameAPIKey
}

func (ak *APIKeyAuthenticator) Authenticate(r *http.Request) (Claims, error) {
	key := r.Header.Get(ak.headerName)
	if key == "" {
		return nil, fmt.Errorf("missing api key header: %s", ak.headerName)
	}

	ak.mu.RLock()
	owner, ok := ak.keys[key]
	ak.mu.RUnlock()

	if !ok {
		return nil, fmt.Errorf("invalid api key")
	}

	return Claims{
		"sub":     owner,
		"api_key": key,
	}, nil
}
