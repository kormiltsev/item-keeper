package commandui

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	app "github.com/kormiltsev/item-keeper/internal/app"
	appstorage "github.com/kormiltsev/item-keeper/internal/app/appstorage"
	"golang.org/x/term"
	//	appstorage "github.com/kormiltsev/item-keeper/internal/app/appstorage"
)

// manager is operates STDIN.
type manager struct {
	stopSignal chan os.Signal
	chclose    chan struct{}
	ctx        context.Context
	reader     *bufio.Reader
}

// appmanager is interface to operate STDIN and print to terminal.
var appmanager = manager{}

// var chclose chan struct{}

// quiter wait for signals from system
func (appman *manager) quiter() {
	<-appman.stopSignal
	appman.quit()
}

// StartTui starts app. First page - get secret word and authorisation page.
func StartTui(ctx context.Context, ch chan struct{}, stopSignal chan os.Signal) {

	//upload configs
	fmt.Println(app.UploadConfigsApp())

	appmanager = manager{
		stopSignal: stopSignal,
		chclose:    ch,
		ctx:        ctx,
		reader:     bufio.NewReader(os.Stdin),
	}

	fmt.Print("Enter your secret word: ")
	password, err := term.ReadPassword(int(os.Stdin.Fd()))
	// response, err := appmanager.reader.ReadString('\n')
	if err != nil {
		log.Println("unexp err:", err)
		close(appmanager.chclose)
		return
	}
	// response = deleteEndOfString(response)

	err = app.LoadFromFile(password)
	if err != nil {
		appmanager.regPage()
		return
	}

	app.SaveUserCryptoPass(password)

	appmanager.openMenu()
}

// regPage shows authorization page.
func (appman *manager) regPage() {
	fmt.Print("Please login or register\n[r] Registration\n[a] Autorisation\n[s] Secret word again\n[q] for quit\n")
	response, err := appman.reader.ReadString('\n')
	if err != nil {
		log.Println("unexp err:", err)
		appman.quit()
		return
	}
	response = deleteEndOfString(response)

	switch response {
	case "r":
		appman.regUser()
	case "a":
		appman.autUser()
	case "q":
		appman.quit()
	case "s":
		StartTui(appman.ctx, appman.chclose, appman.stopSignal)
	default:
		log.Println("unknown command:", response)
		appman.regPage()
	}
}

// quit start save data to file and close closing chanel
func (appman *manager) quit() {
	// appman.ctx.Done()

	fmt.Print("saving...")
	app.SaveToFile()
	fmt.Print("\nSee ya")
	close(appman.chclose)
}

// regUser request to register User
func (appman *manager) regUser() {
	login, pass, secret, err := appman.inputCreds()
	if err != nil {
		log.Println("unexp err:", err)
		close(appmanager.chclose)
		return
	}

	app.SaveUserCryptoPass(secret)

	err = app.RegUser(appman.ctx, login, pass)
	if err != nil {
		fmt.Println("Error:", err)
		appman.regPage()
		return
	}

	appman.openMenu()
}

// inputCreds request user to inpus login/password/secret.
func (appman *manager) inputCreds() (string, string, []byte, error) {
	fmt.Print("enter LOGIN:")
	login, err := appman.reader.ReadString('\n')
	if err != nil {
		return "", "", nil, err
	}

	fmt.Print("enter PASSWORD:")
	pass, err := appman.reader.ReadString('\n')
	if err != nil {
		return "", "", nil, err
	}

	fmt.Print("enter your SECRET WORD:")
	keyword, err := term.ReadPassword(int(os.Stdin.Fd()))
	if err != nil {
		return "", "", nil, err
	}
	return deleteEndOfString(login), deleteEndOfString(pass), keyword, nil
}

// autUser start authorization to server.
func (appman *manager) autUser() {
	login, pass, secret, err := appman.inputCreds()
	if err != nil {
		log.Println("unexp err:", err)
		close(appmanager.chclose)
		return
	}
	app.SaveUserCryptoPass(secret)

	err = app.AuthUser(appman.ctx, login, pass)
	if err != nil {
		fmt.Print("Error:", err)
		appman.regPage()
		return
	}

	appman.openMenu()
}

