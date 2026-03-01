package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLogEntryCreatesFile(t *testing.T) {
	dir := t.TempDir()
	text := "hello world"

	if err := logEntry(dir, text); err != nil {
		t.Fatalf("logEntry returned error: %v", err)
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("reading dir: %v", err)
	}
	if len(entries) != 1 {
		t.Fatalf("expected 1 file, got %d", len(entries))
	}

	name := entries[0].Name()
	if !strings.HasPrefix(name, "ql-") || !strings.HasSuffix(name, ".md") {
		t.Errorf("unexpected filename pattern: %s", name)
	}

	content, err := os.ReadFile(filepath.Join(dir, name))
	if err != nil {
		t.Fatalf("reading file: %v", err)
	}
	if string(content) != text {
		t.Errorf("expected %q, got %q", text, string(content))
	}
}

func TestLogEntryInvalidDir(t *testing.T) {
	err := logEntry("/nonexistent/path/that/should/not/exist", "test")
	if err == nil {
		t.Fatal("expected error for invalid directory, got nil")
	}
}

func TestLogEntryEmptyText(t *testing.T) {
	dir := t.TempDir()
	if err := logEntry(dir, ""); err != nil {
		t.Fatalf("logEntry with empty text returned error: %v", err)
	}

	entries, _ := os.ReadDir(dir)
	if len(entries) != 1 {
		t.Fatalf("expected 1 file, got %d", len(entries))
	}

	content, _ := os.ReadFile(filepath.Join(dir, entries[0].Name()))
	if len(content) != 0 {
		t.Errorf("expected empty file, got %d bytes", len(content))
	}
}
