package serverstorage

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
)

var storageaddress = "./data/ServerStorage"

func deleteFilesByItemID(userid string, itemid int64) error {
	err := os.RemoveAll(itemFolderPass(userid, itemid))
	if err != nil {
		return fmt.Errorf("can't delete folder:%v For itemID =%d", err, itemid)
	}
	return nil
}

func deleteFileByFileID(itemid int64, userid, fileid string) error {
	return os.Remove(filepath.Join(itemFolderPass(userid, itemid), fileid))
}

func userFolderPass(userid string) string {
	return filepath.Join(storageaddress, userid)
}

func itemFolderPass(userid string, itemid int64) string {
	return filepath.Join(userFolderPass(userid), strconv.FormatInt(itemid, 10))
}
