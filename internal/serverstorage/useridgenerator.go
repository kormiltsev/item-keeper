package serverstorage

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/sha256"
	"encoding/hex"
	"strconv"
)

// variables is for generate uniq user ID
var (
	useridgen = []byte("usersids")
	key       [32]byte
	block     cipher.Block
	aesgcm    cipher.AEAD
	nonce     = []byte("awsome_nonce")
	initDone  bool
)

// initialize is for start.
func initialize() error {
	if initDone {
		return nil
	}

	key = sha256.Sum256(useridgen)

	var err error
	block, err = aes.NewCipher(key[:])
	if err != nil {
		return err
	}

	aesgcm, err = cipher.NewGCM(block)
	if err != nil {
		return err
	}

	initDone = true
	return nil
}

// shifu encrypts User ID
func shifu(a int) (string, error) {
	err := initialize()
	if err != nil {
		return "", err
	}

	ciphertext := aesgcm.Seal(nil, nonce, []byte(strconv.Itoa(a)), nil)

	export := hex.EncodeToString(ciphertext)
	return export, nil
}
