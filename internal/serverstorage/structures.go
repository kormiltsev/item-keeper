package serverstorage

import "errors"

type Item struct {
	ItemID  string
	UserID  string
	Body    string
	FilesID []string
}

type File struct {
	FileID  string
	ItemID  string
	Address string
}

type User struct {
	Login         string
	Password      string
	UserID        string
	LastUpdate    int64
	OldLastUpdate int64
	ItemIDs       []string
}

// Operational errors
var (
	ErrLoginNotFound = errors.New(`login not found`)
	ErrPasswordWrong = errors.New(`wrong password`)
	ErrUserExists    = errors.New(`user exists`)
	ErrItemNotFound  = errors.New(`item not found`)
)
