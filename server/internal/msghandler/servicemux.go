package msghandler

import (
	"errors"

	"github.com/horahoradev/YNO10k/internal/client"
	"github.com/panjf2000/gnet"
	log "github.com/sirupsen/logrus"
)

// Similar to chain of responsibility pattern
// Passes message to the appropriate service
type ServiceMux struct {
	cm client.PubSubManager
	gh Handler
	ch Handler
	lh Handler
}

func NewServiceMux(gh, ch, lh Handler, cm client.PubSubManager) ServiceMux {
	return ServiceMux{
		cm: cm,
		gh: gh,
		ch: ch,
		lh: lh,
	}
}

func (sm ServiceMux) HandleMessage(clientPayload []byte, c gnet.Conn, cinfo *client.ClientSockInfo) error {
	// Do we have existing context? Then it's a normal message
	var clientInfo *client.ClientSockInfo
	ctx := c.Context()
	if ctx != nil {
		var ok bool
		clientInfo, ok = ctx.(*client.ClientSockInfo)
		if !ok {
			log.Errorf("Failed to typecast context to client")
			c.Close()
			return errors.New("failed to typecast context to client")
		}
	}

	switch {
	case clientInfo == nil:
		// This is the servicename packet, use it to initialize the client info

		clientInfo, err := sm.cm.SubscribeClientToRoom(string(clientPayload), c)
		if err != nil {
			log.Errorf("Failed to add client for room. Err: %s", err)
		}

		// Store the client info with the connection
		c.SetContext(clientInfo)
		return nil

	default:
		// We've already received the service packet, so this is regular message
		switch clientInfo.ServiceType {
		case client.GlobalChat, client.Chat:
			return sm.ch.HandleMessage(clientPayload, c, clientInfo)
		case client.Game:
			return sm.gh.HandleMessage(clientPayload, c, clientInfo)
		case client.List:
			return sm.lh.HandleMessage(clientPayload, c, clientInfo)
		default:
			return errors.New("Could not handle message, client socket servicetype was not set")
		}
	}
}