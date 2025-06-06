package work

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	mrand "math/rand"
	"net"
	"os"
	"server-tcp-go/src/client"
	"server-tcp-go/src/handler/message"
	"server-tcp-go/src/utils"
	"strings"
	"sync"
	"time"
)

type File struct {
	ID        int       `json:"id"`
	Name      string    `json:"name"`
	Size      int       `json:"size"`
	CreatedAt time.Time `json:"created_at"`
}

func (f *File) SerializeFile() ([]byte, error) {
	delim := "\n"
	b, err := json.Marshal(f)
	if err != nil {
		return nil, err
	}
	return append(b, delim...), nil
}

func (f *File) DeserializeFile(data []byte) error {
	b := []byte(strings.TrimSpace(string(data)))
	return json.Unmarshal(b, f)
}

type FileContent struct {
	Mutex     sync.Mutex
	Id        uint32
	Name      string
	Size      uint32
	TotalSize uint32
	Data      map[uint32][]byte
	DataChan  chan Content
	Done      chan bool
}

type Content struct {
	Data     []byte
	Sequence uint32
}

func saveFile(data string) {
	// f, err := os.Create(KeyPath + "key")

	// if err != nil {
	// 	logging.Badln(err)
	// }

	// defer f.Close()

	// _, err2 := f.WriteString(data)

	// if err2 != nil {
	// 	logging.Badln(err2)
	// }

	// logging.Goodln("Saved unique key!")
}

func sendFile(conn net.Conn, key, addr, command string) error {
	strs := strings.Split(command, " ")
	if len(strs) < 3 {
		return fmt.Errorf("invalid command")
	}
	filename := strs[2]
	// Check file exists
	f, err := os.Open(filename)
	if err != nil {
		return fmt.Errorf("error opening file %v", err)
	}
	defer f.Close()

	// Send file header
	fileStat, err := os.Stat(filename)
	if err != nil {
		return fmt.Errorf("error getting file info %v", err)
	}
	//fmt.Println("File Name:", fileStat.Name())        // Base name of the file
	//fmt.Println("Size:", fileStat.Size())             // Length in bytes for regular files
	//fmt.Println("Permissions:", fileStat.Mode())      // File mode bits
	//fmt.Println("Last Modified:", fileStat.ModTime()) // Last modification time
	//fmt.Println("Is Directory: ", fileStat.IsDir())   // Abbreviation for Mode().IsDir()
	if fileStat.IsDir() {
		return fmt.Errorf("no support to get directory") // Needs a function for getting directory !!
	}
	fileId := uint32(mrand.Int31())
	// Send file connent
	cmsg, err := client.NewClientMsg(client.CLIENT_MESSAGE_TYPE_FILENAME, []byte(fmt.Sprintf("%s;%d", fileStat.Name(), fileStat.Size())), fileId)

	if err != nil {
		return fmt.Errorf("error creating client message %v", err)
	}
	byteData, err := cmsg.SerializeClientMsg()

	if err != nil {
		return fmt.Errorf("error serializing client message %v", err)
	}

	encryptedCommand, err := utils.Encrypt([]byte(key), byteData)
	if err != nil {
		return fmt.Errorf("error encrypting command: %v", err)
	}

	encodedCommand := base64.StdEncoding.EncodeToString(encryptedCommand)
	msg := message.NewMsg(message.MessageTypeCommand, []byte(encodedCommand+"\n"+"-"+addr), utils.GetLocalIP())

	byteCommand, err := msg.Encode()
	if err != nil {
		return fmt.Errorf("error marshal command to client: %v", err)
	}
	byteCommand = append(byteCommand, 0x00)

	_, err = conn.Write(byteCommand)
	if err != nil {
		return fmt.Errorf("error sending command to client: %v", err)
	}

	time.Sleep(1 * time.Second)
	var sizeSend int64 = 0
	fileBuffer := make([]byte, BUFFER_SIZE)
	var sqn uint32 = 0
	var errStr string
	defer func() {
		if errStr != "" {
			errStr += "! Stop sendding!"
			// jobSend := models.NewJob(models.JobType_SendServer, []byte(errStr), "", "")
			// Dispatch(jobSend)
		}
	}()
	for sizeSend < fileStat.Size() {
		n, errSend := f.ReadAt(fileBuffer, sizeSend)

		sizeSend += int64(n)
		bData := fileBuffer[:n]
		cmsg, err := client.NewClientMsg(client.CLIENT_MESSAGE_TYPE_FILEDATA, bData, fileId)
		if err != nil {
			errStr = "Error creating client message"
			log.Println(errStr, err)
			return err
		}
		cmsg.Sequence = sqn
		sqn = sqn + 1

		byteData, err := cmsg.SerializeClientMsg()

		if err != nil {
			errStr = "Error serializing client message"
			log.Println(errStr, err)
			return err
		}

		// Encrypt the command output
		encryptedOutput, err := utils.Encrypt([]byte(key), byteData)
		if err != nil {
			errStr = "Error encrypting output"
			log.Println(errStr, err)
			return err
		}
		// Encode the command output to base64
		outputBase64 := base64.StdEncoding.EncodeToString(encryptedOutput)

		msg := message.NewMsg(message.MessageTypeFileData, []byte(outputBase64+"\n"+"-"+addr), utils.GetLocalIP())

		byteCommand, _ := msg.Encode()
		byteCommand = append(byteCommand, 0x00)
		conn.Write(byteCommand)
		if errSend == io.EOF {
			break
		}
	}
	return nil
}
