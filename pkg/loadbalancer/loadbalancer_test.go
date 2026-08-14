package loadbalancer

import (
	"net/url"
	"testing"
)

func mustURL(t *testing.T, s string) *url.URL {
	t.Helper()
	u, err := url.Parse(s)
	if err != nil {
		t.Fatalf("failed to parse %s: %v", s, err)
	}
	return u
}

func TestRoundRobin_Cycles(t *testing.T) {
	backends := []*url.URL{
		mustURL(t, "http://a:80"),
		mustURL(t, "http://b:80"),
		mustURL(t, "http://c:80"),
	}
	rr := NewRoundRobin(backends)

	seen := map[string]int{}
	for i := 0; i < 6; i++ {
		b := rr.Next()
		seen[b.Host]++
	}

	for _, host := range []string{"a:80", "b:80", "c:80"} {
		if seen[host] != 2 {
			t.Errorf("expected %s hit twice, got %d", host, seen[host])
		}
	}
}

func TestRoundRobin_Empty(t *testing.T) {
	rr := NewRoundRobin(nil)
	if b := rr.Next(); b != nil {
		t.Error("expected nil for empty backends")
	}
}

func TestWeightedRoundRobin_Distribution(t *testing.T) {
	backends := []*url.URL{
		mustURL(t, "http://big:80"),
		mustURL(t, "http://small:80"),
	}
	wrr := NewWeightedRoundRobin(backends, []int{3, 1})

	seen := map[string]int{}
	for i := 0; i < 8; i++ {
		seen[wrr.Next().Host]++
	}

	if seen["big:80"] != 6 || seen["small:80"] != 2 {
		t.Errorf("expected 6:2 distribution, got %v", seen)
	}
}

func TestWeightedRoundRobin_ZeroWeights(t *testing.T) {
	backends := []*url.URL{mustURL(t, "http://a:80")}
	wrr := NewWeightedRoundRobin(backends, []int{0})

	if b := wrr.Next(); b == nil || b.Host != "a:80" {
		t.Error("expected first backend when total weight is 0")
	}
}

func TestHealthAwareRoundRobin_MarkUnhealthy(t *testing.T) {
	backends := []*url.URL{
		mustURL(t, "http://a:80"),
		mustURL(t, "http://b:80"),
	}
	hr := NewHealthAwareRoundRobin(backends)

	hr.MarkUnhealthy(0)

	if hr.HealthyCount() != 1 {
		t.Errorf("expected 1 healthy, got %d", hr.HealthyCount())
	}

	for i := 0; i < 5; i++ {
		if b := hr.Next(); b.Host != "b:80" {
			t.Errorf("expected b:80 (healthy), got %s", b.Host)
		}
	}
}

func TestHealthAwareRoundRobin_AllUnhealthy(t *testing.T) {
	backends := []*url.URL{mustURL(t, "http://a:80")}
	hr := NewHealthAwareRoundRobin(backends)
	hr.MarkUnhealthy(0)

	if b := hr.Next(); b != nil {
		t.Error("expected nil when all backends unhealthy")
	}
}

func TestHealthAwareRoundRobin_Recover(t *testing.T) {
	backends := []*url.URL{
		mustURL(t, "http://a:80"),
		mustURL(t, "http://b:80"),
	}
	hr := NewHealthAwareRoundRobin(backends)
	hr.MarkUnhealthy(0)
	hr.MarkHealthy(0)

	if hr.HealthyCount() != 2 {
		t.Errorf("expected 2 healthy after recovery, got %d", hr.HealthyCount())
	}
}
