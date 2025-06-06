package utils

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io"
	"net"
)

// AESKeySize represents the AES key size in bytes
const AESKeySize = 32

func GenerateKey() ([]byte, error) {
	key := make([]byte, AESKeySize)
	_, err := rand.Read(key)
	if err != nil {
		return nil, fmt.Errorf("error generating key: %v", err)
	}
	return key, nil
}

func Encrypt(key, text []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("encryption error: %v", err)
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("encryption error: %v", err)
	}
	nonce := make([]byte, gcm.NonceSize())
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, fmt.Errorf("encryption error: %v", err)
	}
	ciphertext := gcm.Seal(nil, nonce, text, nil)
	ciphertext = append(nonce, ciphertext...)
	return ciphertext, nil
}

func Decrypt(key, ciphertext []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("decryption error: %v", err)
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("decryption error: %v", err)
	}
	nonceSize := gcm.NonceSize()
	if len(ciphertext) < nonceSize {
		return nil, fmt.Errorf("decryption error: invalid ciphertext")
	}
	nonce, ciphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, fmt.Errorf("decryption error: %v", err)
	}
	return plaintext, nil
}

func SendKey(conn net.Conn, key []byte) error {
	encodedKey := base64.StdEncoding.EncodeToString(key)
	_, err := fmt.Fprint(conn, encodedKey+"\n")
	if err != nil {
		return fmt.Errorf("error sending key to client: %v", err)
	}
	return nil
}

func MarengoDecrypt(key, ciphertext []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("decryption error: %v", err)
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("decryption error: %v", err)
	}
	nonceSize := gcm.NonceSize()
	if len(ciphertext) < nonceSize {
		return nil, fmt.Errorf("decryption error: invalid ciphertext")
	}
	nonce, ciphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, fmt.Errorf("decryption error: %v", err)
	}
	return plaintext, nil
}
