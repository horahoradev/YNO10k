package msghandler

import (
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/go-kit/kit/metrics"
	"github.com/horahoradev/YNO10k/internal/client"
	"github.com/horahoradev/YNO10k/internal/clientmessages"
	ynmetrics "github.com/horahoradev/YNO10k/internal/metrics"
	"github.com/horahoradev/YNO10k/internal/protocol"

	"github.com/horahoradev/YNO10k/internal/servermessages"
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
	pubsubManager            client.PubSubManager
	activeSockInfoFlushMap   map[string]*client.ClientSockInfo
	inactiveSockInfoFlushMap map[string]*client.ClientSockInfo
	activeRWLock             *sync.RWMutex
	inactiveRWLock           *sync.RWMutex

	// Metrics
	spriteCounter metrics.Counter
	moveCounter   metrics.Counter
	soundCounter  metrics.Counter
}

func NewGameHandler(ps client.PubSubManager) *GameHandler {

	g := GameHandler{
		pubsubManager:            ps,
		activeSockInfoFlushMap:   make(map[string]*client.ClientSockInfo),
		inactiveSockInfoFlushMap: make(map[string]*client.ClientSockInfo),
		inactiveRWLock:           &sync.RWMutex{},
		activeRWLock:             &sync.RWMutex{},
		spriteCounter:            ynmetrics.ConcreteCounter("yn10k.set_sprite_count"),
		moveCounter:              ynmetrics.ConcreteCounter("yn10k.set_pos_count"),
		soundCounter:             ynmetrics.ConcreteCounter("yn10k.set_sound_count"),
	}
	g.flushWorker()
	return &g
}

func (ch *GameHandler) HandleMessage(payload []byte, c gnet.Conn, s *client.ClientSockInfo) error {
	log.Debugf("Handling game message")
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
		return ch.handleAnimFrame(payload, s)
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
	case 'D':
		// This is a disconnect packet
		return ch.handleDisconnect(payload, s)
	default:
		return fmt.Errorf("Received unknown message %s", payload[0])
	}
}

func (ch *GameHandler) flushWorker() {
	go func() {
		timer := time.NewTicker(time.Second / 60)
		defer timer.Stop()

		loopG := ynmetrics.ConcreteGauge("yno10k.game_handler_flush_latency_ms")
		for true {
			start := time.Now()
			// This simply ensures that the event loop doesn't occur more than 60 hz
			// It isn't a 1/60 second sleep
			<-timer.C

			// Inactive map (which was being written to for the previous cycle) becomes active
			ch.swapActiveAndInactiveMaps()

			// Ensure that all clients gamestate is flushed before moving to next loop,
			// which prevents interleaving of client gamestate broadcasts
			// (which would lead to state issues)
			wg := &sync.WaitGroup{}
			wg.Add(len(ch.activeSockInfoFlushMap))

			ch.activeRWLock.RLock()
			for _, si := range ch.activeSockInfoFlushMap {
				ch.flushSocketInfo(wg, si)
			}

			wg.Wait()
			ch.activeRWLock.RUnlock() // Can only unlock after all workers have returned

			end := time.Since(start)
			loopG.Set(float64(end.Milliseconds()))

			// Just reassign and let GC take care of it
			ch.activeSockInfoFlushMap = make(map[string]*client.ClientSockInfo, len(ch.activeSockInfoFlushMap))
		}
	}()
}

func (ch *GameHandler) swapActiveAndInactiveMaps() {
	// Need to ensure that the inactive map is NOT being written to before we're done swapping
	ch.inactiveRWLock.Lock()
	oldActiveMap := ch.activeSockInfoFlushMap
	ch.activeSockInfoFlushMap = ch.inactiveSockInfoFlushMap
	ch.inactiveSockInfoFlushMap = oldActiveMap
	ch.inactiveRWLock.Unlock()
}

func (ch *GameHandler) flushSocketInfo(wg *sync.WaitGroup, si *client.ClientSockInfo) {
	go func(wg *sync.WaitGroup) {
		defer wg.Done()
		g := ynmetrics.ConcreteGauge("yno10k.flush_duration_ms")

		start := time.Now()
		flushedSO := si.SyncObject.GetFlushedChanges()
		err := ch.pubsubManager.Broadcast(flushedSO, si, false)
		if err != nil {
			log.Errorf("Received error when broadcasting SO: %s", err)
		}
		// Can lead to state problems if send fails, TODO
		duration := time.Since(start)
		g.Set(float64(duration.Milliseconds()))
	}(wg)
}

func (ch *GameHandler) handleDisconnect(payload []byte, s *client.ClientSockInfo) error {
	return ch.pubsubManager.Broadcast(&servermessages.DisconnectMessage{
		Type: "disconnect",
		UUID: s.SyncObject.UID,
	}, s, true)
}

func (ch *GameHandler) handleMovement(payload []byte, c *client.ClientSockInfo) error {
	ch.moveCounter.Add(1)
	t := clientmessages.Movement{}
	matched, seq, err := protocol.Marshal(payload, &t, true)
	switch {
	case err != nil:
		return err
	case !matched:
		return errors.New("Failed to match in handleMovement")
	}

	c.SyncObject.SetPos(t.X, t.Y, seq)
	ch.scheduleChanges(c.SyncObject.UID, c)
	return nil
}

