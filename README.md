# quicklogger

![Quicklogger](./logo-small.png)

This is a tiny GUI app written in Go using the Fyne framework to quickly log a message to a file.

The purpose of this is to have a small Android app to quickly log Ideas into a folder as plain text files.  From there, Syncthing will sync it to my computer at home. 

This is a screenshot of the App running on Fedora Linux. But it also works seamlessly on my Android phone.

![Screenshot](./screenshot.png)

## Installation

1. Download and install the Android NDK. I personally installed it to `~/android/android-ndk-r26b` as of this writing.
2. Clone Quicklogger: `git clone https://codeberg.org/snonux/quicklogger; cd quicklogger`
3. Build it `./build.sh` - Note, you may need to set the `ANDROID_NDK_HOME` environment variable accordingly.
4. Copy `quicklogger.apk` to your Android phone and install it (You may need to allow installing APKs from this source - just follow the instructions Android is prompting you with).

