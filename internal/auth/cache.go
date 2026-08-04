package auth

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/http"
	"sync"
	"time"
)

type TokenCache struct {
	entries sync.Map
	ttl     time.Duration
}

type cacheEntry struct {
	Claims    Claims
	ExpiresAt time.Time
}

func NewTokenCache(ttl time.Duration) *TokenCache {
	tc := &TokenCache{ttl: ttl}
	go tc.cleanupLoop()
	return tc
}

func (tc *TokenCache) cacheKey(tokenString string) string {
	hash := sha256.Sum256([]byte(tokenString))
	return hex.EncodeToString(hash[:16])
}

func (tc *TokenCache) Get(tokenString string) (Claims, bool) {
	key := tc.cacheKey(tokenString)
	val, ok := tc.entries.Load(key)
	if !ok {
		return nil, false
	}

	entry := val.(*cacheEntry)
	if time.Now().After(entry.ExpiresAt) {
		tc.entries.Delete(key)
		return nil, false
	}

	claims := make(Claims)
	for k, v := range entry.Claims {
		claims[k] = v
	}
	return claims, true
}

func (tc *TokenCache) Set(tokenString string, claims Claims) {
	key := tc.cacheKey(tokenString)
	entry := &cacheEntry{
		Claims:    claims,
		ExpiresAt: time.Now().Add(tc.ttl),
	}

	expClaim, ok := claims["exp"]
	if ok {
		var expTime time.Time
		switch v := expClaim.(type) {
		case float64:
			expTime = time.Unix(int64(v), 0)
		case int64:
			expTime = time.Unix(v, 0)
		case time.Time:
			expTime = v
		}
		if !expTime.IsZero() {
			duration := time.Until(expTime)
			if duration < tc.ttl {
				entry.ExpiresAt = time.Now().Add(duration)
			}
		}
	}

	tc.entries.Store(key, entry)
}

func (tc *TokenCache) Delete(tokenString string) {
	tc.entries.Delete(tc.cacheKey(tokenString))
}

func (tc *TokenCache) Invalidate() {
	tc.entries.Range(func(key, _ any) bool {
		tc.entries.Delete(key)
		return true
	})
}

func (tc *TokenCache) Size() int {
	count := 0
	tc.entries.Range(func(_, _ any) bool {
		count++
		return true
	})
	return count
}

func (tc *TokenCache) cleanupLoop() {
	ticker := time.NewTicker(tc.ttl / 2)
	defer ticker.Stop()

	for range ticker.C {
		now := time.Now()
		tc.entries.Range(func(key, value any) bool {
			entry := value.(*cacheEntry)
			if now.After(entry.ExpiresAt) {
				tc.entries.Delete(key)
			}
			return true
		})
	}
}

type CachedJWTAuthenticator struct {
	delegate *JWTAuthenticator
	cache    *TokenCache
}

func NewCachedJWT(delegate *JWTAuthenticator, cacheTTL time.Duration) *CachedJWTAuthenticator {
	if cacheTTL == 0 {
		cacheTTL = 5 * time.Minute
	}
	return &CachedJWTAuthenticator{
		delegate: delegate,
		cache:    NewTokenCache(cacheTTL),
	}
}

func (cja *CachedJWTAuthenticator) Name() string {
	return "cached_jwt"
}

func (cja *CachedJWTAuthenticator) Authenticate(r *http.Request) (Claims, error) {
	tokenString := extractBearerToken(r)
	if tokenString == "" {
		return nil, fmt.Errorf("missing or malformed authorization header")
	}

	if cached, ok := cja.cache.Get(tokenString); ok {
		return cached, nil
	}

	claims, err := cja.delegate.Authenticate(r)
	if err != nil {
		return nil, err
	}

	cja.cache.Set(tokenString, claims)
	return claims, nil
}
