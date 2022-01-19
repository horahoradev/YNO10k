package servermessages

type GetUUID struct {
}

// TODO: used for e.g. error indications
type ServerMessage struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

type UserMessage struct {
	Type string `json:"type"`
	Text string `json:"text"`
	Name string `json:"name"`
	Trip string `json:"trip"`
}

type DisconnectMessage struct {
	Type string `json:"type"`
	UUID string `json:"uuid"`
}
