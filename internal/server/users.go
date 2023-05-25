package server

import (
	"context"
	"errors"
	"log"

	pb "github.com/kormiltsev/item-keeper/internal/server/proto"
	serverstorage "github.com/kormiltsev/item-keeper/internal/serverstorage"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type ItemServer struct {
	pb.UnimplementedItemKeeperServer
}

// RegUser returns new user's id
func (itemserv *ItemServer) RegUser(ctx context.Context, in *pb.RegUserRequest) (*pb.RegUserResponse, error) {
	// create truct from storage
	tostorage := serverstorage.NewToStorage()

	// add user creds to struct
	user := serverstorage.User{
		Login:    in.Login,
		Password: in.Password,
	}
	tostorage.User = &user

	// select DB interface
	tostorage.DB = serverstorage.NewStorager(tostorage)
	if tostorage.Error != nil {
		return nil, tostorage.Error
	}

	// reg user on server
	err := tostorage.DB.RegUser()
	if err != nil {
		if errors.Is(err, serverstorage.ErrUserExists) {
			log.Println("Creating new user:", err)
			return nil, status.Errorf(codes.AlreadyExists, `User exists`)
			//ErrEmptyRequest

		}
		log.Println("Unknown error from storage: ", err)
		return nil, status.Errorf(codes.Internal, `unknown error`)
	}

	// create response
	var response = pb.RegUserResponse{
		Userid:     tostorage.User.UserID,
		Lastupdate: tostorage.User.LastUpdate,
	}

	return &response, nil
}

// AuthUser returns existed user's id
func (itemserv *ItemServer) AuthUser(ctx context.Context, in *pb.AuthUserRequest) (*pb.AuthUserResponse, error) {
	// create truct from storage
	tostorage := serverstorage.NewToStorage()

	// add user creds to struct
	user := serverstorage.User{
		Login:    in.Login,
		Password: in.Password,
	}
	tostorage.User = &user

	// select DB interface
	tostorage.DB = serverstorage.NewStorager(tostorage)
	if tostorage.Error != nil {
		return nil, tostorage.Error
	}

	// reg user on server
	err := tostorage.DB.AuthUser()
	if err != nil {
		if errors.Is(err, serverstorage.ErrPasswordWrong) {
			log.Println("authorization error:", err)
			return nil, status.Errorf(codes.InvalidArgument, `Wrong Password`)
		}
		if errors.Is(err, serverstorage.ErrLoginNotFound) {
			log.Println("authorization error:", err)
			return nil, status.Errorf(codes.InvalidArgument, `User not exists`)
		}
		log.Println("Unknown error from storage: ", err)
		return nil, status.Errorf(codes.Internal, `unknown error`)
	}

	// create response
	var response = pb.AuthUserResponse{
		Userid: tostorage.User.UserID,
	}

	return &response, nil
}
