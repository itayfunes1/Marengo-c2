package server

import (
	"log"
	"net"
	"server-tcp-go/src/config"
	"server-tcp-go/src/handler/message"
	"server-tcp-go/src/utils"
	"server-tcp-go/src/work"

	"github.com/joho/godotenv"
)

func Init(fileName ...string) {
	initHandleChan()

	err := godotenv.Load(fileName...)
	if err != nil {
		log.Fatal("❌ Error loading .env file:", err)
		panic("error loading .env file")
	}

	meConf := config.NewServerEndpointConf(
		"master endpoint conf",
		utils.GetEnv("MasterIp", "127.0.0.1"),
		utils.GetEnv("MasterPort", "2605"),
	)

	config.NewConfig(meConf, utils.GetEnv("BridgePort", "15319"))

	// 🔐 Wait for master connection before starting menu
	sendInitToMaster()

	go work.StartUp()
}

func initHandleChan() {
	messageMasterChan = make(chan Msg)
	go handleMasterMess()
}

func sendInitToMaster() {
	masterHost := config.GetMasterEndPoint()
	log.Println("[🔌] Connecting to master at:", masterHost)

	conn, err := net.Dial("tcp", masterHost)
	if err != nil {
		log.Fatal("❌ Cannot connect to master:", err)
	}

	log.Println("[✅] Connected to master.")

	keyVerify := utils.GetEnv("Key", "V4lqNUFdqXleFWg0")
	mess := message.NewMsg(message.MessageTypeInitServer, []byte(keyVerify), utils.GetLocalIP())
	messByte, err := mess.Encode()
	if err != nil {
		log.Fatal("❌ Failed to encode init message:", err)
	}

	_, err = conn.Write(messByte)
	if err != nil {
		log.Fatal("❌ Failed to send init message:", err)
	}

	work.MasterConnection = conn
	go listenToMaster(conn)

	log.Println("[🚀] Master connection established and initialized.")
}
