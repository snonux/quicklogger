#!/bin/sh

if [ -z "$ANDORID_NDK_HOME" ]; then
    # This is where I personally have installed the Android NDK
    export ANDROID_NDK_HOME=~/android-ndk/android-ndk-r26b
fi

fyne package -os android

if [ -d ~/Documents/APKs ]; then
    # Will be synced to my phone by Syncthing, so I can install it there.
    cp -v quicklogger.apk ~/Documents/APKs
fi

