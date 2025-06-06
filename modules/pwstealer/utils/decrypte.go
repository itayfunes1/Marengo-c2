package utils

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"errors"
	"syscall"
	"unsafe"
)

var (
	dllcrypt32  = syscall.NewLazyDLL("Crypt32.dll")
	dllkernel32 = syscall.NewLazyDLL("Kernel32.dll")

	procEncryptData = dllcrypt32.NewProc("CryptProtectData")
	procDecryptData = dllcrypt32.NewProc("CryptUnprotectData")
	procLocalFree   = dllkernel32.NewProc("LocalFree")
)

func DecryptWinApi(data []byte) ([]byte, error) {
	var outblob DATA_BLOB
	r, _, err := procDecryptData.Call(uintptr(unsafe.Pointer(NewBlob(data))), 0, 0, 0, 0, 0x1, uintptr(unsafe.Pointer(&outblob)))
	if r == 0 {
		return nil, err
	}
	defer procLocalFree.Call(uintptr(unsafe.Pointer(outblob.pbData)))
	return outblob.ToByteArray(), nil
}

type DATA_BLOB struct {
	cbData uint32
	pbData *byte
}

func NewBlob(d []byte) *DATA_BLOB {
	if len(d) == 0 {
		return &DATA_BLOB{}
	}
	return &DATA_BLOB{
		pbData: &d[0],
		cbData: uint32(len(d)),
	}
}

func (b *DATA_BLOB) ToByteArray() []byte {
	d := make([]byte, b.cbData)
	copy(d, (*[1 << 30]byte)(unsafe.Pointer(b.pbData))[:])
	return d
}

func decrypt(encrypted []byte, secretedKey []byte) ([]byte, error) {
	var decrypt func(encrypted, password []byte) ([]byte, error)
	switch {
	case bytes.HasPrefix(encrypted, prefixDPAPI[:]):
		// present before Chrome v80 on Windows
		decrypt = func(encrypted, _ []byte) ([]byte, error) {
			return DecryptWinApi(encrypted)
		}
	case bytes.HasPrefix(encrypted, []byte(`v10`)):
		fallthrough
	default:
		decrypt = decryptAES256GCM
	}

	decrypted, err := decrypt(encrypted, secretedKey)
	if err != nil {
		return nil, err
	}

	return decrypted, nil
}

var (
	fallbackPasswordLinux = [...]byte{'p', 'e', 'a', 'n', 'u', 't', 's'}
	fallbackPasswordMacOS = [...]byte{'m', 'o', 'c', 'k', '_', 'p', 'a', 's', 's', 'w', 'o', 'r', 'd'}                     // mock keychain
	prefixDPAPI           = [...]byte{1, 0, 0, 0, 208, 140, 157, 223, 1, 21, 209, 17, 140, 122, 0, 192, 79, 194, 151, 235} // 0x01000000D08C9DDF0115D1118C7A00C04FC297EB
)

func decryptAES256GCM(encrypted, password []byte) ([]byte, error) {
	// https://stackoverflow.com/a/60423699

	if len(encrypted) < 3+12+16 {
		return nil, errors.New(`encrypted value too short`)
	}

	/* encoded value consists of: {
		"v10" (3 bytes)
		nonce (12 bytes)
		ciphertext (variable size)
		tag (16 bytes)
	}
	*/
	nonce := encrypted[3 : 3+12]
	ciphertextWithTag := encrypted[3+12:]

	block, err := aes.NewCipher(password)
	if err != nil {
		return nil, err
	}

	// default size for nonce and tag match
	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	plaintext, err := aesgcm.Open(nil, nonce, ciphertextWithTag, nil)
	if err != nil {
		return nil, err
	}

	return plaintext, nil
}
