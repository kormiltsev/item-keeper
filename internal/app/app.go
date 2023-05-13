package app

import (
	"crypto/sha1"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"log"
	"strconv"

	serverstorage "github.com/kormiltsev/item-keeper/internal/serverstorage"

	appstorage "github.com/kormiltsev/item-keeper/internal/app/appstorage"
)

var (
	currentuser            string = "AppUser"
	currentuserpassword    string = "password"
	currentuserencryptokey string = "manualLocalKey"
	currentlastupdate      int64  = 0
)

func RegUser(login, password string) error {
	// create truct from storage
	tostorage := serverstorage.NewToStorage()

	// encrypt login and password here
	login, password = encodeLoginPass(login, password)
	currentuser = login
	currentuserpassword = password

	// add user creds to struct
	user := serverstorage.User{
		Login:    login,
		Password: password,
	}
	tostorage.User = &user

	// select DB interface
	tostorage.DB = serverstorage.NewStorager(tostorage)
	if tostorage.Error != nil {
		return tostorage.Error
	}

	// reg user on server
	err := tostorage.DB.RegUser()
	if err != nil {
		log.Println("reg user error from server:", err)
	}

	// save local
	appstorage.NewUser(tostorage.User.UserID, 0)

	// save current user id
	currentuser = tostorage.User.UserID
	currentlastupdate = tostorage.User.LastUpdate

	return nil
}

func encodeLoginPass(login, password string) (string, string) {
	h := sha1.New()
	h.Write([]byte(login))
	login = hex.EncodeToString(h.Sum(nil))

	sum := sha256.Sum256([]byte(login + password))
	password = hex.EncodeToString(sum[:])

	return login, password
}

func AuthUser(login, password string) error {
	// create truct from storage
	tostorage := serverstorage.NewToStorage()

	// encrypt login and password here
	login, password = encodeLoginPass(login, password)
	currentuser = login
	currentuserpassword = password

	// add user creds to struct
	user := serverstorage.User{
		Login:    login,
		Password: password,
	}
	tostorage.User = &user

	// select DB interface
	tostorage.DB = serverstorage.NewStorager(tostorage)
	if tostorage.Error != nil {
		return tostorage.Error
	}

	// reg user on server
	err := tostorage.DB.AuthUser()
	if err != nil {
		log.Println("auth user error from server:", err)
	}

	// save local
	// using NewUser() due to oneClient = oneUser rule (for now)
	appstorage.NewUser(tostorage.User.UserID, 0)

	// save current user id
	currentuser = tostorage.User.UserID
	currentlastupdate = 0

	// and go get catalog from server
	UpdateDataFromServer() // used currentuser and currentlastupdate

	return nil
}

func presetItem() *appstorage.Item {
	return &appstorage.Item{
		UserID:        currentuser,
		Parameters:    []appstorage.Parameter{{Name: "Parameter1", Value: "val1"}, {Name: "Parameter2", Value: "val2"}},
		UploadAddress: []string{"./data/sourceClient/test.txt", "./data/sourceClient/Jocker.jpeg"},
	}
}

