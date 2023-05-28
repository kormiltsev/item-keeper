package serverstorage

import (
	"bufio"
	"context"
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
	"sync/atomic"
	"time"
)

type ToMock struct {
	Data *ToStorage
}

type onechange struct {
	userid  string
	updated int64
	itemid  int64
}

// database realosation
var (
	mu    = sync.Mutex{}
	Users = map[string]*User{} // key= login
	Items = map[int64]*Item{}  // key = ItemID
	Files = map[string]*File{} // key = FileID
)

// log of changes
var listOfChanges = []onechange{}

// for itemID uniq number
var itemIDcounter int64

var errFileIDExists = errors.New("file id exists")

// RegUser register new user and returns UserID
func (mock *ToMock) RegUser(ctx context.Context) error {

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

// AuthUser returns UserID, or error
func (mock *ToMock) AuthUser(ctx context.Context) error {

	mu.Lock()
	defer mu.Unlock()

	user, ok := Users[mock.Data.User.Login]
	if !ok {
		log.Println("user not found:", mock.Data.User.Login)
		return ErrLoginNotFound
	}

	if user.Password != mock.Data.User.Password {
		log.Println("wrong password")
		return ErrPasswordWrong
	}

	mock.Data.User.UserID = user.UserID
	mock.Data.User.LastUpdate = user.LastUpdate
	return nil
}

// createUserID for internal use only
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

// PutItems add or update item. just replaced by new one
func (mock *ToMock) PutItems(ctx context.Context) error {
	// error if empty
	if len(mock.Data.List) == 0 {
		return ErrEmptyRequest
	}

	// if items more than 1
	for i := range mock.Data.List {
		// empty ItemID means add new item
		if mock.Data.List[i].ItemID == 0 {
			mock.Data.List[i].ItemID = atomic.AddInt64(&itemIDcounter, 1)

			// old: string
			// sum := sha256.Sum256([]byte(item.UserID + strconv.FormatInt(time.Now().UnixNano(), 16)))
			// mock.Data.List[i].ItemID = hex.EncodeToString(sum[:])
		}
	}

	//add to DB
	mu.Lock()
	defer mu.Unlock()

	for _, item := range mock.Data.List {

		// keep old files ids if exists
		if _, ok := Items[item.ItemID]; ok {
			olditemFiles := Items[item.ItemID].FilesID
			item.FilesID = make([]string, len(olditemFiles))
			copy(item.FilesID, olditemFiles)
		}

		// save to catalog
		Items[item.ItemID] = &item
		logToListOfChanges(item.UserID, item.ItemID)
	}

	return nil
}

// internal use logger
func logToListOfChanges(userid string, itemid int64) {

	newlog := onechange{
		userid:  userid,
		updated: time.Now().UnixMilli(),
		itemid:  itemid,
	}
	listOfChanges = append(listOfChanges, newlog)
}

// UploadFile accepts file, save it in storage and
func (mock *ToMock) UploadFile(ctx context.Context) error {

	// check body
	if len(mock.Data.File.Body) == 0 {
		return fmt.Errorf("empty file.Body in request uploadFile: %s", mock.Data.File.FileID)
	}

	// generates file ID
	sum := sha256.Sum256([]byte(mock.Data.File.FileID + strconv.FormatInt(time.Now().UnixNano(), 16)))
	mock.Data.File.FileID = hex.EncodeToString(sum[:])

	// check for doubles
	err := checkForDoublesFileID(mock.Data.File.FileID)
	if errors.Is(err, errFileIDExists) {
		return mock.UploadFile(ctx) // recreate id if fileID is already exists
	}

	// create path localstorage/userid/itemid
	path := filepath.Join(storageaddress, mock.Data.File.UserID)
	path = filepath.Join(path, strconv.FormatInt(mock.Data.File.ItemID, 10))

	// create folder if not exists
	err = os.MkdirAll(path, os.ModePerm)
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

	// logg changes
	logToListOfChanges(mock.Data.File.UserID, mock.Data.File.ItemID)
	return err
}

func checkForDoublesFileID(fileid string) error {
	mu.Lock()
	defer mu.Unlock()

	if _, ok := Files[fileid]; ok {
		log.Println("fileID is doubled...")
		return errFileIDExists
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
		return fmt.Errorf("wrong ItemID, not found item %d", file.ItemID)
	}
	item.FilesID = append(item.FilesID, file.FileID)
	Items[file.ItemID] = item
	return nil
}

// UpdateByLastUpdate accepts lastUpdate date from client and returns list if items that edited after lastUpdate
func (mock *ToMock) UpdateByLastUpdate(ctx context.Context) error {
	log.Println("lastupdate requested:", mock.Data.User.LastUpdate)
	mock.Data.List = mock.Data.List[:0]

	// check last update date
	listOfItems, lu := returnItemIDChanged(mock.Data.User.UserID, mock.Data.User.LastUpdate)
	if len(listOfItems) == 0 {
		return nil
	}
	mock.Data.User.LastUpdate = lu
	mock.Data.List = returnItemsByIDs(listOfItems...)
	return nil
}

func returnItemIDChanged(userid string, lastupdate int64) ([]int64, int64) {
	mu.Lock()
	defer mu.Unlock()

	var luanswer int64
	answer := make([]int64, 0)
	for i := len(listOfChanges) - 1; i >= 0; i-- {
		if listOfChanges[i].userid == userid && listOfChanges[i].updated > lastupdate {
			answer = append(answer, listOfChanges[i].itemid)
			if listOfChanges[i].updated > luanswer {
				luanswer = listOfChanges[i].updated
			}
		}
	}
	return answer, luanswer
}

func returnItemsByIDs(itemsids ...int64) []Item {
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
			Deleted: itm.Deleted,
		}

		copy(item.FilesID, itm.FilesID)
		// for i, fid := range itm.FilesID {
		// 	item.FilesID[i] = fid
		// }

		answer = append(answer, item)
	}
	return answer
}

