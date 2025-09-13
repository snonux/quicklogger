package main

import (
    "fmt"
    "os"
    "time"

    "fyne.io/fyne/v2"
    "fyne.io/fyne/v2/app"
    "fyne.io/fyne/v2/container"
    "fyne.io/fyne/v2/dialog"
    "fyne.io/fyne/v2/widget"
)

const (
	appId           = "org.buetow.quicklogger"
	placeholderText = "Enter text here..."
	maxTextLength   = 5000 // Limit text length to prevent performance issues
)

var (
    defaultDirectory = "."
)

var windowSize = fyne.NewSize(400, 100)

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
	// Create main window
	window := a.NewWindow("Quick logger")

	input := widget.NewMultiLineEntry()
	// Optimization 1: Disable word wrapping on Android to improve performance
	// Word wrapping causes expensive recalculations on every text change
	if fyne.CurrentDevice().IsMobile() {
		input.Wrapping = fyne.TextWrapOff
	} else {
		input.Wrapping = fyne.TextWrapWord
	}
	input.SetPlaceHolder(placeholderText)
	// Optimization 2: Reduce visible rows on mobile to limit rendering area
	if fyne.CurrentDevice().IsMobile() {
		input.SetMinRowsVisible(10)
	} else {
		input.SetMinRowsVisible(30)
	}

	// Optimization 3: Add text length indicator
	charCount := widget.NewLabel("0 chars")

	// Optimization 4: Throttle text changes with validation
	input.OnChanged = func(text string) {
		// Update character count
		charCount.SetText(fmt.Sprintf("%d chars", len(text)))

		// Warn if text is getting too long
		if len(text) > maxTextLength {
			dialog.ShowInformation("Text Limit",
				fmt.Sprintf("Text is getting long (%d chars). Consider logging to avoid performance issues.", len(text)),
				window)
		}
	}

    logTextButton := widget.NewButton("Log text", func() {
        filename := fmt.Sprintf("%s/ql-%s.md",
            a.Preferences().StringWithFallback("Directory", defaultDirectory),
            time.Now().Format("060102-150405"),
        )
        if err := os.WriteFile(filename, []byte(input.Text), 0644); err != nil {
            dialog.ShowError(err, window)
            return
        }
        input.SetText("")
        input.SetPlaceHolder(placeholderText)
        // Reset character count
        charCount.SetText("0 chars")
    })

	// Optimization 5: Add clear button for quick text clearing
	clearButton := widget.NewButton("Clear", func() {
		input.SetText("")
		charCount.SetText("0 chars")
		window.Canvas().Focus(input)
	})

    window.SetContent(container.NewVBox(
        input,
        container.NewHBox(
            logTextButton,
            clearButton,
            widget.NewButton("Preferences", func() {
                createPreferenceWindow(a).Show()
            }),
            charCount, // Show character count
        ),
    ))
	window.Resize(windowSize)
	window.Canvas().Focus(input)

	return window
}

func main() {
	createMainWindow(app.NewWithID(appId)).ShowAndRun()
}
