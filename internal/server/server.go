package server

import (
	"context"
	"log"

	configs "github.com/kormiltsev/item-keeper/internal/configs"
	storage "github.com/kormiltsev/item-keeper/internal/storage"
)

func StartServer(ctx context.Context) {
	con := configs.UploadConfigs()
	log.Println("configs uploaded:", con)

	uitem := storage.NewItem()
	uitem.DB = storage.NewToStorage(uitem)
	err := uitem.DB.Connect(ctx)
	if err != nil {
		log.Fatal(err)
	}
	defer uitem.DB.Disconnect()

	uitem.User.Login = "correct"
	uitem.User.Pass = "wrong"
	uitem.DB = storage.NewToStorage(uitem)
	uitem.DB.LoginUser(ctx)
	log.Println(uitem.User.Error)

	uitem.User.Login = "wrong"
	uitem.User.Pass = "any"
	uitem.DB = storage.NewToStorage(uitem)
	uitem.DB.LoginUser(ctx)
	log.Println(uitem.User.Error)

	uitem.User.Login = "correct"
	uitem.User.Pass = "correct"
	uitem.DB = storage.NewToStorage(uitem)
	uitem.DB.LoginUser(ctx)
	log.Println("error =", uitem.User.Error)
}

// StartServerGRPC run grpc server
func StartServerGRPC(port string) {
}
