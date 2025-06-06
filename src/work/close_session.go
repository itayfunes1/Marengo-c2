package work

import (
	"server-tcp-go/src/handler/message"
	"server-tcp-go/src/utils"
)

func CloseSession() {
	mess := message.Message{
		Type:          message.MessageTypeCloseSession,
		Data:          []byte("0"),
		ClientAddress: utils.GetLocalIP(),
	}

	messByte, _ := mess.Encode()
	messByte = append(messByte, 0x00)
	BridgeConnection.Write(messByte)
	BridgeConnection.Close()
}
