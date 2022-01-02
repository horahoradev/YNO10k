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
	MatchPrefix   string `ynoproto:"0"`
	UnignoredUUID string
}

type UnignoreGameEvents struct {
	MatchPrefix   string `ynoproto:"1"`
	UnignoredUUID string
}

type IgnoreChatEvents struct {
	MatchPrefix string `ynoproto:"2"`
	IgnoredUUID string
}

type IgnoreGameEvents struct {
	MatchPrefix string `ynoproto:"3"`
	IgnoredUUID string
}

type SetUsername struct {
	MatchPrefix string `ynoproto:"4"`
	Username    string
}

type SendMessage struct {
	MatchPrefix string `ynoproto:"5"`
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
	MatchPrefix string `ynoproto:"1"`
	X           uint16
	Y           uint16
}

// TODOS:
// 1. Change packet type length to uint8
