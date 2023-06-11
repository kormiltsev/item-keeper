package main

import (
	"github.com/rivo/tview"
)

type Item struct {
	UserID     string
	ItemID     int64
	Parameters []string
}

func main() {
	app := tview.NewApplication()

	// Create a modal
	modal := tview.NewModal().
		SetText("This is a modal").
		AddButtons([]string{"OK"}).
		SetDoneFunc(func(buttonIndex int, buttonLabel string) {
			// Hide the modal when the "OK" button is pressed
			app.SetRoot(nil, false)
			app.Draw()
		})

	// Show the modal
	app.SetRoot(modal, true)

	if err := app.Run(); err != nil {
		panic(err)
	}
}
