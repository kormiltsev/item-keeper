package serverstorage

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"
)

var storageaddress = "./data/ServerStorage"

func fileUploadToFileStorage(id int64, file *File) {

	// storage switcher TODO here
	//
	// =========================

	err := fileUploadToFileStorageLocal(id, file)
	if err != nil {
		log.Println("Can't save file local:", err)
	}
}

func fileUploadToFileStorageLocal(id int64, file *File) error {
	// create path localstorage/userid/itemid
	path := filepath.Join(storageaddress, file.UserID)
	path = filepath.Join(path, strconv.FormatInt(file.ItemID, 10))

	// create folder if not exists
	err := os.MkdirAll(path, os.ModePerm)
	if err != nil {
		return fmt.Errorf("can't create local directory %s, error:%v", path, err)
	}

	// write file
	path = filepath.Join(path, strconv.FormatInt(id, 10))
	err = os.WriteFile(path, file.Body, 0644)
	if err != nil {
		return fmt.Errorf("write file %s error:%v", path, err)
	}
	return nil
}

func deleteFilesByItemID(userid string, itemid int64) error {
	err := os.RemoveAll(itemFolderPass(userid, itemid))
	if err != nil {
		return fmt.Errorf("can't delete folder:%v For itemID =%d", err, itemid)
	}
	return nil
}

func deleteFileByFileID(itemid int64, userid string, fileid int64) error {
	return os.Remove(filepath.Join(itemFolderPass(userid, itemid), strconv.FormatInt(fileid, 10)))
}

func userFolderPass(userid string) string {
	return filepath.Join(storageaddress, userid)
}

func itemFolderPass(userid string, itemid int64) string {
	return filepath.Join(userFolderPass(userid), strconv.FormatInt(itemid, 10))
}
