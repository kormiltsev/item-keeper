package appstorage

import (
	"fmt"
	"log"
	"strings"
	"sync"
)

// internal application Data (snapshot)
var localcatalogaddress = "./data/Catalog.dat"

// maximum size of files
var maxFileSize int64 = 5242880

type List struct {
	LastUpdate int64
	UserID     string
	mu         sync.Mutex
	Items      map[int64]*Item    // key = itemid
	Files      map[int64]*File    // key = fileid
	parameters map[string][]int64 // only for search and display; key = param.name
}

var Catalog = List{
	LastUpdate: 0,
	UserID:     "AppUser",
	mu:         sync.Mutex{},
	Items:      map[int64]*Item{},
	Files:      map[int64]*File{},
	parameters: map[string][]int64{},
}

type Operator struct {
	UserID          string
	Mapa            *List
	LastUpdate      int64
	ID              int64
	QuickSearch     string
	Search          map[string][]string
	Answer          map[int64]*Item
	AnswerAddresses map[int64][]string
}

var ErrEmptyRequest = fmt.Errorf("empty request")
var ErrNotFound = fmt.Errorf("not found")

func newCatalog(userid string, lastUpdate int64) {
	Catalog.mu.Lock()
	defer Catalog.mu.Unlock()

	Catalog.LastUpdate = lastUpdate
	Catalog.UserID = userid
	Catalog.Items = map[int64]*Item{}
	Catalog.Files = map[int64]*File{}
	Catalog.parameters = map[string][]int64{}
}

func NewUser(userid string, lastUpdate int64) {
	newCatalog(userid, lastUpdate)

	//erase file storage
	deleteAllFilesAllUsers()
}

func NewItem(userid string) *Item {
	return &Item{
		UserID:         userid,
		Parameters:     make([]Parameter, 0),
		FileIDs:        make([]int64, 0),
		UploadAddress:  make([]string, 0),
		LocalAddresses: make([]string, 0),
	}
}

func ReturnOperator(userid string) (*Operator, error) {
	if Catalog.UserID != userid {
		return nil, fmt.Errorf("user not found")
	}
	return &Operator{
		UserID:          userid,
		Mapa:            &Catalog,
		Search:          map[string][]string{},
		Answer:          map[int64]*Item{},
		AnswerAddresses: map[int64][]string{},
	}, nil
}

func (op *Operator) PutItems(items ...*Item) error {
	op.Mapa.mu.Lock()
	defer op.Mapa.mu.Unlock()

	Catalog.LastUpdate = op.LastUpdate

	for _, item := range items {
		// add item to catalog
		// olditem, ok := op.Mapa.Items[item.ItemID]
		// if ok && olditem.FileIDs != nil {
		// 	item.FileIDs = olditem.FileIDs
		// }
		op.Mapa.Items[item.ItemID] = item

		// register item in parameters map
		for _, par := range item.Parameters {
			op.Mapa.parameters[par.Name] = append(op.Mapa.parameters[par.Name], item.ItemID)
		}
	}
	return nil
}

func (op *Operator) FindItemByID() error {
	var ok bool
	// if empty request
	if op.ID == 0 {
		return ErrEmptyRequest
	}

	op.Mapa.mu.Lock()
	defer op.Mapa.mu.Unlock()

	op.Answer[op.ID], ok = op.Mapa.Items[op.ID]
	if !ok {
		return ErrNotFound
	}

	// add file local addresses
	op.AnswerAddresses[op.ID] = op.addFilesAddresses(op.ID)

	return nil
}

func (op *Operator) FindItemQuick() error {
	// if empty request
	if op.QuickSearch == "" {
		return ErrEmptyRequest
	}

	op.Mapa.mu.Lock()
	defer op.Mapa.mu.Unlock()

	// for every item
	for id, itm := range op.Mapa.Items {
		for _, par := range itm.Parameters {

			if strings.Contains(par.Value, op.QuickSearch) {
				op.Answer[id] = op.Mapa.Items[id]
				op.AnswerAddresses[id] = op.addFilesAddresses(id)
			}
		}
	}
	return nil
}

