package appstorage

type Parameter struct {
	Name  string
	Value string
}

type Item struct {
	UserID        string
	ItemID        int64
	Parameters    []Parameter
	FileIDs       []string
	UploadAddress []string
}

type File struct {
	FileID  string
	UserID  string
	ItemID  int64
	Address string // for storage
	Body    []byte // not for store
	Hash    string // not for store
}
