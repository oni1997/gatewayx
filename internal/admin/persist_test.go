package admin

import (
	"path/filepath"
	"testing"
	"time"
)

func TestPersistence_KeysAndCerts(t *testing.T) {
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "test.db")

	store, err := NewStoreWithPersistence(dbPath)
	if err != nil {
		t.Fatalf("failed to create persistent store: %v", err)
	}

	key, fullKey := store.CreateKey("test-key", "owner-1")
	if key.ID == "" {
		t.Fatal("expected key ID")
	}

	store.AddCertificate("api.example.com", "Lets Encrypt", time.Now().Add(90*24*time.Hour))

	store2, err := NewStoreWithPersistence(dbPath)
	if err != nil {
		t.Fatalf("failed to reopen store: %v", err)
	}

	if _, ok := store2.ValidateKey(fullKey); !ok {
		t.Error("expected key to persist across store instances")
	}

	if certs := store2.ListCertificates(); len(certs) != 1 {
		t.Errorf("expected 1 cert to persist, got %d", len(certs))
	}
}

func TestPersistence_RevokeKey(t *testing.T) {
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "test.db")

	store, err := NewStoreWithPersistence(dbPath)
	if err != nil {
		t.Fatalf("failed: %v", err)
	}

	key, _ := store.CreateKey("temp", "owner")
	store.RevokeKey(key.ID)

	store2, _ := NewStoreWithPersistence(dbPath)
	if len(store2.ListKeys()) != 0 {
		t.Error("expected revoked key to not persist")
	}
}
