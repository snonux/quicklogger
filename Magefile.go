//go:build mage
// +build mage

package main

import (
    "io"
    "fmt"
    "os"
    "path/filepath"
    "runtime"
    "strconv"
    "time"

    "github.com/magefile/mage/sh"
)

// Build compiles the quicklogger app into ./bin.
func Build() error {
    outDir := "bin"
    if err := os.MkdirAll(outDir, 0o755); err != nil {
        return err
    }

    binName := "quicklogger"
    if runtime.GOOS == "windows" {
        binName += ".exe"
    }

    out := filepath.Join(outDir, binName)
    fmt.Printf("Building %s\n", out)
    return sh.RunV("go", "build", "-o", out, ".")
}

// Run launches the app with `go run .`.
func Run() error {
    fmt.Println("Running quicklogger (verbose build)")
    // Show Go version and key env for diagnostics
    _ = sh.RunV("go", "version")
    _ = sh.RunV("go", "env", "GOVERSION", "GOOS", "GOARCH", "GOMOD", "GOMODCACHE", "GOCACHE")
    // Compile+run with verbose and trace flags to show build steps
    return sh.RunV("go", "run", "-v", "-x", ".")
}

// Clean removes build artifacts (bin/ and local APK).
func Clean() error {
    fmt.Println("Cleaning build artifacts")
    // Remove bin directory
    if err := os.RemoveAll("bin"); err != nil {
        return err
    }
    // Remove generated APK if present
    _ = os.Remove("quicklogger.apk")
    return nil
}

// Android packages an Android APK via Fyne and copies it to ~/Documents/APKs if present.
func Android() error {
    env := map[string]string{}

    // Respect existing ANDROID_NDK_HOME. If not set, try legacy var or a common default path.
    ndk := os.Getenv("ANDROID_NDK_HOME")
    if ndk == "" {
        // Some setups export a misspelled var; honor it if present.
        ndk = os.Getenv("ANDORID_NDK_HOME")
    }
    if ndk == "" {
        if home, err := os.UserHomeDir(); err == nil {
            guess := filepath.Join(home, "android-ndk", "android-ndk-r26b")
            env["ANDROID_NDK_HOME"] = guess
            fmt.Printf("ANDROID_NDK_HOME not set, guessing %s\n", guess)
        }
    }

    fmt.Println("Packaging Android APK via Fyne")
    if err := sh.RunWithV(env, "fyne", "package", "-os", "android"); err != nil {
        return err
    }

    // If ~/Documents/APKs exists, copy the APK there.
    home, err := os.UserHomeDir()
    if err != nil {
        return nil // packaging succeeded; copying is best-effort
    }
    destDir := filepath.Join(home, "Documents", "APKs")
    if st, err := os.Stat(destDir); err == nil && st.IsDir() {
        src := "quicklogger.apk"
        if _, err := os.Stat(src); err == nil {
            dst := filepath.Join(destDir, filepath.Base(src))
            if err := copyFile(src, dst); err != nil {
                return err
            }
            fmt.Printf("Copied %s to %s\n", src, dst)
        }
    }
    return nil
}

// AndroidCross builds an Android APK using fyne-cross with Docker/Podman.
// Mirrors steps from README using a Podman socket if available.
func AndroidCross() error {
    env := map[string]string{}

    // If DOCKER_HOST is not set, try a sensible Podman socket default.
    if os.Getenv("DOCKER_HOST") == "" {
        uid := os.Geteuid()
        sock := filepath.Join("/run/user", strconv.Itoa(uid), "podman", "podman.sock")
        if _, err := os.Stat(sock); err == nil {
            env["DOCKER_HOST"] = "unix://" + sock
            fmt.Printf("Using Podman socket: %s\n", env["DOCKER_HOST"])
        } else {
            fmt.Println("DOCKER_HOST is not set and no Podman socket found; relying on default Docker daemon.")
        }
    }

    // Ensure fyne-cross is available; attempt install if not.
    if _, err := sh.Output("which", "fyne-cross"); err != nil {
        fmt.Println("Installing fyne-cross (requires network access)...")
        if err := sh.RunV("go", "install", "github.com/fyne-io/fyne-cross@latest"); err != nil {
            return fmt.Errorf("fyne-cross not found and installation failed: %w", err)
        }
    }

    // Pull/update builder image then build.
    if err := sh.RunWithV(env, "fyne-cross", "android", "--pull"); err != nil {
        return err
    }
    if err := sh.RunWithV(env, "fyne-cross", "android"); err != nil {
        return err
    }

    // Copy newest APK to ~/Documents/APKs if available.
    apk, err := newestAPK("fyne-cross/dist/android")
    if err != nil {
        // Not fatal; print a note and finish
        fmt.Printf("Note: could not locate built APK: %v\n", err)
        return nil
    }
    home, herr := os.UserHomeDir()
    if herr != nil {
        return nil
    }
    destDir := filepath.Join(home, "Documents", "APKs")
    if st, err := os.Stat(destDir); err == nil && st.IsDir() {
        dst := filepath.Join(destDir, filepath.Base(apk))
        if err := copyFile(apk, dst); err != nil {
            return err
        }
        fmt.Printf("Copied %s to %s\n", apk, dst)
    }
    return nil
}

func newestAPK(dir string) (string, error) {
    entries, err := os.ReadDir(dir)
    if err != nil {
        return "", err
    }
    var newest string
    var newestMod time.Time
    for _, e := range entries {
        if e.IsDir() {
            continue
        }
        name := e.Name()
        if filepath.Ext(name) != ".apk" {
            continue
        }
        info, err := e.Info()
        if err != nil {
            continue
        }
        if info.ModTime().After(newestMod) {
            newestMod = info.ModTime()
            newest = filepath.Join(dir, name)
        }
    }
    if newest == "" {
        return "", fmt.Errorf("no .apk found in %s", dir)
    }
    return newest, nil
}

func copyFile(src, dst string) error {
    in, err := os.Open(src)
    if err != nil {
        return err
    }
    defer in.Close()

    out, err := os.Create(dst)
    if err != nil {
        return err
    }
    defer func() { _ = out.Close() }()

    if _, err = io.Copy(out, in); err != nil {
        return err
    }
    return out.Sync()
}
