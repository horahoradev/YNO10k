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
