package ratelimit

import (
	"net/http"
	"sync"
	"time"
)

type Limiter interface {
	Allow() bool
	AllowN(n int) bool
}

type MemoryStore struct {
	mu       sync.RWMutex
	limiters map[string]Limiter
	config   Config
}

func NewMemoryStore(cfg Config) *MemoryStore {
	return &MemoryStore{
		limiters: make(map[string]Limiter),
		config:   cfg,
	}
}

func (ms *MemoryStore) GetLimiter(key string) Limiter {
	ms.mu.RLock()
	lim, ok := ms.limiters[key]
	ms.mu.RUnlock()

	if ok {
		return lim
	}

	ms.mu.Lock()
	defer ms.mu.Unlock()

	if lim, ok = ms.limiters[key]; ok {
		return lim
	}

	lim = ms.createLimiter()
	ms.limiters[key] = lim
	return lim
}

func (ms *MemoryStore) createLimiter() Limiter {
	switch ms.config.Strategy {
	case "sliding_window":
		return NewSlidingWindow(ms.config.Rate, ms.config.Burst)
	default:
		return NewTokenBucket(ms.config.Rate, ms.config.Burst)
	}
}

func (ms *MemoryStore) Allow(key string) bool {
	return ms.GetLimiter(key).Allow()
}

func (ms *MemoryStore) AllowN(key string, n int) bool {
	return ms.GetLimiter(key).AllowN(n)
}

func (ms *MemoryStore) Cleanup(interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for range ticker.C {
		ms.mu.Lock()
		for k, lim := range ms.limiters {
			switch v := lim.(type) {
			case *TokenBucket:
				v.mu.Lock()
				if v.tokens >= float64(v.burst) {
					delete(ms.limiters, k)
				}
				v.mu.Unlock()
			case *SlidingWindow:
				v.mu.Lock()
				active := false
				for _, count := range v.counters {
					if count > 0 {
						active = true
						break
					}
				}
				if !active {
					delete(ms.limiters, k)
				}
				v.mu.Unlock()
			}
		}
		ms.mu.Unlock()
	}
}

func DefaultKeyExtractor(r *http.Request) string {
	return r.RemoteAddr
}
