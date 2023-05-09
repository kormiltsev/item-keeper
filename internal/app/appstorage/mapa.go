package appstorage

import (
	"log"
	"sync"
)

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
	UserID  string
	Mapa    *List
	NewItem *Item
}

func NewUser(userid string) {
	Catalog.mu.Lock()
	defer Catalog.mu.Unlock()

	// create new Catalog for new user

	// delete old user's files?
	// or make array of user's Catalogs here
	//
	// =====================================

	Catalog.UserID = userid
	Catalog.Items = map[string]*Item{}
	Catalog.Files = map[string]*File{} // need to delete files
	Catalog.parameters = map[string][]string{}

	// NewOperator(userid)

	log.Println(Catalog)
}

func NewOperator(userid string) *Operator {
	return &Operator{
		UserID:  userid,
		Mapa:    &Catalog,
		NewItem: NewItem(userid),
	}
}

func NewItem(userid string) *Item {
	return &Item{
		UserID:     userid,
		Parameters: make([]Parameter, 0),
		FileIDs:    make([]string, 0),
	}
}

func (ct *List) PutItems(items ...Item) error {
	ct.mu.Lock()
	defer ct.mu.Unlock()

	for _, item := range items {
		// add item to catalog
		Catalog.Items[item.ItemID] = &item

		// register item in parameters map
		for _, par := range item.Parameters {
			ct.parameters[par.Name] = append(ct.parameters[par.Name], item.ItemID)
		}
	}
	return nil
}
