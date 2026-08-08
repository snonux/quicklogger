# GrapheneOS setup & the "read-only filesystem" error

Quicklogger writes notes as `ql-<timestamp>.md` files into a **Directory** you
choose (Preferences), and Syncthing syncs that folder home. On GrapheneOS two
things must be in place, and a fresh install will *appear* broken until they are.

## Symptom: "read-only filesystem" after installing

The Directory preference defaults to `"."`. On Android the app process's working
directory is `/`, which is read-only, so `os.WriteFile("./ql-*.md")` fails with
**EROFS** ("read-only file system") — *not* a permission denial (EACCES).

This most often shows up right after an install because **installing with a
different signing key requires an uninstall first, which wipes the app's
preferences** (`files/fyne/preferences.json` resets to `{}`), so Directory falls
back to the default `"."`.

Two notes on that:

- **The persistent `scripts/debug.keystore`** means subsequent
  `adb install -r` rebuilds use the *same* key, so they no longer need an
  uninstall and your preferences survive. The EROFS-on-reinstall was a one-time
  effect of the first key change.
- A hardening default (writing to the app's writable private files dir via
  `$FILESDIR`) was considered but deliberately **not** adopted: notes there are
  app-private, not synced, and silently logging to an unsynced location is worse
  than an obvious error. Keep Directory pointed at your synced folder.

## The notes folder

For this setup the Syncthing-synced notes folder is:

```
/storage/emulated/0/Notes/Vault
```

(existing `ql-*.md` files and Syncthing `.stversions/Vault/` live there).

## Setup steps (GrapheneOS)

1. **Grant storage access (storage scopes — required).** `adb shell pm grant
   … WRITE_EXTERNAL_STORAGE` is **not** enough on GrapheneOS; the app still
   can't write shared storage until you select the folder in the UI:
   - Settings → Apps → Quicklogger → Permissions → **Storage & media**
   - Enable **Storage scopes** and add the folder: navigate to
     Internal storage → **Notes → Vault** (or **Notes**) → Use this folder →
     Allow.
2. **Set the Directory.** Open Quicklogger → **Preferences** → Directory →
   `/storage/emulated/0/Notes/Vault` → Save.
   (Equivalently, while the app is stopped, write
   `files/fyne/preferences.json` to `{"Directory":"/storage/emulated/0/Notes/Vault"}`.)
3. Test: type a note → **Log text** → a `ql-*.md` should appear in the Vault;
   Syncthing then syncs it home.

## Auto-log shared text

Optional: Preferences → enable **Auto-log shared text**. With it on, text
shared *into* Quicklogger is written straight to the Directory instead of
prefilling the editor. Useful with the share pipeline
([android-share.md](android-share.md)); off by default so you review first.

## Verifying after granting scopes

A quick end-to-end check (needs `android:debuggable=true`, i.e. before the
debuggable flag is stripped):

```sh
PKG=org.buetow.quicklogger
adb shell am force-stop $PKG
# temporarily enable auto-log and point at the Vault
printf '%s' '{"Directory":"/storage/emulated/0/Notes/Vault","AutoLogSharedText":true}' \
  | adb shell run-as $PKG tee /data/data/$PKG/files/fyne/preferences.json >/dev/null
adb shell "am start -a android.intent.action.SEND -t text/plain \
  --es android.intent.extra.TEXT 'qltest' -n $PKG/.ShareActivity"
adb shell ls /storage/emulated/0/Notes/Vault/ | grep ql-    # expect a new file
# restore your normal prefs afterwards
```

> Note: `run-as` only works while the app is `debuggable=true`. Once the
> debuggable-flag strip ([android-share.md](android-share.md) TODO) lands,
> you'll need `DEBUGGABLE=1 mage androidshare` to keep this inspection path.