package ml

import (
	"fmt"
	"sort"
	"time"

	"github.com/oni1997/gatewayx/internal/metrics"
)

type BottleneckReport struct {
	SlowRoutes  []SlowRoute      `json:"slow_routes"`
	Spikes      []LatencySpike   `json:"spikes"`
	Summary     string           `json:"summary"`
}

type SlowRoute struct {
	Name     string  `json:"name"`
	AvgMs    float64 `json:"avg_ms"`
	MaxMs    float64 `json:"max_ms"`
	Count    int64   `json:"count"`
	Severity string  `json:"severity"`
}

type LatencySpike struct {
	Route     string  `json:"route"`
	CurrentMs float64 `json:"current_ms"`
	AvgMs     float64 `json:"avg_ms"`
	Ratio     float64 `json:"ratio"`
}

type RecommendationReport struct {
	RateLimits  []RateLimitSuggestion `json:"rate_limits"`
	CacheHints  []CacheSuggestion     `json:"cache_hints"`
	Summary     string                `json:"summary"`
}

type RateLimitSuggestion struct {
	Route       string  `json:"route"`
	CurrentRPS  float64 `json:"current_rps"`
	SuggestedRPS float64 `json:"suggested_rps"`
	Reason      string  `json:"reason"`
}

type CacheSuggestion struct {
	Route         string  `json:"route"`
	ReadRatio     float64 `json:"read_ratio"`
	AvgLatencyMs  float64 `json:"avg_latency_ms"`
	Reason        string  `json:"reason"`
}

func AnalyzeBottlenecks(c *metrics.Collector) BottleneckReport {
	report := BottleneckReport{}

	snapshot := c.Snapshot()

	for _, route := range snapshot.Routes {
		severity := "normal"
		if route.AvgMs > 500 {
			severity = "critical"
		} else if route.AvgMs > 200 {
			severity = "warning"
		} else if route.AvgMs > 100 {
			severity = "elevated"
		}

		if severity != "normal" {
			report.SlowRoutes = append(report.SlowRoutes, SlowRoute{
				Name:     route.Name,
				AvgMs:    route.AvgMs,
				MaxMs:    route.MaxMs,
				Count:    route.Count,
				Severity: severity,
			})
		}
	}

	report.Spikes = detectSpikes(c)

	sort.Slice(report.SlowRoutes, func(i, j int) bool {
		return report.SlowRoutes[i].AvgMs > report.SlowRoutes[j].AvgMs
	})

	if len(report.SlowRoutes) == 0 && len(report.Spikes) == 0 {
		report.Summary = "No bottlenecks detected. All routes performing normally."
	} else if len(report.Spikes) > 0 {
		report.Summary = "Latency spikes detected on one or more routes. Investigate backend health."
	} else {
		report.Summary = "Slow routes detected. Review backend performance or add caching."
	}

	return report
}

func detectSpikes(c *metrics.Collector) []LatencySpike {
	var spikes []LatencySpike
	snapshot := c.Snapshot()

	for _, route := range snapshot.Routes {
		ratio := float64(0)
		if route.AvgMs > 0 && route.Count > 10 {
			ratio = route.LastMs / route.AvgMs
		}

		if ratio > 3.0 && route.LastMs > 50 {
			spikes = append(spikes, LatencySpike{
				Route:     route.Name,
				CurrentMs: route.LastMs,
				AvgMs:     route.AvgMs,
				Ratio:     ratio,
			})
		}
	}

	return spikes
}

func GenerateRecommendations(c *metrics.Collector, uptime time.Duration) RecommendationReport {
	report := RecommendationReport{}
	snapshot := c.Snapshot()

	for _, route := range snapshot.Routes {
		if route.Count < 10 {
			continue
		}

		rps := float64(route.Count) / uptime.Seconds()

		if rps > 5 {
			suggested := rps * 1.5
			report.RateLimits = append(report.RateLimits, RateLimitSuggestion{
				Route:        route.Name,
				CurrentRPS:   rps,
				SuggestedRPS: suggested,
				Reason:       "high traffic route, consider rate limiting at " + formatRPS(suggested) + " req/s",
			})
		}

		if route.AvgMs > 50 && rps > 2 {
			report.CacheHints = append(report.CacheHints, CacheSuggestion{
				Route:        route.Name,
				ReadRatio:    0.8,
				AvgLatencyMs: route.AvgMs,
				Reason:       "high latency with moderate traffic, consider response caching",
			})
		}
	}

	if len(report.RateLimits) == 0 && len(report.CacheHints) == 0 {
		report.Summary = "No recommendations at this time. Traffic patterns are within normal range."
	} else {
		report.Summary = "Recommendations generated based on current traffic analysis."
	}

	return report
}

func formatRPS(rps float64) string {
	return fmt.Sprintf("%.0f req/s", rps)
}
