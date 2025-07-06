package main

import (
	"fmt"
	"os"
	"strings"
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
	defaultTagItems  = []string{
		"infra",
		"log",
		"share",
		"share:li",
		"share:ma",
		"share:no",
		"track",
		"track 10",
		"track 15",
		"track 20",
		"track 25",
		"track 30",
		"track 5",
		"work",
	}
	defaultWhatItems = []string{
		"Breathing",
		"Bulgarian",
		"Ema",
		"Exercise",
		"Meditation",
		"Music",
		"Reading Articles",
		"Reading Books",
		"Stretching",
		"Tech",
	}
)

var windowSize = fyne.NewSize(400, 100)

func createPreferenceWindow(a fyne.App) fyne.Window {
	window := a.NewWindow("Preferences")
	directoryPreference := widget.NewEntry()
	directoryPreference.SetText(a.Preferences().StringWithFallback("Directory", defaultDirectory))

	tagDropdownPreference := widget.NewEntry()
	tagDropdownPreference.SetText(a.Preferences().StringWithFallback("Tags", strings.Join(defaultTagItems, ",")))

	whatDropdownPreference := widget.NewEntry()
	whatDropdownPreference.SetText(a.Preferences().StringWithFallback("Whats", strings.Join(defaultWhatItems, ",")))

	window.SetContent(container.NewVBox(
		container.NewVBox(
			widget.NewLabel("Directory:"),
			directoryPreference,
			widget.NewLabel("Tags:"),
			tagDropdownPreference,
			widget.NewLabel("Whats:"),
			whatDropdownPreference,
		),
		container.NewHBox(
			widget.NewButton("Save", func() {
				a.Preferences().SetString("Directory", directoryPreference.Text)
				a.Preferences().SetString("Tags", tagDropdownPreference.Text)
				a.Preferences().SetString("Whats", whatDropdownPreference.Text)
				window.Hide()
			}),
			widget.NewButton("Reset dropdowns", func() {
				// directoryPreference.SetText(defaultDirectory)
				tagDropdownPreference.SetText(strings.Join(defaultTagItems, ","))
				whatDropdownPreference.SetText(strings.Join(defaultWhatItems, ","))
			},
			),
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

	// Dropdown with pre-selectable items
	daysDropdown := widget.NewSelect([]string{"0", "1", "3", "7", "14", "30", "60", "99"}, func(selected string) {
		input.SetText(selected + " ")
		window.Canvas().Focus(input)
	})
	daysDropdown.PlaceHolder = "Days"

	tagDropdownItems := strings.Split(a.Preferences().StringWithFallback("Tags", strings.Join(defaultTagItems, ",")), ",")
	tagDropdown := widget.NewSelect(tagDropdownItems, func(selected string) {
		input.Append(selected + " ")
		window.Canvas().Focus(input)
	})
	tagDropdown.PlaceHolder = "Tag"

	whatDropdownItems := strings.Split(a.Preferences().StringWithFallback("Whats", strings.Join(defaultWhatItems, ",")), ",")
	whatDropdown := widget.NewSelect(whatDropdownItems, func(selected string) {
		input.Append(selected + " ")
		window.Canvas().Focus(input)
		input.Cursor()
	})
	whatDropdown.PlaceHolder = "What"

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
		container.NewHBox(daysDropdown, tagDropdown, whatDropdown, logTextButton),
		input,
		container.NewHBox(
			widget.NewButton("Preferences", func() {
				createPreferenceWindow(a).Show()
			}),
			clearButton,
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
