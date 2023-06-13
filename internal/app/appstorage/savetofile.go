package appstorage

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/sha256"
	"encoding/gob"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"

	"golang.org/x/crypto/scrypt"
)

func (op *Operator) SaveEncryptedCatalog(password string) error {

	// create folder if not exists
	err := os.MkdirAll(filepath.Dir(localcatalogaddress), os.ModePerm)
	if err != nil {
		return fmt.Errorf("can't create directory for catalog %s, error:%v", localcatalogaddress, err)
	}

	os.Remove(localcatalogaddress)

	// Serialize the catalog structure
	data, err := op.serialize()
	if err != nil {
		return err
	}

	// Generate the encryption key and IV from the password
	seawater := sha256.Sum256([]byte(clientUniqueID))
	key, iv := deriveKeyAndIV(seawater[:], []byte(password))

	// Create the AES cipher block
	block, err := aes.NewCipher(key)
	if err != nil {
		return err
	}

	// Encrypt the serialized data
	encryptedData := make([]byte, len(data))
	cipher.NewCFBEncrypter(block, iv).XORKeyStream(encryptedData, data)

	// Save the encrypted data to a file
	if err := ioutil.WriteFile(localcatalogaddress, encryptedData, 0644); err != nil {
		return err
	}

	return nil
}

type Listik struct {
	LastUpdate int64
	UserID     string
	Items      map[int64]*Item
	Files      map[int64]*File
	// ItemIDs    []int64
	// Items      []Item // key = itemid
	// FileIDs    []int64
	// Files      []File // key = fileid
}

func (op *Operator) serialize() ([]byte, error) {
	// Create a buffer to hold the serialized data
	var buffer bytes.Buffer

	// Create an encoder for writing to the buffer
	encoder := gob.NewEncoder(&buffer)

	op.Mapa.mu.Lock()
	defer op.Mapa.mu.Unlock()

	tosave := Listik{
		LastUpdate: op.Mapa.LastUpdate,
		UserID:     op.Mapa.UserID,
		Items:      op.Mapa.Items,
		Files:      op.Mapa.Files,
		// ItemIDs:    make([]int64, 0),
		// Items:      make([]Item, 0),
		// FileIDs:    make([]int64, 0),
		// Files:      make([]File, 0),
	}

	// for k, v := range op.Mapa.Items {
	// 	tosave.ItemIDs = append(tosave.ItemIDs, k)
	// 	tosave.Items = append(tosave.Items, *v)
	// }
	// for k, v := range op.Mapa.Files {
	// 	tosave.FileIDs = append(tosave.FileIDs, k)
	// 	tosave.Files = append(tosave.Files, *v)
	// }

	// Encode the catalog structure
	if err := encoder.Encode(tosave); err != nil {
		return nil, err
	}

	return buffer.Bytes(), nil
}

func deriveKeyAndIV(salts, password []byte) ([]byte, []byte) {
	salt := make([]byte, 16)
	copy(salt, salts[:16])

	// Derive the key from the password and salt using Scrypt
	key, err := scrypt.Key(password, salt, 16384, 8, 1, 32)
	if err != nil {
		log.Fatal(err)
	}

	iv := make([]byte, aes.BlockSize)
	if len(password) < aes.BlockSize {
		copy(iv, password)
	} else {
		copy(iv, password[:aes.BlockSize])
	}

	return key, iv
}

func ReadDecryptedCatalog(password string) (string, error) {

	// create folder if not exists
	err := os.MkdirAll(filepath.Dir(localcatalogaddress), os.ModePerm)
	if err != nil {
		return "", fmt.Errorf("can't create directory for catalog %s, error:%v", localcatalogaddress, err)
	}

	// Read the encrypted data from the file
	encryptedData, err := ioutil.ReadFile(localcatalogaddress)
	if err != nil {
		return "", err
	}

	// Generate the encryption key and IV from the password
	seawater := sha256.Sum256([]byte(clientUniqueID))
	key, iv := deriveKeyAndIV(seawater[:], []byte(password))

	// Create the AES cipher block
	block, err := aes.NewCipher(key)
	if err != nil {
		log.Println("NewCipher", err)
		return "", err
	}

	// Decrypt the encrypted data
	decryptedData := make([]byte, len(encryptedData))
	cipher.NewCFBDecrypter(block, iv).XORKeyStream(decryptedData, encryptedData)

	// Deserialize the decrypted data into the catalog structure
	return deserialize(decryptedData)
}

func deserialize(data []byte) (string, error) {
	// Create a buffer with the serialized data
	buffer := bytes.NewBuffer(data)

	// Create a decoder for reading from the buffer
	decoder := gob.NewDecoder(buffer)

	// Decode the catalog structure
	catalogue := Listik{
		Items: map[int64]*Item{},
		Files: map[int64]*File{},
		// ItemIDs: make([]int64, 0),
		// Items:   make([]Item, 0),
		// FileIDs: make([]int64, 0),
		// Files:   make([]File, 0),
	}

	err := decoder.Decode(&catalogue)
	if err != nil {
		return "", err
	}

	Catalog.LastUpdate = catalogue.LastUpdate
	Catalog.UserID = catalogue.UserID

	Catalog.Items = catalogue.Items
	Catalog.Files = catalogue.Files

	// for i, v := range catalogue.ItemIDs {
	// 	Catalog.Items[v] = &catalogue.Items[i]
	// }
	// for i, v := range catalogue.FileIDs {
	// 	Catalog.Files[v] = &catalogue.Files[i]
	// }

	return Catalog.UserID, nil
}
