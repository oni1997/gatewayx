package ratelimit

import (
	"sync"
	"time"
)

type SlidingWindow struct {
	mu       sync.Mutex
	rate     float64
	window   time.Duration
	counters map[int64]int
}

func NewSlidingWindow(rate float64, burst int) *SlidingWindow {
	_ = burst
	return &SlidingWindow{
		rate:     rate,
		window:   time.Second,
		counters: make(map[int64]int),
	}
}

func (sw *SlidingWindow) Allow() bool {
	sw.mu.Lock()
	defer sw.mu.Unlock()

	now := time.Now().UnixNano()
	bucket := now / int64(sw.window)

	sw.prune(bucket)

	maxPerWindow := int(sw.rate)

	if sw.counters[bucket] >= maxPerWindow {
		return false
	}

	sw.counters[bucket]++
	return true
}

func (sw *SlidingWindow) AllowN(n int) bool {
	sw.mu.Lock()
	defer sw.mu.Unlock()

	now := time.Now().UnixNano()
	bucket := now / int64(sw.window)

	sw.prune(bucket)

	maxPerWindow := int(sw.rate)

	if sw.counters[bucket]+n > maxPerWindow {
		return false
	}

	sw.counters[bucket] += n
	return true
}

func (sw *SlidingWindow) prune(currentBucket int64) {
	cutoff := currentBucket - 2
	for bucket := range sw.counters {
		if bucket < cutoff {
			delete(sw.counters, bucket)
		}
	}
}

func (sw *SlidingWindow) Rate() float64 { return sw.rate }
