package work

import "log"

func ReceiveKey(key []byte) {
	log.Println("Receive key from server")
	if key == nil || len(key) < 32 {
		log.Println("Key invalid, using default key")
		clientKey = []byte(DefaultKey)
	} else {
		clientKey = key
	}
}
