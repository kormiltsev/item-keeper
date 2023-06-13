package commandui

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"os"
	"strconv"

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
	}

	if response == "q\n" {
		appmanager.quit()
		return
	}

	err = app.LoadFromFile(response[:len(response)-1])
	if err != nil {
		appmanager.regPage()
		return
	}
	appmanager.openMenu()
}

func (appman *manager) regPage() {
	fmt.Print("Please login or register\n[r] Registration\n[a] Autorisation\n[s] Secret word\n[q] for quit\n")
	response, err := appman.reader.ReadString('\n')
	if err != nil {
		log.Println("unexp err:", err)
		appman.quit()
		return
	}
	switch response {
	case "r\n":
		appman.regUser()
	case "a\n":
		appman.autUser()
	case "q\n":
		appman.quit()
	case "s\n":
		StartTui(appman.ctx, chclose)
	default:
		log.Println("unknown command")
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
	}

	fmt.Print("enter PASSWORD:")
	pass, err := appman.reader.ReadString('\n')
	if err != nil {
		log.Println("unexp err:", err)
		close(appmanager.chclose)
	}

	fmt.Print("enter your SECRET WORD:")
	secretword, err := appman.reader.ReadString('\n')
	if err != nil {
		log.Println("unexp err:", err)
		close(appmanager.chclose)
	}

	app.SaveUserCryptoPass(secretword[:len(secretword)-1])

	err = app.RegUser(appman.ctx, login, pass)
	if err != nil {
		fmt.Println("Error:", err)
		appman.regPage()
	}

	appman.openMenu()
}

func (appman *manager) autUser() {
	fmt.Print("enter LOGIN:")
	login, err := appman.reader.ReadString('\n')
	if err != nil {
		log.Println("unexp err:", err)
		close(appmanager.chclose)
	}

	fmt.Print("enter PASSWORD:")
	pass, err := appman.reader.ReadString('\n')
	if err != nil {
		log.Println("unexp err:", err)
		close(appmanager.chclose)
	}

	err = app.AuthUser(appman.ctx, login, pass)
	if err != nil {
		fmt.Print("Error:", err)
		appman.regPage()
	}

	appman.openMenu()
}

func (appman *manager) openMenu() {

	fmt.Print("\n[ITEM KEEPER]\n[s] Search\n[a] Add new item\n[d] Delete item\n[c] Catalog print\n[q] Quit\n")
	response, err := appman.reader.ReadString('\n')
	if err != nil {
		log.Println("unexp err:", err)
		fmt.Print("unexpected err:", err)
		appman.openMenu()
	}
	switch response {
	case "s\n":
		appman.search()
	case "a\n":
		appman.addNewItem()
	case "d\n":
		appman.delItem()
	case "q\n":
		appman.quit()
	case "c\n":
		appman.showMapa()
	default:
		appman.openMenu()
	}
}

func (appman *manager) showMapa() {
	mapa, err := app.ShowCatalog()
	if err != nil {
		fmt.Println("nothing found")
	}

	fmt.Println("+---------------------------------------+")
	for _, item := range mapa {
		fmt.Printf("ID: %d\n", item.ItemID)
		for _, param := range item.Parameters {
			fmt.Printf("%s: %s\n", param.Name, param.Value)
		}
		if len(item.UploadAddress) != 0 {
			fmt.Print("Files: \n")
			for _, fileaddress := range item.UploadAddress {
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

	fmt.Printf("In [%s] looking for:", pkey[:len(pkey)-1])
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

	searchSlise = append(searchSlise, searchWord[:len(searchWord)-1])
	searcher.Mapa[pkey[:len(pkey)-1]] = searchSlise
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
		fmt.Print("One more parameter? [y/n]")
		response, err := appman.reader.ReadString('\n')
		if err != nil {
			log.Println("response, err := appman.reader.ReadString('\n'):", err)
			appman.openMenu()
		}
		if response == "n\n" {
			break
		}

		fmt.Print("Next parameter NAME: ")
		pname, err := appman.reader.ReadString('\n')
		if err != nil {
			pname = "UNKNOWN"
		}
		newparam := appstorage.Parameter{
			Name: pname[:len(pname)-1],
		}

		fmt.Print("Next parameter VALUE: ")
		pvalue, err := appman.reader.ReadString('\n')
		if err != nil {
			pvalue = "UNKNOWN"
		}
		newparam.Value = pvalue[:len(pvalue)-1]

		newitem.Parameters = append(newitem.Parameters, newparam)
	}

	// Add files
	for {
		fmt.Print("Add one more file? [y/n]")
		response, err := appman.reader.ReadString('\n')
		if err != nil {
			log.Println("response, err := appman.reader.ReadString('\n'):", err)
			appman.openMenu()
		}
		if response == "n\n" {
			break
		}

		fmt.Print("Input file address: ")
		faddress, err := appman.reader.ReadString('\n')
		if err != nil {
			fmt.Print("Error, try again")
			continue
		}
		newitem.UploadAddress = append(newitem.UploadAddress, faddress[:len(faddress)-1])
	}

	fmt.Print("New Item is ready:\n")
	for _, param := range newitem.Parameters {
		fmt.Printf("%s: %s\n", param.Name, param.Value)
	}

	fmt.Print("\nSave Item? [y/n]")
	response, err := appman.reader.ReadString('\n')
	if err != nil {
		log.Println("response, err := appman.reader.ReadString('\n'):", err)
		appman.openMenu()
	}
	if response == "y\n" || response == "Y\n" {
		err = app.AddNewItem(appman.ctx, newitem)
		if err != nil {
			fmt.Print("Error:", err)
		}
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

	listItemIDToDelete[0], err = strconv.ParseInt(itemid[:len(itemid)-1], 10, 64)
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
