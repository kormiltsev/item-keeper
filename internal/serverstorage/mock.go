package serverstorage

import (
	"crypto/sha1"
	"encoding/hex"
	"log"
	"strconv"
	"sync"
	"time"
)

type ToMock struct {
	Data *ToStorage
}

var mu = sync.Mutex{}
var Users = map[string]*User{}
var Items = map[string]*Item{}
var Files = map[string]*File{}

func (mock *ToMock) RegUser() error {
	mu.Lock()
	defer mu.Unlock()

	if _, ok := Users[mock.Data.User.Login]; ok {
		return ErrUserExists
	}

	mock.createUserID()

	// save in users catalog
	Users[mock.Data.User.UserID] = mock.Data.User

	log.Println("SERVER: ", Users)
	return nil
}

func (mock *ToMock) AuthUser() error {
	mu.Lock()
	defer mu.Unlock()

	user, ok := Users[mock.Data.User.Login]
	if !ok {
		return ErrLoginNotFound
	}

	if user.Password != mock.Data.User.Password {
		return ErrPasswordWrong
	}

	mock.Data.User.UserID = user.UserID
	return nil
}

func (mock *ToMock) createUserID() {
	// create uniq userID
	h := sha1.New()
	h.Write([]byte(mock.Data.User.Password + strconv.FormatInt(time.Now().UnixNano(), 16)))
	sha1_hash := hex.EncodeToString(h.Sum(nil))

	if _, ok := Users[sha1_hash]; ok {
		mock.createUserID()
	}

	mock.Data.User.UserID = sha1_hash
}