func (ch *GameHandler) handleSprite(payload []byte, c *client.ClientSockInfo) error {
	ch.spriteCounter.Add(1)
	t := clientmessages.Sprite{}
	matched, seq, err := protocol.Marshal(payload, &t, true)
	switch {
	case err != nil:
		return fmt.Errorf("Failed to handleSprite. Err: %s", err)
	case !matched:
		return errors.New("Failed to match in handleSprite")
	}

	c.SyncObject.SetSprite(t.SpriteID, t.Spritesheet, seq)
	ch.scheduleChanges(c.SyncObject.UID, c)
	return nil
}

func (ch *GameHandler) scheduleChanges(uid string, c *client.ClientSockInfo) {
	ch.inactiveRWLock.Lock()
	ch.inactiveSockInfoFlushMap[uid] = c
	ch.inactiveRWLock.Unlock()
}

func (ch *GameHandler) handleSound(payload []byte, c *client.ClientSockInfo) error {
	ch.soundCounter.Add(1)
	t := clientmessages.Sound{}
	matched, _, err := protocol.Marshal(payload, &t, true)
	switch {
	case err != nil:
		return err
	case !matched:
		return errors.New("Failed to match in handleSound")
	}

	c.SyncObject.SetSound(t.Volume, t.Tempo, t.Balance, t.SoundFile)
	ch.scheduleChanges(c.SyncObject.UID, c)
	return nil
}

func (ch *GameHandler) handleWeather(payload []byte, c *client.ClientSockInfo) error {
	t := clientmessages.Weather{}
	matched, _, err := protocol.Marshal(payload, &t, true)
	switch {
	case !matched:
		return errors.New("Failed to match in handleWeather")
	case err != nil:
		return err
	}

	c.SyncObject.SetWeather(t.WeatherType, t.WeatherStrength)
	ch.scheduleChanges(c.SyncObject.UID, c)
	return nil
}

func (ch *GameHandler) handleName(payload []byte, c *client.ClientSockInfo) error {
	t := clientmessages.Name{}
	matched, _, err := protocol.Marshal(payload, &t, true)
	switch {
	case !matched:
		return errors.New("Failed to match in handleWeather")
	case err != nil:
		return err
	}

	c.SyncObject.SetName(t.Name)
	ch.scheduleChanges(c.SyncObject.UID, c)
	return nil
}

func (ch *GameHandler) handleVariable(payload []byte, c *client.ClientSockInfo) error {
	t := clientmessages.Variable{}
	matched, _, err := protocol.Marshal(payload, &t, true)
	switch {
	case !matched:
		return errors.New("Failed to match in handleVariable")
	case err != nil:
		return err
	}

	c.SyncObject.SetVariable(t.ID, t.Value)
	ch.scheduleChanges(c.SyncObject.UID, c)
	return nil
}

func (ch *GameHandler) handleSwitchSync(payload []byte, c *client.ClientSockInfo) error {
	t := clientmessages.SwitchSync{}
	matched, _, err := protocol.Marshal(payload, &t, true)
	switch {
	case !matched:
		return errors.New("Failed to match in handleSwitchSync")
	case err != nil:
		return err
	}

	c.SyncObject.SetSwitch(t.ID, t.Value)
	ch.scheduleChanges(c.SyncObject.UID, c)
	return nil
}

func (ch *GameHandler) handleAnimFrame(payload []byte, c *client.ClientSockInfo) error {
	t := clientmessages.AnimFrame{}
	matched, seq, err := protocol.Marshal(payload, &t, true)
	switch {
	case err != nil:
		return err
	case !matched:
		return errors.New("failed to match in handleAnimFrame")
	}

	c.SyncObject.SetAnimFrame(t.Frame, seq)
	ch.scheduleChanges(c.SyncObject.UID, c)
	return nil
}

func (ch *GameHandler) handleFacing(payload []byte, c *client.ClientSockInfo) error {
	t := clientmessages.Facing{}
	matched, _, err := protocol.Marshal(payload, &t, true)
	switch {
	case err != nil:
		return err
	case !matched:
		return errors.New("Failed to match in handleFacing")
	}

	c.SyncObject.SetFacing(t.Facing)
	ch.scheduleChanges(c.SyncObject.UID, c)
	return nil
}

func (ch *GameHandler) handleTypingStatus(payload []byte, c *client.ClientSockInfo) error {
	t := clientmessages.TypingStatus{}
	matched, _, err := protocol.Marshal(payload, &t, true)
	switch {
	case !matched:
		return errors.New("Failed to match in handleTypingStatus")
	case err != nil:
		return err
	}

	c.SyncObject.SetTypingStatus(t.TypingStatus)
	ch.scheduleChanges(c.SyncObject.UID, c)
	return nil
}

func (ch *GameHandler) handleMovementAnimSpeed(payload []byte, c *client.ClientSockInfo) error {
	t := clientmessages.MovementAnimationSpeed{}
	matched, seq, err := protocol.Marshal(payload, &t, true)
	switch {
	case err != nil:
		return err
	case !matched:
		return errors.New("Failed to match in handleMovementAnimSpeed")
	}

	c.SyncObject.SetMovementAnimationSpeed(t.MovementSpeed, seq)
	ch.scheduleChanges(c.SyncObject.UID, c)
	return nil
}
