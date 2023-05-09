package app

import (
	"context"
	"log"

	serverstorage "github.com/kormiltsev/item-keeper/internal/serverstorage"

	appstorage "github.com/kormiltsev/item-keeper/internal/app/appstorage"
)

// const appuser = "AppUser"

func RegUser(ctx context.Context, login, password string) error {
	// create truct from storage
	tostorage := serverstorage.NewToStorage()

	// Shifu login and password here
	//
	// =============================

	// add user creds to struct
	user := serverstorage.User{
		Login:    login,
		Password: password,
	}
	tostorage.User = &user

	// select DB interface
	tostorage.DB = serverstorage.NewStorager(tostorage)
	if tostorage.Error != nil {
		return tostorage.Error
	}

	// reg user on server
	err := tostorage.DB.RegUser()
	if err != nil {
		log.Println("reg user error from server:", err)
	}

	// save local
	appstorage.NewUser(tostorage.User.UserID)
	return nil
}

func AddNewItem() {
	// operator := appstorage.NewOperator(appuser)
	// operator.
	// 	appstorage.CatalogByUserID()
}
