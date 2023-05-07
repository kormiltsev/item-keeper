package server

import (
	"context"
	"crypto/sha1"
	"encoding/hex"
	"errors"
	"fmt"
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

	return &response, nil
}

func (itemserv *ItemServer) LogUser(ctx context.Context, in *pb.LogUserRequest) (*pb.LogUserResponse, error) {
	return &pb.LogUserResponse{}, nil
}

// AddItem recieves new item and responds item's id, lastUpdate time (in item), old lastUpdate or internal error
func (itemserv *ItemServer) AddItem(ctx context.Context, in *pb.AddItemRequest) (*pb.AddItemResponse, error) {
	// create new empty uitem (contains DB, user, list of items and error)
	uitem := storage.NewItem()
	itm := storage.Item{
		Name:       in.Uitem.Name,
		Parameters: map[string]string{},
		ImageLink:  make([]string, 0),
		TitleImage: make([]byte, 0),
		UserID:     in.Uitem.Userid,
	}
	// uitem.List[0].Name = in.Uitem.Name
	// copy(uitem.List[0].Tags, in.Uitem.Tags)
	// uitem.List[0].UserID = in.Uitem.Userid

	// save parameters
	for _, val := range in.Uitem.Params {
		itm.Parameters[val.Name] = val.Value
	}

	// send file to filestorage
	newfile := storage.NewFileToStorage()
	newfile.UserID = in.Uitem.Userid

	log.Println("list of files:", in.Uitem.Images)

	for _, file := range in.Uitem.Images {
		// change title and byte for every file
		newfile.Title = file.Title
		newfile.Data = &file.Body

		h := sha1.New()
		h.Write([]byte(file.Title + file.Body[:10]))
		sha1_hash := hex.EncodeToString(h.Sum(nil))

		newfile.ID =
			// save every file
			newfile.SaveNewFile()
		// copy new file address to item
		itm.ImageLink = append(itm.ImageLink, newfile.FileAddress)
		log.Println("FILE addre:", newfile.FileAddress)
	}

	// save new item to list
	uitem.List = append(uitem.List[:0], itm)

	// return DB structure
	uitem.DB = storage.NewToStorage(uitem)
	// run DB method
	uitem.DB.NewItems(ctx)
	if uitem.Err != nil {
		return nil, status.Errorf(codes.Internal, `unknown storage error`)
	}

	// create response
	var response = pb.AddItemResponse{
		Uitem:         &pb.Uitem{},
		OldLastUpdate: uitem.User.LastUpdate,
	}

	response.Uitem.Id = uitem.List[0].ID
	response.Uitem.Lastupdate = uitem.List[0].LastUpdate

	return &response, nil
}

func (itemserv *ItemServer) UpdateItem(ctx context.Context, in *pb.UpdateItemRequest) (*pb.UpdateItemResponse, error) {
	// create new empty uitem (contains DB, user, list of items and error)
	uitem := storage.NewItem()
	itm := storage.Item{
		ID:         in.Uitem.Id,
		Name:       in.Uitem.Name,
		Parameters: map[string]string{},
		UserID:     in.Uitem.Userid,
	}

	for _, val := range in.Uitem.Params {
		itm.Parameters[val.Name] = val.Value
	}
	uitem.List = append(uitem.List[:0], itm)

	// return DB structure
	uitem.DB = storage.NewToStorage(uitem)
	// run DB method
	uitem.DB.UpdateItems(ctx)
	if uitem.Err != nil {
		switch uitem.Err {
		case storage.ErrItemNotFound:
			return nil, status.Errorf(codes.NotFound, `item not found`)
		default:
			return nil, status.Errorf(codes.Internal, `unknown storage error`)
		}
	}

	// create response
	var response = pb.UpdateItemResponse{
		Uitem:         &pb.Uitem{},
		OldLastUpdate: uitem.User.LastUpdate,
	}

	response.Uitem.Id = uitem.List[0].ID
	response.Uitem.Lastupdate = uitem.List[0].LastUpdate

	return &pb.UpdateItemResponse{}, nil
}

func (itemserv *ItemServer) DeleteItem(ctx context.Context, in *pb.DeleteItemRequest) (*pb.DeleteItemResponse, error) {
	// create DB interface
	uitem := storage.NewItem()
	uitem.User.UserID = in.Userid
	uitem.List = append(uitem.List, storage.Item{ID: in.Itemid})

	// return DB structure
	uitem.DB = storage.NewToStorage(uitem)
	// run DB method
	uitem.DB.DeleteItem(ctx)
	if uitem.Err != nil {
		switch uitem.Err {
		case storage.ErrItemNotFound:
			return nil, status.Errorf(codes.NotFound, `item not found`)
		default:
			return nil, status.Errorf(codes.Internal, `unknown storage error`)
		}
	}

	response := pb.DeleteItemResponse{
		OldLastUpdate: uitem.User.LastUpdate,
		LastUpdate:    uitem.List[0].LastUpdate,
	}

	return &response, nil
}

func (itemserv *ItemServer) GetCatalog(ctx context.Context, in *pb.GetCatalogRequest) (*pb.GetCatalogResponse, error) {
	if in.Userid == "" {
		return nil, status.Errorf(codes.InvalidArgument, `user id not found`)
	}

	// create DB interface
	uit := storage.NewItem()
	uit.User.UserID = in.Userid
	uit.User.LastUpdate = in.LastUpdate

	// return DB structure
	uit.DB = storage.NewToStorage(uit)

	// run getCatalog
	uit.DB.GetCatalogByUser(ctx)

	response := pb.GetCatalogResponse{Uitem: make([]*pb.Uitem, 0)}

	for _, itemfromDB := range uit.List {

		// create an item
		nextpbUitem := pb.Uitem{}

		// load in item
		nextpbUitem.Id = itemfromDB.ID
		nextpbUitem.Name = itemfromDB.Name
		nextpbUitem.Userid = itemfromDB.UserID
		nextpbUitem.Lastupdate = itemfromDB.LastUpdate

		for pkey, pval := range itemfromDB.Parameters {
			nextpbUitem.Params = append(nextpbUitem.Params, &pb.Parameter{Name: pkey, Value: pval})
		}

		for i, flink := range itemfromDB.ImageLink {
			body, err := storage.GetTitle(flink)
			if err == nil {
				nextpbUitem.Images = append(nextpbUitem.Images, &pb.Image{Title: fmt.Sprintf("%s_%d", itemfromDB.ID, i), Body: body})
			}
		}
		// add item to slice
		response.Uitem = append(response.Uitem, &nextpbUitem)
	}

	response.LastUpdate = uit.User.LastUpdate
	return &response, nil
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
		return nil, status.Errorf(codes.Unauthenticated, "tokens list empty")
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
