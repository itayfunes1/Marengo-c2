package message

type MessageType int

const (
	MessageTypeInit         MessageType = 1
	MessageTypeCommand      MessageType = 2
	MessageTypeDelete       MessageType = 3
	MessageTypeInitServer   MessageType = 4
	MessageTypeFileData     MessageType = 5
	MessageTypeCreateBridge MessageType = 6
	MessageTypeInitClient   MessageType = 7
	// MessageTypeInitServer   MessageType = 8
	MessageTypeInitBridge       MessageType = 9
	MessageTypeDeleteClient     MessageType = 10
	MessageTypeDeleteServer     MessageType = 11
	MessageTypeDeleteBridge     MessageType = 12
	MessageTypeCreateBridgeResp MessageType = 13
	MessageTypeCheckInterval    MessageType = 14
	MessageTypeGetListClient    MessageType = 15
	MessageTypeSendKey          MessageType = 16
	MessageTypeCloseSession     MessageType = 17
	MessageTypeClientJoined     MessageType = 18
)
