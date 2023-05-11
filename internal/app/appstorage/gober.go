package appstorage

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/sha256"
	"encoding/gob"
	"encoding/hex"
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
