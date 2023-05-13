package serverstorage

type ToStorage struct {
	DB    Storager
	User  *User
	List  []Item
	File  File
	Error error
}

func NewStorager(tostor *ToStorage) Storager {

	// if configs.ServiceConfig.DBlink == "mock" || configs.ServiceConfig.DBlink == "" {
	stora := ToMock{}
	tostor.DB = &stora
	stora.Data = tostor
	return &stora
	// }

	// postgra := ToPostgres{
	// 	db: db,
	// }
	// tostor.DB = &postgra
	// postgra.Data = tostor
	// return &postgra
}

func NewToStorage() *ToStorage {
	return &ToStorage{
		User: &User{},
		List: make([]Item, 0),
	}
}

func NewItem() *Item {
	return &Item{FilesID: make([]string, 0)}
}

func NewFile() *File {
	return &File{Body: make([]byte, 0)}
}
