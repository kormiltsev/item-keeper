package serverstorage

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"

	configs "github.com/kormiltsev/item-keeper/internal/configs"
)

var storageaddress = "./data/ServerStorage"

func fileUploadToFileStorage(file *File) {

	if configs.ServiceConfig.S3key != "" {
		err := uploadFileS3(file)
		if err != nil {
			log.Println("Can't save file S3:", err)
		}
		return
	}

	err := fileUploadToFileStorageLocal(file)
	if err != nil {
		log.Println("Can't save file local:", err)
	}
}

func fileUploadToFileStorageLocal(file *File) error {
	// create path localstorage/userid/itemid
	path := filepath.Join(storageaddress, file.UserID)
	path = filepath.Join(path, strconv.FormatInt(file.ItemID, 10))

	// create folder if not exists
	err := os.MkdirAll(path, os.ModePerm)
	if err != nil {
		return fmt.Errorf("can't create local directory %s, error:%v", path, err)
	}

	// write file
	path = filepath.Join(path, strconv.FormatInt(file.FileID, 10))
	err = os.WriteFile(path, file.Body, 0644)
	if err != nil {
		return fmt.Errorf("write file %s error:%v", path, err)
	}
	return nil
}

func fileDownloadFromStorage(file *File) ([]byte, error) {
	// storage switcher
	if configs.ServiceConfig.S3key != "" {
		return downloadS3(file)
	}

	// if local file server
	// create path localstorage/userid/itemid
	path := filepath.Join(storageaddress, file.UserID)
	path = filepath.Join(path, strconv.FormatInt(file.ItemID, 10))
	path = filepath.Join(path, strconv.FormatInt(file.FileID, 10))

	return readFileLocal(path)
}

func readFileLocal(fileaddress string) ([]byte, error) {
	file, err := os.Open(fileaddress)

	if err != nil {
		return nil, err
	}
	defer file.Close()

	stats, statsErr := file.Stat()
	if statsErr != nil {
		return nil, statsErr
	}

	var size int64 = stats.Size()
	bytes := make([]byte, size)

	bufr := bufio.NewReader(file)
	_, err = bufr.Read(bytes)

	return bytes, err
}

func deleteFilesByID(tostor *ToStorage) {
	// storage switcher
	if configs.ServiceConfig.S3key != "" {
		err := deleteS3(tostor.FilesNoBody)
		if err != nil {
			log.Println("S3 can't delete files")
		}
	}
	// if local file server
	for _, file := range tostor.FilesNoBody {
		err := deleteFileFromStorage(&file)
		if err != nil {
			log.Printf("can't delete file [%d] from server storage:%v", file.FileID, err)
		}
	}
}

func deleteFileFromStorage(file *File) error {
	// create path localstorage/userid/itemid

	path := filepath.Join(storageaddress, file.UserID)
	path = filepath.Join(path, strconv.FormatInt(file.ItemID, 10))
	path = filepath.Join(path, strconv.FormatInt(file.FileID, 10))

	err := os.Remove(path)
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("can't delete file:%v For itemID =%d", err, file.ItemID)
	}
	return nil
}
