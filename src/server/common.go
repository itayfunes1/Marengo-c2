package server

import "net"

const NumberOfMessageChan = 1_000_000

type Msg struct {
	from    string
	payload []byte
	conn    net.Conn
}

type ServerTcp interface {
	Start()
	acceptConnection()
	handlerConnection(net.Conn)
	handleMess()
}
