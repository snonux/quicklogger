#!/bin/sh
# patch-apk.sh — add Android "Share" support to the Quicklogger APK.
#
# Why this exists:
#   Fyne v2.4.3's GoNativeActivity does not read ACTION_SEND intent extras,
#   and `fyne package` / `fyne-cross` do NOT compile custom Java into the APK
#   (classes.dex is a pre-baked blob baked into the Fyne release). So the only
#   way to receive shared text is to post-process the built APK:
#     1. decode it with apktool,
#     2. merge in a small ShareActivity (compiled Java -> smali),
#     3. add a SEND intent-filter to the manifest,
#     4. rebuild, zipalign and re-sign.
#
# The ShareActivity writes EXTRA_TEXT to <cacheDir>/quicklogger-shared.txt,
# which the Go side (android_shared_android.go: readSharedFromCache) reads on
# startup and on SetOnEnteredForeground.
#
# Prerequisites (already present on this host, paths autodetected):
#   - apktool, javac, keytool (on PATH)
#   - Android SDK build-tools (d8, aapt, apksigner, zipalign) and a platform
#     android.jar under $ANDROID_HOME or ~/Android/Sdk
#   - a base APK built by:  mage androidcross  ->  fyne-cross/dist/android/quicklogger.apk
#
# Usage:
#   ./scripts/patch-apk.sh
# Output:
#   fyne-cross/dist/android/quicklogger-share.apk   (signed, ready to install)
set -eu

ROOT=$(cd "$(dirname "$0")/.." && pwd)
cd "$ROOT"

BASE_APK="fyne-cross/dist/android/quicklogger.apk"
OUT_APK="fyne-cross/dist/android/quicklogger-share.apk"
KEYSTORE="$ROOT/scripts/debug.keystore"
SHARE_SRC="android/src/main/java/org/buetow/quicklogger/ShareActivity.java"

if [ ! -f "$BASE_APK" ]; then
    echo "error: base APK not found at $BASE_APK" >&2
    echo "       run:  mage androidcross" >&2
    exit 1
fi
if [ ! -f "$SHARE_SRC" ]; then
    echo "error: $SHARE_SRC not found" >&2
    exit 1
fi

# --- locate Android SDK / build-tools / platform ---------------------------
SDK="${ANDROID_HOME:-}"
if [ -z "$SDK" ] && [ -d "$HOME/Android/Sdk" ]; then SDK="$HOME/Android/Sdk"; fi
if [ -z "$SDK" ]; then echo "error: ANDROID_HOME not set and ~/Android/Sdk missing" >&2; exit 1; fi

BT=$(ls -d "$SDK"/build-tools/*/ 2>/dev/null | sort -V | tail -1)
if [ -z "$BT" ]; then echo "error: no build-tools under $SDK/build-tools" >&2; exit 1; fi
PLAT=$(ls -d "$SDK"/platforms/android-*/ 2>/dev/null | sort -V | tail -1)android.jar
if [ ! -f "$PLAT" ]; then echo "error: no platform android.jar under $SDK/platforms" >&2; exit 1; fi

echo "SDK        : $SDK"
echo "build-tools: $BT"
echo "platform  : $PLAT"

# --- working dirs ----------------------------------------------------------
WORK=$(mktemp -d)
trap 'rm -rf "$WORK"' EXIT
CLASSES="$WORK/classes"; mkdir -p "$CLASSES"
DEXOUT="$WORK/dexout";   mkdir -p "$DEXOUT"
WRAP="$WORK/wrap";       mkdir -p "$WRAP"

echo ">> compiling ShareActivity.java"
javac -source 1.8 -target 1.8 -nowarn \
    -bootclasspath "$PLAT" -d "$CLASSES" "$SHARE_SRC"

echo ">> dexing ShareActivity"
"$BT/d8" --lib "$PLAT" --min-api 21 --output "$DEXOUT" \
    "$CLASSES/org/buetow/quicklogger/ShareActivity.class"

echo ">> generating smali for ShareActivity"
cp "$DEXOUT/classes.dex" "$WRAP/classes.dex"
cat > "$WRAP/AndroidManifest.xml" <<'XML'
<?xml version="1.0" encoding="utf-8"?>
<manifest xmlns:android="http://schemas.android.com/apk/res/android" package="org.buetow.quickloader">
    <application android:label="t"/>
</manifest>
XML
"$BT/aapt" package -f -M "$WRAP/AndroidManifest.xml" -I "$PLAT" -F "$WRAP/wrap.apk" >/dev/null
( cd "$WRAP" && "$BT/aapt" add wrap.apk classes.dex >/dev/null )
apktool d -f "$WRAP/wrap.apk" -o "$WRAP/decoded" >/dev/null
SHARE_SMALI="$WRAP/decoded/smali/org/buetow/quicklogger/ShareActivity.smali"
if [ ! -f "$SHARE_SMALI" ]; then echo "error: smali generation failed" >&2; exit 1; fi

echo ">> decoding base APK with apktool"
apktool d -f "$BASE_APK" -o "$WORK/decoded" >/dev/null

