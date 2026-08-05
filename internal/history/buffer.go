package history

import (
	"encoding/json"
	"net/http"
	"sync"
	"time"
)

type Entry struct {
	Timestamp  time.Time     `json:"timestamp"`
	TraceID    string        `json:"trace_id"`
	SpanID     string        `json:"span_id"`
	Method     string        `json:"method"`
	Path       string        `json:"path"`
	Host       string        `json:"host"`
	RemoteAddr string        `json:"remote_addr"`
	Status     int           `json:"status"`
	Duration   time.Duration `json:"duration_ms"`
	BytesSent  int64         `json:"bytes_sent"`
}

type Buffer struct {
	mu       sync.RWMutex
	entries  []Entry
	capacity int
	pos      int
	full     bool
}

func NewBuffer(capacity int) *Buffer {
	return &Buffer{
		entries:  make([]Entry, capacity),
		capacity: capacity,
	}
}

func (b *Buffer) Push(entry Entry) {
	b.mu.Lock()
	defer b.mu.Unlock()

	b.entries[b.pos] = entry
	b.pos++
	if b.pos >= b.capacity {
		b.pos = 0
		b.full = true
	}
}

func (b *Buffer) Snapshot() []Entry {
	b.mu.RLock()
	defer b.mu.RUnlock()

	if !b.full {
		result := make([]Entry, b.pos)
		copy(result, b.entries[:b.pos])
		return result
	}

	result := make([]Entry, b.capacity)
	copy(result, b.entries[b.pos:])
	copy(result[b.capacity-b.pos:], b.entries[:b.pos])
	return result
}

func (b *Buffer) Handler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		entries := b.Snapshot()
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(entries)
	})
}
