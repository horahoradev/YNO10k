package client

import "github.com/panjf2000/gnet"

type PubSubManager interface {
	SubscribeClientToRoom(serviceName string, conn gnet.Conn) (*ClientSockInfo, error)
}