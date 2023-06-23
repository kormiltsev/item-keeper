package serverstorage

import configs "github.com/kormiltsev/item-keeper/internal/configs"

// ToStorage is an interface with pointer on DB.
type ToStorage struct {
	DB          Storager
	User        *User
	List        []Item
	File        File
	FilesNoBody []File // response on Update request, files has no bodyes
	Error       error
}

// NewStorager build DB pointer depends on settings.
func NewStorager(tostor *ToStorage) Storager {

	if configs.ServiceConfig.DBlink == "mock" || configs.ServiceConfig.DBlink == "" {
		stora := ToMock{}
		tostor.DB = &stora
		stora.Data = tostor
		return &stora
	}

	postgra := ToPostgres{}
	tostor.DB = &postgra
	postgra.Data = tostor
	return &postgra
}

// NewToStorage returns new interface.
func NewToStorage() *ToStorage {
	return &ToStorage{
		User: &User{},
		List: make([]Item, 0),
	}
}

// NewItem returns empty Item.
func NewItem() *Item {
	return &Item{FilesID: make([]int64, 0)}
}

// NewItem returns empty File.
func NewFile() *File {
	return &File{Body: make([]byte, 0)}
}
