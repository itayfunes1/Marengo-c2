package work

import (
	"encoding/base64"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"server-tcp-go/src/client"
	"server-tcp-go/src/utils"
	"strconv"
	"strings"
	"sync"
)

var knownAgents = make(map[string]bool)

func ReceiveInitBridge(cause string) {
	MasterCreateBridgeResp <- cause
}

func ReceiveCommandResp(data []byte, addr string) {

	log.Println("Receive Command Response")
	// Handle key here
	key := string(clientKey)
	if len(key) < 32 {
		fmt.Println("key invalid")
	}
	log.Println("Receive Command Response data length: ", len(data))
	// commandBase64 := strings.TrimSpace(string(data))
	commandBase64 := string(data)

	// if len(commandBase64) != len(string(data)) {
	// 	fmt.Println("have trim space")
	// 	fmt.Println(commandBase64)
	// }

	// Decode the command from base64
	encryptedCommand, err := base64.StdEncoding.DecodeString(commandBase64)
	if err != nil {
		log.Printf("Error decoding command: %v %v", err, commandBase64)
		return
	}

	command, err := utils.MarengoDecrypt([]byte(key), encryptedCommand)
	if err != nil {
		log.Printf("Error decrypting command: %v", err)
		return
	}

	// Deserialize client message
	clientMsg, err := client.DeserializeClientMsg(command)
	if err != nil || clientMsg == nil {
		log.Printf("Error deserializing command: %v", err)
		return
	}
	// ✅ New agent connection logging
	clientAddr := addr
	if clientAddr == "" {
		clientAddr = "unknown"
	}
	if _, exists := knownAgents[clientAddr]; !exists {
		knownAgents[clientAddr] = true
	}
	t, err := client.ParseClientMessageType(clientMsg.TypeLength)
	if err != nil {
		log.Printf("Error parsing command: %v", err)
		return
	}

	if t == client.CLIENT_MESSAGE_TYPE_FILENAME {
		go handleReceiveFileInfoClient(*clientMsg)
	}
	if t == client.CLIENT_MESSAGE_TYPE_FILEDATA {
		go handleReceiveFileDataClient(*clientMsg)
	}
	// if strings.HasPrefix(string(clientMsg.Data), "[camera]") {
	// 	rspChan := queueHandle.Get(msg.ClientAddress)
	// 	if rspChan != nil {
	// 		// fmt.Println("Found camera channel ")
	// 		rspChan <- string(clientMsg.Data)
	// 		return
	// 	}
	// 	//fmt.Println("Not found camera channel ", msg.ClientAddress)
	// 	return
	// }
	// fmt.Println("\n")
	if t == client.CLIENT_MESSAGE_TYPE_COMMAND {
		fmt.Printf("\n")
		fmt.Println(string(clientMsg.Data))
	}
}

func handleReceiveFileDataClient(msg client.ClientMessage) {
	value, ok := FileMap.Load(msg.FileId)
	if !ok {
		log.Println("New file content ")
		return
	}
	fileContentLoaded := value.(*(FileContent))
	fileContentLoaded.Mutex.Lock()
	fileContentLoaded.DataChan <- Content{
		Sequence: msg.Sequence,
		Data:     msg.Data,
	}
	fileContentLoaded.Mutex.Unlock()
}

func handleReceiveFileInfoClient(msg client.ClientMessage) {
	folderName := ""
	strs := strings.Split(string(msg.Data), ";")
	if len(strs) < 2 {
		log.Println("Invalid file information")
		return
	}
	log.Println("Receive file: ", strs[0], "with ", strs[1])
	if len(strs) >= 3 {
		folderName = strs[2]
	}
	filesize, err := strconv.ParseInt(strs[1], 10, 64)
	if err != nil {
		log.Println("Invalid file size: ", err)
		return
	}
	fileContent := FileContent{
		Mutex:     sync.Mutex{},
		TotalSize: uint32(filesize),
		Name:      strs[0],
		Id:        msg.FileId,
		DataChan:  make(chan Content, 100000),
		Data:      make(map[uint32][]byte),
		Done:      make(chan bool, 1),
	}
	go func() {
		isDone := <-fileContent.Done
		// fmt.Println("Receive noti")
		if isDone {
			log.Println("Fully downloaded")

			if _, err := os.Stat(savePath); errors.Is(err, os.ErrNotExist) {
				err := os.Mkdir(savePath, os.ModePerm)
				if err != nil {
					log.Println("Cannot create file", err)
				}
			}
			var f *os.File
			savedFileName := ""
			if folderName != "" {
				savedFileName = savePath + "/" + folderName
				dir := filepath.Dir(savedFileName)
				err := os.MkdirAll(dir, os.ModePerm)
				if err != nil {
					log.Println("Cannot create folder", err)
					return
				}
				f, err = os.Create(savedFileName)
				if err != nil {
					log.Println("Cannot create forder file", err)
					return

				}

			} else {
				savedFileName = savePath + "/" + fileContent.Name
				f, err = os.Create(savedFileName)
				if err != nil {
					log.Println("Cannot create forder file", err)
					return
				}
			}

			log.Println("Saved file to", savedFileName)
			for i := 0; i < len(fileContent.Data); i++ {
				_, err := f.Write(fileContent.Data[uint32(i)])
				if err != nil {
					log.Println("Write file fail: ", err)
				}
				f.Sync()
			}

			f.Close()

		} else {
			log.Println("Get file failed")
		}
		for i := 0; i < len(fileContent.Data); i++ {
			delete(fileContent.Data, uint32(i))
		}
		fileContent.Data = make(map[uint32][]byte)
		FileMap.Delete(fileContent.Id)
	}()
	_, ok := FileMap.LoadOrStore(msg.FileId, &fileContent)
	if !ok {
		log.Println("New file content ")
		receiverFileJob <- msg.FileId
		return
	}
}
