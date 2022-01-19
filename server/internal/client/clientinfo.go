package client

import (
	"fmt"
	"net"

	log "github.com/sirupsen/logrus"

	"github.com/panjf2000/gnet"
)

type ClientID struct {
	Name string
	Conn gnet.Conn
}

func (cid *ClientID) GetAddr() net.Addr {
	return cid.Conn.RemoteAddr()
}

// Dumb
func (cid *ClientID) IsClosed() bool {
	return cid.Conn.RemoteAddr() == nil || cid.Conn.RemoteAddr().String() == ""
}

func (cid *ClientID) GetUsername() string {
	return cid.Name
}

func (cid *ClientID) SetUsername(name string) {
	cid.Name = name
}

type Client interface {
	Ignore(net.Addr)
	Unignore(net.Addr)
	Send(payload []byte, sender net.Addr) error
	GetAddr() net.Addr
	GetUsername() string
	SetUsername(name string)
	IsClosed() bool
}

type GameClient struct {
	ClientID
	GameIgnores []net.Addr
}

func (gc *GameClient) Ignore(ipv4 net.Addr) {
	// TODO: singleton
	gc.GameIgnores = append(gc.GameIgnores, ipv4)
}

func (gc *GameClient) Unignore(ipv4 net.Addr) {
	for i, addr := range gc.GameIgnores {
		if addr != nil && addr.String() == ipv4.String() {
			// Lol
			gc.GameIgnores[i] = nil
		}
	}
}

func (gc *GameClient) Send(payload []byte, sender net.Addr) error {
	// if sender == nil {
	// 	return errors.New("Sender has disconnected")
	// }

	if gc.ClientID.Conn.RemoteAddr() == nil {
		fmt.Errorf("client's remote addr is nil, has likely already disconnected. Dropping message.")
		return nil
	}

	// If the sender is the current user, just return
	if sender != nil && gc.ClientID.Conn.RemoteAddr().String() == sender.String() {
		return nil
	}

	// is the sender ignored? If so, return without an error
	for _, ignoredAddr := range gc.GameIgnores {
		if ignoredAddr.String() == sender.String() {
			return nil
		}
	}

	return gc.Conn.AsyncWrite(payload)
}

type ChatClient struct {
	ClientID
	ChatIgnores []net.Addr // So this needs to be a singleton or something
}

// TODO: DRY lmao

func (gc *ChatClient) Ignore(ipv4 net.Addr) {
	// TODO: singleton
	gc.ChatIgnores = append(gc.ChatIgnores, ipv4)
}

func (gc *ChatClient) Send(payload []byte, sender net.Addr) error {
	// If the sender is the current user, we want to receive it anyway

	// is the sender ignored? If so, return without an error
	for _, ignoredAddr := range gc.ChatIgnores {
		if ignoredAddr.String() == sender.String() {
			return nil
		}
	}

	return gc.Conn.AsyncWrite(payload)
}

func (gc *ChatClient) Unignore(ipv4 net.Addr) {
	for i, addr := range gc.ChatIgnores {
		if addr != nil && addr.String() == ipv4.String() {
			// Lol
			gc.ChatIgnores[i] = nil
		}
	}
}

func newClient(t ServiceType, conn gnet.Conn) Client {
	clientID := ClientID{
		Conn: conn,
	}

	// TODO: default
	switch t {
	case Chat, GlobalChat:
		return &ChatClient{
			ClientID: clientID,
		}
	case Game:
		return &GameClient{
			ClientID: clientID,
		}
	default:
		// TODO
		log.Errorf("Invalid client type")
		return nil
	}
}
