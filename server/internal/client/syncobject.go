package client

import (
	"strconv"
)

var serial int = 0

type Position struct {
	X uint16 `json:"x"`
	Y uint16 `json:"y"`
}

//			return {id: data.readUInt32LE(2), value: data.readUInt32LE(6)};
type Variable struct {
	ID    uint32 `json:"id"`
	Value uint32 `json:"value"`
}

type Sound struct {
	Volume  uint16 `json:"volume"`
	Tempo   uint16 `json:"tempo"`
	Balance uint16 `json:"balance"`
	Name    string `json:"name"`
}

type Sprite struct {
	ID    uint16 `json:"id"`
	Sheet string `json:"sheet"`
}

// 			return {frame: data.readUInt16LE(2)};
type AnimFrame struct {
	Frame uint16 `json:"frame"`
}

// 			return {id: data.readUInt32LE(2), value: data.readUInt32LE(6)};

type Switch struct {
	ID    uint32 `json:"id"`
	Value uint32 `json:"value"`
}

type Weather struct {
	Type     uint16 `json:"type"`
	Strength uint16 `json:"strength"`
}

type SyncObject struct {
	Type string `json:"type"`
	UID  string `json:"uid"` // Actually a UUID

	Pos        *Position `json:"pos,omitempty"`
	posChanged bool

	Sprite        *Sprite `json:"sprite,omitempty"`
	spriteChanged bool

	Weather        *Weather `json:"weather,omitempty"`
	weatherChanged bool

	Variable        *Variable `json:"variable,omitempty"`
	variableChanged bool

	Sound        *Sound `json:"sound,omitempty"`
	soundChanged bool

	Name        string `json:"name,omitempty"`
	nameChanged bool

	AnimFrame        *uint16 `json:"animframe,omitempty"`
	animframeChanged bool

	Switch        *Switch `json:"switch,omitempty"`
	switchChanged bool

	MovementAnimationSpeed        *uint16 `json:"movementAnimationSpeed,omitempty"`
	movementAnimationSpeedChanged bool

	Facing        *uint16 `json:"facing,omitempty"`
	facingChanged bool

	TypingStatus        *uint16 `json:"typingstatus,omitempty"`
	typingStatusChanged bool
}

func NewSyncObject() *SyncObject {
	serial += 1
	var defaultSpeed uint16 = 4
	return &SyncObject{UID: strconv.Itoa(serial), Type: "objectSync", MovementAnimationSpeed: &defaultSpeed}
}

func (so *SyncObject) SetPos(x, y uint16) {
	so.posChanged = true
	so.Pos = &Position{X: x, Y: y}
}

func (so *SyncObject) SetSprite(id uint16, sheet string) {
	// "why do I have to do this?"
	if id == 0 && sheet == "" {
		return
	}
	so.Sprite = &Sprite{ID: id, Sheet: sheet}
	so.spriteChanged = true
	// TODO: sprite validation goes here
}

func (so *SyncObject) SetSound(volume uint16, tempo uint16, balance uint16, name string) {
	so.soundChanged = true
	so.Sound = &Sound{Volume: volume, Tempo: tempo, Balance: balance, Name: name}
}

func (so *SyncObject) SetWeather(t, strength uint16) {
	so.weatherChanged = true
	so.Weather = &Weather{Type: t, Strength: strength}
}

func (so *SyncObject) SetSwitch(id, value uint32) {
	so.switchChanged = true
	so.Switch = &Switch{ID: id, Value: value}
}

func (so *SyncObject) SetAnimFrame(frame uint16) {
	so.AnimFrame = &frame
	so.animframeChanged = true
}

func (so *SyncObject) SetName(name string) {
	so.nameChanged = true
	so.Name = name
}

func (so *SyncObject) SetMovementAnimationSpeed(animationSpeed uint16) {
	so.movementAnimationSpeedChanged = true
	so.MovementAnimationSpeed = &animationSpeed
}

func (so *SyncObject) SetFacing(facing uint16) {
	so.Facing = &facing
	so.facingChanged = true
}

func (so *SyncObject) SetTypingStatus(typingStatus uint16) {
	so.typingStatusChanged = true
	so.TypingStatus = &typingStatus
}

func (so *SyncObject) SetVariable(id, value uint32) {
	so.variableChanged = true
	so.Variable = &Variable{ID: id, Value: value}
}

func (so *SyncObject) GetAllChanges() interface{} {
	return so
}

func (so *SyncObject) clearChanges() {
	so.posChanged = false
	so.spriteChanged = false
	so.weatherChanged = false
	so.soundChanged = false
	so.nameChanged = false
	so.movementAnimationSpeedChanged = false
	so.facingChanged = false
	so.typingStatusChanged = false
	so.variableChanged = false
	so.switchChanged = false
	so.animframeChanged = false
}

// So this is horrific BUT idk what to do about it lol
func (so *SyncObject) GetFlushedChanges() interface{} {
	s := SyncObject{UID: so.UID, Type: so.Type}

	if so.posChanged {
		s.Pos = so.Pos
	}

	if so.spriteChanged {
		s.Sprite = so.Sprite
	}

	if so.weatherChanged {
		s.Weather = so.Weather
	}

	if so.variableChanged {
		s.Variable = so.Variable
	}

	if so.switchChanged {
		s.Variable = so.Variable
	}

	if so.soundChanged {
		s.Sound = so.Sound
	}

	if so.nameChanged {
		s.Name = so.Name
	}

	if so.movementAnimationSpeedChanged {
		s.MovementAnimationSpeed = so.MovementAnimationSpeed
	}

	if so.facingChanged {
		s.Facing = so.Facing
	}

	if so.typingStatusChanged {
		s.TypingStatus = so.TypingStatus
	}

	if so.animframeChanged {
		s.AnimFrame = so.AnimFrame
	}

	so.clearChanges()

	return s
}
