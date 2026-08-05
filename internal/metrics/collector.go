package metrics

import (
	"sync"
	"sync/atomic"
	"time"
)

type Collector struct {
	Requests      atomic.Int64
	ActiveConns   atomic.Int64
	BytesSent     atomic.Int64
	BytesReceived atomic.Int64
	StatusCodes   sync.Map
	RouteLatency  sync.Map
	Uptime        time.Time
}

type LatencyStats struct {
	mu              sync.Mutex
	Count           int64
	TotalMs         float64
	MinMs           float64
	MaxMs           float64
	LastMs          float64
	P50Ms           float64
	P95Ms           float64
	P99Ms           float64
	recentDurations []float64
}

func NewCollector() *Collector {
	return &Collector{
		Uptime: time.Now(),
	}
}

func (c *Collector) RecordRequest(route, method string, statusCode int, duration time.Duration, bytesRead, bytesWritten int64) {
	c.Requests.Add(1)
	c.BytesReceived.Add(bytesRead)
	c.BytesSent.Add(bytesWritten)

	key := method + ":" + route
	val, _ := c.StatusCodes.LoadOrStore(key, &sync.Map{})
	statusMap := val.(*sync.Map)
	count, _ := statusMap.LoadOrStore(statusCode, new(int64))
	atomic.AddInt64(count.(*int64), 1)

	latVal, _ := c.RouteLatency.LoadOrStore(route, &LatencyStats{})
	stats := latVal.(*LatencyStats)
	stats.mu.Lock()
	stats.Count++
	ms := float64(duration.Microseconds()) / 1000.0
	stats.TotalMs += ms
	stats.LastMs = ms
	if stats.MinMs == 0 || ms < stats.MinMs {
		stats.MinMs = ms
	}
	if ms > stats.MaxMs {
		stats.MaxMs = ms
	}
	stats.recentDurations = append(stats.recentDurations, ms)
	if len(stats.recentDurations) > 1000 {
		stats.recentDurations = stats.recentDurations[1:]
	}
	stats.mu.Unlock()
}

func (c *Collector) GetStatusCodes(routeMethod string) map[int]int64 {
	result := make(map[int]int64)
	val, ok := c.StatusCodes.Load(routeMethod)
	if !ok {
		return result
	}
	statusMap := val.(*sync.Map)
	statusMap.Range(func(k, v any) bool {
		result[k.(int)] = *(v.(*int64))
		return true
	})
	return result
}

func (c *Collector) GetLatency(route string) *LatencyStats {
	val, ok := c.RouteLatency.Load(route)
	if !ok {
		return &LatencyStats{}
	}
	stats := val.(*LatencyStats)
	stats.mu.Lock()
	defer stats.mu.Unlock()

	if stats.Count > 0 {
		stats.P50Ms = stats.recentDurations[len(stats.recentDurations)/2]
		stats.P95Ms = stats.recentDurations[len(stats.recentDurations)*95/100]
		stats.P99Ms = stats.recentDurations[len(stats.recentDurations)*99/100]
	}

	return stats
}

func (c *Collector) Snapshot() CollectorSnapshot {
	snapshot := CollectorSnapshot{
		Requests:      c.Requests.Load(),
		ActiveConns:   c.ActiveConns.Load(),
		BytesSent:     c.BytesSent.Load(),
		BytesReceived: c.BytesReceived.Load(),
		Uptime:        c.Uptime,
		Routes:        make([]RouteSnapshot, 0),
	}

	c.StatusCodes.Range(func(k, v any) bool {
		statusMap := v.(*sync.Map)
		statusMap.Range(func(code, count any) bool {
			return true
		})
		return true
	})

	c.RouteLatency.Range(func(k, v any) bool {
		route := k.(string)
		stats := v.(*LatencyStats)
		stats.mu.Lock()
		avgMs := float64(0)
		if stats.Count > 0 {
			avgMs = stats.TotalMs / float64(stats.Count)
		}
		snapshot.Routes = append(snapshot.Routes, RouteSnapshot{
			Name:   route,
			Count:  stats.Count,
			AvgMs:  avgMs,
			MinMs:  stats.MinMs,
			MaxMs:  stats.MaxMs,
			LastMs: stats.LastMs,
		})
		stats.mu.Unlock()
		return true
	})

	return snapshot
}

type CollectorSnapshot struct {
	Requests      int64           `json:"requests"`
	ActiveConns   int64           `json:"active_connections"`
	BytesSent     int64           `json:"bytes_sent"`
	BytesReceived int64           `json:"bytes_received"`
	Uptime        time.Time       `json:"uptime"`
	Routes        []RouteSnapshot `json:"routes"`
}

type RouteSnapshot struct {
	Name   string  `json:"name"`
	Count  int64   `json:"count"`
	AvgMs  float64 `json:"avg_ms"`
	MinMs  float64 `json:"min_ms"`
	MaxMs  float64 `json:"max_ms"`
	LastMs float64 `json:"last_ms"`
}
