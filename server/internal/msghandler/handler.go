package msghandler

import (
	"github.com/horahoradev/YNO10k/internal/client"
	"github.com/panjf2000/gnet"
)

type Handler interface {
	HandleMessage([]byte, gnet.Conn, *client.ClientSockInfo) error
}
