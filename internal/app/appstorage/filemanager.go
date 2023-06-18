package appstorage

import (
	"bufio"
	"encoding/base64"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"

	configs "github.com/kormiltsev/item-keeper/internal/configsClient"
)

func userFolderPass(userid string) string {
	return filepath.Join(configs.ClientConfig.FileFolder, userid)
}

func itemFolderPass(userid string, itemid int64) string {
	return filepath.Join(userFolderPass(userid), strconv.FormatInt(itemid, 10))
}

func NewFileStruct() *File {
	return &File{Body: make([]byte, 0)}
}

func deleteAllFilesAllUsers() {
	err := os.RemoveAll(configs.ClientConfig.FileFolder)
	if err != nil {
		log.Println("can't delete Catalog directory:", err)
	}
}

func deleteFolderByItemID(userid string, itemid int64) error {
	err := os.RemoveAll(itemFolderPass(userid, itemid))
	log.Println(itemFolderPass(userid, itemid))
	if err != nil {
		return fmt.Errorf("can't delete folder:%v For itemID =%d", err, itemid)
	}
	return nil
}

func (file *File) PrepareFile(pass []byte) error {
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

func (file *File) SaveFileLocal(pass []byte) error {
	// file name decrypto
	flenamebytes, err := base64.StdEncoding.DecodeString(file.FileName)
	if err != nil {
		file.FileName = "file"
	} else {
		flenamebytes, err = FileDecrypt(flenamebytes, pass)
		if err != nil {
			file.FileName = "file"
		} else {
			file.FileName = string(flenamebytes)
		}
	}

	log.Println("starts SaveFileLocal, file name:", file.FileName)
	if file.FileID == 0 {
		return fmt.Errorf("fileID not ready, file not saved local, need run request for files")
	}

	// check body
	if len(file.Body) == 0 {
		return fmt.Errorf("empty file.Body in file %s", file.FileName)
	}

	// create path localstorage/userid/itemid
	log.Println("save file USERID:", file.UserID)
	path := itemFolderPass(file.UserID, file.ItemID)

	// create folder if not exists
	err = os.MkdirAll(path, os.ModePerm)
	if err != nil {
		return fmt.Errorf("can't create local directory %s, error:%v", path, err)
	}

	// decode file
	decodedfile, err := FileDecrypt(file.Body, pass)

	// write file
	path = filepath.Join(path, strconv.FormatInt(file.FileID, 10)+"-"+file.FileName)
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

func (file *File) DeleteFileLocal() {
	// deregister file from local mapa
	Catalog.mu.Lock()
	defer Catalog.mu.Unlock()

	// dereg file from item card in catalog
	itm, ok := Catalog.Items[file.ItemID]
	if ok {
		for i, fleID := range itm.FileIDs {
			if fleID == file.FileID {
				newFileIDs := itm.FileIDs[:i]
				if i != len(itm.FileIDs)-1 {
					itm.FileIDs = append(newFileIDs, itm.FileIDs[i+1:]...)
				}
			}
		}
		Catalog.Items[file.ItemID] = itm
	}

	fle, ok := Catalog.Files[file.ItemID]
	if !ok {
		log.Printf("trying delete unregistered file. Address unknown. ItemID = %d, FileID = %d\n", file.ItemID, file.FileID)
		return
	}

	// dereg from file catalog
	delete(Catalog.Files, file.FileID)

	deleteFileFromLocalStorageByAddress(fle.Address)
}

func deleteFileFromLocalStorageByAddress(path string) {

	err := os.Remove(path)
	if err != nil {
		log.Printf("can't delete file:%v File address =%d\n", err, path)
	}
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
