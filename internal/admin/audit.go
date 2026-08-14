package admin

import (
	"log/slog"
	"sync"
	"time"
)

type AuditLog struct {
	mu      sync.Mutex
	logger  *slog.Logger
	entries []AuditEntry
}

type AuditEntry struct {
	Timestamp time.Time `json:"timestamp"`
	Action    string    `json:"action"`
	Resource  string    `json:"resource"`
	Detail    string    `json:"detail"`
}

func NewAuditLog(logger *slog.Logger) *AuditLog {
	if logger == nil {
		logger = slog.Default()
	}
	return &AuditLog{
		logger:  logger.With("component", "audit"),
		entries: make([]AuditEntry, 0),
	}
}

func (a *AuditLog) Record(action, resource, detail string) {
	entry := AuditEntry{
		Timestamp: time.Now(),
		Action:    action,
		Resource:  resource,
		Detail:    detail,
	}

	a.mu.Lock()
	a.entries = append(a.entries, entry)
	if len(a.entries) > 1000 {
		a.entries = a.entries[len(a.entries)-1000:]
	}
	a.mu.Unlock()

	a.logger.Info("audit",
		"action", action,
		"resource", resource,
		"detail", detail,
	)
}

func (a *AuditLog) Snapshot() []AuditEntry {
	a.mu.Lock()
	defer a.mu.Unlock()
	result := make([]AuditEntry, len(a.entries))
	copy(result, a.entries)
	return result
}

func (a *AuditLog) Clear() {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.entries = make([]AuditEntry, 0)
}
