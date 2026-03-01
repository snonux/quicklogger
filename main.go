package main

import (
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
)

const (
	appID           = "org.buetow.quicklogger"
	placeholderText = "Enter text here..."
	maxTextLength   = 5000 // Limit text length to prevent performance issues
)

const defaultDirectory = "."

var windowSize = fyne.NewSize(400, 100)

// logEntry writes text to a timestamped markdown file in dir.
// Separates persistence logic from the UI so it can be tested independently.
func logEntry(dir, text string) error {
	filename := filepath.Join(dir, "ql-"+time.Now().Format("060102-150405")+".md")
	return os.WriteFile(filename, []byte(text), 0o644)
}

// newInputWidget creates the multi-line text entry with platform-appropriate
// wrapping and row count settings.
func newInputWidget() *widget.Entry {
	input := widget.NewMultiLineEntry()
	input.SetPlaceHolder(placeholderText)

	// On mobile, disable word wrapping and reduce visible rows to limit
	// expensive recalculations and rendering area.
	if fyne.CurrentDevice().IsMobile() {
		input.Wrapping = fyne.TextWrapOff
		input.SetMinRowsVisible(10)
	} else {
		input.Wrapping = fyne.TextWrapWord
		input.SetMinRowsVisible(30)
	}

	return input
}

func createPreferenceWindow(a fyne.App) fyne.Window {
	window := a.NewWindow("Preferences")
	directoryPreference := widget.NewEntry()
	directoryPreference.SetText(a.Preferences().StringWithFallback("Directory", defaultDirectory))

	window.SetContent(container.NewVBox(
		container.NewVBox(
			widget.NewLabel("Directory:"),
			directoryPreference,
		),
		container.NewHBox(
			widget.NewButton("Save", func() {
				a.Preferences().SetString("Directory", directoryPreference.Text)
				window.Hide()
			}),
		)))
	window.Resize(windowSize)

	return window
}

func createMainWindow(a fyne.App) fyne.Window {
	window := a.NewWindow("Quick logger")
	input := newInputWidget()
	charCount := widget.NewLabel("0 chars")

	// Track whether the length warning has been shown so we don't fire a
	// modal dialog on every keystroke above the limit.
	warnShown := false
	input.OnChanged = func(text string) {
		charCount.SetText(fmt.Sprintf("%d chars", len(text)))
		if len(text) > maxTextLength && !warnShown {
			warnShown = true
			dialog.ShowInformation("Text Limit",
				fmt.Sprintf("Text is getting long (%d chars). Consider logging to avoid performance issues.", len(text)),
				window)
		} else if len(text) <= maxTextLength {
			warnShown = false
		}
	}

	// resetInput clears the text entry and character count.
	resetInput := func() {
		input.SetText("")
		charCount.SetText("0 chars")
	}

	logTextButton := widget.NewButton("Log text", func() {
		dir := a.Preferences().StringWithFallback("Directory", defaultDirectory)
		if err := logEntry(dir, input.Text); err != nil {
			dialog.ShowError(err, window)
			return
		}
		resetInput()
	})

	clearButton := widget.NewButton("Clear", func() {
		resetInput()
		window.Canvas().Focus(input)
	})

	// loadSharedText reads Android-shared text from cache and populates the input.
	// Used both at startup and when the app returns to the foreground.
	// A missing cache file is expected (no share pending); real errors are logged.
	loadSharedText := func() {
		txt, err := readSharedFromCache()
		if err != nil {
			if !errors.Is(err, os.ErrNotExist) {
				log.Printf("readSharedFromCache: %v", err)
			}
			return
		}
		if txt != "" {
			input.SetText(txt)
			charCount.SetText(fmt.Sprintf("%d chars", len(txt)))
			window.Canvas().Focus(input)
		}
	}

	if fyne.CurrentDevice().IsMobile() {
		loadSharedText()
	}

	window.SetContent(container.NewVBox(
		input,
		container.NewHBox(
			logTextButton,
			clearButton,
			widget.NewButton("Preferences", func() {
				createPreferenceWindow(a).Show()
			}),
			charCount,
		),
	))
	window.Resize(windowSize)
	window.Canvas().Focus(input)

	// On Android, also check for new shared text whenever app returns to foreground.
	if lc := a.Lifecycle(); lc != nil {
		lc.SetOnEnteredForeground(loadSharedText)
	}

	return window
}

func main() {
	createMainWindow(app.NewWithID(appID)).ShowAndRun()
}
