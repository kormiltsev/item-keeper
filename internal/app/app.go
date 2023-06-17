package app

import (
	"bytes"
	"context"
	"crypto/sha256"
	"fmt"
	"log"
	"time"

	appstorage "github.com/kormiltsev/item-keeper/internal/app/appstorage"
	clientconnector "github.com/kormiltsev/item-keeper/internal/client"
	pb "github.com/kormiltsev/item-keeper/internal/server/proto"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func LoadFromFile(cryptokey string) error {
	var err error
	currentuserencryptokey = cryptokey
	currentuser, currentlastupdate, err = appstorage.ReadDecryptedCatalog(cryptokey)
	return err
}

func ShowCatalog() (map[int64]*appstorage.Item, error) {
	operator, err := appstorage.ReturnOperator(currentuser)
	if err != nil {
		return nil, fmt.Errorf("operator local:%v", err)
	}

	err = operator.UploadFilesAddresses()
	if err != nil {
		return nil, fmt.Errorf("can't show catalog:%v", err)
	}
	return operator.Mapa.Items, nil
}

func SaveToFile() error {
	op, err := appstorage.ReturnOperator(currentuser)
	if err != nil {
		log.Println("ReturnOperator:", err)
		return err
	}

	return op.SaveEncryptedCatalog(currentuserencryptokey)
}

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
		log.Println("put item: can't update from server:", errup)
		return nil
	}

	// run upload file in goroutine
	// REDO to channel and worker ?
	go uploadFileFromItemToServer(appitem)

	return nil

}

func uploadFileFromItemToServer(appitem *appstorage.Item) {
	if len(appitem.UploadAddress) == 0 {
		return
	}

	log.Println("starts file upload: ", appitem.UploadAddress)

	ctx := context.Background()
	// prepare and send files after NewItemID was created by server
	for i, fileaddress := range appitem.UploadAddress {
		file := appstorage.NewFileStruct()
		file.FileID = int64(i) // temporary id to upload
		file.ItemID = appitem.ItemID
		file.UserID = appitem.UserID
		file.Address = fileaddress

		err := encodeAndUploadFileToServer(ctx, file)
		if err != nil {
			// send error to error channel
			//
			// ===========================

			log.Printf("File %s not uploaded:%v", fileaddress, err)
			return
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
		// appitem.FileIDs = append(appitem.FileIDs, file.FileID)

		// reg fileID to item to Catalog, request interface
		operator, erro := appstorage.ReturnOperator(currentuser)
		if erro != nil {
			log.Printf("can't save local:%v\n", erro)
		}
		err = operator.RegisterFilesToItems(*file)
		if err != nil {
			log.Println("Can't register file to list of Items:", err)
		}
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

				if again == 3 {
					return fmt.Errorf("UploadFile error, server not responding")
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

	if len(response.Item) == 0 && len(response.File) == 0 {
		currentlastupdate = response.Lastupdate
		log.Println("response empty, looks like everithing updated")
		return nil
	}

	// save item to Catalog, request interface
	operator, erro := appstorage.ReturnOperator(currentuser)
	if erro != nil {
		return fmt.Errorf("can't save local:%v", erro)
	}

	// set last update same as data from server
	operator.LastUpdate = response.Lastupdate

	answer := make([]*appstorage.Item, 0, len(response.Item))

	// for every item decode and add file ids into appstorage item
	for _, itm := range response.Item {

		// if item deleted
		if itm.Deleted {

			if err := operator.DeleteItemByID(itm.Itemid); err != nil {
				log.Println("error with local delete:", err)
			}
			continue
		}

		// decode to local item struct
		if len(itm.Body) == 0 {
			log.Println("update.response empty item.body")
			return fmt.Errorf("update.response empty item.body")
		}

		newitem, err := appstorage.Decode(itm.Body, currentuserencryptokey)
		if err != nil {
			log.Println("error on decoding:", err)
			continue
		}

		// add itemid from server
		newitem.ItemID = itm.Itemid

		// making answer slice of items
		answer = append(answer, newitem)
	}

	// save local
	// log.Println("answer: ", answer)
	err = operator.PutItems(answer...)
	if err != nil {
		log.Println("can't save item local:", err)
	}

	// for every fileNoBody add in id to item
	fls := make([]appstorage.File, 0, len(response.File))
	flsids := make([]int64, 0, len(response.File))

	for _, fle := range response.File {
		log.Println("response file: ", fle)
		if fle.Deleted {
			// copy to appstorage type
			appfile := appstorage.NewFileStruct()
			appfile.FileID = fle.Fileid
			appfile.ItemID = fle.Itemid
			appfile.UserID = fle.Userid

			// DeleteFileLocal deregister and delete file
			appfile.DeleteFileLocal()
			continue
		}
		// register files id to items in mapa of items
		var fl = appstorage.File{
			FileID: fle.Fileid,
			ItemID: fle.Itemid,
		}
		fls = append(fls, fl)

		flsids = append(flsids, fl.FileID)
	}

	// set lust update as on server
	currentlastupdate = response.Lastupdate

	err = operator.RegisterFilesToItems(fls...)
	if err != nil {
		log.Println("can't register files to item local:", err)
		return err
	}

	// download files (run goroutine)
	go requestFilesByFileID(0, flsids)

	return nil
}

func requestFilesByFileID(tryNumber int, listOfFileids []int64) {
	if len(listOfFileids) == 0 {
		return
	}

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
func RequestFilesByFileID(ctx context.Context, fileids ...int64) ([]int64, []int64, error) {
	log.Println("requestfiles RequestFilesByFileID")
	var err error
	if len(fileids) == 0 {
		log.Println("0 files requested")
		return nil, nil, appstorage.ErrEmptyRequest
	}

	readyfiles := make([]int64, 0)
	errorfiles := make([]int64, 0)
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

func requestFileByFileID(ctx context.Context, fileid int64) error {
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
		return fmt.Errorf("response body damaged, file id:%d", req.Fileid)
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
		Userid: currentuser,
		Itemid: itemids,
	}

	// gRPC
	cc := clientconnector.NewClientConnector(ctx)
	cl := *cc.Client

	// run request
	response, err := cl.DeleteEntity(cc.Ctx, &req)
	if len(response.Itemid) == 0 && len(response.Fileid) == 0 && err != nil {
		return nil, fmt.Errorf("no error items returned, but error:%v", err)
	}

	// Need to retry or just inform user about error with deletion?
	for _, itemIDNotDeleted := range response.Itemid {
		log.Println("ITEM not deleted:", itemIDNotDeleted)
	}
	for _, fileIDNotDeleted := range response.Fileid {
		log.Println("FILE not deleted:", fileIDNotDeleted)
	}

	operator, err := appstorage.ReturnOperator(currentuser)
	if err != nil {
		return nil, fmt.Errorf("operator local:%v", err)
	}

	// delete local
	for _, itmid := range itemids {
		operator.DeleteItemByID(itmid)
	}

	// upload new status from server
	// go UpdateDataFromServer(ctx)
	if err := UpdateDataFromServer(ctx); err != nil {
		log.Println("deleted. but update from server error:", err)
	}
	return response.Itemid, nil
}

func UploadConfigsApp() string {
	return clientconnector.UploadConfigsCli()
}
