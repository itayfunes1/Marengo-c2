package work

import (
	"encoding/base64"
	"fmt"
	"log"
)

func SaveKeytoFile() {
	log.Println("Saving client key to file...")
	log.Println("Listing clients...")
	clientList := getListClient()
	for i, c := range clientList {
		fmt.Printf("%d - %s", i+1, c)
	}
	// log.Println("Please Input Client:")
	// clientSelected, err := getInputString()
	// if err != nil {
	// 	fmt.Errorf("Invalid Input: %v", err)
	// 	return
	// }

	key := ""
	saveFile(base64.StdEncoding.EncodeToString([]byte(key)))
}
