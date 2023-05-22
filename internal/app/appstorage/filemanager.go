package appstorage

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
)

func userFolderPass(userid string) string {
	return filepath.Join(localstorageaddress, userid)
}

func itemFolderPass(userid, itemid string) string {
	return filepath.Join(userFolderPass(userid), itemid)
}

func NewFileStruct() *File {
	return &File{Body: make([]byte, 0)}
}

func (file *File) PrepareFile(pass string) error {
	//read file
	bytes, size, err := readFile(file.Address)
	if err != nil {
		return fmt.Errorf("read file %s error:%v", file.Address, err)
	}

	// size limit. Redo to batch
	if size > maxFileSize {
		return fmt.Errorf("file is too big. Max size = %d bytes", maxFileSize)
	}

	// encrypt file
	file.Body, err = FileEncrypt(bytes, pass)
	if err != nil {
		return fmt.Errorf("encrypt file %s error:%v", file.Address, err)
	}
	return nil
}

func (file *File) SaveFileLocal(pass string) error {
	if file.FileID == "" {
		return fmt.Errorf("fileID not ready, file not saved local, need run request for files")
	}

	// check body
	if len(file.Body) == 0 {
		return fmt.Errorf("empty file.Body in file %s", file.FileID)
	}

	// create path localstorage/userid/itemid
	path := filepath.Join(localstorageaddress, file.UserID)
	path = filepath.Join(path, file.ItemID)

	// create folder if not exists
	err := os.MkdirAll(path, os.ModePerm)
	if err != nil {
		return fmt.Errorf("can't create local directory %s, error:%v", path, err)
	}

	// decode file
	decodedfile, err := FileDecrypt(file.Body, pass)

	// write file
	path = filepath.Join(path, file.FileID)
	err = os.WriteFile(path, decodedfile, 0644)
	if err != nil {
		return fmt.Errorf("write file %s error:%v", path, err)
	}

	// add address
	file.Address = path

	// register new file to local mapa
	Catalog.mu.Lock()
	Catalog.Files[file.FileID] = file
	Catalog.mu.Unlock()

	return nil
}

func readFile(fileaddress string) ([]byte, int64, error) {
	file, err := os.Open(fileaddress)

	if err != nil {
		return nil, 0, err
	}
	defer file.Close()

	stats, statsErr := file.Stat()
	if statsErr != nil {
		return nil, 0, statsErr
	}

	var size int64 = stats.Size()
	bytes := make([]byte, size)

	bufr := bufio.NewReader(file)
	_, err = bufr.Read(bytes)

	return bytes, size, err
}

func deleteFolderByItemID(userid, itemid string) error {
	err := os.RemoveAll(itemFolderPass(userid, itemid))
	if err != nil {
		return fmt.Errorf("can't delete folder:%v For itemID =%s", err, itemid)
	}
	return nil
}

// func saveFilesLocal(cc *client.ClientConnector) {
// 	path := filepath.Join(localstorageaddress, cc.UserID)

// 	for _, item := range cc.Items {
// 		// create folder if not exists
// 		err := os.MkdirAll(path, os.ModePerm)
// 		if err != nil {
// 			log.Println("file storage can't create directory to store file:", err)
// 			return
// 		}

// 		for _, file := range item.Images {
// 			if len(file.Body) != 0 {
// 				path = filepath.Join(path, item.Id)

// 				err = os.WriteFile(path, file.Body, 0644)
// 				if err != nil {
// 					log.Println("file storage error, check FILESERVERADDRESS available:", err)
// 				}

// 			}
// 		}
// 	}
// }