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
	a := app.NewWithID("org.buetow.quicklogger")
	w := a.NewWindow("Quick logger")
	// Same dir as my Obsidian
	storageDir := "/storage/emulated/0/Notes/Vault"

	input := widget.NewMultiLineEntry()
	input.SetPlaceHolder("Enter text here!")

	button := widget.NewButton("Log text", func() {
		content := input.Text
		filename := fmt.Sprintf("%s/quicklog-%s.md", storageDir, getSHA256Hash(content))
		err := os.WriteFile(filename, []byte(content), 0644)
		if err != nil {
			log.Println("Error writing to file:", err)
			input.SetText(err.Error())
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
	w.ShowAndRun()
}

func getSHA256Hash(text string) string {
	hasher := sha256.New()
	hasher.Write([]byte(text))
	return hex.EncodeToString(hasher.Sum(nil))
}
