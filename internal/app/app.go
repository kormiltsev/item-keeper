package app

import (
	"bytes"
	"context"
	"crypto/sha256"
	"fmt"
	"log"
	"strconv"
	"time"

	appstorage "github.com/kormiltsev/item-keeper/internal/app/appstorage"
	clientconnector "github.com/kormiltsev/item-keeper/internal/client"
	pb "github.com/kormiltsev/item-keeper/internal/server/proto"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func NewAppItem() *appstorage.Item {
	return &appstorage.Item{}
}

func AddNewItem(ctx context.Context, appitem *appstorage.Item) error {
	// appitem := presetItem()
	appitem.UserID = currentuser

	// set context with time limit
	ctxto, cancel := context.WithTimeout(ctx, 5000*time.Millisecond)
	defer cancel()

	// encode item data
	body, err := appitem.Encode(currentuserencryptokey)
	if err != nil {
		log.Println("error item encoding:", err)
		return err
	}

	// buil request
	req := pb.PutItemsRequest{
		Item: &pb.Item{
			Itemid: appitem.ItemID,
			Userid: appitem.UserID,
			Body:   body,
		},
	}

	// gRPC
	cc := clientconnector.NewClientConnector(ctxto)
	cl := *cc.Client

	// run request
	response, err := cl.PutItems(cc.Ctx, &req)
	if err != nil {
		if e, ok := status.FromError(err); ok {
			if e.Code() == codes.InvalidArgument {
				return fmt.Errorf(`empty request:%v`, e.Message())
			} else {
				return fmt.Errorf(`PutItem:%v:%v`, e.Code(), e.Message())
			}
		}
		return fmt.Errorf(`PutItem error:%v`, err)
	}

	// if empty response
	if response.Item.Itemid == 0 {
		log.Println("FAIL: PutItems() empty response from server")
		return fmt.Errorf(`server internal error`)
	}

	// save new itemID into local item
	appitem.ItemID = response.Item.Itemid

	// update data from server
	if errup := UpdateDataFromServer(ctx); errup != nil {
		log.Println("put item: can't update from server:", err)
		return nil
	}

	// run upload file in goroutine
	// REDO to channel and worker ?
	go uploadFileFromItemToServer(appitem)

	return nil

	// // save item to Catalog, request interface
	// operator, erro := appstorage.ReturnOperator(appitem.UserID)
	// if erro != nil {
	// 	log.Println(erro)
	// 	return err
	// }

	// // save local
	// err = operator.PutItems(appitem)
	// if err != nil {
	// 	log.Println("can't save item local:", err)
	// 	return err
	// }
}

func uploadFileFromItemToServer(appitem *appstorage.Item) {

	log.Println("starts file upload: ", appitem.UploadAddress)

	ctx := context.Background()
	// prepare and send files after NewItemID was created by server
	for i, fileaddress := range appitem.UploadAddress {
		file := appstorage.NewFileStruct()
		file.FileID = strconv.Itoa(i) // temporary id to upload
		file.ItemID = appitem.ItemID
		file.UserID = appitem.UserID
		file.Address = fileaddress

		err := encodeAndUploadFileToServer(ctx, file)
		if err != nil {
			// send error to error channel
			//
			// ===========================

			log.Printf("File %s not uploaded:%v", fileaddress, err)
		}

		// check if other user authorized local already? then no need to save file local
		if file.UserID != currentuser {
			return
		}

		// save original file local and register in Catalog.Files
		err = file.SaveFileLocal(currentuserencryptokey)
		if err != nil {
			log.Println("Can't save file local:", err)
			continue
		}

		// copy file address into new appitem
		appitem.FileIDs = append(appitem.FileIDs, file.FileID)
	}
}

func encodeAndUploadFileToServer(ctx context.Context, file *appstorage.File) error {
	// read and encode file
	err := file.PrepareFile(currentuserencryptokey)
	if err != nil {
		log.Println("can't prepare file", file.Address, "error:", err)
		return fmt.Errorf("file encoding error:%s", file.Address)
	}

	return uploadEncryptedFileToServer(ctx, file)
}

func uploadEncryptedFileToServer(ctx context.Context, file *appstorage.File) error {

	hash := sha256.Sum256(file.Body)

	// grpc
	// buil request
	req := pb.UploadFileRequest{
		File: &pb.File{
			Itemid: file.ItemID,
			Userid: file.UserID,
			Fileid: file.FileID,
			Body:   file.Body,
			Hash:   hash[:],
		},
	}

	// set tokens
	cc := clientconnector.NewClientConnector(ctx)
	cl := *cc.Client

	// run request
	response, err := cl.UploadFile(cc.Ctx, &req)
	if err != nil {
		if e, ok := status.FromError(err); ok {
			if e.Code() == codes.DataLoss {

				// retry sending 3 times
				again := 0
				for err != nil || again < 3 {
					response, err = cl.UploadFile(cc.Ctx, &req)
					again++
				}
				if err != nil {
					return fmt.Errorf(`recieving data lost: %v`, e.Message())
				}
			} else {
				return fmt.Errorf(`UploadFile error:%v:%v`, e.Code(), e.Message())
			}
		}
		if err != nil {
			return fmt.Errorf(`UploadFile error:%v`, err)
		}
	}

	file.FileID = response.Fileid
	file.UserID = response.Userid
	file.ItemID = response.Itemid
	return nil
}

// UpdateDataFromServer request lastUpdate from server and save items rfom response local
func UpdateDataFromServer(ctx context.Context) error {

	// buil request
	req := pb.UpdateByLastUpdateRequest{
		Userid:     currentuser,
		Lastupdate: currentlastupdate,
	}

	// gRPC
	cc := clientconnector.NewClientConnector(ctx)
	cl := *cc.Client

	// run request
	response, err := cl.UpdateByLastUpdate(cc.Ctx, &req)
	if err != nil {
		return fmt.Errorf(`update request error:%v`, err)
	}

	if len(response.Item) == 0 {
		if response.Lastupdate > currentlastupdate {
			log.Println("response empty, but lastupdate date different")
			return fmt.Errorf("last update not equal, but no items resieved")
		}
		log.Println("response empty, looks like everithing updated")
		return nil
	}

	// save item to Catalog, request interface
	operator, erro := appstorage.ReturnOperator(currentuser)
	if erro != nil {
		return fmt.Errorf("can't save local:%v", erro)
	}

	// set last update same as data from server
	currentlastupdate = response.Lastupdate
	operator.LastUpdate = response.Lastupdate

	answer := make([]*appstorage.Item, 0, len(response.Item))

	// for every item decode and add file ids into appstorage item
	for _, itm := range response.Item {

		// if item deleted
		if itm.Deleted {
			// log.Println("deleted da:", itm.Itemid)
			// delete item in Catalog, request interface
			operator, erro := appstorage.ReturnOperator(currentuser)
			if erro != nil {
				log.Println("cant delete local, Operator error:", erro)
				continue
			}

			if err := operator.DeleteItemByID(itm.Itemid); err != nil {
				log.Println("error with local delete:", err)
			}
			continue
		}

		// decode to local item struct
		newitem, err := appstorage.Decode(itm.Body, currentuserencryptokey)
		if err != nil {
			log.Println("error on decoding:", err)
			continue
		}

		// add itemid from server
		newitem.ItemID = itm.Itemid

		// upload file ids into local item
		newitem.FileIDs = make([]string, 0, len(itm.Filesid))
		for _, fileid := range itm.Filesid {
			if len(fileid) == 0 {
				continue
			}
			newitem.FileIDs = append(newitem.FileIDs, fileid)
		}

		// making answer slice of items
		answer = append(answer, newitem)
	}

	// save local
	// log.Println("answer: ", answer)
	err = operator.PutItems(answer...)
	if err != nil {
		log.Println("can't save item local:", err)
	}

	// download files
	fileIDs := make([]string, 0)
	for _, item := range answer {
		fileIDs = append(fileIDs, item.FileIDs...)
	}

	// download files (run goroutine)
	if len(fileIDs) != 0 {
		go requestFilesByFileID(0, fileIDs)
	}

	return nil
}

func requestFilesByFileID(tryNumber int, listOfFileids []string) {
	doneFilesID, nondoneFilesID, err := RequestFilesByFileID(context.Background(), listOfFileids...)
	if err != nil {
		log.Printf("%d/%d files downloaded, error: %v", len(doneFilesID), len(listOfFileids), err)
		if (nondoneFilesID == nil || len(nondoneFilesID) != 0) && tryNumber < 3 {

			time.Sleep(1 * time.Second)

			log.Println("retry download file, attempt #", tryNumber)

			requestFilesByFileID(tryNumber+1, nondoneFilesID)
		}
	} else {
		log.Printf("%d/%d files downloaded", len(doneFilesID), len(listOfFileids))
	}
}

// RequestFilesByFileID returns files ids dawnloaded successfully and error (if some of them not recieved)
func RequestFilesByFileID(ctx context.Context, fileids ...string) ([]string, []string, error) {
	var err error
	if len(fileids) == 0 {
		return nil, nil, fmt.Errorf("empty request")
	}

	readyfiles := make([]string, 0)
	errorfiles := make([]string, 0)
	for _, fileid := range fileids {

		err = requestFileByFileID(ctx, fileid)
		if err == nil {
			readyfiles = append(readyfiles, fileid)
		} else {
			errorfiles = append(errorfiles, fileid)
			log.Println("file not recieved:", fileid)
		}
	}

	if len(readyfiles) != len(fileids) {
		return readyfiles, errorfiles, fmt.Errorf("some of file not recieved")
	}

	return readyfiles, nil, nil
}

func requestFileByFileID(ctx context.Context, fileid string) error {
	// buil request
	req := pb.GetFileByFileIDRequest{
		Userid: currentuser,
		Fileid: fileid,
	}

	// gRPC
	cc := clientconnector.NewClientConnector(ctx)
	cl := *cc.Client

	// run request
	response, err := cl.GetFileByFileID(cc.Ctx, &req)
	if err != nil {
		if e, ok := status.FromError(err); ok {
			if e.Code() == codes.Internal {
				return fmt.Errorf(`request file error: %v`, e.Message())
			}
		}
		if err != nil {
			return fmt.Errorf(`request file error:%v`, err)
		}
	}

	if !checkHas(response.File.Body, response.File.Hash) {
		return fmt.Errorf("response body damaged, file id:%s", req.Fileid)
	}

	// check if other user authorized local already? then no need to save file local
	if response.File.Userid != currentuser {
		return nil
	}

	// copy to appstorage tipe
	appfile := appstorage.NewFileStruct()
	appfile.FileID = response.File.Fileid
	appfile.ItemID = response.File.Itemid
	appfile.UserID = response.File.Userid
	appfile.Body = make([]byte, len(response.File.Body))
	copy(appfile.Body, response.File.Body)

	// encode and save file local on client
	err = appfile.SaveFileLocal(currentuserencryptokey)
	if err != nil {
		log.Println("can't save file on client:", err)
	}
	return nil
}

func checkHas(body []byte, hash []byte) bool {
	sum := sha256.Sum256(body)
	return bytes.Equal(sum[:], hash)
}

// DeleteItems return errored items and error
func DeleteItems(ctx context.Context, itemids []int64) ([]int64, error) {
	// if empty
	if len(itemids) == 0 {
		return nil, fmt.Errorf("empty request")
	}

	// build request
	req := pb.DeleteEntityRequest{
		Itemid: itemids,
	}

	// gRPC
	cc := clientconnector.NewClientConnector(ctx)
	cl := *cc.Client

	// run request
	response, err := cl.DeleteEntity(cc.Ctx, &req)
	if len(response.Itemid) == 0 && err != nil {
		return nil, fmt.Errorf("no error items returned, but error:%v", err)
	}

	// upload new status from server
	// go UpdateDataFromServer(ctx)
	if err := UpdateDataFromServer(ctx); err != nil {
		log.Println("deleted. but update from server error:", err)
	}

	return response.Itemid, nil
}
