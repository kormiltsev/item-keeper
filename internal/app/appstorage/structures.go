package appstorage

type Parameter struct {
	Name  string
	Value string
}

type Item struct {
	UserID     string
	ItemID     string
	Parameters []Parameter
	FileIDs    []string
}

type File struct {
	ItemID  string
	Address string // for storage
	Body    []byte // not for store
}
