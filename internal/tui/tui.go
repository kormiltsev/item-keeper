package tui

import (
	"context"

	tcell "github.com/gdamore/tcell/v2"
	client "github.com/kormiltsev/item-keeper/internal/client"
	pb "github.com/kormiltsev/item-keeper/internal/server/proto"
	"github.com/rivo/tview"
)

var userSettings = struct {
	userID      string
	lastUpdate  int64
	datastorage string
}{userID: "NewClient", lastUpdate: 0, datastorage: "./data/client"}

var ind int

// Tview
var pages = tview.NewPages()
var itemData = tview.NewTextView()
var app = tview.NewApplication()
var form = tview.NewForm()
var itemsList = tview.NewList().ShowSecondaryText(false)
var flex = tview.NewFlex()

// var load = tview.NewFlex()
var text = tview.NewTextView().
	SetTextColor(tcell.ColorGreen).
	SetText("(a) Add new item \n(d) to Delete \n(e) to Edit item \n(u) to Update \n(q) Quit")

func StartTui(ctx context.Context) {
	// NewItem returns Uitems.List = []Item
	cli := client.NewClientConnector(userSettings.userID)
	defer client.CloseConnection()

	cli.LastUpdate = userSettings.lastUpdate

	// send lastupdate date from client to server and request catalog if LUclient != LUserver
	cli.ListOfItems(ctx)

	// if there new data on server, upload it to ram
	if cli.LastUpdate != userSettings.lastUpdate {
		saveNewCatalog(cli)
	}

	itemsList.SetSelectedFunc(func(index int, name string, second_name string, shortcut rune) {
		ind = index
		setItemText(ram[index])
	})

	makeListOfItems()

	flex.SetDirection(tview.FlexRow).
		AddItem(tview.NewFlex().
			AddItem(itemsList, 0, 1, true).
			AddItem(itemData, 0, 4, false), 0, 6, false).
		AddItem(text, 0, 1, false)

	// load.SetDirection(tview.FlexRow).
	// 	AddItem(text, 0, 1, false)

	flex.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Rune() {
		case 113:
			app.Stop()
		case 97:
			form.Clear(true)
			addItemForm(cli)
			pages.SwitchToPage("Add Item")
		case 100:
			form.Clear(true)
			delItemMapa(cli)
			makeListOfItems()
			pages.SwitchToPage("Menu")
		case 101:
			form.Clear(true)
			editItemMapa(cli)
			makeListOfItems()
			pages.SwitchToPage("Menu")
		case 117:
			form.Clear(true)
			updateCatalog(cli)
			makeListOfItems()
			pages.SwitchToPage("Menu")

		}
		return event
	})

	pages.AddPage("Menu", flex, true, true)
	pages.AddPage("Add Item", form, true, false)

	if err := app.SetRoot(pages, true).EnableMouse(true).Run(); err != nil {
		panic(err)
	}

}

func makeListOfItems() {

	// add to list
	// uitems.List = append(uitems.List, uitems.EditItem)

	itemsList.Clear()

	for index, ite := range *catalog.List {
		itemsList.AddItem(ite.Name+" ", " ", rune(49+index), nil)
	}
}

func setItemText(item *pb.Uitem) {
	itemData.Clear()
	text := item.Name + " id:" + item.Id + "\n"
	for _, v := range item.Params {
		text += v.Name + ": " + v.Value + "\n"
	}
	itemData.SetText(text)
}

func createpbItem() pb.Uitem {
	return pb.Uitem{Params: make([]*pb.Parameter, 0), Images: make([]*pb.Image, 0)}
}

func addItemForm(cc *client.ClientConnector) *tview.Form {

	newitem := createpbItem()

	form.AddInputField("Name", newitem.Name, 20, nil, func(name string) {
		newitem.Name = name
	})

	form.AddInputField("Parameter1", "", 100, nil, func(par1 string) {
		newitem.Params = append(newitem.Params, &pb.Parameter{Name: "Parameter1", Value: par1})
	})

	form.AddInputField("Parameter2", "", 100, nil, func(par2 string) {
		newitem.Params = append(newitem.Params, &pb.Parameter{Name: "Parameter2", Value: par2})
	})

	form.AddInputField("Parameter3", "", 100, nil, func(par3 string) {
		newitem.Params = append(newitem.Params, &pb.Parameter{Name: "Parameter3", Value: par3})
	})

	// form.AddInputField("File address", "", 100, nil, func(file1 string) {
	// 	newitem.Images = append(newitem.Images, &pb.Image{Title: file1})
	// })

	form.AddCheckbox("send file", false, func(fileadr bool) {
		newitem.Images = append(newitem.Images, &pb.Image{Title: "./data/sourceClient/test.txt"})
	})

	form.AddButton("Save", func() {
		// uitems.List = append(uitems.List, uitems.EditItem)
		cc.Items = make([]*pb.Uitem, 1)
		cc.Items[0] = &newitem

		AddNewItemsToMapa(cc)
		makeListOfItems()
		pages.SwitchToPage("Menu")
	})

	return form
}
