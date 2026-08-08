# Android keyboard suggestions (FUTO)

Stock Fyne disables IME suggestions on Android. This doc is the root cause, the
patch we apply, and why it's only a partial fix.

## Root cause

In Fyne v2.4.3 the Android activity hardcodes the no-suggestions input type:

`internal/driver/mobile/app/GoNativeActivity.java`
```java
private static final int DEFAULT_INPUT_TYPE = InputType.TYPE_TEXT_FLAG_NO_SUGGESTIONS; // 0x80000
```

Both `setupEntry()` and `doShowKeyboard()` set the hidden `EditText` to
`DEFAULT_INPUT_TYPE`. In `doShowKeyboard` the keyboard-type switch only overrides
the input type for the **number** and **password** keyboards; the **default**
and **singleline** branches leave it at `0x80000`.

Quicklogger's multi-line entry returns `mobile.DefaultKeyboard`
(`widget/entry.go`: `if e.MultiLine { return mobile.DefaultKeyboard }`), so it
hits the default branch → `TYPE_TEXT_FLAG_NO_SUGGESTIONS`.

`TYPE_TEXT_FLAG_NO_SUGGESTIONS` explicitly tells the IME not to show
suggestions/autocorrect. **Gboard often ignores it; FUTO honors it** and turns
its suggestion bar off — which is the behaviour that prompted this work.

## The patch (opt-in via `SUGGESTIONS=1`, default for `mage androidshare`)

Because `DEFAULT_INPUT_TYPE` is a `static final int`, it is **inlined** as a
`const` at every call site, so patching the field would do nothing. The actual
input type lives in the decompiled smali for the `doShowKeyboard` anonymous
`Runnable`: `org/golang/app/GoNativeActivity$3.smali`.

The switch uses a `packed-switch` on the keyboard type, with `v4` initialised to
`0x80000` and only the number (`pswitch_1` → `0x80002`) and password
(`pswitch_0` → `0x80080`) cases overriding it. The patch inserts
`const/4 v4, 0x1` (`TYPE_CLASS_TEXT`) in **only** the `:pswitch_3` (default) and
`:pswitch_2` (singleline) blocks:

```diff
     :pswitch_3
+    const/4 v4, 0x1
     nop
     .line 165
     move-object v0, v1
```

Number and password keyboards are untouched, so only the default text entry
becomes suggestion-capable. The change is verified by re-decompiling the signed
APK and grepping for `const/4 v4, 0x1` in `GoNativeActivity$3.smali` (expect 2).

## Why it's only a partial fix

Fyne's Android text entry is a **hidden, one-character `EditText`** (kept at
`"0"`) used purely as a keystroke conduit; the real text is drawn on Fyne's own
canvas. Two consequences:

1. The IME never sees the real document, so context-aware engines (like FUTO)
   can't offer completions based on what's already typed.
2. Fyne consumes only **committed** text — `onTextChanged` → native
   `keyboardTyped` — and ignores the IME **composing region**
   (`setComposingText`/`commitText` composition). So inline/composing suggestions
   that FUTO inserts as you type can still misbehave (doubled or mangled chars).

The patch makes the suggestion **bar** appear; inserting suggestions may still
be imperfect. Test on the target device (`SUGGESTIONS=1 mage androidshare`,
reinstall, tap the entry, type with FUTO).

## Disabling

```sh
SUGGESTIONS=0 mage androidshare
```

leaves the keyboard behaviour at stock Fyne (suggestions off).