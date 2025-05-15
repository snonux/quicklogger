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
)

var (
	defaultDirectory = "."
	defaultTagItems  = []string{
		"log",
		"share",
		"share:li",
		"share:ma",
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
	input.Wrapping = fyne.TextWrapWord
	input.SetPlaceHolder(placeholderText)
	input.SetMinRowsVisible(30)

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
	})

	window.SetContent(container.NewVBox(
		container.NewHBox(daysDropdown, tagDropdown, whatDropdown, logTextButton),
		input,
		widget.NewButton("Preferences", func() {
			createPreferenceWindow(a).Show()
		}),
	))
	window.Resize(windowSize)
	window.Canvas().Focus(input)

	return window
}

func main() {
	createMainWindow(app.NewWithID(appId)).ShowAndRun()
}
