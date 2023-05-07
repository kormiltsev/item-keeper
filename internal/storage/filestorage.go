package storage

import (
	"bufio"
	"crypto/md5"
	"fmt"
	"log"
	"os"
	"path/filepath"

	configs "github.com/kormiltsev/item-keeper/internal/configs"
)

type FileStorageFile struct {
	ID          string
	UserID      string
	ItemID      string
	Title       string
	FileAddress string
	Avatar      []byte
	Data        *[]byte
}

func NewFileToStorage() *FileStorageFile {
	return &FileStorageFile{
		Avatar: make([]byte, 0),
	}
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

func (fsf *FileStorageFile) SaveNewFile() {

	path := filepath.Join(configs.ServiceConfig.FileServerAddress, fsf.UserID)

	// create random file name
	hash := md5.New()
	b := []byte(fsf.Title)
	fsf.ID = fmt.Sprintf("%x", hash.Sum(b))
	path = filepath.Join(path, fsf.ID)
	// path = filepath.Join(path, fsf.Title)

	log.Println("path1:", path)

	err := os.MkdirAll(path, os.ModePerm)
	if err != nil {
		log.Println("file storage can't create directory wile creating:", err)
		return
	}

	// path := configs.ServiceConfig.FileServerAddress + item.UserID + item.Title

	// create random file name
	// hash := md5.New()
	b = []byte(fsf.UserID + fsf.Title)
	path = filepath.Join(path, fmt.Sprintf("%x", hash.Sum(b)))

	log.Println("path2:", path)

	err = os.WriteFile(path, *fsf.Data, 0644)
	if err != nil {
		log.Println("file storage error, check FILESERVERADDRESS available:", err)
	}
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
