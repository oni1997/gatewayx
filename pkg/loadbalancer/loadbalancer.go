package loadbalancer

import (
	"net/url"
	"sync"
	"sync/atomic"
)

type Balancer interface {
	Next() *url.URL
}

type RoundRobin struct {
	backends []*url.URL
	counter  atomic.Uint64
}

func NewRoundRobin(backends []*url.URL) *RoundRobin {
	return &RoundRobin{backends: backends}
}

func (rr *RoundRobin) Next() *url.URL {
	if len(rr.backends) == 0 {
		return nil
	}
	idx := rr.counter.Add(1) % uint64(len(rr.backends))
	return rr.backends[idx]
}

type HealthAwareRoundRobin struct {
	mu       sync.RWMutex
	backends []*url.URL
	healthy  []bool
	counter  atomic.Uint64
}

func NewHealthAwareRoundRobin(backends []*url.URL) *HealthAwareRoundRobin {
	healthy := make([]bool, len(backends))
	for i := range healthy {
		healthy[i] = true
	}
	return &HealthAwareRoundRobin{
		backends: backends,
		healthy:  healthy,
	}
}

func (hr *HealthAwareRoundRobin) Next() *url.URL {
	hr.mu.RLock()
	defer hr.mu.RUnlock()

	var healthyIdx []int
	for i, h := range hr.healthy {
		if h {
			healthyIdx = append(healthyIdx, i)
		}
	}

	if len(healthyIdx) == 0 {
		return nil
	}

	idx := healthyIdx[hr.counter.Add(1)%uint64(len(healthyIdx))]
	return hr.backends[idx]
}

func (hr *HealthAwareRoundRobin) MarkHealthy(index int) {
	hr.mu.Lock()
	defer hr.mu.Unlock()
	if index >= 0 && index < len(hr.healthy) {
		hr.healthy[index] = true
	}
}

func (hr *HealthAwareRoundRobin) MarkUnhealthy(index int) {
	hr.mu.Lock()
	defer hr.mu.Unlock()
	if index >= 0 && index < len(hr.healthy) {
		hr.healthy[index] = false
	}
}

func (hr *HealthAwareRoundRobin) HealthyCount() int {
	hr.mu.RLock()
	defer hr.mu.RUnlock()
	count := 0
	for _, h := range hr.healthy {
		if h {
			count++
		}
	}
	return count
}

func (hr *HealthAwareRoundRobin) Backends() []*url.URL {
	hr.mu.RLock()
	defer hr.mu.RUnlock()
	return append([]*url.URL(nil), hr.backends...)
}

type WeightedRoundRobin struct {
	backends []*url.URL
	weights  []int
	counter  atomic.Uint64
}

func NewWeightedRoundRobin(backends []*url.URL, weights []int) *WeightedRoundRobin {
	return &WeightedRoundRobin{
		backends: backends,
		weights:  weights,
	}
}

func (wrr *WeightedRoundRobin) Next() *url.URL {
	if len(wrr.backends) == 0 {
		return nil
	}

	totalWeight := 0
	for _, w := range wrr.weights {
		totalWeight += w
	}
	if totalWeight == 0 {
		return wrr.backends[0]
	}

	pos := wrr.counter.Add(1) % uint64(totalWeight)
	var accumulated int
	for i, w := range wrr.weights {
		accumulated += w
		if uint64(accumulated) > pos {
			return wrr.backends[i]
		}
	}
	return wrr.backends[len(wrr.backends)-1]
}
