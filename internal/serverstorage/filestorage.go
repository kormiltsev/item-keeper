package serverstorage

import (
	"fmt"
	"os"
	"path/filepath"
)

var storageaddress = "./data/ServerStorage"

func deleteFilesByItemID(userid, itemid string) error {
	err := os.RemoveAll(itemFolderPass(userid, itemid))
	if err != nil {
		return fmt.Errorf("can't delete folder:%v For itemID =%s", err, itemid)
	}
	return nil
}

func deleteFileByFileID(userid, itemid, fileid string) error {
	return os.Remove(filepath.Join(itemFolderPass(userid, itemid), fileid))
}

func userFolderPass(userid string) string {
	return filepath.Join(storageaddress, userid)
}

func itemFolderPass(userid, itemid string) string {
	return filepath.Join(userFolderPass(userid), itemid)
}