# apktool 2.9 cannot decode this app's resource table (the only resource is a
# single mipmap/icon), so the launcher icon is dropped on the round-trip
# (manifest decodes android:icon as "@null"). Restore it manually so the app
# keeps its home-screen icon.
if grep -q 'android:icon="@null"' "$WORK/decoded/AndroidManifest.xml"; then
    echo ">> restoring launcher icon dropped by apktool"
    icon_rel=$(unzip -l "$BASE_APK" | awk '/mipmap.*icon\.png/{print $4; exit}')
    if [ -n "$icon_rel" ]; then
        mkdir -p "$WORK/decoded/$(dirname "$icon_rel")"
        unzip -p "$BASE_APK" "$icon_rel" > "$WORK/decoded/$icon_rel"
        cat > "$WORK/decoded/res/values/public.xml" <<'XML'
<?xml version="1.0" encoding="utf-8"?>
<resources>
    <public type="mipmap" name="icon" id="0x7f020000" />
</resources>
XML
        sed -i 's#android:icon="@null"#android:icon="@mipmap/icon"#' "$WORK/decoded/AndroidManifest.xml"
    else
        echo "warning: no mipmap icon found in base APK" >&2
    fi
fi

echo ">> merging ShareActivity smali into decoded APK"
mkdir -p "$WORK/decoded/smali/org/buetow/quicklogger"
cp "$SHARE_SMALI" "$WORK/decoded/smali/org/buetow/quicklogger/ShareActivity.smali"

# Optional: enable IME suggestions for the default+singleline keyboards.
# Fyne hardcodes InputType.TYPE_TEXT_FLAG_NO_SUGGESTIONS (0x80000) which FUTO
# (and other keyboards) honor by disabling suggestions. This patch sets the
# default/singleline keyboard branches to TYPE_CLASS_TEXT (0x1) only; the
# number and password keyboards keep their own input types untouched.
#
# NOTE: this is a PARTIAL fix. Fyne's Android entry is a hidden 1-char EditText
# used purely as a keystroke conduit (the real text is drawn on Fyne's canvas)
# and it ignores the IME composing region, so inline/composing suggestions from
# FUTO may still misbehave. Enable with:  SUGGESTIONS=1 ./scripts/patch-apk.sh
if [ "${SUGGESTIONS:-0}" = "1" ]; then
    echo ">> enabling IME suggestions (FUTO) in GoNativeActivity smali"
    python3 - "$WORK/decoded/smali/org/golang/app/GoNativeActivity\$3.smali" <<'PY'
import sys
p = sys.argv[1]; s = open(p).read()
for label in (":pswitch_3", ":pswitch_2"):
    needle = "    " + label + "\n    nop\n"
    repl = "    " + label + "\n    const/4 v4, 0x1\n    nop\n"
    if label + "\n    const/4 v4, 0x1" in s:
        continue
    if needle not in s:
        raise SystemExit("error: %s anchor not found" % label)
    s = s.replace(needle, repl, 1)
open(p, "w").write(s)
print("patched default+singleline keyboard input types")
PY
fi

echo ">> patching AndroidManifest.xml"
python3 - "$WORK/decoded/AndroidManifest.xml" <<'PY'
import sys, re
p = sys.argv[1]
xml = open(p).read()
share = """        <activity android:name="org.buetow.quicklogger.ShareActivity" android:exported="true" android:excludeFromRecents="true" android:launchMode="singleTask" android:taskAffinity="" android:theme="@android:style/Theme.Translucent.NoTitleBar">
            <intent-filter>
                <action android:name="android.intent.action.SEND"/>
                <category android:name="android.intent.category.DEFAULT"/>
                <data android:mimeType="text/plain"/>
                <data android:mimeType="text/*"/>
            </intent-filter>
        </activity>
"""
anchor = "        </activity>\n        <receiver"
if "org.buetow.quicklogger.ShareActivity" in xml:
    print("manifest already patched")
else:
    if anchor not in xml:
        raise SystemExit("error: manifest anchor not found")
    xml = xml.replace(anchor, "        </activity>\n" + share + "        <receiver", 1)
    open(p, "w").write(xml)
    print("manifest patched")
PY

echo ">> rebuilding APK with apktool"
apktool b "$WORK/decoded" -o "$WORK/unsigned.apk" >/dev/null

echo ">> zipalign"
"$BT/zipalign" -p -f 4 "$WORK/unsigned.apk" "$WORK/aligned.apk"

if [ ! -f "$KEYSTORE" ]; then
    echo ">> creating debug keystore (one-time)"
    keytool -genkeypair -v -keystore "$KEYSTORE" \
        -storepass android -keypass android -alias quicklogger \
        -keyalg RSA -keysize 2048 -validity 10000 \
        -dname "CN=Quicklogger, O=Debug, C=XX" >/dev/null
fi

echo ">> signing"
"$BT/apksigner" sign --ks "$KEYSTORE" --ks-pass pass:android \
    --key-pass pass:android --in "$WORK/aligned.apk" --out "$OUT_APK"
"$BT/apksigner" verify "$OUT_APK" >/dev/null && echo ">> signature verified"

echo ">> done: $OUT_APK"