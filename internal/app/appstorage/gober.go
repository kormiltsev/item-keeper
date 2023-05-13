package appstorage

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/sha256"
	"encoding/gob"
	"encoding/hex"
	"fmt"
)

func (item *Item) Encode(currentuserpassword string) (string, error) {

	var inputBuffer bytes.Buffer
	gob.NewEncoder(&inputBuffer).Encode(item)

	destBytes := inputBuffer.Bytes()

	tostor, err := shifu(currentuserpassword, destBytes)
	if err != nil {
		return "", err
	}

	return tostor, nil
}

func shifu(currentuserpassword string, data []byte) (string, error) {
	key := sha256.Sum256([]byte(currentuserpassword))

	block, err := aes.NewCipher(key[:])
	if err != nil {
		return "", err
	}
	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonce := []byte("awsome_nonce")

	ciphertext := aesgcm.Seal(nil, nonce, data, nil) //[]uint8

	return hex.EncodeToString(ciphertext), nil
}

func Decode(sourse string, currentuserpassword string) (*Item, error) {
	bts, err := deshifu(currentuserpassword, sourse)
	if err != nil {
		return nil, fmt.Errorf("can't decode string to bytes in gob decoder:%v", err)
	}

	buf := bytes.NewBuffer(bts)
	dec := gob.NewDecoder(buf)

	answer := Item{}

	if err := dec.Decode(&answer); err != nil {
		return nil, fmt.Errorf("can't decode bytes to Item in gob decoder:%v", err)
	}

	return &answer, nil
}

func deshifu(currentuserpassword, data string) ([]byte, error) {
	key := sha256.Sum256([]byte(currentuserpassword))

	block, err := aes.NewCipher(key[:])
	if err != nil {
		return nil, err
	}
	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	// TODO upgrade
	nonce := []byte("awsome_nonce")

	encrypted, err := hex.DecodeString(data)
	if err != nil {
		return nil, err
	}
	// decode
	decrypted, err := aesgcm.Open(nil, nonce, encrypted, nil)
	if err != nil {
		return nil, err
	}
	return decrypted, nil
}
