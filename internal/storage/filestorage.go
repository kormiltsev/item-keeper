package storage

import (
	"bufio"
	"crypto/sha1"
	"encoding/hex"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"time"

	configs "github.com/kormiltsev/item-keeper/internal/configs"
)

type FileStorageFile struct {
	ID          string
	UserID      string
	ItemID      string
	Title       string
	FileAddress string
	Data        *[]byte
}

func NewFileToStorage() *FileStorageFile {
	return &FileStorageFile{}
}

func FileStoragePing(fileaddress string) error {

	err := os.MkdirAll(fileaddress, os.ModePerm)
	if err != nil {
		log.Println("file storage can't create directory, check FILESERVERADDRESS available:", err)
		return err
	}

	err = os.WriteFile(fileaddress+"/testfilename", []byte("0"), 0644)
	if err != nil {
		log.Println("file storage can't create file, check FILESERVERADDRESS available:", err)
		return err
	}

	err = os.Remove(fileaddress + "/testfilename")
	if err != nil {
		log.Println("file storage error with removing, check FILESERVERADDRESS available:", err)
		return err
	}
	return nil
}

// remove directory with all files and subdirectories in /<UserID>/<ItemID>
func (fsf *FileStorageFile) DeleteOldFilesByItemID() {
	path := filepath.Join(configs.ServiceConfig.FileServerAddress, fsf.UserID)
	path = filepath.Join(path, fsf.ItemID)

	// remove directory with all files and subdirectories
	err := os.RemoveAll(path)
	if err != nil {
		log.Printf("can't delete files for item ID=%s, error:%v", fsf.ItemID, err)
	}
}

func (fsf *FileStorageFile) SaveNewFile() {

	// create uniq id for image
	h := sha1.New()
	h.Write([]byte(fsf.ItemID + strconv.FormatInt(time.Now().UnixNano(), 16)))
	sha1_hash := hex.EncodeToString(h.Sum(nil))
	fsf.ID = sha1_hash

	path := filepath.Join(configs.ServiceConfig.FileServerAddress, fsf.UserID)
	path = filepath.Join(path, fsf.ItemID)

	log.Println("path1:", path)

	err := os.MkdirAll(path, os.ModePerm)
	if err != nil {
		log.Println("file storage can't create directory wile creating:", err)
		return
	}

	path = filepath.Join(path, fsf.ID)

	err = os.WriteFile(path, *fsf.Data, 0644)
	if err != nil {
		log.Println("file storage error, check FILESERVERADDRESS available:", err)
	}
	chanIDofNewUploadedFiles() <- Item{ID: fsf.ItemID, ImageLink: []string{path}}
}

func GetTitle(filelink string) ([]byte, error) {
	file, err := os.Open(filelink)

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
