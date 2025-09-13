//go:build android

package main

import (
    "os"
    "path/filepath"
)

// readSharedFromCache tries to read the shared text written by the Android activity.
// The activity writes into getCacheDir()/quicklogger-shared.txt; on Go, os.TempDir()
// maps to the same location in Android (app cache directory).
func readSharedFromCache() (string, error) {
    dir, derr := os.UserCacheDir()
    if derr != nil || dir == "" {
        dir = os.TempDir()
    }
    path := filepath.Join(dir, "quicklogger-shared.txt")
    b, err := os.ReadFile(path)
    if err != nil {
        return "", err
    }
    // best-effort cleanup; ignore errors
    _ = os.Remove(path)
    return string(b), nil
}

func debugSharedPath() string {
    dir, _ := os.UserCacheDir()
    if dir == "" {
        dir = os.TempDir()
    }
    return filepath.Join(dir, "quicklogger-shared.txt")
}
