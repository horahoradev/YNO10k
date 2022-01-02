package client

import (
	guuid "github.com/google/uuid"
	"github.com/panjf2000/gnet"
)

type Client struct {
	Name     string
	Tripcode string
	UUID     guuid.UUID

	// Roughly the same for chat and game events
	ChatRoomID string
	GameRoomID string

	// O(N) for search but list will be small and cache friendly
	GameIgnores []guuid.UUID
	ChatIgnores []guuid.UUID

	GameEventSocket  gnet.Conn
	RoomChatSocket   gnet.Conn
	GlobalChatSocket gnet.Conn
}
