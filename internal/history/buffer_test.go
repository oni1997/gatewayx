package history

import (
	"encoding/json"
	"net/http/httptest"
	"testing"
	"time"
)

func TestBuffer_PushAndSnapshot(t *testing.T) {
	b := NewBuffer(5)

	b.Push(Entry{Method: "GET", Path: "/a"})
	b.Push(Entry{Method: "POST", Path: "/b"})
	b.Push(Entry{Method: "PUT", Path: "/c"})

	entries := b.Snapshot()
	if len(entries) != 3 {
		t.Fatalf("expected 3 entries, got %d", len(entries))
	}
	if entries[0].Method != "GET" {
		t.Errorf("expected GET, got %s", entries[0].Method)
	}
	if entries[2].Method != "PUT" {
		t.Errorf("expected PUT, got %s", entries[2].Method)
	}
}

func TestBuffer_Overflow(t *testing.T) {
	b := NewBuffer(3)

	b.Push(Entry{Method: "1"})
	b.Push(Entry{Method: "2"})
	b.Push(Entry{Method: "3"})
	b.Push(Entry{Method: "4"})

	entries := b.Snapshot()
	if len(entries) != 3 {
		t.Fatalf("expected 3 entries after overflow, got %d", len(entries))
	}
	if entries[0].Method != "2" {
		t.Errorf("expected 2, got %s", entries[0].Method)
	}
	if entries[2].Method != "4" {
		t.Errorf("expected 4, got %s", entries[2].Method)
	}
}

func TestBuffer_Handler(t *testing.T) {
	b := NewBuffer(10)
	b.Push(Entry{
		Timestamp:  time.Now(),
		TraceID:    "abc123",
		Method:     "GET",
		Path:       "/api/test",
		Status:     200,
		Duration:   5 * time.Millisecond,
	})

	req := httptest.NewRequest("GET", "/history", nil)
	rec := httptest.NewRecorder()

	b.Handler().ServeHTTP(rec, req)

	if rec.Code != 200 {
		t.Errorf("expected 200, got %d", rec.Code)
	}

	var entries []Entry
	if err := json.Unmarshal(rec.Body.Bytes(), &entries); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}
	if len(entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(entries))
	}
	if entries[0].TraceID != "abc123" {
		t.Errorf("expected abc123, got %s", entries[0].TraceID)
	}
}

func TestBuffer_Empty(t *testing.T) {
	b := NewBuffer(10)
	entries := b.Snapshot()
	if len(entries) != 0 {
		t.Error("expected empty snapshot")
	}
}
