package serverstorage

type ToStorage struct {
	DB    Storager
	User  *User
	List  []Item
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
		List: make([]Item, 0),
	}
}
