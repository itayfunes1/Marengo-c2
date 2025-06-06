package message

import (
	"encoding/json"
	"strings"
)

type Message struct {
	Type          MessageType `json:"type"`
	Data          []byte      `json:"data"`
	ClientAddress string      `json:"client"`
}

func NewMsg(t MessageType, data []byte, addr string) *Message {
	return &Message{
		Type:          t,
		Data:          data,
		ClientAddress: addr,
	}
}

func (m *Message) Encode() ([]byte, error) {
	delim := "\n"
	// Create a buffer to hold the serialized data
	jsonData, err := json.Marshal(m)
	if err != nil {
		return nil, err
	}
	jsonData = append(jsonData, []byte(delim)...)
	return jsonData, err
}

func (m *Message) Decode(b []byte) error {
	b = []byte(strings.TrimSpace(string(b)))
	err := json.Unmarshal(b, &m)
	if err != nil {
		return err
	}
	return nil
}
