package tui

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"

	client "github.com/kormiltsev/item-keeper/internal/client"
	pb "github.com/kormiltsev/item-keeper/internal/server/proto"
)

var ram = []*pb.Uitem{}

type catal struct {
	List *[]*pb.Uitem
}

var catalog = catal{List: &ram}

func updateCatalog(cc *client.ClientConnector) {
	cc.LastUpdate = userSettings.lastUpdate

	// send lastupdate date from client to server and request catalog if LUclient != LUserver
	cc.ListOfItems(context.Background())

	// if there new data on server, upload it to ram
	if cc.LastUpdate != userSettings.lastUpdate {
		saveNewCatalog(cc)

		saveFilesFromResponse(cc)
	}
}

func saveFilesFromResponse(cc *client.ClientConnector) {
	path := filepath.Join(userSettings.datastorage, cc.UserID)
	// hash := md5.New()
	for i, item := range cc.Items {
		for j, file := range item.Images {
			if len(file.Body) != 0 {
				path = file.Title
			}
		}
	}
	b := []byte(cc.Title)
	path = filepath.Join(path, fmt.Sprintf("%x", hash.Sum(b)))
	// path = filepath.Join(path, fsf.Title)

	log.Println("path1:", path)

	err := os.MkdirAll(path, os.ModePerm)
	if err != nil {
		log.Println("file storage can't create directory wile creating:", err)
		return
	}

	// path := configs.ServiceConfig.FileServerAddress + item.UserID + item.Title

	// create random file name
	// hash := md5.New()
	b = []byte(fsf.UserID + fsf.Title)
	path = filepath.Join(path, fmt.Sprintf("%x", hash.Sum(b)))

	log.Println("path2:", path)

	err = os.WriteFile(path, *fsf.Data, 0644)
	if err != nil {
		log.Println("file storage error, check FILESERVERADDRESS available:", err)
	}
}

func saveNewCatalog(cc *client.ClientConnector) {
	if cc.LastUpdate == userSettings.lastUpdate {
		return
	}

	userSettings.lastUpdate = cc.LastUpdate

	ram = make([]*pb.Uitem, len(cc.Items))
	copy(ram, cc.Items)

}

func AddNewItemsToMapa(cc *client.ClientConnector) {

	// upload file
	// for i, item := range cc.Items {

	for j, file := range cc.Items[0].Images {

		dat, err := readFile(file.Title)
		if err != nil {
			log.Println("can't find file", file.Title, " Err:", err)
		} else {
			cc.Items[0].Images[j].Body = dat
		}

	}

	// send to server
	err := cc.AddNewItem(context.Background())
	if err != nil {
		log.Println("error in additind item to server:", err)
		return
	}

	// if server's LA after client's LA, then upload all items from server
	if cc.LastUpdate > userSettings.lastUpdate {
		cc.ListOfItems(context.Background())
		log.Println("get catalog in Additing new item:", cc.Items)
		ram = cc.Items
		userSettings.lastUpdate = cc.LastUpdate
		return
	}

	// update local date
	userSettings.lastUpdate = cc.Items[0].Lastupdate

	// else add only las added item
	ram = append(ram, cc.Items...)
}

func delItemMapa(cc *client.ClientConnector) {

	// log.Println("in mapa start delete", cc.Items[0].Id)
	// log.Println("mapa", ram)
	cc.Items = cc.Items[:0]
	cc.Items = append(cc.Items, &pb.Uitem{Id: ram[ind].Id})

	err := cc.DeleteItem(context.Background())
	if err != nil {
		log.Println("try delete error:", err, " gRPC STATUS:", cc.Error)
		return
	}

	// if server's LA after client's LA, then upload all items from server
	if cc.LastUpdate > userSettings.lastUpdate {
		cc.ListOfItems(context.Background())
		log.Println("get catalog in Additing new item:", cc.Items)
		ram = cc.Items
		return
	}

	userSettings.lastUpdate = cc.Items[0].Lastupdate

	//delete in ram
	if len(ram) < 2 {
		ram = ram[:0]
		return
	}

	if ind == len(ram) {
		ram = ram[:len(ram)-1]
	} else {
		ram = append(ram[:ind], ram[ind+1:]...)
	}

}

func editItemMapa(cc *client.ClientConnector) {

	// create edited item
	editedItem := pb.Uitem{
		Id:   ram[ind].Id,
		Name: ram[ind].Name + "_Edited",
	}

	for _, param := range ram[ind].Params {
		editedItem.Params = append(editedItem.Params, &pb.Parameter{Name: param.Name, Value: "Edited"})
	}
	// ==================

	cc.Items = cc.Items[:0]
	cc.Items = append(cc.Items, &editedItem)

	err := cc.EditItem(context.Background())
	if err != nil {
		log.Println("try edit error:", err, " gRPC STATUS:", cc.Error)
		return
	}

	// if server's LA after client's LA, then upload all items from server
	if cc.LastUpdate > userSettings.lastUpdate {
		cc.ListOfItems(context.Background())
		log.Println("get catalog in Additing new item:", cc.Items)
		ram = cc.Items
		return
	}

	userSettings.lastUpdate = cc.Items[0].Lastupdate

	for i, it := range ram {
		if it.Id == cc.Items[0].Id {
			ram[i] = cc.Items[0]
			return
		}
	}

}

func readFile(filename string) ([]byte, error) {
	file, err := os.Open(filename)

	if err != nil {
		return nil, err
	}
	defer file.Close()

	stats, statsErr := file.Stat()
	if statsErr != nil {
		return nil, statsErr
	}

	var size int64 = stats.Size()
	bytes := make([]byte, size)

	bufr := bufio.NewReader(file)
	_, err = bufr.Read(bytes)

	return bytes, err
}