// openMenu prints menu.
func (appman *manager) openMenu() {

	fmt.Print("\n[ITEM KEEPER]\n[s] Search\n[qs] Quick search\n[a] Add new item\n[d] Delete item\n[e] Edit item\n[c] Catalog print\n[u] Update\n[info] Print app version\n[q] Quit\n")
	response, err := appman.reader.ReadString('\n')
	if err != nil {
		log.Println("unexp err:", err)
		fmt.Print("unexpected err:", err)
		appman.openMenu()
	}
	response = deleteEndOfString(response)

	switch response {
	case "s":
		appman.search()
	case "qs":
		appman.quicksearch()
	case "a":
		appman.addNewItem()
	case "d":
		appman.delItem()
	case "e":
		appman.editItem()
	case "q":
		appman.quit()
	case "c":
		appman.showMapa()
	case "u":
		appman.update()
	case "info":
		appman.info()
	default:
		log.Println("unknown command:", response)
		appman.openMenu()
	}
}

// info prints app info
func (appman *manager) info() {
	fmt.Println(app.PrintVersion())
	appman.openMenu()
}

// showMapa print all items from catalog.
func (appman *manager) showMapa() {
	mapa, err := app.ShowCatalog()
	if err != nil {
		fmt.Println("nothing was found:", err)
		go appman.openMenu()
		return
	}

	fmt.Println("+---------------------------------------+")
	for _, item := range mapa {
		printItem(item)
	}
	appman.openMenu()
}

// search returns item with requested words in parameters
func (appman *manager) search() {

	// here can be more than one parameter to search and more than one word. TODO "for" if required
	fmt.Print("Parameter's name:")
	pkey, err := appman.reader.ReadString('\n')
	if err != nil {
		log.Println("unexp err:", err)
		appman.openMenu()
	}

	pkey = deleteEndOfString(pkey)

	fmt.Printf("In [%s] looking for:", pkey)
	searchWord, err := appman.reader.ReadString('\n')
	if err != nil {
		log.Println("unexp err:", err)
		appman.openMenu()
	}

	searcher := app.NewSearchByParameter()
	searchSlise, ok := searcher.Mapa[pkey]
	if !ok {
		searchSlise = make([]string, 0)
	}

	searchSlise = append(searchSlise, deleteEndOfString(searchWord))
	searcher.Mapa[pkey] = searchSlise
	// =====================================================================================

	err = searcher.SearchItemByParameters()
	if err != nil {
		fmt.Print("nothing was found")
		appman.openMenu()
	}

	fmt.Println("+------------------------------------------+")
	for _, item := range searcher.Answer {
		printItem(item)
	}
	appman.openMenu()
}

// quicksearch returns item with word requested in any of parameter.
func (appman *manager) quicksearch() {

	fmt.Print("Looking for:")
	searchWord, err := appman.reader.ReadString('\n')
	if err != nil {
		log.Println("unexp err:", err)
		appman.openMenu()
		return
	}

	searchWord = deleteEndOfString(searchWord)

	if searchWord == "" || searchWord == " " {
		log.Println("unexp err:", err)
		appman.openMenu()
		return
	}

	searcher := app.NewSearchByParameter()
	searcher.QuickSearch = searchWord

	err = searcher.SearchItemQuick()
	if err != nil {
		fmt.Print("nothing was found")
		appman.openMenu()
		return
	}

	fmt.Println("+------------------------------------------+")
	for _, item := range searcher.Answer {
		printItem(item)
	}
	go appman.openMenu()
}

// printItem prints item data.
func printItem(item *appstorage.Item) {
	fmt.Printf("ID: %d\n", item.ItemID)
	for _, param := range item.Parameters {
		fmt.Printf("%s: %s\n", param.Name, param.Value)
	}
	if len(item.LocalAddresses) != 0 {
		fmt.Print("--Files: \n")
		for _, fileaddress := range item.LocalAddresses {
			fmt.Printf("   %s\n", filepath.Base(fileaddress))
			// FUN print image
			printImageASCII(fileaddress)
			// ===============
		}
	}
	fmt.Println("+------------------------------------------+")
}

