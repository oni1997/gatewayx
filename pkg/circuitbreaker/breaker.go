package circuitbreaker

import (
	"sync"
	"time"
)

type State int

const (
	StateClosed   State = iota
	StateOpen
	StateHalfOpen
)

func (s State) String() string {
	switch s {
	case StateClosed:
		return "closed"
	case StateOpen:
		return "open"
	case StateHalfOpen:
		return "half-open"
	default:
		return "unknown"
	}
}

type Breaker struct {
	mu                sync.Mutex
	state             State
	failureCount      int
	consecutiveSuccess int
	failureThreshold  int
	successThreshold  int
	openDuration      time.Duration
	lastFailure       time.Time
	lastStateChange   time.Time
	totalFailures     int64
	totalSuccess      int64
}

func New(failureThreshold, successThreshold int, openDuration time.Duration) *Breaker {
	if failureThreshold <= 0 {
		failureThreshold = 5
	}
	if successThreshold <= 0 {
		successThreshold = 3
	}
	if openDuration <= 0 {
		openDuration = 30 * time.Second
	}

	return &Breaker{
		state:            StateClosed,
		failureThreshold: failureThreshold,
		successThreshold: successThreshold,
		openDuration:     openDuration,
		lastStateChange:  time.Now(),
	}
}

func (b *Breaker) Allow() bool {
	b.mu.Lock()
	defer b.mu.Unlock()

	switch b.state {
	case StateClosed:
		return true
	case StateOpen:
		if time.Since(b.lastStateChange) > b.openDuration {
			b.state = StateHalfOpen
			b.lastStateChange = time.Now()
			b.consecutiveSuccess = 0
			return true
		}
		return false
	case StateHalfOpen:
		return true
	default:
		return true
	}
}

func (b *Breaker) Success() {
	b.mu.Lock()
	defer b.mu.Unlock()

	b.totalSuccess++

	switch b.state {
	case StateClosed:
		b.failureCount = 0
	case StateHalfOpen:
		b.consecutiveSuccess++
		if b.consecutiveSuccess >= b.successThreshold {
			b.state = StateClosed
			b.failureCount = 0
			b.lastStateChange = time.Now()
		}
	}
}

func (b *Breaker) Failure() {
	b.mu.Lock()
	defer b.mu.Unlock()

	b.totalFailures++
	b.lastFailure = time.Now()

	switch b.state {
	case StateClosed:
		b.failureCount++
		if b.failureCount >= b.failureThreshold {
			b.state = StateOpen
			b.lastStateChange = time.Now()
		}
	case StateHalfOpen:
		b.state = StateOpen
		b.lastStateChange = time.Now()
	}
}

func (b *Breaker) State() State {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.state
}

func (b *Breaker) Stats() map[string]any {
	b.mu.Lock()
	defer b.mu.Unlock()
	return map[string]any{
		"state":             b.state.String(),
		"failure_count":     b.failureCount,
		"total_failures":    b.totalFailures,
		"total_success":     b.totalSuccess,
		"last_failure":      b.lastFailure.Format(time.RFC3339),
		"last_state_change": b.lastStateChange.Format(time.RFC3339),
	}
}
