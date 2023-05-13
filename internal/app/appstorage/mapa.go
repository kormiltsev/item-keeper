package appstorage

import (
	"fmt"
	"log"
	"os"
	"strings"
	"sync"
)

var localstorageaddress = "./data/ClientLocalStorage"

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
	UserID string
	Mapa   *List
	Search map[string][]string
	Answer map[string]*Item
}

var ErrEmptyRequest = fmt.Errorf("empty request")

func NewUser(userid string, lastUpdate int64) {
	Catalog.mu.Lock()
	defer Catalog.mu.Unlock()

	// create new Catalog for new user
	// remove user's directory with all files and subdirectories
	err := os.RemoveAll(userFolderPass(userid))
	if err != nil {
		log.Printf("can't delete directory for userid = %s, error:%v", userid, err)
	}

	Catalog.LastUpdate = lastUpdate
	Catalog.UserID = userid
	Catalog.Items = map[string]*Item{}
	Catalog.Files = map[string]*File{} // need to delete files
	Catalog.parameters = map[string][]string{}

	log.Println(Catalog)
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
		UserID: userid,
		Mapa:   &Catalog,
		Search: map[string][]string{},
		Answer: map[string]*Item{},
	}, nil
}

func (op *Operator) PutItems(items ...*Item) error {
	op.Mapa.mu.Lock()
	defer op.Mapa.mu.Unlock()

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
		for _, itemid := range op.Mapa.parameters[key] {

			// is item already in answer list?
			if _, ok := op.Answer[itemid]; ok {
				continue
			}

			// search pKeysearch in item's Parameters
			for _, par := range op.Mapa.Items[itemid].Parameters {

				// for every searchVal in search list
				for _, valsearche := range searchstrings {

					// find!
					if strings.Contains(par.Value, valsearche) {
						op.Answer[itemid] = op.Mapa.Items[itemid]
					}
				}
			}
		}
	}
	return nil
}

func ReturnIDs() []string {
	answer := make([]string, len(Catalog.Items))
	i := 0
	for k := range Catalog.Items {
		answer[i] = k
	}
	return answer
}
