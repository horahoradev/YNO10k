package clientmessages

/*
const (
	pardonChat = iota
	pardonGame
	ignoreChat
	ignoreGame
	getUUID
	userMessage
)
*/

type UnignoreChatEvents struct {
	MatchPrefix       string `ynoproto:"00"`
	UnignoredUsername string
}

type UnignoreGameEvents struct {
	MatchPrefix       string `ynoproto:"01"`
	UnignoredUsername string
}

type IgnoreChatEvents struct {
	MatchPrefix     string `ynoproto:"02"`
	IgnoredUsername string
}

type IgnoreGameEvents struct {
	MatchPrefix     string `ynoproto:"03"`
	IgnoredUsername string
}

type SetUsername struct {
	MatchPrefix string `ynoproto:"04"`
	Username    string
}

type SendMessage struct {
	MatchPrefix string `ynoproto:"05"`
	Message     string
}

/*
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
*/

// uint16 packet type, uint16 X, uint16_t Y
type Movement struct {
	MatchPrefix string `ynoproto:"01"`
	X           uint16
	Y           uint16
}

// TODOS:
// 1. Change packet type length to uint8

// uint16 packet type, uint16 sprite 'id', string spritesheet
type Sprite struct {
	MatchPrefix string `ynoproto:"02"`
	SpriteID    uint16
	Spritesheet uint32
}

//uint16 packet type, uint16 volume, uint16 tempo, uint16 balance, string sound file
type Sound struct {
	MatchPrefix string `ynoproto:"03"`
	Volume      uint16
	Tempo       uint16
	Balance     uint16
	SoundFile   string
}

//uint16 packet type, uint16 weather type, uint16 weather strength
type Weather struct {
	MatchPrefix     string `ynoproto:"04"`
	WeatherType     uint16
	WeatherStrength uint16
}

//uint16 packet type, string name
type Name struct {
	MatchPrefix string `ynoproto:"05"`
	Name        string
}

//uint16 packet type, uint16 movement speed
type MovementAnimationSpeed struct {
	MatchPrefix   string `ynoproto:"4"`
	MovementSpeed uint16
}

//uint16 packet type, uint32 var id, int32 value
type Variable struct {
	MatchPrefix string `ynoproto:"5"`
	ID          uint32
	Value       uint32
}

//uint16 packet type, uint32 switch id, int32 value
type SwitchSync struct {
	MatchPrefix string `ynoproto:"6"`
	ID          uint32
	Value       uint32
}

//uint16 packet type, uint16 type
type AnimType struct {
	MatchPrefix string `ynoproto:"7"`
	Type        uint16
}

//uint16 packet type, uint16 frame
type AnimFrame struct {
	MatchPrefix string `ynoproto:"8"`
	Frame       uint16
}

type Facing struct {
	MatchPrefix string `ynoproto:"9"`
	Facing      uint16
}

type TypingStatus struct {
	MatchPrefix  string `ynoproto:"10"`
	TypingStatus uint16
}
