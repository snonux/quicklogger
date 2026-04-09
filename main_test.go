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

func TestPrepareSharedTextLoadSkipsWhitespaceOnly(t *testing.T) {
	mode, text, ok := prepareSharedTextLoad("  \n\t  ", false)
	if ok {
		t.Fatal("expected whitespace-only text to be skipped")
	}
	if mode != sharedTextLoadPrefill {
		t.Fatalf("expected prefill mode default, got %v", mode)
	}
	if text != "" {
		t.Fatalf("expected empty text, got %q", text)
	}
}

func TestPrepareSharedTextLoadPrefillMode(t *testing.T) {
	mode, text, ok := prepareSharedTextLoad("hello", false)
	if !ok {
		t.Fatal("expected shared text to be accepted")
	}
	if mode != sharedTextLoadPrefill {
		t.Fatalf("expected prefill mode, got %v", mode)
	}
	if text != "hello" {
		t.Fatalf("expected original text, got %q", text)
	}
}

func TestPrepareSharedTextLoadAutoLogMode(t *testing.T) {
	mode, text, ok := prepareSharedTextLoad("hello", true)
	if !ok {
		t.Fatal("expected shared text to be accepted")
	}
	if mode != sharedTextLoadAutoLog {
		t.Fatalf("expected auto-log mode, got %v", mode)
	}
	if text != "hello" {
		t.Fatalf("expected original text, got %q", text)
	}
}

func TestPrepareSharedTextLoadAllowsLongText(t *testing.T) {
	text := strings.Repeat("x", maxTextLength+1)
	mode, gotText, ok := prepareSharedTextLoad(text, true)
	if !ok {
		t.Fatal("expected long shared text to be accepted")
	}
	if mode != sharedTextLoadAutoLog {
		t.Fatalf("expected auto-log mode, got %v", mode)
	}
	if gotText != text {
		t.Fatalf("expected original text to be preserved, got %d bytes", len(gotText))
	}
}