func AddNewItem() {
	var err error
	appitem := presetItem()

	// encode data
	serveritem := serverstorage.NewItem()
	serveritem.UserID = appitem.UserID
	serveritem.Body, err = appitem.Encode(currentuserencryptokey)
	if err != nil {
		log.Println("error encrypt item:", err)
		return
	}

	// prepare to server
	tostor := serverstorage.NewToStorage()
	tostor.List = append(tostor.List, *serveritem)
	tostor.DB = serverstorage.NewStorager(tostor)

	// to server
	err = tostor.DB.PutItems()
	if err != nil {
		log.Println("put item to server err:", err)
		return
	}

	// if empty response
	if len(tostor.List) == 0 || tostor.List[0].ItemID == "" {
		log.Println("FAIL: PutItems() empty response from server")
		return
	}

	// save new itemID into local item
	appitem.ItemID = tostor.List[0].ItemID

	// prepare and send files after NewItemID was created by server
	for i, fileaddress := range appitem.UploadAddress {
		file := appstorage.NewFileStruct()
		file.FileID = strconv.Itoa(i) // temporary id to upload
		file.ItemID = tostor.List[0].ItemID
		file.UserID = tostor.List[0].UserID
		file.Address = fileaddress

		// read and encode file
		err := file.PrepareFile(currentuserencryptokey)
		if err != nil {
			log.Println("can't prepare file", fileaddress, "error:", err)
			continue
		}

		// to server storage
		tostorfile := serverstorage.NewFile()
		tostorfile.FileID = file.FileID
		tostorfile.ItemID = file.ItemID
		tostorfile.UserID = file.UserID
		tostorfile.Body = file.Body

		// create tostor
		tostor := serverstorage.NewToStorage()
		tostor.File = *tostorfile
		tostor.DB = serverstorage.NewStorager(tostor)

		// upload file to server
		err = tostor.DB.UploadFile()
		if err != nil {
			log.Println("can't upload file to server:", tostor.File.Address, "error:", err)
			// retry delivery
			//
			// ==============
		}

		// parse result (reuse local body)
		file.FileID = tostor.File.FileID

		// save file local and register in Catalog.Files
		err = file.SaveFileLocal(currentuserencryptokey)
		if err != nil {
			log.Println("Can't save file local:", err)
			continue
		}

		// copy file address into new appitem
		appitem.FileIDs = append(appitem.FileIDs, file.FileID)
	}

	// save item to Catalog, request interface
	operator, erro := appstorage.ReturnOperator(appitem.UserID)
	if erro != nil {
		log.Println(erro)
	}

	// save local
	err = operator.PutItems(appitem)
	if err != nil {
		log.Println("can't save item local:", err)
	}
}

func SearchItemByParameters() {

	searchlist := map[string][]string{
		"Parameter1": {"val", "foo"},
		"Parameter2": {"2", "boo"},
	}

	// prepare to server
	oper, err := appstorage.ReturnOperator(currentuser)
	if err != nil {
		log.Println("user not found in local memory. RegUser before SearchItemByParameters()")
		return
	}

	for key, val := range searchlist {
		if _, ok := oper.Search[key]; !ok {
			oper.Search[key] = make([]string, 0)
		}
		oper.Search[key] = append(oper.Search[key], val...)
	}

	err = oper.FindItemByParameter()
	if err != nil {
		log.Printf("FAIL search error: %v, looking for:%v", err, searchlist)
	}

	ans := "search results:\n"

	for _, item := range oper.Answer {
		ans = fmt.Sprintf("%s%v\n", ans, item.Parameters)
	}
	log.Println(ans)
}

func UpdateDataFromServer() {

	// prepare to server
	tostor := serverstorage.NewToStorage()
	tostor.User.UserID = currentuser
	tostor.User.LastUpdate = currentlastupdate
	tostor.DB = serverstorage.NewStorager(tostor)

	// to server
	err := tostor.DB.UpdateByLastUpdate()
	if err != nil {
		log.Println("put item to server err:", err)
		return
	}

	if len(tostor.List) == 0 {
		log.Println("everything updated")
		return
	}

	// save item to Catalog, request interface
	operator, erro := appstorage.ReturnOperator(tostor.User.UserID)
	if erro != nil {
		log.Println(erro)
	}

	answer := make([]*appstorage.Item, 0, len(tostor.List))

	// for every item decode and add file ids into appstorage item
	for _, itm := range tostor.List {

		// decode to local item struct
		newitem, err := appstorage.Decode(itm.Body, currentuserencryptokey)
		if err != nil {
			log.Println("error on decoding:", err)
			continue
		}

		// upload file ids into local item
		newitem.FileIDs = make([]string, 0, len(itm.FilesID))
		for _, fileid := range itm.FilesID {
			if len(fileid) == 0 {
				continue
			}
			newitem.FileIDs = append(newitem.FileIDs, fileid)
		}

		log.Println(newitem)

		// making answer slice of items
		answer = append(answer, newitem)

		log.Println(newitem)
	}

	// save local
	err = operator.PutItems(answer...)
	if err != nil {
		log.Println("can't save item local:", err)
	}

}
