package serverstorage

import "errors"

type Item struct {
	ItemID  int64
	UserID  string
	Body    string
	FilesID []int64
	Deleted bool
}

type File struct {
	FileID  int64
	UserID  string
	ItemID  int64
	Address string
	Body    []byte
	Deleted bool
}

type User struct {
	Login         string
	Password      string
	UserID        string
	LastUpdate    int64
	OldLastUpdate int64
	// FileIDs       []string // used for request, not for storage
}

// Operational errors
var (
	ErrLoginNotFound = errors.New(`login not found`)
	ErrPasswordWrong = errors.New(`wrong password`)
	ErrUserExists    = errors.New(`user exists`)
	ErrItemNotFound  = errors.New(`item not found`)
	ErrEmptyRequest  = errors.New(`empty request`)
)
