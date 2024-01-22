package main

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"log"
	"os"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

func main() {
	myApp := app.NewWithID("org.buetow.quicklogger")
	myWindow := myApp.NewWindow("Quick logger")

	storageDir := fyne.CurrentApp().Storage().RootURI().Path()
	label := widget.NewLabel(storageDir)

	input := widget.NewMultiLineEntry()
	input.SetPlaceHolder("Enter text here.")

	button := widget.NewButton("Log text", func() {
		content := input.Text
		filename := fmt.Sprintf("%s/quicklog-%s.txt", storageDir, getSHA256Hash(content))
		err := os.WriteFile(filename, []byte(content), 0644)
		if err != nil {
			log.Println("Error writing to file:", err)
			input.SetText(err.Error())
		} else {
			input.SetText("")
		}
	})

	myWindow.SetContent(container.NewVBox(
		label,
		input,
		button,
	))
	myWindow.Resize(fyne.NewSize(200, 100))
	myWindow.ShowAndRun()
}

func getSHA256Hash(text string) string {
	hasher := sha256.New()
	hasher.Write([]byte(text))
	return hex.EncodeToString(hasher.Sum(nil))
}
