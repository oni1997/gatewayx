package healthcheck

import (
	"net/http"
	"net/url"
	"time"

	"github.com/oni1997/gatewayx/pkg/loadbalancer"
)

type Monitor struct {
	client    *http.Client
	balancer  *loadbalancer.HealthAwareRoundRobin
	path      string
	interval  time.Duration
	timeout   time.Duration
	healthy   int
	unhealthy int
	stopCh    chan struct{}
}

type Config struct {
	Path      string
	Interval  time.Duration
	Timeout   time.Duration
	Healthy   int
	Unhealthy int
}

func NewMonitor(balancer *loadbalancer.HealthAwareRoundRobin, cfg Config) *Monitor {
	if cfg.Path == "" {
		cfg.Path = "/health"
	}
	if cfg.Interval <= 0 {
		cfg.Interval = 10 * time.Second
	}
	if cfg.Timeout <= 0 {
		cfg.Timeout = 5 * time.Second
	}
	if cfg.Healthy <= 0 {
		cfg.Healthy = 3
	}
	if cfg.Unhealthy <= 0 {
		cfg.Unhealthy = 3
	}

	return &Monitor{
		client:    &http.Client{Timeout: cfg.Timeout},
		balancer:  balancer,
		path:      cfg.Path,
		interval:  cfg.Interval,
		timeout:   cfg.Timeout,
		healthy:   cfg.Healthy,
		unhealthy: cfg.Unhealthy,
		stopCh:    make(chan struct{}),
	}
}

func (m *Monitor) Start() {
	go m.loop()
}

func (m *Monitor) Stop() {
	close(m.stopCh)
}

func (m *Monitor) loop() {
	ticker := time.NewTicker(m.interval)
	defer ticker.Stop()

	failures := make(map[int]int)
	successes := make(map[int]int)

	for {
		select {
		case <-m.stopCh:
			return
		case <-ticker.C:
			m.check(failures, successes)
		}
	}
}

func (m *Monitor) check(failures, successes map[int]int) {
	backends := m.balancer.Backends()

	for i, backend := range backends {
		ok := m.ping(backend)

		if ok {
			successes[i]++
			failures[i] = 0
			if successes[i] >= m.healthy {
				m.balancer.MarkHealthy(i)
			}
		} else {
			failures[i]++
			successes[i] = 0
			if failures[i] >= m.unhealthy {
				m.balancer.MarkUnhealthy(i)
			}
		}
	}
}

func (m *Monitor) ping(backend *url.URL) bool {
	u := *backend
	u.Path = m.path

	resp, err := m.client.Get(u.String())
	if err != nil {
		return false
	}
	defer func() { _ = resp.Body.Close() }()

	return resp.StatusCode >= 200 && resp.StatusCode < 400
}
