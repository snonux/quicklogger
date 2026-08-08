package org.buetow.quicklogger;

import android.app.Activity;
import android.content.Intent;
import android.os.Bundle;
import java.io.File;
import java.io.FileOutputStream;
import java.io.IOException;

/**
 * ShareActivity receives ACTION_SEND text intents (the Android "Share" sheet)
 * and hands the text to the Go/Fyne app via a cache file.
 *
 * Why this exists: Fyne v2.4.3's GoNativeActivity does not read SEND intent
 * extras, and `fyne package` / `fyne-cross` do not compile custom Java into the
 * APK at build time (classes.dex is a pre-baked blob). This activity is merged
 * into the APK by the post-processing script in scripts/patch-apk.sh.
 *
 * Flow: write EXTRA_TEXT to <cacheDir>/quicklogger-shared.txt, then launch the
 * main GoNativeActivity so the Go side's loadSharedText() (runs on startup and
 * on SetOnEnteredForeground) picks it up and clears the cache.
 */
public class ShareActivity extends Activity {

    private static final String CACHE_NAME = "quicklogger-shared.txt";

    @Override
    protected void onCreate(Bundle savedInstanceState) {
        super.onCreate(savedInstanceState);

        Intent intent = getIntent();
        String action = intent.getAction();
        String type = intent.getType();
        if (Intent.ACTION_SEND.equals(action) && type != null && type.startsWith("text/")) {
            CharSequence shared = intent.getCharSequenceExtra(Intent.EXTRA_TEXT);
            if (shared != null) {
                writeCache(shared.toString());
            }
        }

        // Bring the main app to the foreground so it reads the cache file.
        // GoNativeActivity lives in the app's own classes.dex, not on the
        // compile-time classpath, so reference it by name.
        Intent main = new Intent();
        main.setClassName(this, "org.golang.app.GoNativeActivity");
        main.addFlags(Intent.FLAG_ACTIVITY_NEW_TASK | Intent.FLAG_ACTIVITY_CLEAR_TOP);
        startActivity(main);
        finish();
    }

    private void writeCache(String text) {
        File dir = getCacheDir();
        if (dir == null) {
            return;
        }
        File file = new File(dir, CACHE_NAME);
        FileOutputStream fos = null;
        try {
            fos = new FileOutputStream(file, false);
            fos.write(text.getBytes("UTF-8"));
        } catch (IOException e) {
            // Best effort; the main app simply won't see shared text.
        } finally {
            if (fos != null) {
                try {
                    fos.close();
                } catch (IOException e) {
                    // ignore
                }
            }
        }
    }
}