// addNewItem recieve input item data and push to server
func (appman *manager) addNewItem() {

	newitem := app.NewAppItem()

	// Add parameters
	for {
		fmt.Print("Enter name of parameter ([n] to skip): ")
		pname, err := appman.reader.ReadString('\n')
		if err != nil {
			pname = "UNKNOWN"
		}
		pname = deleteEndOfString(pname)

		if pname == "n" {
			break
		}

		newparam := appstorage.Parameter{
			Name: pname,
		}

		fmt.Print("Next parameter's value: ")
		pvalue, err := appman.reader.ReadString('\n')
		if err != nil {
			pvalue = "UNKNOWN"
		}

		newparam.Value = deleteEndOfString(pvalue)

		newitem.Parameters = append(newitem.Parameters, newparam)
	}

	// Add files
	for {
		fmt.Print("Input file address ([n] to skip): ")
		faddress, err := appman.reader.ReadString('\n')
		if err != nil {
			fmt.Print("Error, try again")
			continue
		}
		faddress = deleteEndOfString(faddress)

		if faddress == "n" {
			break
		}

		newitem.UploadAddress = append(newitem.UploadAddress, faddress)
	}

	fmt.Print("New Item is ready:\n")
	for _, param := range newitem.Parameters {
		fmt.Printf("%s: %s\n", param.Name, param.Value)
	}
	for _, files := range newitem.UploadAddress {
		fmt.Printf("+ file: %s\n", files)
	}

	fmt.Print("\nSave Item? [y/n]")
	response, err := appman.reader.ReadString('\n')
	if err != nil {
		log.Println("response, err := appman.reader.ReadString('\n'):", err)
		appman.openMenu()
	}
	response = deleteEndOfString(response)

	if response == "y" || response == "Y" {
		err = app.AddNewItem(appman.ctx, newitem)
		if err != nil {
			fmt.Print("Error:", err)
		}
	} else {
		fmt.Print("Item is not saved")
	}

	appman.openMenu()
}

// delItem delete Item by ID
func (appman *manager) delItem() {

	listItemIDToDelete := make([]int64, 1)

	fmt.Print("Enter itemID to delete: ")
	itemid, err := appman.reader.ReadString('\n')
	if err != nil {
		appman.openMenu()
		return
	}

	listItemIDToDelete[0], err = strconv.ParseInt(deleteEndOfString(itemid), 10, 64)
	if err != nil {
		fmt.Print("Wrong id:", itemid)
		appman.openMenu()
		return
	}

	log.Println("listItemIDToDelete", listItemIDToDelete)
	notdeletedlist, err := app.DeleteItems(appman.ctx, listItemIDToDelete)
	if err != nil {
		fmt.Print("This id was not deleted:", notdeletedlist, "the reason is:", err)
	}
	appman.openMenu()
}

// update request for updates from server
func (appman *manager) update() {

	err := app.UpdateDataFromServer(appman.ctx)
	if err != nil {
		fmt.Print("trying to update, but error: ", err)
	} else {
		fmt.Print("everithing updated")
	}

	appman.openMenu()
}

// deleteEndOfString delete symbols of end of line in linux and windows
func deleteEndOfString(original string) string {
	original = strings.ReplaceAll(original, "\r\n", "")
	return strings.ReplaceAll(original, "\n", "")
}

