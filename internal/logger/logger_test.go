package logger

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/oni1997/gatewayx/internal/config"
)

func TestNew_JSONFormat(t *testing.T) {
	var buf bytes.Buffer
	_ = buf
	log := New(config.LoggingConfig{Level: "info", Format: "json", Output: "stdout"})

	if log == nil {
		t.Fatal("expected non-nil logger")
	}
}

func TestNew_TextFormat(t *testing.T) {
	log := New(config.LoggingConfig{Level: "debug", Format: "text", Output: "stdout"})
	if log == nil {
		t.Fatal("expected non-nil logger")
	}
}

func TestNew_FileOutput(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.log")

	log := New(config.LoggingConfig{
		Level:  "error",
		Format: "text",
		Output: "file",
		File:   path,
	})
	if log == nil {
		t.Fatal("expected non-nil logger")
	}

	log.Error("test message")

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read log file: %v", err)
	}
	if !strings.Contains(string(data), "test message") {
		t.Errorf("expected log file to contain message, got %s", data)
	}
}

func TestNew_DefaultLevel(t *testing.T) {
	log := New(config.LoggingConfig{Format: "json", Output: "stdout"})
	if log == nil {
		t.Fatal("expected non-nil logger")
	}
}
