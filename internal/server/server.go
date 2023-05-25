package server

import (
	"context"
	"log"
	"net"

	configs "github.com/kormiltsev/item-keeper/internal/configs"
	pb "github.com/kormiltsev/item-keeper/internal/server/proto"
	serverstorage "github.com/kormiltsev/item-keeper/internal/serverstorage"
	"google.golang.org/grpc"
)

// StartServer connect to DB
func StartServer(ctx context.Context, close chan struct{}) {

	con := configs.UploadConfigs()
	log.Println("configs uploaded:", con)

	// prepare to server
	tostor := serverstorage.NewToStorage()
	tostor.DB = serverstorage.NewStorager(tostor)

	err := tostor.DB.Connect(ctx)
	if err != nil {
		log.Fatal(err)
	}
	defer tostor.DB.Disconnect()

	<-close
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

	log.Println("gRPC started")

	if err := s.Serve(listen); err != nil {
		log.Println("gRPC server crushed:", err)
	}
}
