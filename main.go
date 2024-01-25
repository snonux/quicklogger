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

func main() {
	a := app.NewWithID("org.buetow.quicklogger")
	a.Preferences().SetString("Directory", "/storage/emulated/0/Notes/Vault")
	w := a.NewWindow("Quick logger")

	// Same dir as my Obsidian
	storageDir := a.Preferences().String("Directory")

	input := widget.NewMultiLineEntry()
	input.SetPlaceHolder("Enter text here!")

	button := widget.NewButton("Log text", func() {
		content := input.Text
		filename := fmt.Sprintf("%s/quicklog-%d.md", storageDir, time.Now().Unix())
		err := os.WriteFile(filename, []byte(content), 0644)
		if err != nil {
			dialog.ShowError(err, w)
		} else {
			input.SetText("")
		}
	})

	w.SetContent(container.NewVBox(
		widget.NewLabel("To be in the .zone!"),
		input,
		button,
	))
	w.Resize(fyne.NewSize(200, 100))
	w.Canvas().Focus(input)
	w.ShowAndRun()
}
