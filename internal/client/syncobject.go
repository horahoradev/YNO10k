package client

import guuid "github.com/google/uuid"

type Position struct {
	X uint16
	Y uint16
}

type Sound struct {
	Volume  uint16
	Tempo   uint16
	Balance uint16
	Name    string
}

type SyncObject struct {
	UID string `json:"uid"` // Actually a UUID

	Pos        Position `json:"pos,omitempty"`
	posChanged bool

	Sprite        string `json:"sprite,omitempty"`
	spriteChanged bool

	Sound        Sound `json:"sound,omitempty"`
	soundChanged bool

	Name        string `json:"name,omitempty"`
	nameChanged bool

	MovementAnimationSpeed        uint16 `json:"movementAnimationSpeed,omitempty"`
	movementAnimationSpeedChanged bool

	Facing        uint16 `json:"facing,omitempty"`
	facingChanged bool

	TypingStatus        uint16 `json:"typingstatus,omitempty"`
	typingStatusChanged bool
}

func NewSyncObject() *SyncObject {
	uuid := guuid.New()
	return &SyncObject{UID: uuid.String()}
}

func (so *SyncObject) SetPos(x, y uint16) {
	so.posChanged = true
	so.Pos = Position{X: x, Y: y}
}

func (so *SyncObject) SetSprite(sprite string) {
	so.spriteChanged = true
	// TODO: sprite validation goes here
	so.Sprite = sprite
}

func (so *SyncObject) SetSound(volume uint16, tempo uint16, balance uint16, name string) {
	so.soundChanged = true
	so.Sound = Sound{Volume: volume, Tempo: tempo, Balance: balance, Name: name}
}

func (so *SyncObject) SetName(name string) {
	so.nameChanged = true
	so.Name = name
}

func (so *SyncObject) SetMovementAnimationSpeed(animationSpeed uint16) {
	so.movementAnimationSpeedChanged = true
	so.MovementAnimationSpeed = animationSpeed
}

func (so *SyncObject) SetFacing(facing uint16) {
	so.facingChanged = true
	so.Facing = facing
}

func (so *SyncObject) SetTypingStatus(typingStatus uint16) {
	so.typingStatusChanged = true
	so.TypingStatus = typingStatus
}

func (so *SyncObject) GetAllChanges() interface{} {
	return so
}

// So this is horrific BUT idk what to do about it lol
func (so *SyncObject) GetFlushedChanges() interface{} {
	s := SyncObject{UID: so.UID}

	if so.posChanged {
		s.Pos = so.Pos
	}

	if so.spriteChanged {
		s.Sprite = so.Sprite
	}

	if so.soundChanged {
		s.Sound = so.Sound
	}

	if so.nameChanged {
		s.Name = so.Name
	}

	if s.movementAnimationSpeedChanged {
		s.MovementAnimationSpeed = so.MovementAnimationSpeed
	}

	if s.facingChanged {
		s.Facing = so.Facing
	}

	if s.typingStatusChanged {
		s.TypingStatus = so.TypingStatus
	}

	return s
}
