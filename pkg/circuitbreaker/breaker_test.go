package circuitbreaker

import (
	"testing"
	"time"
)

func TestBreaker_ClosedToOpen(t *testing.T) {
	b := New(3, 2, time.Second)

	for i := 0; i < 3; i++ {
		if !b.Allow() {
			t.Fatal("should allow in closed state")
		}
		b.Failure()
	}

	if b.Allow() {
		t.Error("should not allow after threshold reached")
	}
	if b.State() != StateOpen {
		t.Errorf("expected open state, got %s", b.State())
	}
}

func TestBreaker_Success(t *testing.T) {
	b := New(3, 2, 10*time.Millisecond)

	for i := 0; i < 3; i++ {
		b.Allow()
		b.Failure()
	}

	if b.State() != StateOpen {
		t.Fatal("should be open")
	}

	time.Sleep(15 * time.Millisecond)

	if !b.Allow() {
		t.Fatal("should allow in half-open after cooldown")
	}
	b.Success()
	b.Allow()
	b.Success()

	if b.State() != StateClosed {
		t.Errorf("should be closed after successes, got %s", b.State())
	}
}

func TestBreaker_HalfOpenFailure(t *testing.T) {
	b := New(2, 2, 5*time.Millisecond)

	b.Allow()
	b.Failure()
	b.Allow()
	b.Failure()

	time.Sleep(10 * time.Millisecond)

	if !b.Allow() {
		t.Fatal("should allow in half-open")
	}
	b.Failure()

	if b.State() != StateOpen {
		t.Errorf("should be open after half-open failure, got %s", b.State())
	}
}

func TestBreaker_Stats(t *testing.T) {
	b := New(3, 2, time.Second)
	b.Allow()
	b.Success()
	b.Allow()
	b.Failure()

	stats := b.Stats()
	if stats["total_success"].(int64) != 1 {
		t.Errorf("expected 1 success")
	}
	if stats["total_failures"].(int64) != 1 {
		t.Errorf("expected 1 failure")
	}
}