// editItem rewrite data in Item
func (appman *manager) editItem() {
	fmt.Print("enter item ID to edit:")
	itemid, err := appman.reader.ReadString('\n')
	if err != nil {
		log.Println("STDIN error:", err)
		fmt.Println("item id not found")
		appman.openMenu()
		return
	}

	itemid = deleteEndOfString(itemid)

	searcher := app.NewSearchByParameter()
	searcher.ID, err = strconv.ParseInt(deleteEndOfString(itemid), 10, 64)
	if err != nil {
		fmt.Println("for item ID use digits")
		appman.openMenu()
		return
	}

	err = searcher.ReturnItemByID()
	if err != nil {
		if errors.Is(err, appstorage.ErrNotFound) {
			fmt.Println("Item not found")
		}
		log.Println("ReturnItemByID:", err)
		fmt.Println("item id not found:", err)
		appman.openMenu()
		return
	}

	if len(searcher.Answer) > 1 {
		log.Println("ReturnItemByID found more than 1 item by ID")
	}

	// show old data
	// printItem(searcher.Answer[searcher.ID])

	itemEdited := appstorage.NewItem(searcher.Answer[searcher.ID].UserID)
	itemEdited.ItemID = searcher.Answer[searcher.ID].ItemID

	fmt.Printf("Parameters (leave empty to delete):\n")
	for _, param := range searcher.Answer[searcher.ID].Parameters {
		fmt.Printf("OLD %s:%s\nNEW %s:", param.Name, param.Value, param.Name)

		pvaluenew, err := appman.reader.ReadString('\n')
		if err != nil {
			pvaluenew = "UNKNOWN"
		}
		pvaluenew = deleteEndOfString(pvaluenew)

		if pvaluenew != "" {
			param.Value = pvaluenew
			itemEdited.Parameters = append(itemEdited.Parameters, param)
		}
	}

	// Add new parameters
	for {
		fmt.Print("Add new parameter? (enter [parameter name] or [n] to skip): ")
		pname, err := appman.reader.ReadString('\n')
		if err != nil {
			pname = "UNKNOWN"
		}
		pname = deleteEndOfString(pname)

		if pname == "n" {
			break
		}

		newparam := appstorage.Parameter{
			Name: pname,
		}

		fmt.Print("Parameter's value: ")
		pvalue, err := appman.reader.ReadString('\n')
		if err != nil {
			pvalue = "UNKNOWN"
		}

		newparam.Value = deleteEndOfString(pvalue)

		itemEdited.Parameters = append(itemEdited.Parameters, newparam)
	}

	// FILES
	var filesToDelete = appstorage.NewItem(searcher.Answer[searcher.ID].UserID)

	// file addresses of item
	sl, ok := searcher.FileAddresses[searcher.ID]
	if !ok {
		fmt.Println("no files")
	}

	for _, fileaddress := range sl { //searcher.FileAddresses[searcher.ID] {
		// FUN print image
		printImageASCII(fileaddress)
		// ===============
		fmt.Printf("current file: %s\nDelete file? [y/n]:", filepath.Base(fileaddress))

		response, err := appman.reader.ReadString('\n')
		if err != nil {
			fmt.Println("file is keeped")
			continue
		}
		response = deleteEndOfString(response)

		if response == "y" || response == "Y" {
			flif := app.ReturnFileIDByAddress(fileaddress)
			if flif != 0 {
				filesToDelete.FileIDs = append(filesToDelete.FileIDs, flif)
			}
		}
	}

	// Add new files
	for {
		fmt.Print("Input file address ([n] to skip): ")
		faddress, err := appman.reader.ReadString('\n')
		if err != nil {
			fmt.Print("Error, try again")
			continue
		}
		faddress = deleteEndOfString(faddress)

		if faddress == "n" {
			break
		}

		itemEdited.UploadAddress = append(itemEdited.UploadAddress, faddress)
	}

	fmt.Print("Item is ready:\n")
	for _, param := range itemEdited.Parameters {
		fmt.Printf("%s: %s\n", param.Name, param.Value)
	}
	for _, files := range itemEdited.UploadAddress {
		fmt.Printf("+ file: %s\n", files)
	}
OLDSLICE:
	for _, fileaddress := range sl {
		for _, flid := range filesToDelete.FileIDs {
			if flid == app.ReturnFileIDByAddress(fileaddress) {
				continue OLDSLICE
			}
		}
		fmt.Printf("  file: %s\n", filepath.Base(fileaddress))
	}

	fmt.Print("\nSave Item? [y/n]")
	response, err := appman.reader.ReadString('\n')
	if err != nil {
		log.Println("response, err := appman.reader.ReadString('\n'):", err)
		appman.openMenu()
	}
	response = deleteEndOfString(response)

	if response == "y" || response == "Y" {
		err = app.EditItem(appman.ctx, itemEdited, filesToDelete)
		if err != nil {
			fmt.Print("Error:", err)
		}
	} else {
		fmt.Print("Item is not saved")
	}

	appman.openMenu()
}
