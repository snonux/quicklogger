//go:build !android

package main

import "testing"

func TestReadSharedFromCacheStub(t *testing.T) {
	txt, err := readSharedFromCache()
	if err != nil {
		t.Fatalf("stub returned error: %v", err)
	}
	if txt != "" {
		t.Errorf("stub returned non-empty text: %q", txt)
	}
}

func TestHandleSharedTextLoadAutoLogSkipsEmptyStubText(t *testing.T) {
	var clearCalls int

	handleSharedTextLoad(
		"",
		true,
		t.TempDir(),
		func(string) {
			t.Fatal("prefill should not be called for empty shared text")
		},
		func() {
			t.Fatal("focus should not be called for empty shared text")
		},
		func() {
			t.Fatal("resetInput should not be called for empty shared text")
		},
		func() {
			clearCalls++
		},
		func(string, string) error {
			t.Fatal("logEntry should not be called for empty shared text")
			return nil
		},
		func(string, string) {
			t.Fatal("info dialog should not be shown for empty shared text")
		},
		func(error) {
			t.Fatal("error dialog should not be shown for empty shared text")
		},
	)

	if clearCalls != 1 {
		t.Fatalf("expected cache cleanup once, got %d", clearCalls)
	}
}
