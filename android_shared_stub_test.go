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
