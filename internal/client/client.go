package client

import (
	"context"
	"log"
	"sync"

	configs "github.com/kormiltsev/item-keeper/internal/configsClient"
	pb "github.com/kormiltsev/item-keeper/internal/server/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
)

type ClientSettings struct {
	port        string
	tokenName   string
	clientToken string
	Connection  *grpc.ClientConn
}

var ClientSet = ClientSettings{
	port:        ":3333",
	tokenName:   "CLIENT_TOKEN",
	clientToken: "clientToken",
}

func UploadConfigsCli() string {
	con, welcomLetter := configs.UploadConfigsClient()
	log.Println("configs uploaded:", con)
	ClientSet.port = con.GRPCaddress
	ClientSet.clientToken = con.ClientToken
	return welcomLetter
}

// NewClientConnector make connection once and then returns interface with client
func NewClientConnector(ctx context.Context) *ClientConnector {
	var err error
	var cli pb.ItemKeeperClient
	var once sync.Once
	once.Do(func() {
		ClientSet.Connection, err = grpc.Dial(ClientSet.port, grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err != nil {
			log.Fatal(err)
		}
		cli = pb.NewItemKeeperClient(ClientSet.Connection)
	})

	// set token
	md := metadata.New(map[string]string{ClientSet.tokenName: ClientSet.clientToken})
	ctx = metadata.NewOutgoingContext(ctx, md)

	return &ClientConnector{Ctx: ctx, Client: &cli}
}

// CloseConnection before stop application
func CloseConnection() {
	ClientSet.Connection.Close()
}
