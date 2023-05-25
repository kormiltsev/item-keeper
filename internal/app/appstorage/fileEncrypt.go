package appstorage

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"fmt"
	"io"
)

func FileEncrypt(plaintext []byte, keystring string) ([]byte, error) {

	// Key
	key := sha256.Sum256([]byte(keystring))

	block, err := aes.NewCipher(key[:])
	if err != nil {
		return nil, fmt.Errorf("NewCipher error:%v", err)
	}

	ciphertext := make([]byte, aes.BlockSize+len(plaintext))

	iv := ciphertext[:aes.BlockSize]

	// 16 random bytes
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return nil, fmt.Errorf("make random error:%v", err)
	}

	// Return an encrypted stream
	stream := cipher.NewCFBEncrypter(block, iv)

	// Encrypt bytes from plaintext to ciphertext
	stream.XORKeyStream(ciphertext[aes.BlockSize:], plaintext)

	return ciphertext, nil
}

func FileDecrypt(ciphertext []byte, keystring string) ([]byte, error) {

	// Key
	key := sha256.Sum256([]byte(keystring))

	block, err := aes.NewCipher(key[:])
	if err != nil {
		return nil, fmt.Errorf("AES error:%v", err)
	}

	// if the text is too small, then it is incorrect
	if len(ciphertext) < aes.BlockSize {
		return nil, fmt.Errorf("income []byte too short:%v", err)
	}

	iv := ciphertext[:aes.BlockSize]

	// Remove the IV from the ciphertext
	ciphertext = ciphertext[aes.BlockSize:]

	// Return a decrypted stream
	stream := cipher.NewCFBDecrypter(block, iv)

	// Decrypt bytes from ciphertext
	stream.XORKeyStream(ciphertext, ciphertext)

	return ciphertext, nil
}
