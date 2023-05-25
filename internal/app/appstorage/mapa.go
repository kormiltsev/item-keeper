package appstorage

import (
	"fmt"
	"log"
	"strings"
	"sync"
)

var localstorageaddress = "./data/ClientLocalStorage"

var maxFileSize int64 = 5242880

type List struct {
	LastUpdate int64
	UserID     string
	mu         sync.Mutex
	Items      map[string]*Item
	Files      map[string]*File
	parameters map[string][]string // only for searche and display
}

var Catalog = List{
	LastUpdate: 0,
	UserID:     "AppUser",
	mu:         sync.Mutex{},
	Items:      map[string]*Item{},
	Files:      map[string]*File{},
	parameters: map[string][]string{},
}

type Operator struct {
	UserID          string
	Mapa            *List
	LastUpdate      int64
	Search          map[string][]string
	Answer          map[string]*Item
	AnswerAddresses map[string][]string
}

var ErrEmptyRequest = fmt.Errorf("empty request")

func NewUser(userid string, lastUpdate int64) {
	Catalog.mu.Lock()
	defer Catalog.mu.Unlock()

	// create new Catalog for new user
	// remove user's directory with all files and subdirectories
	// err := os.RemoveAll(userFolderPass(userid))
	// if err != nil {
	// 	log.Printf("can't delete directory for userid = %s, error:%v", userid, err)
	// }

	Catalog.LastUpdate = lastUpdate
	Catalog.UserID = userid
	Catalog.Items = map[string]*Item{}
	Catalog.Files = map[string]*File{} // need to delete files
	Catalog.parameters = map[string][]string{}

	//erase file storage
	deleteAllFilesAllUsers()

	// log.Println(Catalog)
}

func NewItem(userid string) *Item {
	return &Item{
		UserID:        userid,
		Parameters:    make([]Parameter, 0),
		FileIDs:       make([]string, 0),
		UploadAddress: make([]string, 0),
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
		Answer:          map[string]*Item{},
		AnswerAddresses: map[string][]string{},
	}, nil
}

func (op *Operator) PutItems(items ...*Item) error {
	op.Mapa.mu.Lock()
	defer op.Mapa.mu.Unlock()

	Catalog.LastUpdate = op.LastUpdate

	for _, item := range items {
		// add item to catalog
		op.Mapa.Items[item.ItemID] = item

		// register item in parameters map
		for _, par := range item.Parameters {
			op.Mapa.parameters[par.Name] = append(op.Mapa.parameters[par.Name], item.ItemID)
		}
	}
	return nil
}

func (op *Operator) FindItemByParameter() error {
	op.Mapa.mu.Lock()
	defer op.Mapa.mu.Unlock()

	// if empty request
	if len(op.Search) == 0 {
		return ErrEmptyRequest
	}

	// for every searchKey
	for key, searchstrings := range op.Search {

		// for every itemID in Catalog.parameters[searchKey]
		for i, itemid := range op.Mapa.parameters[key] {

			// is item already in answer list?
			if _, ok := op.Answer[itemid]; ok {
				continue
			}

			// if item deleted, so need to delete it from paramneters list
			itm, ok := op.Mapa.Items[itemid]
			if !ok {
				if i == len(op.Mapa.parameters[key])-1 {
					op.Mapa.parameters[key] = op.Mapa.parameters[key][:len(op.Mapa.parameters[key])-1]
				} else {
					list := op.Mapa.parameters[key]
					op.Mapa.parameters[key] = append(list[:i], list[i+1:]...)
				}
				continue
			}
			// search pKeysearch in item's Parameters
			for _, par := range itm.Parameters {

				// for every searchVal in search list
				for _, valsearche := range searchstrings {

					// find!
					if strings.Contains(par.Value, valsearche) {
						op.Answer[itemid] = op.Mapa.Items[itemid]
						op.AnswerAddresses[itemid] = op.addFilesAddresses(itemid)
					}
				}
			}
		}
	}
	return nil
}

func (op *Operator) addFilesAddresses(itemid string) []string {
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

func ReturnIDs() []string {
	answer := make([]string, len(Catalog.Items))
	i := 0
	for k := range Catalog.Items {
		answer[i] = k
	}
	return answer
}

func (op *Operator) DeleteItemByID(itemid string) error {
	if len(itemid) == 0 {
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
