package client

import (
	guuid "github.com/google/uuid"
	"github.com/horahoradev/YNO10k/internal/messages"
)

type ClientSockInfo struct {
	ServiceType ServiceType
	ClientInfo  *Client
	SyncObject  *SyncObject
}

// Not idempotent
func (csi *ClientSockInfo) IgnoreGameEvents(req messages.IgnoreGameEvents) error {
	uuid, err := guuid.FromBytes([]byte(req.IgnoredUUID))
	if err != nil {
		return err
	}

	// Can have duplicates... but it's fine
	csi.ClientInfo.GameIgnores = append(csi.ClientInfo.GameIgnores, uuid)
	return nil
}

func (csi *ClientSockInfo) IgnoreChatEvents(req messages.IgnoreChatEvents) error {
	uuid, err := guuid.FromBytes([]byte(req.IgnoredUUID))
	if err != nil {
		return err
	}

	// Can have duplicates... but it's fine
	csi.ClientInfo.ChatIgnores = append(csi.ClientInfo.ChatIgnores, uuid)
	return nil
}

// Technically this method of deletion is a slow memory leak
// Will fix... later
// This is also O(N), maybe better to just use a map... but arrays are cache friendly
func (csi *ClientSockInfo) UnignoreChatEvents(req messages.UnignoreChatEvents) error {
	uuid, err := guuid.FromBytes([]byte(req.UnignoredUUID))
	if err != nil {
		return err
	}
	for i, ignoredUUID := range csi.ClientInfo.ChatIgnores {
		if ignoredUUID == uuid {
			// Replace with blank UUID, LOL
			csi.ClientInfo.ChatIgnores[i] = guuid.UUID{}
		}
	}

	return nil
}

func (csi *ClientSockInfo) UnignoreGameEvents(req messages.UnignoreGameEvents) error {
	uuid, err := guuid.FromBytes([]byte(req.UnignoredUUID))
	if err != nil {
		return err
	}
	for i, ignoredUUID := range csi.ClientInfo.GameIgnores {
		if ignoredUUID == uuid {
			// Replace with blank UUID, LOL
			csi.ClientInfo.GameIgnores[i] = guuid.UUID{}
		}
	}

	return nil
}
