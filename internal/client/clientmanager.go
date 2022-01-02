package client

import (
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"strings"

	"log"

	guuid "github.com/google/uuid"
	"github.com/panjf2000/gnet"
)

func newClient() *Client {
	return &Client{
		GameIgnores: make([]guuid.UUID, 0),
		ChatIgnores: make([]guuid.UUID, 0),
	}
}

type ClientPubsubManager struct {
	// Game name : client info
	gameClientMap map[string]GameClientInfo
}

type GameClientInfo struct {
	// Room name: remote addr without port : client info
	clientRoomRemoteAddrMap map[string]map[string]*Client

	// remote addr without port: client info
	clientRemoteAddrMap map[string]*Client
}

func (cm *ClientPubsubManager) SubscribeClientToRoom(serviceName string, conn gnet.Conn) (*ClientSockInfo, error) {
	// TODO: cleanup old sock addr in clientRemoteAddrMap

	// Split provided servicename into something we can use
	gameName, roomName, err := cm.splitServiceName(serviceName)
	if err != nil {
		return nil, err
	}

	serviceType, err := getTypeFromRoomName(roomName)
	if err != nil {
		return nil, err
	}

	// Does the game client manager exist? if not, create
	gameServ, ok := cm.gameClientMap[gameName]
	if !ok {
		cm.gameClientMap[gameName] = GameClientInfo{
			clientRoomRemoteAddrMap: make(map[string]map[string]*Client),
		}
	}

	// Replace relevant socket (or just assign if need be)
	err = cm.replaceClientSock(gameName, serviceName, roomName, conn)
	if err != nil {
		return nil, err
	}

	// Does the client info already exist somewhere? If so, move it
	// Otherwise, just add it
	ip := getIPFromConn(conn)
	k, ok := gameServ.clientRemoteAddrMap[ip]

	switch ok {
	case true:
		k = newClient()
	default:
		// do nothing
	}

	// Deletion in old room will be handled in an async worker
	gameServ.clientRoomRemoteAddrMap[roomName][ip] = k
	return &ClientSockInfo{
		ClientInfo:  k,
		ServiceType: serviceType,
		GameName:    gameName,
	}, nil
}

// Splits the service name into constituent parts
func (cm *ClientPubsubManager) splitServiceName(serviceName string) (gameName, serviceType string, err error) {
	validID := regexp.MustCompile(`^([a-zA-Z\d]*)(gchat|game\d*|chat\d*)\z`)
	rs := validID.FindStringSubmatch(serviceName)

	switch {
	case len(rs) > 0:
		return rs[0], rs[1], nil

	default:
		return "", "", fmt.Errorf("invalid servicename pattern")
	}
}

//func (cm *ClientPubsubManager) RetrieveClientInfo(serviceName string, conn gnet.Conn) (c *Client, ok bool) {
//	// Split provided servicename into something we can use
//	gameName, _, err := cm.splitServiceName(serviceName)
//	if err != nil {
//		log.Errorf("Failed to split service name. Err: %s", err)
//		return nil, false
//	}
//
//	gs, ok := cm.gameClientMap[gameName]
//	if !ok {
//		return nil, ok
//	}
//
//	c, ok = gs.clientRemoteAddrMap[getIPFromConn(conn)]
//}

func (cm *ClientPubsubManager) Broadcast(payload interface{}, sockinfo ClientSockInfo, broadcastType ServiceType) error {
	clients, ok := cm.gameClientMap[sockinfo.GameName]
	if !ok {
		return fmt.Errorf("Failed to broadcast, could not find clients for game name %s", sockinfo.GameName)
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	roomName := ""
	switch broadcastType {
	case GlobalChat:
		roomName = "gchat"
	case Chat:
		roomName = sockinfo.ClientInfo.ChatRoomID
	case Game:
		roomName = sockinfo.ClientInfo.GameRoomID
	default:
		return errors.New("Invalid broadcast type provided")
	}

OUTER:
	for _, client := range clients.clientRoomRemoteAddrMap[roomName] {
		// If the target is the sender
		if client.UUID == sockinfo.ClientInfo.UUID {
			continue
		}

		// Handle ignores
		var ignores []guuid.UUID
		switch broadcastType {
		case GlobalChat, Chat:
			ignores = sockinfo.ClientInfo.ChatIgnores
		case Game:
			ignores = sockinfo.ClientInfo.GameIgnores
		}

		for _, ignoredUUID := range ignores {
			if client.UUID == ignoredUUID {
				continue OUTER
			}
		}

		err = client.GlobalChatSocket.AsyncWrite(payloadBytes)
		if err != nil {
			log.Errorf("Failed to async write global chat message. Err: %s. Continuing...", err)
		}
	}
}

func (cm *ClientPubsubManager) replaceClientSock(gameName, serviceType, roomName string, conn gnet.Conn) error {
	gs, ok := cm.gameClientMap[gameName]
	if !ok {
		return fmt.Errorf("Failed to retrieve game server while replacing client socket")
	}

	clientIP := getIPFromConn(conn)

	c, ok := gs.clientRemoteAddrMap[clientIP]
	if !ok {
		return fmt.Errorf("failed to retrieve client info while replacing client socket")
	}

	t, err := getTypeFromRoomName(serviceType)
	if err != nil {
		return err
	}

	switch t {
	case GlobalChat:
		if c.GlobalChatSocket != nil {
			c.GlobalChatSocket.Close()
		}
		c.GlobalChatSocket = conn
	case Chat:
		if c.RoomChatSocket != nil {
			c.RoomChatSocket.Close()
		}
		c.RoomChatSocket = conn
		c.ChatRoomID = roomName
	case Game:
		if c.GameEventSocket != nil {
			c.GameEventSocket.Close()
		}
		c.GameEventSocket = conn
		c.GameRoomID = roomName
	default:
		return fmt.Errorf("Invalid service type prefix in socket replacement: %s", serviceType)
	}

	return nil
}

func getIPFromConn(conn gnet.Conn) string {
	return strings.Split(conn.RemoteAddr().String(), ":")[0]
}

type ServiceType string

const (
	GlobalChat ServiceType = "GlobalChat"
	Chat       ServiceType = "Chat"
	Game       ServiceType = "Game"
	List       ServiceType = "List"
)

func getTypeFromRoomName(roomname string) (ServiceType, error) {
	// FIXME
	// This code block is a little sus, come back to later
	switch {
	case strings.HasPrefix(roomname, "gchat"):
		return GlobalChat, nil
	case strings.HasPrefix(roomname, "chat"):
		return Chat, nil
	case strings.HasPrefix(roomname, "list"):
		return List, nil
	case strings.HasPrefix(roomname, "game"):
		return Game, nil
	default:
		return "", fmt.Errorf("Invalid service type prefix in socket replacement: %s", roomname)
	}
}