func (mock *ToMock) GetFileByFileID(ctx context.Context) error {
	var err error

	mu.Lock()
	defer mu.Unlock()

	fle, ok := Files[mock.Data.File.FileID]
	if !ok {
		return nil // if item deleted, itemID exists, but FileID not exists
	}

	//read file
	mock.Data.File.Body, err = readFile(fle.Address)
	if err != nil {
		return fmt.Errorf("read file %s error:%v", fle.Address, err)
	}

	// add itemId, coze itemId is not mandatory in request
	mock.Data.File.ItemID = fle.ItemID

	return nil
}

func readFile(fileaddress string) ([]byte, error) {
	file, err := os.Open(fileaddress)

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

func (mock *ToMock) DeleteItems(ctx context.Context) error {

	mu.Lock()
	defer mu.Unlock()

	// delete items
	if len(mock.Data.List) == 0 {
		return ErrEmptyRequest
	}

	for i, itemtodelete := range mock.Data.List {
		if itemtodelete.ItemID == 0 {
			continue
		}

		olditem, ok := Items[itemtodelete.ItemID]
		if ok {

			// delete item folder
			err := deleteFilesByItemID(olditem.UserID, itemtodelete.ItemID)
			if err != nil {
				log.Printf("files deletion for itemid =%d, error:%v\n", itemtodelete.ItemID, err)
			}

			// file deregistration
			for _, fileid := range olditem.FilesID {
				delete(Files, fileid)
			}

			// mark as deleted
			olditem.Deleted = true
			Items[itemtodelete.ItemID] = olditem
		}
		// log changes
		logToListOfChanges(mock.Data.File.UserID, mock.Data.File.ItemID)
		mock.Data.List[i] = Item{}
	}

	return nil
}

// func deregisterAndDeleteFileByID(filesidtodelete string) error {
// 	// deregister
// 	file, ok := Files[filesidtodelete]
// 	if ok {
// 		delete(Files, filesidtodelete)
// 	} else {
// 		return fmt.Errorf("file not register in Files")
// 	}

// 	// remove from storage
// 	return deleteFilesByItemID(file.UserID, file.ItemID)
// }

func (mock *ToMock) DeleteFile(ctx context.Context) error {
	mu.Lock()
	defer mu.Unlock()

	file, ok := Files[mock.Data.File.FileID]
	if !ok {
		return ErrItemNotFound
	}
	return deleteFileByFileID(file.ItemID, file.UserID, file.FileID)
}

func (mock *ToMock) Connect(ctx context.Context) error {
	return nil
}

func (mock *ToMock) Disconnect() {
}
