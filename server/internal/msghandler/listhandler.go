package msghandler

import (
	"github.com/horahoradev/YNO10k/internal/client"
	"github.com/panjf2000/gnet"
)

type ListHandler struct{}

func NewListHandler() *ListHandler {
	return &ListHandler{}
}

func (ch *ListHandler) HandleMessage(payload []byte, c gnet.Conn, s *client.ClientSockInfo) error {
	// TODO
	return nil
}