func (op *Operator) FindItemByParameter() error {
	// if empty request
	if len(op.Search) == 0 {
		return ErrEmptyRequest
	}

	op.Mapa.mu.Lock()
	defer op.Mapa.mu.Unlock()

	// for every searchKey

	for id, itm := range op.Mapa.Items {
		for _, par := range itm.Parameters {

			for key, searchstrings := range op.Search {
				if key == par.Name {
					for _, searchword := range searchstrings {
						if strings.Contains(par.Value, searchword) {
							op.Answer[id] = op.Mapa.Items[id]
							op.AnswerAddresses[id] = op.addFilesAddresses(id)
						}
					}

				}
			}
		}
	}
	return nil
}

func (op *Operator) addFilesAddresses(itemid int64) []string {
	answer := make([]string, 0)
	item := op.Mapa.Items[itemid]
	for _, flid := range item.FileIDs {

		file, ok := op.Mapa.Files[flid]
		if !ok {
			log.Println("can't find file in Files local by id (unregistered?):", flid)
			continue
		}

		answer = append(answer, file.Address)
	}
	return answer
}

func (op *Operator) RegisterFilesToItems(files ...File) error {
	op.Mapa.mu.Lock()
	defer op.Mapa.mu.Unlock()

	if op.LastUpdate > Catalog.LastUpdate {
		Catalog.LastUpdate = op.LastUpdate
	}

	for _, fle := range files {
		// add file id to item in catalog
		itm, ok := op.Mapa.Items[fle.ItemID]
		if ok {
			for _, ids := range itm.FileIDs {
				if ids == fle.FileID {
					continue
				}
			}
			op.Mapa.Items[fle.ItemID].FileIDs = append(itm.FileIDs, fle.FileID)
		}
	}
	return nil
}

func (op *Operator) DeleteItemByID(itemid int64) error {
	if itemid <= 0 {
		return fmt.Errorf("empty request")
	}

	op.Mapa.mu.Lock()
	defer op.Mapa.mu.Unlock()

	item, ok := op.Mapa.Items[itemid]
	if !ok {
		return nil
	}

	// delete folder
	err := deleteFolderByItemID(op.Mapa.UserID, itemid)
	if err != nil {
		log.Println("can't delete local folder for item:", itemid)
	}

	// unregister files
	for _, fileid := range item.FileIDs {
		delete(op.Mapa.Files, fileid)
	}

	// unregister item
	delete(op.Mapa.Items, itemid)

	return nil
}

func (op *Operator) UploadFilesAddresses() error {
	op.Mapa.mu.Lock()
	defer op.Mapa.mu.Unlock()

	for k, item := range op.Mapa.Items {
		op.Mapa.Items[k].LocalAddresses = op.Mapa.Items[k].LocalAddresses[:0]
		for _, fileid := range item.FileIDs {
			fileRegistered, ok := op.Mapa.Files[fileid]
			if !ok {
				return fmt.Errorf("file not found")
			}
			op.Mapa.Items[k].LocalAddresses = append(op.Mapa.Items[k].LocalAddresses, fileRegistered.Address)
		}
	}
	return nil
}

func (op *Operator) ReturnFileIDByAddress(address string) int64 {
	op.Mapa.mu.Lock()
	defer op.Mapa.mu.Unlock()

	for k, fle := range op.Mapa.Files {
		if fle.Address == address {
			return k
		}
	}
	return 0
}

func (op *Operator) DeleteFileByID(fileid int64) {
	op.Mapa.mu.Lock()

	fle, ok := op.Mapa.Files[fileid]
	if !ok {
		return
	}

	op.Mapa.mu.Unlock()

	fle.DeleteFileLocal()
}
