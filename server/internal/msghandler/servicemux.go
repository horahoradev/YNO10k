package msghandler

import (
	"errors"
	"net"
	"sync"

	"github.com/horahoradev/YNO10k/internal/client"
	"github.com/panjf2000/gnet"
	log "github.com/sirupsen/logrus"
)

// Similar to chain of responsibility pattern
// Passes message to the appropriate service
type ServiceMux struct {
	cm          client.PubSubManager
	gh          Handler
	ch          Handler
	lh          Handler
	SyncChanMap map[net.Addr]chan struct{}
	m           *sync.Mutex
}

func NewServiceMux(gh, ch, lh Handler, cm client.PubSubManager) ServiceMux {
	return ServiceMux{
		cm:          cm,
		gh:          gh,
		ch:          ch,
		lh:          lh,
		SyncChanMap: make(map[net.Addr]chan struct{}),
		m:           &sync.Mutex{},
	}
}

func (sm ServiceMux) HandleMessage(clientPayload []byte, c gnet.Conn, cinfo *client.ClientSockInfo) error {
	log.Debug("Handling service message")

	syncChan, ok := sm.SyncChanMap[c.RemoteAddr()]
	sm.m.Lock()
	if !ok {
		syncChan = make(chan struct{})
		sm.SyncChanMap[c.RemoteAddr()] = syncChan
	}
	sm.m.Unlock()

	// Do we have existing context? Then it's a normal message
	_, _, err := client.SplitServiceName(string(clientPayload))
	switch {
	case getClientSockInfo(c) == nil && err == nil:
		// This is a packet I have no need for. FIXME on the client
		if string(clientPayload) == "chat" {
			return nil
		}

		// This is the servicename packet, use it to initialize the client info
		log.Debugf("Subscribing client to room %s", string(clientPayload))
		cInfo, err := sm.cm.SubscribeClientToRoom(string(clientPayload), c)
		if err != nil {
			log.Errorf("Failed to add client for room. Err: %s", err)
			return err
		}
		log.Debugf("Subscribed client to room %s", string(clientPayload))
		// Store the client info with the connection
		c.SetContext(cInfo)
		close(syncChan)
		return nil

	default:
		<-syncChan
		clientInfo := getClientSockInfo(c)
		if clientInfo == nil {
			return errors.New("Clientinfo nil in message handler main code path")
		}
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

func getClientSockInfo(c gnet.Conn) *client.ClientSockInfo {
	ctx := c.Context()
	if ctx != nil {
		var ok bool
		clientInfo, ok := ctx.(*client.ClientSockInfo)
		if !ok {
			log.Errorf("Failed to typecast context to client")
			c.Close()
			return nil
		}

		return clientInfo
	}

	return nil
}
