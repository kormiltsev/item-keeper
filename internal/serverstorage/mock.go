package serverstorage

import (
	"crypto/sha1"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"sync"
	"time"
)

type ToMock struct {
	Data *ToStorage
}

type onechange struct {
	userid  string
	updated int64
	itemid  string
}

// database realosation
var mu = sync.Mutex{}
var Users = map[string]*User{} // key= login
var Items = map[string]*Item{} // key = ItemID
var Files = map[string]*File{} // key = FileID

// log of changes
var listOfChanges = []onechange{}

var storageaddress = "./data/serverstorage"

var errFileIDExists = errors.New("file id exists")

func (mock *ToMock) RegUser() error {

	if mock.Data.User.Login == "" || mock.Data.User.Password == "" {
		return ErrEmptyRequest
	}

	mu.Lock()
	defer mu.Unlock()

	if _, ok := Users[mock.Data.User.Login]; ok {
		return ErrUserExists
	}

	mock.createUserID()

	mock.Data.User.LastUpdate = time.Now().UnixMilli()

	// save in users catalog
	Users[mock.Data.User.Login] = mock.Data.User

	return nil
}

func (mock *ToMock) AuthUser() error {

	mu.Lock()
	defer mu.Unlock()

	user, ok := Users[mock.Data.User.Login]
	if !ok {
		log.Println("USERNOTFOUND:", Users, "requested:", mock.Data.User.Login)
		return ErrLoginNotFound
	}

	if user.Password != mock.Data.User.Password {
		log.Println("wrong password")
		return ErrPasswordWrong
	}

	log.Println("AUTH register: ", user.UserID)
	mock.Data.User.UserID = user.UserID
	mock.Data.User.LastUpdate = user.LastUpdate
	return nil
}

func (mock *ToMock) createUserID() {
	// create uniq userID
	h := sha1.New()
	h.Write([]byte(mock.Data.User.Login + strconv.FormatInt(time.Now().UnixNano(), 16)))
	newID := hex.EncodeToString(h.Sum(nil))

	// check for doubles
	for _, user := range Users {
		if user.UserID == newID && user.Login != mock.Data.User.Login {
			// try again
			mock.createUserID()
			return
		}
	}
	// write to user new user ID
	mock.Data.User.UserID = newID
}

func (mock *ToMock) PutItems() error {
	// error if empty
	if len(mock.Data.List) == 0 {
		return fmt.Errorf("empty request PutItem")
	}
	// if items more than 1
	for i, item := range mock.Data.List {
		// empty ItemID means add new item
		if item.ItemID == "" {
			sum := sha256.Sum256([]byte(item.UserID + strconv.FormatInt(time.Now().UnixNano(), 16)))
			mock.Data.List[i].ItemID = hex.EncodeToString(sum[:])
		}
	}

	//add to DB
	mu.Lock()
	defer mu.Unlock()

	for _, item := range mock.Data.List {
		Items[item.ItemID] = &item
		logToListOfChanges(item.UserID, item.ItemID) //mock.Data.User.UserID, item.ItemID)
	}

	return nil
}

func logToListOfChanges(userid, itemid string) {

	newlog := onechange{
		userid:  userid,
		updated: time.Now().UnixMilli(),
		itemid:  itemid,
	}
	listOfChanges = append(listOfChanges, newlog)

	log.Println("listOfChanges", listOfChanges)
}
func (mock *ToMock) UploadFile() error {

	// check body
	if len(mock.Data.File.Body) == 0 {
		return fmt.Errorf("empty file.Body in request uploadFile: %s", mock.Data.File.FileID)
	}

	// generates file ID
	sum := sha256.Sum256([]byte(mock.Data.File.FileID + strconv.FormatInt(time.Now().UnixNano(), 16)))
	mock.Data.File.FileID = hex.EncodeToString(sum[:])

	// create path localstorage/userid/itemid
	path := filepath.Join(storageaddress, mock.Data.File.UserID)
	path = filepath.Join(path, mock.Data.File.ItemID)

	// create folder if not exists
	err := os.MkdirAll(path, os.ModePerm)
	if err != nil {
		return fmt.Errorf("can't create local directory %s, error:%v", path, err)
	}

	// write file
	path = filepath.Join(path, mock.Data.File.FileID)
	err = os.WriteFile(path, mock.Data.File.Body, 0644)
	if err != nil {
		return fmt.Errorf("write file %s error:%v", path, err)
	}

	// add address
	mock.Data.File.Address = path

	// register file to Files and Items
	err = addFileAddressToItems(&mock.Data.File)
	if errors.Is(err, errFileIDExists) {
		return mock.UploadFile() // recreate id if fileID is already exists
	}

	return nil
}

// add goroutines here
func addFileAddressToItems(file *File) error {
	mu.Lock()
	defer mu.Unlock()

	// register file in Files, if fileID exists return error
	if _, ok := Files[file.FileID]; ok {
		log.Println("fileID is doubled...")
		return errFileIDExists
	}
	Files[file.FileID] = file

	// register files id to item
	item, ok := Items[file.ItemID]
	if !ok {
		return fmt.Errorf("wrong ItemID, not found item %s", file.ItemID)
	}
	item.FilesID = append(item.FilesID, file.FileID)
	Items[file.ItemID] = item
	return nil
}

func (mock *ToMock) UpdateByLastUpdate() error {
	log.Println("lastupdate requested:", mock.Data.User.LastUpdate)
	mock.Data.List = mock.Data.List[:0]

	// check last update date
	listOfItems := returnItemIDChanged(mock.Data.User.UserID, mock.Data.User.LastUpdate)
	if len(listOfItems) == 0 {
		return nil
	}

	mock.Data.List = returnItemsByIDs(listOfItems...)
	return nil
}

func returnItemIDChanged(userid string, lastupdate int64) []string {
	mu.Lock()
	defer mu.Unlock()

	answer := make([]string, 0)
	for i := len(listOfChanges) - 1; i >= 0; i-- {
		if listOfChanges[i].userid == userid && listOfChanges[i].updated > lastupdate {
			answer = append(answer, listOfChanges[i].itemid)
		}
	}
	return answer
}

func returnItemsByIDs(itemsids ...string) []Item {
	mu.Lock()
	defer mu.Unlock()

	answer := make([]Item, 0)
	for _, id := range itemsids {
		itm, ok := Items[id]
		if !ok {
			log.Println("requested item ID not found, id =", id)
			continue
		}

		// copy (not pointer)
		item := Item{
			ItemID:  itm.ItemID,
			UserID:  itm.UserID,
			Body:    itm.Body,
			FilesID: make([]string, len(itm.FilesID)),
		}

		copy(item.FilesID, itm.FilesID)
		// for i, fid := range itm.FilesID {
		// 	item.FilesID[i] = fid
		// }

		answer = append(answer, item)
	}
	return answer
}
