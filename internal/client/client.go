package client

import (
	"context"
	"log"
	"time"

	pb "github.com/kormiltsev/item-keeper/internal/server/proto"

	storage "github.com/kormiltsev/item-keeper/internal/storage"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

type ClientSettings struct {
	port      string
	tokenName string
	token     string
}

var ClientSet = ClientSettings{
	port:      ":9999",
	tokenName: "cli_user_token",
	token:     "TokenNotSet",
}

func RunTestClient(port string) {
	ClientSet.port = port
	conn, err := grpc.Dial(ClientSet.port, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	c := pb.NewItemKeeperClient(conn)

	ctxto, cancel := context.WithTimeout(context.Background(), 5000*time.Millisecond)
	defer cancel()

	md := metadata.New(map[string]string{ClientSet.tokenName: ClientSet.token})
	ctx := metadata.NewOutgoingContext(ctxto, md)

	// tc1
	PushUser := storage.User{
		Login: "ErrUserExists",
		Pass:  "password",
	}
	RegUser(ctx, PushUser, c)

	// tc2
	ClientSet.token = "TokenNotSet"
	md = metadata.New(map[string]string{ClientSet.tokenName: ClientSet.token})
	ctx = metadata.NewOutgoingContext(ctxto, md)
	PushUser = storage.User{
		Login: "NewLogin",
		Pass:  "password",
	}
	RegUser(ctx, PushUser, c)
}

// RegUser set up userid in client token (metadata). Error if user exists.
func RegUser(ctx context.Context, user storage.User, c pb.ItemKeeperClient) {
	newuserid, err := c.RegUser(ctx, &pb.RegUserRequest{
		Login: user.Login,
		Pass:  user.Pass,
	})
	if err != nil {
		if e, ok := status.FromError(err); ok {
			if e.Code() == codes.AlreadyExists {
				log.Println(`user already exists:`, e.Message())
			} else {
				log.Println(e.Code(), e.Message())
			}
		} else {
			log.Printf("Reg user error from server: %v\n", err)
		}
	} else {
		log.Println("user id:", newuserid.Userid)
		ClientSet.token = newuserid.Userid
		log.Println("client settings:", ClientSet)
	}
}
