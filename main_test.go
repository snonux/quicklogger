package main

import (
	"errors"
	"fmt"
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

func TestLogEntryHandlesEdgeCases(t *testing.T) {
	tests := []struct {
		name string
		text string
	}{
		{
			name: "empty",
			text: "",
		},
		{
			name: "very long",
			text: strings.Repeat("x", maxTextLength+1),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			dir := t.TempDir()

			if err := logEntry(dir, tc.text); err != nil {
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
			if string(content) != tc.text {
				t.Fatalf("expected %d bytes, got %d", len(tc.text), len(content))
			}
		})
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

func TestHandleSharedTextLoadAutoLogSuccessRemovesCache(t *testing.T) {
	cacheDir := t.TempDir()
	cachePath := filepath.Join(cacheDir, "quicklogger-shared.txt")
	if err := os.WriteFile(cachePath, []byte("hello"), 0o644); err != nil {
		t.Fatalf("writing cache file: %v", err)
	}

	var infoTitle, infoMessage string
	var resetCalls, clearCalls int
	handleSharedTextLoad(
		"hello",
		true,
		cacheDir,
		func(string) {
			t.Fatal("prefill should not be called in auto-log mode")
		},
		func() {
			t.Fatal("focus should not be called in auto-log mode")
		},
		func() {
			resetCalls++
		},
		func() {
			clearCalls++
			if err := os.Remove(cachePath); err != nil && !errors.Is(err, os.ErrNotExist) {
				t.Fatalf("removing cache file: %v", err)
			}
		},
		func(dir, text string) error {
			if dir != cacheDir {
				t.Fatalf("expected dir %q, got %q", cacheDir, dir)
			}
			if text != "hello" {
				t.Fatalf("expected text %q, got %q", "hello", text)
			}
			return nil
		},
		func(title, message string) {
			infoTitle = title
			infoMessage = message
		},
		func(err error) {
			t.Fatalf("unexpected auto-log error: %v", err)
		},
	)

	if resetCalls != 1 {
		t.Fatalf("expected resetInput to be called once, got %d", resetCalls)
	}
	if clearCalls != 1 {
		t.Fatalf("expected cache cleanup once, got %d", clearCalls)
	}
	if infoTitle != "Logged" || infoMessage != "Shared text has been logged." {
		t.Fatalf("unexpected info dialog: %q / %q", infoTitle, infoMessage)
	}
	if _, err := os.Stat(cachePath); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("expected cache file to be removed, stat err=%v", err)
	}
}

func TestHandleSharedTextLoadAutoLogFailureKeepsCache(t *testing.T) {
	cacheDir := t.TempDir()
	cachePath := filepath.Join(cacheDir, "quicklogger-shared.txt")
	if err := os.WriteFile(cachePath, []byte("hello"), 0o644); err != nil {
		t.Fatalf("writing cache file: %v", err)
	}

	logErr := fmt.Errorf("boom")
	var errorCalled bool
	var clearCalls int
	handleSharedTextLoad(
		"hello",
		true,
		cacheDir,
		func(string) {
			t.Fatal("prefill should not be called in auto-log mode")
		},
		func() {
			t.Fatal("focus should not be called in auto-log mode")
		},
		func() {
			t.Fatal("resetInput should not be called on auto-log failure")
		},
		func() {
			clearCalls++
			if err := os.Remove(cachePath); err != nil {
				t.Fatalf("unexpected cleanup on failure: %v", err)
			}
		},
		func(string, string) error {
			return logErr
		},
		func(string, string) {
			t.Fatal("info dialog should not be shown on failure")
		},
		func(err error) {
			if !errors.Is(err, logErr) {
				t.Fatalf("expected log error %v, got %v", logErr, err)
			}
			errorCalled = true
		},
	)

	if !errorCalled {
		t.Fatal("expected error dialog callback")
	}
	if clearCalls != 0 {
		t.Fatalf("expected cache to remain on failure, cleanup calls=%d", clearCalls)
	}
	if _, err := os.Stat(cachePath); err != nil {
		t.Fatalf("expected cache file to remain, stat err=%v", err)
	}
}
