package msghandler

import (
	"errors"
	"fmt"
	"time"

	"github.com/horahoradev/YNO10k/internal/client"
	"github.com/horahoradev/YNO10k/internal/clientmessages"
	"github.com/horahoradev/YNO10k/internal/protocol"
	"github.com/panjf2000/gnet"
	log "github.com/sirupsen/logrus"
)

const (
	movement = iota + 1
	sprite
	sound
	weather
	name
	movementAnimationSpeed
	variable
	switchsync
	animtype
	animFrame
	facing
	typingstatus
	syncme // Deprecated
)

type GameHandler struct {
	pubsubManager    client.PubSubManager
	sockinfoFlushMap map[string]*client.ClientSockInfo
}

func (ch *GameHandler) HandleMessage(payload []byte, c gnet.Conn, s *client.ClientSockInfo) error {
	return nil

}

func (ch *GameHandler) muxMessage(payload []byte, c gnet.Conn, s *client.ClientSockInfo) error {
	if len(payload) == 0 {
		return errors.New("Payload cannot be empty!")
	}

	switch payload[0] {
	case movement:
		return ch.handleMovement(payload, c, s)

	case sprite:

	case sound:

	case weather:

	case name:

	case movementAnimationSpeed:

	case variable:

	case switchsync:

	case animtype:

	case facing:

	case typingstatus:

	case syncme:

	default:
		return fmt.Errorf("Received unknown message %s", payload[0])
	}

}

func (ch *GameHandler) flushWorker() error {
	timer := time.NewTimer(time.Second)
	defer timer.Stop()

	for true {
		<-timer.C
		for key, si := range ch.sockinfoFlushMap {
			flushedSO := si.SyncObject.GetFlushedChanges()
			err := ch.pubsubManager.Broadcast(flushedSO, si)
			if err != nil {
				log.Errorf("Received error when broadcasting SO: %s", err)
			} else {
				delete(ch.sockinfoFlushMap, key)
			}
		}
	}
}

func (ch *GameHandler) handleMovement(payload []byte, c *client.ClientSockInfo) error {
	t := clientmessages.Movement{}
	matched, err := protocol.Marshal(payload, &t)
	switch {
	case !matched:
		return errors.New("Failed to match in handleMovement")
	case err != nil:
		return err
	}

	c.SyncObject.SetPos(t.X, t.Y)
	ch.sockinfoFlushMap[c.SyncObject.UID] = c
	return nil
}

func (ch *GameHandler) handleSprite(payload []byte, c *client.ClientSockInfo) error {
	t := clientmessage.{}
	matched, err := protocol.Marshal(payload, &t)
	switch {
	case !matched:
		return errors.New("Failed to match in handleSprite")
	case err != nil:
		return err
	}

	c.SyncObject.SetPos(t.X, t.Y)
	ch.sockinfoFlushMap[c.SyncObject.UID] = c
	return nil
}
