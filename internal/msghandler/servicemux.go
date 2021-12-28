package msghandler

import (
	"errors"
	"fmt"

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

func (sm *ServiceMux) HandleMessage(clientPayload []byte, c gnet.Conn) error {
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

	default:
		// We've already received the service packet, so this is regular message
		switch clientInfo.ServiceType {
		case client.GlobalChat, client.Chat:
			err := sm.ch.HandleMessage(clientPayload, c, clientInfo)
			if err != nil {
				return fmt.Errorf("chat handler failed to handle message: %s", err)
			}
		case client.Game:
			err := sm.gh.HandleMessage(clientPayload, c, clientInfo)
			if err != nil {
				return fmt.Errorf("game handler failed to handle message: %s", err)
			}
		case client.List:
			err := sm.lh.HandleMessage(clientPayload, c, clientInfo)
			if err != nil {
				return fmt.Errorf("list handler failed to handle message: %s", err)
			}
		default:
			log.Errorf("Could not handle message, client socket servicetype was not set")
		}
	}
}
