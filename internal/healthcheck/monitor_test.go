package healthcheck

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/oni1997/gatewayx/pkg/loadbalancer"
)

func TestMonitor_MarksUnhealthy(t *testing.T) {
	healthy := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
	}))
	defer healthy.Close()

	unhealthy := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(500)
	}))
	defer unhealthy.Close()

	healthyURL, _ := url.Parse(healthy.URL)
	unhealthyURL, _ := url.Parse(unhealthy.URL)

	lb := loadbalancer.NewHealthAwareRoundRobin([]*url.URL{healthyURL, unhealthyURL})

	monitor := NewMonitor(lb, Config{
		Path:      "/",
		Interval:  50 * time.Millisecond,
		Timeout:   time.Second,
		Healthy:   1,
		Unhealthy: 1,
	})
	monitor.Start()
	defer monitor.Stop()

	time.Sleep(150 * time.Millisecond)

	if lb.HealthyCount() != 1 {
		t.Errorf("expected 1 healthy backend, got %d", lb.HealthyCount())
	}

	monitor.check(make(map[int]int), make(map[int]int))
}

func TestMonitor_Ping(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
	}))
	defer server.Close()

	serverURL, _ := url.Parse(server.URL)

	lb := loadbalancer.NewHealthAwareRoundRobin([]*url.URL{serverURL})
	monitor := NewMonitor(lb, Config{Path: "/", Healthy: 1, Unhealthy: 1})

	if !monitor.ping(serverURL) {
		t.Error("expected ping to succeed")
	}

	badURL, _ := url.Parse("http://127.0.0.1:1")
	if monitor.ping(badURL) {
		t.Error("expected ping to fail for unreachable backend")
	}
}

func TestHealthAwareRoundRobin_SkipsUnhealthy(t *testing.T) {
	u1, _ := url.Parse("http://backend1:80")
	u2, _ := url.Parse("http://backend2:80")
	u3, _ := url.Parse("http://backend3:80")

	lb := loadbalancer.NewHealthAwareRoundRobin([]*url.URL{u1, u2, u3})

	lb.MarkUnhealthy(1)
	lb.MarkUnhealthy(2)

	if lb.HealthyCount() != 1 {
		t.Errorf("expected 1 healthy, got %d", lb.HealthyCount())
	}

	for i := 0; i < 5; i++ {
		backend := lb.Next()
		if backend == nil {
			t.Fatal("expected a backend")
		}
		if backend.Host != "backend1:80" {
			t.Errorf("expected backend1, got %s", backend.Host)
		}
	}
}

func TestHealthAwareRoundRobin_AllUnhealthy(t *testing.T) {
	u1, _ := url.Parse("http://backend1:80")

	lb := loadbalancer.NewHealthAwareRoundRobin([]*url.URL{u1})
	lb.MarkUnhealthy(0)

	if backend := lb.Next(); backend != nil {
		t.Error("expected nil when all backends unhealthy")
	}
}
