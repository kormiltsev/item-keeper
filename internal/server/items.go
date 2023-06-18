package server

import (
	"bytes"
	"context"
	"crypto/sha256"
	"errors"
	"log"

	pb "github.com/kormiltsev/item-keeper/internal/server/proto"
	serverstorage "github.com/kormiltsev/item-keeper/internal/serverstorage"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// RegUser returns new user's id
func (itemserv *ItemServer) PutItems(ctx context.Context, in *pb.PutItemsRequest) (*pb.PutItemsResponse, error) {

	serveritem := serverstorage.NewItem()
	serveritem.UserID = in.Item.Userid
	serveritem.Body = in.Item.Body
	serveritem.ItemID = in.Item.Itemid

	// prepare to server
	tostor := serverstorage.NewToStorage()
	tostor.List = append(tostor.List, *serveritem)

	tostor.DB = serverstorage.NewStorager(tostor)

	// to server
	err := tostor.DB.PutItems(ctx)
	if err != nil {
		if errors.Is(err, serverstorage.ErrEmptyRequest) {
			log.Printf("put items (cli token = %v): %v", getClientToken(ctx), err)
			return nil, status.Errorf(codes.InvalidArgument, `empty request`)
		}
		log.Println("Unknown error from storage: ", err)
		return nil, status.Errorf(codes.Internal, `unknown error`)
	}

	// create response
	var response = pb.PutItemsResponse{
		Item: &pb.Item{
			Itemid: tostor.List[0].ItemID,
			Userid: tostor.List[0].UserID,
			Body:   tostor.List[0].Body,
		},
	}

	return &response, nil
}

// UploadFile recieve File and save it.
func (itemserv *ItemServer) UploadFile(ctx context.Context, in *pb.UploadFileRequest) (*pb.UploadFileResponse, error) {
	// check all body recieved
	if !checkHas(in.File.Body, in.File.Hash) {
		log.Printf("hash error (cli token = %v)", getClientToken(ctx))
		return nil, status.Errorf(codes.DataLoss, `hash not equal`)
	}

	// if file bached assebly file here
	//
	// ================================

	// to server storage
	tostorfile := serverstorage.NewFile()
	tostorfile.FileID = in.File.Fileid
	tostorfile.ItemID = in.File.Itemid
	tostorfile.UserID = in.File.Userid
	tostorfile.FileName = in.File.Filename
	tostorfile.Body = in.File.Body

	// create tostor
	tostor := serverstorage.NewToStorage()
	tostor.File = *tostorfile
	tostor.DB = serverstorage.NewStorager(tostor)

	// upload file to server
	err := tostor.DB.UploadFile(ctx)
	if err != nil {
		log.Printf("can't upload file to server (cli token=%v):%d, error: %v", getClientToken(ctx), tostor.File.FileID, err)
		return nil, status.Errorf(codes.Internal, `can't save file`)
	}

	// create response
	var response = pb.UploadFileResponse{
		Fileid: tostor.File.FileID,
		Userid: tostor.File.UserID,
		Itemid: tostor.File.ItemID,
	}

	return &response, nil

}

// checkHas returns tru is hash the same.
func checkHas(body []byte, hash []byte) bool {
	sum := sha256.Sum256(body)
	return bytes.Equal(sum[:], hash[:])
}

// UpdateByLastUpdate responds last updates (items and files IDs) with last update later than requested time stamp (now just additive int64)
func (itemserv *ItemServer) UpdateByLastUpdate(ctx context.Context, in *pb.UpdateByLastUpdateRequest) (*pb.UpdateByLastUpdateResponse, error) {

	log.Printf("lastupdate requested (cli token = %v):%d, for user: %s", getClientToken(ctx), in.Lastupdate, in.Userid)

	// prepare to server
	tostor := serverstorage.NewToStorage()
	tostor.User.UserID = in.Userid
	tostor.User.LastUpdate = in.Lastupdate

	tostor.DB = serverstorage.NewStorager(tostor)

	// to server
	err := tostor.DB.UpdateByLastUpdate(ctx)
	if err != nil {
		log.Printf("UpdateByLastUpdate (cli token = %v) err:%v", getClientToken(ctx), err)
	}

	// create response
	var response = pb.UpdateByLastUpdateResponse{
		Lastupdate: tostor.User.LastUpdate,
	}

	for _, item := range tostor.List {
		itm := pb.Item{
			Itemid:  item.ItemID,
			Userid:  item.UserID,
			Body:    item.Body,
			Filesid: item.FilesID,
			Deleted: item.Deleted,
		}

		response.Item = append(response.Item, &itm)
	}

	for _, filesNoBody := range tostor.FilesNoBody {
		fle := pb.File{
			Itemid:  filesNoBody.ItemID,
			Userid:  filesNoBody.UserID,
			Fileid:  filesNoBody.FileID,
			Deleted: filesNoBody.Deleted,
		}

		response.File = append(response.File, &fle)
	}
	return &response, nil
}

// GetFileByFileID returns files (binary)
func (itemserv *ItemServer) GetFileByFileID(ctx context.Context, in *pb.GetFileByFileIDRequest) (*pb.GetFileByFileIDResponse, error) {
	// prepare to server
	file := serverstorage.NewFile()
	file.FileID = in.Fileid
	file.UserID = in.Userid

	tostor := serverstorage.NewToStorage()
	tostor.File = *file

	tostor.DB = serverstorage.NewStorager(tostor)

	// to server
	err := tostor.DB.GetFileByFileID(ctx)
	if err != nil {
		return nil, status.Errorf(codes.Internal, `can't download file`)
	}

	// create response
	responsefile := pb.File{
		Itemid:   tostor.File.ItemID,
		Userid:   tostor.File.UserID,
		Fileid:   tostor.File.FileID,
		Filename: tostor.File.FileName,
	}

	// copy body
	responsefile.Body = make([]byte, len(tostor.File.Body))
	copy(responsefile.Body, tostor.File.Body)

	// make hash
	hash := sha256.Sum256(responsefile.Body)
	responsefile.Hash = hash[:]

	// make response
	var response = pb.GetFileByFileIDResponse{
		File: &responsefile,
	}

	return &response, nil
}

// DeleteEntity delete items or/and files. note: to delete files only send empty Itemid in request.
func (itemserv *ItemServer) DeleteEntity(ctx context.Context, in *pb.DeleteEntityRequest) (*pb.DeleteEntityResponse, error) {

	// create response
	var response = pb.DeleteEntityResponse{
		Userid: in.Userid,
		Itemid: make([]int64, 0),
		Fileid: make([]int64, 0),
	}

	// operates files first
	serveritem := serverstorage.NewItem()
	serveritem.UserID = in.Userid
	serveritem.FilesID = make([]int64, len(in.Fileid))
	copy(serveritem.FilesID, in.Fileid)

	// prepare to server
	tostor := serverstorage.NewToStorage()
	tostor.List = append(tostor.List, *serveritem)

	// for every itemid from request
	for _, itemid := range in.Itemid {
		serveritem := serverstorage.NewItem()
		serveritem.UserID = in.Userid
		serveritem.ItemID = itemid
		tostor.List = append(tostor.List, *serveritem)
	}

	// return interface
	tostor.DB = serverstorage.NewStorager(tostor)

	// to server
	err := tostor.DB.DeleteItems(ctx)
	if err != nil {
		// return error?
		log.Printf("error delete items (cli token = %v): %v", getClientToken(ctx), err)
		return nil, status.Errorf(codes.Internal, `can't delete`)
	}

	return &response, nil
}
