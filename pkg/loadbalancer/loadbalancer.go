package loadbalancer

import (
	"net/url"
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
