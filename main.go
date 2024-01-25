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
	appId            = "org.buetow.quicklogger"
	defaultDirectory = "."
)

var windowSize = fyne.NewSize(200, 100)

func createPreferenceWindow(a fyne.App, mainWindow fyne.Window) fyne.Window {
	window := a.NewWindow("Preferences")
	window.Resize(windowSize)
	directoryPreference := widget.NewEntry()
	directoryPreference.SetText(a.Preferences().StringWithFallback("Directory", defaultDirectory))

	saveButton := widget.NewButton("Save", func() {
		a.Preferences().SetString("Directory", directoryPreference.Text)
		window.Hide()
		mainWindow.Show()
	})

	preferencesContent := container.NewVBox(
		widget.NewLabel("Directory"),
		directoryPreference,
		saveButton,
	)
	window.SetContent(preferencesContent)

	return window
}

func createMainWindow(a fyne.App) fyne.Window {
	window := a.NewWindow("Quick logger")

	input := widget.NewMultiLineEntry()
	input.Wrapping = fyne.TextWrapWord
	input.SetPlaceHolder("Enter text here...")

	button := widget.NewButton("Log text", func() {
		filename := fmt.Sprintf("%s/ql-%s.md",
			a.Preferences().StringWithFallback("Directory", defaultDirectory),
			time.Now().Format("060102-150405"))
		err := os.WriteFile(filename, []byte(input.Text), 0644)
		if err != nil {
			dialog.ShowError(err, window)
			return
		}
		input.SetText("")
	})

	window.SetContent(container.NewVBox(
		input,
		button,
		widget.NewButton("Preferences", func() {
			createPreferenceWindow(a, window).Show()
		}),
	))
	window.Resize(windowSize)
	window.Canvas().Focus(input)

	return window
}

func main() {
	createMainWindow(app.NewWithID(appId)).ShowAndRun()
}
