package main

import (
	"log"
	"math/rand"
	"strconv"

	"github.com/rivo/tview"
)

type Item struct {
	UserID     string
	ItemID     int64
	Parameters []string
}

type App struct {
	app          *tview.Application
	pages        *tview.Pages
	list         *tview.List
	form         *tview.Form
	items        []*Item
	itemIndex    int
	modal        *tview.Modal
	modalVisible bool
}

func main() {
	app := NewApp()
	if err := app.Run(); err != nil {
		log.Fatal(err)
	}
}

func NewApp() *App {
	app := &App{
		app:          tview.NewApplication(),
		pages:        tview.NewPages(),
		list:         tview.NewList(),
		form:         tview.NewForm(),
		items:        []*Item{},
		itemIndex:    -1,
		modalVisible: false,
	}

	app.list.
		SetSelectedFunc(app.showItemDetails)

	app.form.
		AddInputField("UserID", "", 30, nil, nil).
		AddButton("Create", app.createItem).
		AddButton("Delete", app.deleteItem).
		AddButton("Quit", func() {
			app.app.Stop()
		})

	app.modal = tview.NewModal().
		SetDoneFunc(func(buttonIndex int, buttonLabel string) {
			app.hideModal()
		})

	app.pages.AddPage("main", app.createMainPage(), true, true)

	return app
}

func (app *App) createMainPage() tview.Primitive {
	grid := tview.NewGrid().SetRows(0, 1).
		AddItem(app.list, 0, 0, 1, 1, 0, 0, true).
		AddItem(app.form, 1, 0, 1, 1, 0, 0, true)

	return grid
}

func (app *App) Run() error {
	app.app.SetRoot(app.pages, true)
	app.app.EnableMouse(true)

	return app.app.Run()
}

func (app *App) createItem() {
	userID := app.form.GetFormItemByLabel("UserID").(*tview.InputField).GetText()

	if userID == "" {
		app.showMessage("Please enter a UserID")
		return
	}

	itemID := generateItemID()
	item := &Item{
		UserID:     userID,
		ItemID:     itemID,
		Parameters: []string{},
	}

	app.items = append(app.items, item)
	app.list.AddItem(item.UserID, "", 0, nil)
	app.list.SetCurrentItem(len(app.items) - 1)

	app.showMessage("Item created successfully")
}

func (app *App) deleteItem() {
	if app.itemIndex >= 0 && app.itemIndex < len(app.items) {
		app.items = append(app.items[:app.itemIndex], app.items[app.itemIndex+1:]...)
		app.list.RemoveItem(app.itemIndex)
		app.itemIndex = -1

		app.showMessage("Item deleted successfully")
	}
}

func (app *App) showItemDetails(index int, mainText string, secondaryText string, shortcut rune) {
	app.itemIndex = index

	if app.itemIndex >= 0 && app.itemIndex < len(app.items) {
		item := app.items[app.itemIndex]

		app.form.Clear(true)
		app.form.SetTitle("Item Details - ItemID: " + strconv.FormatInt(item.ItemID, 10))

		app.form.
			AddInputField("UserID", item.UserID, 30, nil, func(text string) {
				app.updateItemDetails("UserID", text)
			}).
			AddButton("Add Parameter", app.addParameter).
			AddButton("Remove Parameter", app.removeParameter)

		for _, param := range item.Parameters {
			app.form.AddInputField("Parameter", param, 30, nil, nil)
		}

		app.app.SetFocus(app.form)
	}
}

func (app *App) addParameter() {
	app.form.AddInputField("Parameter", "", 30, nil, nil)
	app.app.SetFocus(app.form)
}

func (app *App) removeParameter() {
	items := app.form.GetFormItemCount()
	if items > 3 {
		app.form.RemoveFormItem(items - 1)
		app.app.SetFocus(app.form)
	}
}

func (app *App) updateItemDetails(label string, value string) {
	if app.itemIndex >= 0 && app.itemIndex < len(app.items) {
		item := app.items[app.itemIndex]

		switch label {
		case "UserID":
			item.UserID = value
		}
	}
}

func (app *App) showMessage(message string) {
	app.modal.SetText(message)
	app.showModal()
}

func (app *App) showModal() {
	if app.modalVisible {
		return
	}

	app.pages.AddPage("modal", app.modal, true, true)
	app.modalVisible = true

	app.app.SetFocus(app.modal)
}

func (app *App) hideModal() {
	if !app.modalVisible {
		return
	}

	app.pages.RemovePage("modal")
	app.modalVisible = false

	app.app.SetFocus(app.list)
}

func generateItemID() int64 {
	// Generate a random item ID
	return 1000 + rand.Int63n(9000)
}
