package msghandler

import "github.com/panjf2000/gnet"

type Handler interface {
	HandleMessage([]byte, gnet.Conn) error
}
