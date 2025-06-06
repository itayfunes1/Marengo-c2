package server

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"server-tcp-go/src/handler/message"
	"server-tcp-go/src/work"
	"strings"
)

var messageMasterChan chan Msg

func handleMasterMess() {
	log.Println("Init handle master mess queue")
	for msg := range messageMasterChan {
		// log.Println("handle master message from: ", msg.from)
		mess := &message.Message{}
		err := mess.Decode(msg.payload)
		if err != nil {
			log.Println("cannot decode this: ", string(msg.payload))
			log.Println("cannot decode master message, err: ", err)
			continue
		}

		switch mess.Type {
		case message.MessageTypeGetListClient:
			go work.ListClientResp(string(mess.Data))
		case message.MessageTypeCheckInterval:
			continue
		case message.MessageTypeCreateBridgeResp:
			go work.ReceiveInitBridge(string(mess.Data))
		case message.MessageTypeSendKey:
			go work.ReceiveKey(mess.Data)
		case message.MessageTypeClientJoined:
			fmt.Printf("🟢  New agent connected: %s\n", mess.ClientAddress)
		default:
			log.Println("message type not found")
		}

	}
}

func listenToMaster(conn net.Conn) {
	defer conn.Close()
	masterReader := bufio.NewReader(conn)
	for {
		bytes, err := masterReader.ReadBytes(byte('\n'))
		if err != nil {
			log.Printf("Error reading from connection: %v\n", err)
			return
		}
		if len(bytes) == 1 && bytes[0] == 10 {
			continue
		}
		bytes = []byte(strings.TrimSuffix(string(bytes), "\n"))
		messageMasterChan <- Msg{
			from:    conn.RemoteAddr().String(),
			payload: bytes,
			conn:    conn,
		}
	}
}
