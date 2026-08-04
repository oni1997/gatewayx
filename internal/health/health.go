package health

import (
	"encoding/json"
	"net/http"
	"sync"
	"time"
)

type Status string

const (
	StatusHealthy   Status = "healthy"
	StatusDegraded  Status = "degraded"
	StatusUnhealthy Status = "unhealthy"
)

type Checker struct {
	mu       sync.RWMutex
	checks   map[string]CheckFunc
	status   Status
	started  time.Time
}

type CheckFunc func() error

type Report struct {
	Status    Status            `json:"status"`
	Uptime    string            `json:"uptime"`
	Checks    map[string]string `json:"checks"`
	Timestamp string            `json:"timestamp"`
}

func New() *Checker {
	return &Checker{
		checks:  make(map[string]CheckFunc),
		started: time.Now(),
	}
}

func (c *Checker) Register(name string, fn CheckFunc) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.checks[name] = fn
}

func (c *Checker) Handler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		c.mu.RLock()
		checks := make(map[string]string, len(c.checks))
		healthy := true
		for name, fn := range c.checks {
			if err := fn(); err != nil {
				checks[name] = "unhealthy: " + err.Error()
				healthy = false
			} else {
				checks[name] = "healthy"
			}
		}
		c.mu.RUnlock()

		status := StatusHealthy
		if !healthy {
			status = StatusUnhealthy
		}

		report := Report{
			Status:    status,
			Uptime:    time.Since(c.started).String(),
			Checks:    checks,
			Timestamp: time.Now().UTC().Format(time.RFC3339),
		}

		w.Header().Set("Content-Type", "application/json")
		if status == StatusUnhealthy {
			w.WriteHeader(http.StatusServiceUnavailable)
		}
		json.NewEncoder(w).Encode(report)
	}
}
