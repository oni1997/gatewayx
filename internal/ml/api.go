package ml

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/oni1997/gatewayx/internal/history"
	"github.com/oni1997/gatewayx/internal/metrics"
)

type AnalysisService struct {
	collector *metrics.Collector
	history   *history.Buffer
	scanner   *SecurityScanner
}

func NewAnalysisService(collector *metrics.Collector, history *history.Buffer) *AnalysisService {
	return &AnalysisService{
		collector: collector,
		history:   history,
		scanner:   NewSecurityScanner(),
	}
}

func (svc *AnalysisService) SecurityHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		entries := svc.history.Snapshot()
		report := svc.scanner.Scan(entries)
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(report)
	})
}

func (svc *AnalysisService) BottlenecksHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		report := AnalyzeBottlenecks(svc.collector)
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(report)
	})
}

func (svc *AnalysisService) RecommendationsHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		uptime := time.Since(svc.collector.Uptime)
		report := GenerateRecommendations(svc.collector, uptime)
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(report)
	})
}

func (svc *AnalysisService) FullReportHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		entries := svc.history.Snapshot()
		security := svc.scanner.Scan(entries)
		bottlenecks := AnalyzeBottlenecks(svc.collector)
		uptime := time.Since(svc.collector.Uptime)
		recommendations := GenerateRecommendations(svc.collector, uptime)

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"security":        security,
			"bottlenecks":     bottlenecks,
			"recommendations": recommendations,
		})
	})
}
