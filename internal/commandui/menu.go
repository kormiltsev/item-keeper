package commandui

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	app "github.com/kormiltsev/item-keeper/internal/app"
	appstorage "github.com/kormiltsev/item-keeper/internal/app/appstorage"
	//	appstorage "github.com/kormiltsev/item-keeper/internal/app/appstorage"
)

type manager struct {
	chclose chan struct{}
	ctx     context.Context
	reader  *bufio.Reader
}

var appmanager = manager{}

var chclose chan struct{}

func StartTui(ctx context.Context, ch chan struct{}) {

	//upload configs
	fmt.Println(app.UploadConfigsApp())

	appmanager = manager{
		chclose: ch,
		ctx:     ctx,
		reader:  bufio.NewReader(os.Stdin),
	}

	fmt.Print("Enter your secret word: ")
	response, err := appmanager.reader.ReadString('\n')
	if err != nil {
		log.Println("unexp err:", err)
		close(appmanager.chclose)
		return
	}
	response = deleteEndOfString(response)

	if response == "q" {
		appmanager.quit()
		return
	}

	err = app.LoadFromFile(response)
	if err != nil {
		appmanager.regPage()
		return
	}

	app.SaveUserCryptoPass(response)

	appmanager.openMenu()
}

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
		StartTui(appman.ctx, chclose)
	default:
		log.Println("unknown command:", response)
		appman.regPage()
	}
}

func (appman *manager) quit() {
	// appman.ctx.Done()

	fmt.Print("saving...")
	app.SaveToFile()
	fmt.Print("\nSee ya")
	close(appman.chclose)
}

func (appman *manager) regUser() {
	fmt.Print("enter LOGIN:")
	login, err := appman.reader.ReadString('\n')
	if err != nil {
		log.Println("unexp err:", err)
		close(appmanager.chclose)
		return
	}

	fmt.Print("enter PASSWORD:")
	pass, err := appman.reader.ReadString('\n')
	if err != nil {
		log.Println("unexp err:", err)
		close(appmanager.chclose)
		return
	}

	fmt.Print("enter your SECRET WORD:")
	secretword, err := appman.reader.ReadString('\n')
	if err != nil {
		log.Println("unexp err:", err)
		close(appmanager.chclose)
		return
	}

	app.SaveUserCryptoPass(deleteEndOfString(secretword))

	err = app.RegUser(appman.ctx, deleteEndOfString(login), deleteEndOfString(pass))
	if err != nil {
		fmt.Println("Error:", err)
		appman.regPage()
		return
	}

	appman.openMenu()
}

func (appman *manager) autUser() {
	fmt.Print("enter LOGIN:")
	login, err := appman.reader.ReadString('\n')
	if err != nil {
		log.Println("unexp err:", err)
		close(appmanager.chclose)
		return
	}

	fmt.Print("enter PASSWORD:")
	pass, err := appman.reader.ReadString('\n')
	if err != nil {
		log.Println("unexp err:", err)
		close(appmanager.chclose)
		return
	}

	fmt.Print("enter your SECRET WORD:")
	secretword, err := appman.reader.ReadString('\n')
	if err != nil {
		log.Println("unexp err:", err)
		close(appmanager.chclose)
		return
	}
	app.SaveUserCryptoPass(deleteEndOfString(secretword))

	err = app.AuthUser(appman.ctx, deleteEndOfString(login), deleteEndOfString(pass))
	if err != nil {
		fmt.Print("Error:", err)
		appman.regPage()
		return
	}

	appman.openMenu()
}

func (appman *manager) openMenu() {

	fmt.Print("\n[ITEM KEEPER]\n[s] Search\n[a] Add new item\n[d] Delete item\n[c] Catalog print\n[u] Update\n[q] Quit\n")
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
	case "a":
		appman.addNewItem()
	case "d":
		appman.delItem()
	case "q":
		appman.quit()
	case "c":
		appman.showMapa()
	case "u":
		appman.update()
	default:
		log.Println("unknown command:", response)
		appman.openMenu()
	}
}

func (appman *manager) showMapa() {
	mapa, err := app.ShowCatalog()
	if err != nil {
		fmt.Println("nothing was found")
	}

	fmt.Println("+---------------------------------------+")
	for _, item := range mapa {
		fmt.Printf("ID: %d\n", item.ItemID)
		for _, param := range item.Parameters {
			fmt.Printf("%s: %s\n", param.Name, param.Value)
		}
		if len(item.LocalAddresses) != 0 {
			fmt.Print("Files: \n")
			for _, fileaddress := range item.LocalAddresses {
				fmt.Printf("   %s\n", fileaddress)
			}
		}
		fmt.Println("+---------------------------------------+")
	}
	appman.openMenu()
}

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
		fmt.Print("nothing found")
		appman.openMenu()
	}

	for k, v := range searcher.Answer {
		fmt.Printf("%d: %v\n", k, v.Parameters)
		if len(searcher.FileAddresses) != 0 {
			fmt.Printf("Files: %v\n", searcher.FileAddresses)
		}
	}

	appman.openMenu()
}

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
		fmt.Printf("--file: %s\n", files)
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

func (appman *manager) update() {

	err := app.UpdateDataFromServer(appman.ctx)
	if err != nil {
		fmt.Print("trying to update, but error: ", err)
	} else {
		fmt.Print("everithing updated")
	}

	appman.openMenu()
}

func deleteEndOfString(original string) string {
	original = strings.ReplaceAll(original, "\r\n", "")
	return strings.ReplaceAll(original, "\n", "")
}
