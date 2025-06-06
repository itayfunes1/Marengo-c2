package client

import "fmt"

// This file for client - c2 server interface

type ClientMessage struct {
	TypeLength uint16 // 3 bits for type, 5 bits for length of message
	Data       []byte //
	FileId     uint32
	Sequence   uint32
}

type ClientMessageType uint8

const (
	INVALID_TYPE                 ClientMessageType = 0
	CLIENT_MESSAGE_TYPE_FILEDATA ClientMessageType = 1
	CLIENT_MESSAGE_TYPE_FILENAME ClientMessageType = 2
	CLIENT_MESSAGE_TYPE_COMMAND  ClientMessageType = 3
)

func ParseClientMessageType(data uint16) (ClientMessageType, error) {
	// Get first 3 bit of uint16
	t := data >> 13
	if t > 8 {
		return INVALID_TYPE, fmt.Errorf("invalid type")
	}

	return ClientMessageType(t), nil
}

func Encode2TypeLength(t ClientMessageType, length uint16) (uint16, error) {
	if length > (1 << 13) {
		return 0, fmt.Errorf("over length")
	}
	return (uint16(t) << 13) | length, nil
}

func (cmsg *ClientMessage) SerializeClientMsg() ([]byte, error) {
	data := []byte{}
	if cmsg.TypeLength == 0 || cmsg.Data == nil {
		return nil, fmt.Errorf("invalid message")
	}
	data = append(data, byte(cmsg.TypeLength>>8))
	data = append(data, byte(cmsg.TypeLength&0xFF))
	t, err := ParseClientMessageType(cmsg.TypeLength)
	if err != nil {
		return nil, err
	}
	if t == CLIENT_MESSAGE_TYPE_FILEDATA || t == CLIENT_MESSAGE_TYPE_FILENAME {
		data = append(data, byte(cmsg.FileId>>24))
		data = append(data, byte(cmsg.FileId>>16)&0xFF)
		data = append(data, byte(cmsg.FileId>>8)&0xFF)
		data = append(data, byte(cmsg.FileId)&0xFF)
	}
	if t == CLIENT_MESSAGE_TYPE_FILEDATA {
		data = append(data, byte(cmsg.Sequence>>24))
		data = append(data, byte(cmsg.Sequence>>16)&0xFF)
		data = append(data, byte(cmsg.Sequence>>8)&0xFF)
		data = append(data, byte(cmsg.Sequence)&0xFF)
	}
	data = append(data, cmsg.Data...)

	return data, nil
}

func DeserializeClientMsg(data []byte) (*ClientMessage, error) {
	var cmsg *ClientMessage
	if len(data) < 2 {
		return nil, fmt.Errorf("invalid data")
	}

	cmsg = &ClientMessage{}
	// concate 2 first bytes into uint16 TypeLength
	cmsg.TypeLength = uint16(data[0])<<8 | uint16(data[1])
	if cmsg.TypeLength == 0 {
		return nil, fmt.Errorf("invalid type length")
	}

	t, err := ParseClientMessageType(cmsg.TypeLength)
	if err != nil {
		return nil, err
	}

	length := uint16(data[1]) | (uint16(data[0]&0b11111) << 8)

	data = data[2:]
	if t == CLIENT_MESSAGE_TYPE_FILEDATA || t == CLIENT_MESSAGE_TYPE_FILENAME {
		if len(data) < 4 {
			return nil, fmt.Errorf("invalid data id")
		}
		cmsg.FileId = uint32(data[0])<<24 | uint32(data[1])<<16 | uint32(data[2])<<8 | uint32(data[3])
		data = data[4:]
	}
	if t == CLIENT_MESSAGE_TYPE_FILEDATA {
		if len(data) < 4 {
			return nil, fmt.Errorf("invalid data id")
		}
		cmsg.Sequence = uint32(data[0])<<24 | uint32(data[1])<<16 | uint32(data[2])<<8 | uint32(data[3])
		data = data[4:]
	}
	if uint16(len(data)) != length {
		return nil, fmt.Errorf("invalid data length")
	}
	cmsg.Data = data[:]

	return cmsg, nil
}

func NewClientMsg(t ClientMessageType, data []byte, fileId uint32) (*ClientMessage, error) {
	var cmsg *ClientMessage
	var err error
	cmsg = &ClientMessage{}
	length := uint16(len(data))
	cmsg.TypeLength, err = Encode2TypeLength(t, length)
	if err != nil {
		return nil, err
	}
	if t == CLIENT_MESSAGE_TYPE_FILEDATA || t == CLIENT_MESSAGE_TYPE_FILENAME {
		cmsg.FileId = fileId
	}
	cmsg.Data = data
	return cmsg, nil
}
