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

	// prepare to server
	tostor := serverstorage.NewToStorage()
	tostor.List = append(tostor.List, *serveritem)

	tostor.DB = serverstorage.NewStorager(tostor)

	// to server
	err := tostor.DB.PutItems()
	if err != nil {
		if errors.Is(err, serverstorage.ErrEmptyRequest) {
			log.Println("put items:", err)
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

func (itemserv *ItemServer) UploadFile(ctx context.Context, in *pb.UploadFileRequest) (*pb.UploadFileResponse, error) {
	// check all body recieved
	if !checkHas(in.File.Body, in.File.Hash) {
		log.Println("hash error")
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
	tostorfile.Body = in.File.Body

	// create tostor
	tostor := serverstorage.NewToStorage()
	tostor.File = *tostorfile
	tostor.DB = serverstorage.NewStorager(tostor)

	// upload file to server
	err := tostor.DB.UploadFile()
	if err != nil {
		log.Println("can't upload file to server:", tostor.File.FileID, "error:", err)
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

func checkHas(body []byte, hash []byte) bool {
	sum := sha256.Sum256(body)
	return bytes.Equal(sum[:], hash[:])
}

func (itemserv *ItemServer) UpdateByLastUpdate(ctx context.Context, in *pb.UpdateByLastUpdateRequest) (*pb.UpdateByLastUpdateResponse, error) {

	// prepare to server
	tostor := serverstorage.NewToStorage()
	tostor.User.UserID = in.Userid
	tostor.User.LastUpdate = in.Lastupdate

	tostor.DB = serverstorage.NewStorager(tostor)

	// to server
	err := tostor.DB.UpdateByLastUpdate()
	if err != nil {
		log.Println("UpdateByLastUpdate err:", err)
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
		}

		response.Item = append(response.Item, &itm)
	}
	return &response, nil
}

func (itemserv *ItemServer) GetFileByFileID(ctx context.Context, in *pb.GetFileByFileIDRequest) (*pb.GetFileByFileIDResponse, error) {

	// prepare to server
	file := serverstorage.NewFile()
	file.FileID = in.Fileid
	file.UserID = in.Userid

	tostor := serverstorage.NewToStorage()
	tostor.File = *file

	tostor.DB = serverstorage.NewStorager(tostor)

	// to server
	err := tostor.DB.GetFileByFileID()
	if err != nil {
		return nil, status.Errorf(codes.Internal, `can't download file`)
	}

	// create response
	responsefile := pb.File{
		Itemid: tostor.File.ItemID,
		Userid: tostor.File.UserID,
		Fileid: tostor.File.FileID,
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

// RegUser returns new user's id
func (itemserv *ItemServer) DeleteEntity(ctx context.Context, in *pb.DeleteEntityRequest) (*pb.DeleteEntityResponse, error) {

	// create response
	var response = pb.DeleteEntityResponse{
		Userid: make([]string, 0),
		Itemid: make([]string, 0),
		Fileid: make([]string, 0),
	}

	// operates files first
	for _, fileid := range in.Fileid {
		serverfile := serverstorage.NewFile()
		serverfile.FileID = fileid

		// prepare to server
		tostor := serverstorage.NewToStorage()
		tostor.File = *serverfile

		tostor.DB = serverstorage.NewStorager(tostor)

		// to server
		err := tostor.DB.DeleteFile()
		if err != nil {
			if errors.Is(err, serverstorage.ErrItemNotFound) {
				log.Println("file not found, skipped fileid:", fileid)
				continue
			}
			// return error
			response.Fileid = append(response.Fileid, tostor.File.FileID)
			log.Println("error delete file:", fileid)
		}
	}

	// operates items then
	tostor := serverstorage.NewToStorage()

	// for every itemid from request
	for _, itemid := range in.Itemid {
		serveritem := serverstorage.NewItem()
		serveritem.ItemID = itemid
		tostor.List = append(tostor.List, *serveritem)
	}

	// return interface
	tostor.DB = serverstorage.NewStorager(tostor)

	// to server
	err := tostor.DB.DeleteItems()
	if err != nil {
		// return error?
		log.Println("error delete items")
	}

	for _, item := range tostor.List {
		if item.ItemID != "" {
			response.Itemid = append(response.Itemid, item.ItemID)
		}
	}

	return &response, nil
}
