package work

import (
	"bufio"
	"log"
	"net"
	"server-tcp-go/src/handler/message"
	"strings"
)

func listenToBridge(conn net.Conn) {
	defer conn.Close()
	bridgeReader := bufio.NewReader(conn)
	for {
		bytes, err := bridgeReader.ReadBytes(byte('\n'))
		if err != nil {
			log.Printf("Error reading from connection: %v\n", err)
			return
		}

		if len(bytes) == 1 && bytes[0] == 10 {
			continue
		}
		log.Println("\nReceive from bridge data length: ", len(string(bytes)))
		bytes = []byte(strings.TrimSuffix(string(bytes), "\n"))

		go func(data []byte) {
			// data := bytes
			mess := &message.Message{}
			err = mess.Decode(data)
			if err != nil {
				log.Printf("Cannot decode message: %v", err)
				return
				// ReceiveCommandResp(data)
				// return
			}

			switch mess.Type {
			case message.MessageTypeCreateBridgeResp:
				go ReceiveInitBridge(string(mess.Data))
			case message.MessageTypeCommand:
				go ReceiveCommandResp(mess.Data, mess.ClientAddress)
			}
		}(bytes)

	}
}
