//go:build android

package main

import (
	"os"
	"path/filepath"
)

// sharedTextCachePath returns the cache file used for Android share intents.
func sharedTextCachePath() string {
	dir, derr := os.UserCacheDir()
	if derr != nil || dir == "" {
		dir = os.TempDir()
	}
	return filepath.Join(dir, "quicklogger-shared.txt")
}

// readSharedFromCache tries to read the shared text written by the Android activity.
func readSharedFromCache() (string, error) {
	b, err := os.ReadFile(sharedTextCachePath())
	if err != nil {
		return "", err
	}
	return string(b), nil
}
