package server

import (
	"context"
	"errors"
	"log"

	pb "github.com/kormiltsev/item-keeper/internal/server/proto"
	storage "github.com/kormiltsev/item-keeper/internal/storage"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

type ItemServer struct {
	pb.UnimplementedItemKeeperServer
}

// RegUser returns new user's id
func (itemserv *ItemServer) RegUser(ctx context.Context, in *pb.RegUserRequest) (*pb.RegUserResponse, error) {
	// create new empty uitem (contains DB, user, list of items and error)
	uitem := storage.NewItem()
	uitem.User.Login = in.Login
	uitem.User.Pass = in.Pass
	// return DB structure
	uitem.DB = storage.NewToStorage(uitem)
	// run DB method
	uitem.DB.CreateUser(ctx)
	if uitem.Err != nil {
		if errors.Is(uitem.Err, storage.ErrUserExists) {
			log.Println("Creating new user:", uitem.Err)
			return nil, status.Errorf(codes.AlreadyExists, `User exists`)
		}
		log.Println("Unknown error from storage: ", uitem.Err)
		return nil, status.Errorf(codes.Internal, `unknown error`)
	}

	// create response
	var response = pb.RegUserResponse{
		Userid: uitem.User.UserID,
	}

	log.Println(response)

	return &response, nil
}

func (itemserv *ItemServer) LogUser(ctx context.Context, in *pb.LogUserRequest) (*pb.LogUserResponse, error) {
	return &pb.LogUserResponse{}, nil
}

func (itemserv *ItemServer) AddItem(ctx context.Context, in *pb.AddItemRequest) (*pb.AddItemResponse, error) {
	return &pb.AddItemResponse{}, nil
}

func (itemserv *ItemServer) UpdateItem(ctx context.Context, in *pb.UpdateItemRequest) (*pb.UpdateItemResponse, error) {
	return &pb.UpdateItemResponse{}, nil
}

func (itemserv *ItemServer) GetCatalog(ctx context.Context, in *pb.GetCatalogRequest) (*pb.GetCatalogResponse, error) {
	return &pb.GetCatalogResponse{}, nil
}

func (itemserv *ItemServer) Pictures(in *pb.PicturesRequest, srv pb.ItemKeeper_PicturesServer) error {
	return nil
}

// unaryInterceptor searche for userid in token
func unaryInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	var err error

	// check metadata exists
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, status.Errorf(codes.Unauthenticated, "token not found")
	}

	// check token exists
	values := md.Get("cli_user_token")
	if len(values) == 0 {
		return nil, status.Errorf(codes.Unauthenticated, "token empty")
	}

	// check every of values
	for _, token := range values {
		if len(token) == 0 {
			err = status.Errorf(codes.Unauthenticated, "token empty")
		}
	}
	// all tokens are empty
	if err != nil {
		return nil, err
	}

	// OK
	return handler(ctx, req)
}
