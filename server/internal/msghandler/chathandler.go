package msghandler

import (
	"errors"
	"fmt"

	"github.com/horahoradev/YNO10k/internal/client"
	"github.com/horahoradev/YNO10k/internal/clientmessages"
	"github.com/horahoradev/YNO10k/internal/protocol"
	"github.com/horahoradev/YNO10k/internal/servermessages"
	"github.com/panjf2000/gnet"
	log "github.com/sirupsen/logrus"
)

const (
	pardonChat = iota
	pardonGame
	ignoreChat
	ignoreGame
	setUsername
	userMessage
)

type ChatHandler struct {
	pm client.PubSubManager
}

func NewChatHandler(pm client.PubSubManager) *ChatHandler {
	return &ChatHandler{
		pm: pm,
	}
}

func (ch *ChatHandler) HandleMessage(payload []byte, c gnet.Conn, s *client.ClientSockInfo) error {
	log.Print("Handling chat message")
	return ch.muxMessage(payload, c, s)
}

func (ch *ChatHandler) muxMessage(payload []byte, c gnet.Conn, s *client.ClientSockInfo) error {
	if len(payload) == 0 {
		return errors.New("Payload cannot be empty!")
	}

	switch payload[0] {
	case pardonChat:
		return ch.pardonChat(payload, s)
	case ignoreChat:
		return ch.ignoreChat(payload, s)
	case pardonGame:
		// Unimplemented for now, will require a hack with my modeling
		log.Errorf("Received pardonGame and couldn't handle")
		// return ch.pardonGame(payload, s)
	case ignoreGame:
		log.Errorf("Received ignoreGame and couldn't handle")
		// return ch.ignoreGame(payload, s)
	case setUsername:
		return ch.setUsername(payload, s)
	case userMessage:
		return ch.sendUserMessage(payload, s)
	default:
		return fmt.Errorf("Received unknown message %s", payload[0])
	}

	return nil
}

func (ch *ChatHandler) pardonChat(payload []byte, client *client.ClientSockInfo) error {
	t := clientmessages.UnignoreChatEvents{}
	matched, err := protocol.Marshal(payload, &t)
	switch {
	case !matched:
		return errors.New("Failed to match")
	case err != nil:
		return err
	}

	user, err := ch.pm.GetUsernameForGame(client.GameName, client.RoomName, t.UnignoredUsername)
	if err != nil {
		return err
	}

	client.ClientInfo.Unignore(user.ClientInfo.GetAddr())
	return nil
}

// func (ch *ChatHandler) pardonGame(payload []byte, client *client.ClientSockInfo) error {
// 	t := messages.UnignoreGameEvents{}
// 	matched, err := protocol.Marshal(payload, &t)
// 	switch {
// 	case !matched:
// 		return errors.New("Failed to match")
// 	case err != nil:
// 		return err
// 	}

// 	return client.UnignoreGameEvents(t)
// }

func (ch *ChatHandler) ignoreChat(payload []byte, client *client.ClientSockInfo) error {
	t := clientmessages.IgnoreChatEvents{}
	matched, err := protocol.Marshal(payload, &t)
	switch {
	case !matched:
		return errors.New("Failed to match")
	case err != nil:
		return err
	}

	user, err := ch.pm.GetUsernameForGame(client.GameName, client.RoomName, t.IgnoredUsername)
	if err != nil {
		return err
	}

	client.ClientInfo.Ignore(user.ClientInfo.GetAddr())

	return nil
}

// func (ch *ChatHandler) ignoreGame(payload []byte, client *client.ClientSockInfo) error {
// 	t := clientmessages.IgnoreGameEvents{}
// 	matched, err := protocol.Marshal(payload, &t)
// 	switch {
// 	case !matched:
// 		return errors.New("Failed to match")
// 	case err != nil:
// 		return err
// 	}

// 	user, err := ch.pm.GetUsernameForGame(client.GameName, t.IgnoredUsername)
// 	if err != nil {
// 		return err
// 	}

// 	return client.ClientInfo.Ignore(user.ClientInfo.GetAddr())
// }

func (ch *ChatHandler) setUsername(payload []byte, client *client.ClientSockInfo) error {
	t := clientmessages.SetUsername{}
	matched, err := protocol.Marshal(payload, &t)
	switch {
	case !matched:
		return errors.New("Failed to match")
	case err != nil:
		return err
	}

	client.ClientInfo.SetUsername(t.Username)

	return ch.pm.Broadcast(servermessages.ServerMessage{
		MessageType: "server",
		Message:     fmt.Sprintf("%s#%s has connected to the channel", t.Username, client),
	}, client)
}

func (ch *ChatHandler) sendUserMessage(payload []byte, client *client.ClientSockInfo) error {
	t := clientmessages.SendMessage{}
	matched, err := protocol.Marshal(payload, &t)
	switch {
	case !matched:
		return errors.New("Failed to match")
	case err != nil:
		return err
	}

	if client.ClientInfo.GetUsername() == "" {
		return errors.New("name not set, cannot send message")
	}

	return ch.pm.Broadcast(servermessages.UserMessage{
		Text: t.Message,
		Name: client.ClientInfo.GetUsername(),
	}, client)
}
