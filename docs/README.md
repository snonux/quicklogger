# Quicklogger docs

Background and internals for the Android build/share pipeline. The
[`README.md`](../README.md) covers day-to-day usage; these notes cover the
*why* and the gotchas.

- [android-share.md](android-share.md) — why "Share to Quicklogger" needs APK
  post-processing, how `scripts/patch-apk.sh` works, and the cache-path contract
  between the Java `ShareActivity` and the Go side.
- [android-keyboard.md](android-keyboard.md) — why FUTO/Gboard suggestions are
  off in stock Fyne, the smali patch, and why it is only a partial fix.
- [grapheneos-setup.md](grapheneos-setup.md) — storage scopes, the
  "read-only file system" error after a reinstall, and the notes folder.
- [build-notes.md](build-notes.md) — the `.gitignore` package-dir gotcha, the
  FyneApp.toml build number, and the persistent signing key.