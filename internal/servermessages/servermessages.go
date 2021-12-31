package servermessages

type GetUUID struct {
}

// TODO: used for e.g. error indications
type ServerMessage struct {
	MessageType string
	Message     string
}

type UserMessage struct {
	MessageType string
	Text        string
	Name        string
	Trip        string
}
