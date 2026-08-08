# Build notes

Small but painful build/packaging details learned while wiring up the Android
share pipeline.

## `.gitignore` swallowed the Java package dir

The repo originally had a bare `quicklogger` entry in `.gitignore` (meant to
ignore the built binary/APK). A bare pattern matches any file or directory
**named** `quicklogger` anywhere — including the Java package directory
`android/src/main/java/org/buetow/quicklogger/`. So `ShareActivity.java` was
silently ignored and never staged until fixed.

The ignore is now specific:

```
fyne-cross
tmp-pkg
bin/
quicklogger.apk
scripts/debug.keystore
```

If you add more Java under `org.buetow.quicklogger`, watch for this coming back
if anyone re-broadens the pattern.

## FyneApp.toml build number

`fyne package` rewrites `FyneApp.toml` (normalising indentation) and increments
`Build` on each package. So `Build` drifts upward with every build run, not just
releases. It's harmless (just needs to be higher than the previous release's
build), but don't be surprised to see unrelated `Build` bumps in diffs — they're
a packaging side effect, not a manual edit. Bump `Version` explicitly for
releases (see the `increment-version-and-push` skill).

## Persistent signing key

`scripts/patch-apk.sh` signs with `scripts/debug.keystore` (created on first
run, gitignored). Because the key is **persistent across rebuilds**,
`adb install -r` no longer requires an uninstall, so app preferences survive
reinstalls. Important: this is a self-signed debug key — fine for personal
sideloading, not a release key.

## apktool 2.9 + this APK's resources

`apktool 2.9.3` cannot decode this app's resource table (the APK has exactly one
resource, `mipmap/icon` id `0x7f020000`), emitting
`W: End of chunk hit. Skipping remaining entries (1) in type: mipmap` and
leaving `res/values/public.xml` empty. The icon restore step in
`patch-apk.sh` works around this; if a future apktool version decodes it
correctly, that workaround (guarded by `android:icon="@null"`) becomes a no-op.