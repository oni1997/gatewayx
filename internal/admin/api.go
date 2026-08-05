package admin

import (
	"crypto/rand"
	"encoding/json"
	"fmt"
	"math/big"
	"net/http"
	"sync"
	"time"

	"github.com/oni1997/gatewayx/internal/metrics"
)

type APIKey struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Owner     string    `json:"owner"`
	Key       string    `json:"key,omitempty"`
	Prefix    string    `json:"prefix"`
	CreatedAt time.Time `json:"created_at"`
	LastUsed  time.Time `json:"last_used"`
}

type Certificate struct {
	ID        string    `json:"id"`
	Domain    string    `json:"domain"`
	Issuer    string    `json:"issuer"`
	NotAfter  time.Time `json:"not_after"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
}

type Store struct {
	mu   sync.RWMutex
	keys map[string]*APIKey
	certs map[string]*Certificate
}

func NewStore() *Store {
	return &Store{
		keys:  make(map[string]*APIKey),
		certs: make(map[string]*Certificate),
	}
}

func (s *Store) CreateKey(name, owner string) (*APIKey, string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	id := generateID()
	fullKey := "sk-" + generateSecret(32)
	prefix := fullKey[:11] + "..."

	key := &APIKey{
		ID:        id,
		Name:      name,
		Owner:     owner,
		Key:       fullKey,
		Prefix:    prefix,
		CreatedAt: time.Now(),
	}

	s.keys[fullKey] = key
	return key, fullKey
}

func (s *Store) ValidateKey(rawKey string) (*APIKey, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	key, ok := s.keys[rawKey]
	if ok {
		key.LastUsed = time.Now()
	}
	return key, ok
}

func (s *Store) ListKeys() []APIKey {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make([]APIKey, 0, len(s.keys))
	for _, k := range s.keys {
		clone := *k
		clone.Key = ""
		result = append(result, clone)
	}
	return result
}

func (s *Store) RevokeKey(id string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	for raw, key := range s.keys {
		if key.ID == id {
			delete(s.keys, raw)
			return true
		}
	}
	return false
}

func (s *Store) AddCertificate(domain, issuer string, notAfter time.Time) *Certificate {
	s.mu.Lock()
	defer s.mu.Unlock()

	cert := &Certificate{
		ID:        generateID(),
		Domain:    domain,
		Issuer:    issuer,
		NotAfter:  notAfter,
		CreatedAt: time.Now(),
		Status:    certStatus(notAfter),
	}
	s.certs[cert.ID] = cert
	return cert
}

func (s *Store) ListCertificates() []Certificate {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make([]Certificate, 0, len(s.certs))
	for _, c := range s.certs {
		clone := *c
		clone.Status = certStatus(clone.NotAfter)
		result = append(result, clone)
	}
	return result
}

func certStatus(notAfter time.Time) string {
	now := time.Now()
	if now.After(notAfter) {
		return "expired"
	}
	if now.Add(30 * 24 * time.Hour).After(notAfter) {
		return "expiring"
	}
	return "active"
}

func NewHandler(store *Store, collector *metrics.Collector) http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /api/keys", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(store.ListKeys())
	})

	mux.HandleFunc("POST /api/keys", func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			Name  string `json:"name"`
			Owner string `json:"owner"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, `{"error":"invalid request"}`, 400)
			return
		}
		key, fullKey := store.CreateKey(req.Name, req.Owner)
		result := map[string]any{
			"id":   key.ID,
			"name": key.Name,
			"key":  fullKey,
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(result)
	})

	mux.HandleFunc("DELETE /api/keys/{id}", func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")
		if store.RevokeKey(id) {
			w.WriteHeader(http.StatusNoContent)
		} else {
			http.Error(w, `{"error":"not found"}`, 404)
		}
	})

	mux.HandleFunc("GET /api/certs", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(store.ListCertificates())
	})

	mux.HandleFunc("POST /api/certs", func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			Domain  string `json:"domain"`
			Issuer  string `json:"issuer"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, `{"error":"invalid request"}`, 400)
			return
		}
		cert := store.AddCertificate(req.Domain, req.Issuer, time.Now().Add(90*24*time.Hour))
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(cert)
	})

	mux.HandleFunc("GET /api/config", func(w http.ResponseWriter, r *http.Request) {
		snapshot := collector.Snapshot()
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"uptime":    time.Since(collector.Uptime).String(),
			"requests":  snapshot.Requests,
			"routes":    snapshot.Routes,
		})
	})

	return mux
}

func generateID() string {
	return fmt.Sprintf("%d-%s", time.Now().UnixNano()/1000000, generateSecret(8))
}

func generateSecret(length int) string {
	const chars = "abcdefghijklmnopqrstuvwxyz0123456789"
	b := make([]byte, length)
	for i := range b {
		n, _ := rand.Int(rand.Reader, big.NewInt(int64(len(chars))))
		b[i] = chars[n.Int64()]
	}
	return string(b)
}
