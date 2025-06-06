package work

import (
	"log"
	"net"
	"sync"
)

const BUFFER_SIZE = 5000
const savePath = "./temp"
const DefaultKey = "11111111111111111111111111111111"

var MasterConnection net.Conn
var BridgeConnection net.Conn
var FileMap sync.Map
var receiverFileJob chan uint32

var MasterListClientResp chan string
var MasterCreateBridgeResp chan string
var BridgeCameraResp chan string

var connToMaster net.Conn
var clientKey []byte

func GetConnToMaster() net.Conn {
	log.Println("Get connection to master")
	return connToMaster
}

func UpdateConnToMaster(conn net.Conn) {
	log.Println("Update connection to master")
	connToMaster = conn
}
