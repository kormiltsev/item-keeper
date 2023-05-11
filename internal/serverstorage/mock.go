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

var mu = sync.Mutex{}
var Users = map[string]*User{}
var Items = map[string]*Item{}
var Files = map[string]*File{}

var storageaddress = "./data/serverstorage"

var errFileIDExists = errors.New("file id exists")

func (mock *ToMock) RegUser() error {
	mu.Lock()
	defer mu.Unlock()

	if _, ok := Users[mock.Data.User.Login]; ok {
		return ErrUserExists
	}

	mock.createUserID()

	// save in users catalog
	Users[mock.Data.User.UserID] = mock.Data.User

	// log.Println("SERVER: ", Users)
	return nil
}

func (mock *ToMock) AuthUser() error {
	mu.Lock()
	defer mu.Unlock()

	user, ok := Users[mock.Data.User.Login]
	if !ok {
		return ErrLoginNotFound
	}

	if user.Password != mock.Data.User.Password {
		return ErrPasswordWrong
	}

	mock.Data.User.UserID = user.UserID
	return nil
}

func (mock *ToMock) createUserID() {
	// create uniq userID
	h := sha1.New()
	h.Write([]byte(mock.Data.User.Password + strconv.FormatInt(time.Now().UnixNano(), 16)))
	sha1_hash := hex.EncodeToString(h.Sum(nil))

	if _, ok := Users[sha1_hash]; ok {
		mock.createUserID()
	}

	mock.Data.User.UserID = sha1_hash
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
	}

	log.Println("SERVER: got item:", mock.Data.List[0])
	return nil
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
