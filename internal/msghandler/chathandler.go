package msghandler

import (
	"errors"
	"fmt"

	guuid "github.com/google/uuid"
	"github.com/horahoradev/YNO10k/internal/client"
	"github.com/horahoradev/YNO10k/internal/clientmessages"
	"github.com/horahoradev/YNO10k/internal/messages"
	"github.com/horahoradev/YNO10k/internal/protocol"
	"github.com/horahoradev/YNO10k/internal/servermessages"
	"github.com/panjf2000/gnet"
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

func (ch *ChatHandler) HandleMessage(payload []byte, c gnet.Conn, s *client.ClientSockInfo) error {
	return ch.muxMessage(payload, c, s)
}

func (ch *ChatHandler) muxMessage(payload []byte, c gnet.Conn, s *client.ClientSockInfo) error {
	if len(payload) == 0 {
		return errors.New("Payload cannot be empty!")
	}

	switch payload[0] {
	case pardonChat:
		return ch.pardonChat(payload, s)
	case pardonGame:
		return ch.pardonGame(payload, s)
	case ignoreChat:
		return ch.ignoreChat(payload, s)
	case ignoreGame:
		return ch.ignoreGame(payload, s)
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
	t := messages.UnignoreChatEvents{}
	matched, err := protocol.Marshal(payload, &t)
	switch {
	case !matched:
		return errors.New("Failed to match")
	case err != nil:
		return err
	}

	return client.UnignoreChatEvents(t)
}

func (ch *ChatHandler) pardonGame(payload []byte, client *client.ClientSockInfo) error {
	t := messages.UnignoreGameEvents{}
	matched, err := protocol.Marshal(payload, &t)
	switch {
	case !matched:
		return errors.New("Failed to match")
	case err != nil:
		return err
	}

	return client.UnignoreGameEvents(t)
}

func (ch *ChatHandler) ignoreChat(payload []byte, client *client.ClientSockInfo) error {
	t := messages.IgnoreChatEvents{}
	matched, err := protocol.Marshal(payload, &t)
	switch {
	case !matched:
		return errors.New("Failed to match")
	case err != nil:
		return err
	}

	return client.IgnoreChatEvents(t)
}

func (ch *ChatHandler) ignoreGame(payload []byte, client *client.ClientSockInfo) error {
	t := messages.IgnoreGameEvents{}
	matched, err := protocol.Marshal(payload, &t)
	switch {
	case !matched:
		return errors.New("Failed to match")
	case err != nil:
		return err
	}

	return client.IgnoreGameEvents(t)
}

func (ch *ChatHandler) setUsername(payload []byte, client *client.ClientSockInfo) error {
	t := clientmessages.SetUsername{}
	matched, err := protocol.Marshal(payload, &t)
	switch {
	case !matched:
		return errors.New("Failed to match")
	case err != nil:
		return err
	}

	// guuid := guuid.New()

	client.ClientInfo.Name = t.Username
	// client.ClientInfo.UUID = guuid

	return ch.pm.Broadcast(servermessages.ServerMessage{
		MessageType: "server",
		Message:     fmt.Sprintf("%s#%s has connected to the channel", t.Username, guuid.String()),
	})
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

	if client.ClientInfo.Name == "" {
		return errors.New("name not set, cannot send message")
	}

	return ch.pm.Broadcast(servermessages.UserMessage{
		Text: t.Message,
		Name: client.ClientInfo.Name,
		Trip: client.ClientInfo.UUID.String(),
	})
}
