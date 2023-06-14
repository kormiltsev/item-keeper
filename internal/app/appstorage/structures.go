package appstorage

type Parameter struct {
	Name  string
	Value string
}

type Item struct {
	UserID         string
	ItemID         int64
	Parameters     []Parameter
	FileIDs        []int64
	UploadAddress  []string
	LocalAddresses []string // not for storage
}

type File struct {
	FileID  int64
	UserID  string
	ItemID  int64
	Address string // for storage
	Body    []byte // not for storage
	Hash    string // not for storage
}
