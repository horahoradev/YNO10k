package msghandler

import (
	"errors"
	"fmt"

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
	facing
	typingstatus
	syncme // Deprecated
)

type GameHandler struct {
	pubsubManager   client.PubSubManager
	syncobjFlushMap map[string]*client.SyncObject
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
	for true {
		for key, so := range ch.syncobjFlushMap {
			flushedSO := so.GetFlushedChanges()
			err := ch.pubsubManager.Broadcast(flushedSO)
			// TODO: include uuid in argument to broadcast so we can ignore the sending player
			// Also include room to publish to
			if err != nil {
				log.Errorf("Received error when broadcasting SO: %s", err)
			} else {
				delete(ch.syncobjFlushMap, key)
			}
		}
	}
}

func (ch *GameHandler) handleMovement(payload []byte, c *client.ClientSockInfo) error {
	t := clientmessages.Movement{}
	matched, err := protocol.Marshal(payload, &t)
	switch {
	case !matched:
		return errors.New("Failed to match")
	case err != nil:
		return err
	}

	c.SyncObject.SetPos(t.X, t.Y)
	ch.syncobjFlushMap[c.SyncObject.UID] = c.SyncObject
	return nil
}
