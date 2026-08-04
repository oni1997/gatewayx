package auth

import (
	"fmt"
	"net/http"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

const NameSession = "session"

type SessionAuthenticator struct {
	store    *SessionStore
	config   SessionConfig
}

type SessionConfig struct {
	TTL           time.Duration
	MaxSessions   int64
	CleanupInterval time.Duration
}

type Session struct {
	ID        string
	UserID    string
	Claims    Claims
	CreatedAt time.Time
	ExpiresAt time.Time
	LastUsed  time.Time
	mu        sync.RWMutex
}

func (s *Session) IsExpired() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return time.Now().After(s.ExpiresAt)
}

func (s *Session) Touch() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.LastUsed = time.Now()
}

type SessionStore struct {
	sessions      sync.Map
	maxSessions   int64
	count         atomic.Int64
	ttl           time.Duration
	stopCleanup   chan struct{}
}

func NewSessionStore(maxSessions int64, ttl time.Duration) *SessionStore {
	ss := &SessionStore{
		maxSessions: maxSessions,
		ttl:         ttl,
		stopCleanup: make(chan struct{}),
	}
	go ss.cleanupLoop()
	return ss
}

func (ss *SessionStore) Create(userID string, claims Claims) (*Session, error) {
	if ss.maxSessions > 0 && ss.count.Load() >= ss.maxSessions {
		return nil, fmt.Errorf("max sessions reached")
	}

	now := time.Now()
	session := &Session{
		ID:        generateSessionID(userID),
		UserID:    userID,
		Claims:    claims,
		CreatedAt: now,
		ExpiresAt: now.Add(ss.ttl),
		LastUsed:  now,
	}

	ss.sessions.Store(session.ID, session)
	ss.count.Add(1)

	return session, nil
}

func (ss *SessionStore) Get(id string) (*Session, bool) {
	val, ok := ss.sessions.Load(id)
	if !ok {
		return nil, false
	}
	session := val.(*Session)
	if session.IsExpired() {
		ss.sessions.Delete(id)
		ss.count.Add(-1)
		return nil, false
	}
	session.Touch()
	return session, true
}

func (ss *SessionStore) Delete(id string) {
	if _, ok := ss.sessions.LoadAndDelete(id); ok {
		ss.count.Add(-1)
	}
}

func (ss *SessionStore) Count() int64 {
	return ss.count.Load()
}

func (ss *SessionStore) Stop() {
	close(ss.stopCleanup)
}

func (ss *SessionStore) cleanupLoop() {
	ticker := time.NewTicker(ss.ttl / 2)
	if tickerDuration := ss.ttl / 2; tickerDuration <= 0 {
		ticker = time.NewTicker(time.Minute)
	}
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			now := time.Now()
			ss.sessions.Range(func(key, value any) bool {
				session := value.(*Session)
				if now.After(session.ExpiresAt) {
					ss.sessions.Delete(key)
					ss.count.Add(-1)
				}
				return true
			})
		case <-ss.stopCleanup:
			return
		}
	}
}

func generateSessionID(userID string) string {
	now := time.Now().UnixNano()
	return fmt.Sprintf("%s-%d-%x", userID, now, now%1000000)
}

type SessionOptions struct {
	TTL           time.Duration
	MaxSessions   int64
	CookieName    string
	HeaderName    string
}

func NewSession(opts SessionOptions) (*SessionAuthenticator, error) {
	ttl := opts.TTL
	if ttl == 0 {
		ttl = 30 * time.Minute
	}
	maxSessions := opts.MaxSessions
	if maxSessions == 0 {
		maxSessions = 10000
	}

	sa := &SessionAuthenticator{
		store: NewSessionStore(maxSessions, ttl),
		config: SessionConfig{
			TTL:           ttl,
			MaxSessions:   maxSessions,
			CleanupInterval: ttl / 2,
		},
	}

	return sa, nil
}

func (sa *SessionAuthenticator) Name() string {
	return NameSession
}

func (sa *SessionAuthenticator) Store() *SessionStore {
	return sa.store
}

func (sa *SessionAuthenticator) Authenticate(r *http.Request) (Claims, error) {
	sessionID := extractSessionID(r)

	if sessionID == "" {
		return nil, fmt.Errorf("missing session token")
	}

	session, ok := sa.store.Get(sessionID)
	if !ok {
		return nil, fmt.Errorf("invalid or expired session")
	}

	claims := make(Claims)
	session.mu.RLock()
	for k, v := range session.Claims {
		claims[k] = v
	}
	claims["session_id"] = session.ID
	claims["user_id"] = session.UserID
	session.mu.RUnlock()

	return claims, nil
}

func extractSessionID(r *http.Request) string {
	if sid := r.Header.Get("X-Session-ID"); sid != "" {
		return sid
	}

	cookie, err := r.Cookie("gatewayx_session")
	if err == nil && cookie.Value != "" {
		return cookie.Value
	}

	auth := r.Header.Get("Authorization")
	if strings.HasPrefix(auth, "Session ") {
		return strings.TrimPrefix(auth, "Session ")
	}

	return ""
}

func CreateSession(store *SessionStore, userID string, claims Claims) (*Session, error) {
	return store.Create(userID, claims)
}

func DestroySession(store *SessionStore, id string) {
	store.Delete(id)
}
