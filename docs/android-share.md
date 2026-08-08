# Android share support — why and how

Quicklogger receives `ACTION_SEND` text via a small Java `ShareActivity` that is
merged into the APK **after** Fyne builds it. This doc explains why a normal
build can't do it, how the post-processing works, and the contract the two
halves rely on.

## Why `fyne package` / `fyne-cross` can't add it directly

Two hard limits in Fyne v2.4.3:

1. **No custom Java is compiled at package time.** `classes.dex` is a
   pre-generated base64 blob (`dexStr` in `cmd/fyne/internal/mobile/dex.go`),
   produced once during Fyne's own release by `gendex.go` (which compiles
   `internal/driver/mobile/app/*.java` with `javac` + `dx`). At app-build time,
   `buildAPK` in `cmd/fyne/internal/mobile/build_androidapp.go` just writes that
   blob verbatim into the APK. So adding a `.java` file to your tree does
   nothing, and editing the Fyne module's `GoNativeActivity.java` does nothing
   either.

2. **`GoNativeActivity` ignores `ACTION_SEND`.** Its `onCreate` only loads the
   native library and sets up the hidden `EditText`; there is no code reading
   `Intent.EXTRA_TEXT`. A `SEND` intent-filter on the main activity alone would
   receive the share and then silently drop the text.

There is also a third, related gotcha: `fyne package` only reads a manifest at
the **project root** —

```go
dir := filepath.Dir(pkg.GoFiles[0])                       // package dir
manifestPath := filepath.Join(dir, "AndroidManifest.xml") // root only
```

— so `android/AndroidManifest.xml` is ignored and a default manifest (no
`SEND`, no `ShareActivity`) is generated. The old `android/AndroidManifest.xml`
was therefore dead weight and has been removed.

## How `scripts/patch-apk.sh` works

The script post-processes the base APK produced by `mage androidcross`
(`fyne-cross/dist/android/quicklogger.apk`) rather than changing the build:

1. **Compile `ShareActivity.java` → smali.**
   `javac` against `android.jar`, then `d8` to dex, then wrap the dex in a
   throwaway APK with `aapt` and `apktool d` it to get
   `ShareActivity.smali`. (Direct baksmali isn't assumed to be installed.)
2. **Decode the base APK** with `apktool d` (readable manifest + `smali/`).
3. **Restore the launcher icon.** `apktool 2.9` cannot decode this app's
   resource table (it has a single resource, `mipmap/icon` id `0x7f020000`),
   so it drops the icon PNG and renders `android:icon` as `@null`. The script
   extracts `res/mipmap-xxxhdpi-v4/icon.png` back, declares the resource id in
   `res/values/public.xml`, and fixes the manifest icon ref. See
   [build-notes.md](build-notes.md).
4. **Merge `ShareActivity.smali`** into `smali/org/buetow/quicklogger/`.
5. **Optional FUTO patch** (`SUGGESTIONS=1`, default on for `mage androidshare`)
   — see [android-keyboard.md](android-keyboard.md).
6. **Patch the manifest** to add a `SEND` intent-filter on `ShareActivity`.
7. **Rebuild, zipalign, and re-sign** with `scripts/debug.keystore`.

### Prerequisites (host)
`apktool`, `javac`, `keytool` on `PATH`, plus an Android SDK with build-tools
(`d8`, `aapt`, `apksigner`, `zipalign`) and a platform `android.jar`. The script
auto-detects `$ANDROID_HOME` (falling back to `~/Android/Sdk`).

## The cache-path contract

The two halves communicate through one file:

```
<app getCacheDir()>/quicklogger-shared.txt
```

- **Java (`ShareActivity`)** writes `Intent.EXTRA_TEXT` to
  `getCacheDir()/quicklogger-shared.txt`, then starts `GoNativeActivity`.
- **Go (`android_shared_android.go`)** reads the same path on startup and on
  `SetOnEnteredForeground`, then deletes it.

They match because on Android `os.UserCacheDir()` returns an error (no
`$HOME`/`$XDG_CACHE_HOME`), so the Go code falls back to `os.TempDir()`, which
gomobile sets (`setenv("TMPDIR", getTmpdir())`) to the app's
`getCacheDir()`. Don't change one side without the other.

## Runtime flow

```
other app ──Share──▶ ShareActivity
   writes EXTRA_TEXT → <cacheDir>/quicklogger-shared.txt
   startActivity(GoNativeActivity)   # NEW_TASK | CLEAR_TOP
   finish()
        └▶ Go: loadSharedText()      # runs on startup AND onEnteredForeground
             reads cache → input.SetText(text)   (prefill)
             OR, if "Auto-log shared text" pref: logEntry() → ql-*.md, reset
             clears the cache file
```

With **Auto-log shared text** off (default) the shared text prefills the editor
for review; with it on, the text is written straight to the log directory.

## Troubleshooting

- **Quicklogger doesn't appear in the share sheet.** The manifest's `SEND`
  filter or `ShareActivity` didn't make it in. Check the built APK:
  `aapt dump xmltree <apk> AndroidManifest.xml | grep SEND` (expect 1) and
  `unzip -p <apk> classes.dex | strings | grep ShareActivity` (expect a hit).
- **Share opens Quicklogger but the editor stays empty.** The cache file wasn't
  written or wasn't where Go expects. Re-check the cache-path contract above.
- **Notes fail to write ("read-only filesystem").** That's the *Directory*
  preference / storage-scopes issue, not share — see
  [grapheneos-setup.md](grapheneos-setup.md).

## Known TODOs

- **`android:debuggable="true"` still on.** The base fyne-cross build is a
  debug build, so the installer warns the app is "currently being tested".
  The plan is to strip the flag in `patch-apk.sh` (a `DEBUGGABLE=1` escape hatch
  keeps `adb shell run-as` working during testing), but that change is not
  applied yet.