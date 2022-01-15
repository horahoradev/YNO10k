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

func NewGameHandler(ps client.PubSubManager) *GameHandler {
	return &GameHandler{
		pubsubManager:    ps,
		sockinfoFlushMap: make(map[string]*client.ClientSockInfo),
	}
}

func (ch *GameHandler) HandleMessage(payload []byte, c gnet.Conn, s *client.ClientSockInfo) error {
	return ch.muxMessage(payload, c, s)

}

func (ch *GameHandler) muxMessage(payload []byte, c gnet.Conn, s *client.ClientSockInfo) error {
	if len(payload) == 0 {
		return errors.New("Payload cannot be empty!")
	}

	switch payload[0] {
	case movement:
		return ch.handleMovement(payload, s)
	case sprite:
		return ch.handleSprite(payload, s)
	case sound:
		return ch.handleSound(payload, s)
	case weather:
		return ch.handleWeather(payload, s)
	case name:
		return ch.handleName(payload, s)
	case movementAnimationSpeed:
		return ch.handleMovementAnimSpeed(payload, s)
	case variable:
		return ch.handleVariable(payload, s)
	case animFrame:
		// Unimplemented
		return errors.New("Received unimplemented message type animFrame")
	case switchsync:
		return ch.handleSwitchSync(payload, s)
	case animtype:
		// Unimplemented
		return errors.New("Received unimplemented message type animtype")

	case facing:
		return ch.handleFacing(payload, s)
	case typingstatus:
		return ch.handleTypingStatus(payload, s)
	case syncme:
		// Deprecated
		return errors.New("Deprecated message type syncme")
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
	return nil
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
	t := clientmessages.Sprite{}
	matched, err := protocol.Marshal(payload, &t)
	switch {
	case !matched:
		return errors.New("Failed to match in handleSprite")
	case err != nil:
		return err
	}

	c.SyncObject.SetSprite(t.SpriteID, t.Spritesheet)
	ch.sockinfoFlushMap[c.SyncObject.UID] = c
	return nil
}

func (ch *GameHandler) handleSound(payload []byte, c *client.ClientSockInfo) error {
	t := clientmessages.Sound{}
	matched, err := protocol.Marshal(payload, &t)
	switch {
	case !matched:
		return errors.New("Failed to match in handleSound")
	case err != nil:
		return err
	}

	c.SyncObject.SetSound(t.Volume, t.Tempo, t.Balance, t.SoundFile)
	ch.sockinfoFlushMap[c.SyncObject.UID] = c
	return nil
}

func (ch *GameHandler) handleWeather(payload []byte, c *client.ClientSockInfo) error {
	t := clientmessages.Weather{}
	matched, err := protocol.Marshal(payload, &t)
	switch {
	case !matched:
		return errors.New("Failed to match in handleWeather")
	case err != nil:
		return err
	}

	c.SyncObject.SetWeather(t.WeatherType, t.WeatherStrength)
	ch.sockinfoFlushMap[c.SyncObject.UID] = c
	return nil
}

func (ch *GameHandler) handleName(payload []byte, c *client.ClientSockInfo) error {
	t := clientmessages.Name{}
	matched, err := protocol.Marshal(payload, &t)
	switch {
	case !matched:
		return errors.New("Failed to match in handleWeather")
	case err != nil:
		return err
	}

	c.SyncObject.SetName(t.Name)
	ch.sockinfoFlushMap[c.SyncObject.UID] = c
	return nil
}

func (ch *GameHandler) handleVariable(payload []byte, c *client.ClientSockInfo) error {
	t := clientmessages.Variable{}
	matched, err := protocol.Marshal(payload, &t)
	switch {
	case !matched:
		return errors.New("Failed to match in handleVariable")
	case err != nil:
		return err
	}

	c.SyncObject.SetVariable(t.ID, t.Value)
	ch.sockinfoFlushMap[c.SyncObject.UID] = c
	return nil
}

func (ch *GameHandler) handleSwitchSync(payload []byte, c *client.ClientSockInfo) error {
	t := clientmessages.SwitchSync{}
	matched, err := protocol.Marshal(payload, &t)
	switch {
	case !matched:
		return errors.New("Failed to match in handleSwitchSync")
	case err != nil:
		return err
	}

	c.SyncObject.SetSwitch(t.ID, t.Value)
	ch.sockinfoFlushMap[c.SyncObject.UID] = c
	return nil
}

func (ch *GameHandler) handleAnimFrame(payload []byte, c *client.ClientSockInfo) error {
	t := clientmessages.AnimFrame{}
	matched, err := protocol.Marshal(payload, &t)
	switch {
	case !matched:
		return errors.New("Failed to match in handleWeather")
	case err != nil:
		return err
	}

	c.SyncObject.SetAnimFrame(t.Frame)
	ch.sockinfoFlushMap[c.SyncObject.UID] = c
	return nil
}

func (ch *GameHandler) handleFacing(payload []byte, c *client.ClientSockInfo) error {
	t := clientmessages.Facing{}
	matched, err := protocol.Marshal(payload, &t)
	switch {
	case !matched:
		return errors.New("Failed to match in handleFacing")
	case err != nil:
		return err
	}

	c.SyncObject.SetFacing(t.Facing)
	ch.sockinfoFlushMap[c.SyncObject.UID] = c
	return nil
}

func (ch *GameHandler) handleTypingStatus(payload []byte, c *client.ClientSockInfo) error {
	t := clientmessages.TypingStatus{}
	matched, err := protocol.Marshal(payload, &t)
	switch {
	case !matched:
		return errors.New("Failed to match in handleTypingStatus")
	case err != nil:
		return err
	}

	c.SyncObject.SetTypingStatus(t.TypingStatus)
	ch.sockinfoFlushMap[c.SyncObject.UID] = c
	return nil
}

func (ch *GameHandler) handleMovementAnimSpeed(payload []byte, c *client.ClientSockInfo) error {
	t := clientmessages.MovementAnimationSpeed{}
	matched, err := protocol.Marshal(payload, &t)
	switch {
	case !matched:
		return errors.New("Failed to match in handleMovementAnimSpeed")
	case err != nil:
		return err
	}

	c.SyncObject.SetMovementAnimationSpeed(t.MovementSpeed)
	ch.sockinfoFlushMap[c.SyncObject.UID] = c
	return nil
}
