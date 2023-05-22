package server

import (
	"context"
	"fmt"
	"log"
	"net"

	configs "github.com/kormiltsev/item-keeper/internal/configs"
	pb "github.com/kormiltsev/item-keeper/internal/server/proto"
	storage "github.com/kormiltsev/item-keeper/internal/storage"
	"google.golang.org/grpc"
)

func StartServer(ctx context.Context, close chan struct{}) {

	con := configs.UploadConfigs()
	log.Println("configs uploaded:", con)

	uitem := storage.NewItem()
	uitem.DB = storage.NewToStorage(uitem)
	err := uitem.DB.Connect(ctx)
	if err != nil {
		log.Fatal(err)
	}
	defer uitem.DB.Disconnect()

	// short test
	mockDB(ctx, uitem)
	// ==========

	// test file storage available
	err = storage.FileStoragePing(configs.ServiceConfig.FileServerAddress)
	if err != nil {
		log.Println("file storage fail:", err)
	}
	// ===========================
	<-close
}

func mockDB(ctx context.Context, db *storage.Uitem) {
	uitem := storage.NewItem()
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
	log.Println("error database test =", uitem.User.Error)
}

// StartServerGRPC run grpc server
func StartServerGRPC(port string) {
	listen, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatal(err)
	}

	s := grpc.NewServer(
		grpc.UnaryInterceptor(unaryInterceptor),
	)

	pb.RegisterItemKeeperServer(s, &ItemServer{})

	fmt.Println("gRPC started")

	if err := s.Serve(listen); err != nil {
		log.Println("gRPC server crushed:", err)
	}
}
