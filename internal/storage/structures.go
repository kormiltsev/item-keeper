package storage

import "errors"

// to storage
type User struct {
	Login       string
	Pass        string
	UserID      string
	LastUpdate  string
	DateCreated string
	Error       error
}

// to storage
type Item struct {
	ID          int
	UserID      string
	Name        string
	Tags        []string
	Parameters  map[string]string
	PictureLink []string
	Deleted     bool
}

type Uitem struct {
	DB   Storager
	User User
	List []Item
	Err  error
}

// Operational errors
var (
	ErrLoginNotFound = errors.New(`login not found`)
	ErrPasswordWrong = errors.New(`wrong password`)
	ErrUserExists    = errors.New(`user exists`)
)
