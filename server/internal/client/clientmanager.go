package client

import (
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"sync"

	log "github.com/sirupsen/logrus"

	"github.com/panjf2000/gnet"
)

type ClientPubsubManager struct {
	// Game name : client info
	gameClientMap map[string]GameClientInfo
	lock          *sync.RWMutex
}

func NewClientPubsubManager() *ClientPubsubManager {
	return &ClientPubsubManager{
		gameClientMap: make(map[string]GameClientInfo),
		lock:          &sync.RWMutex{},
	}
}

type GameClientInfo struct {
	// Room name: client info
	clientRoomRemoteAddrMap map[string][]*ClientSockInfo
}

func (cm *ClientPubsubManager) GetUsernameForGame(game, room, username string) (*ClientSockInfo, error) {
	gameClientInfo, ok := cm.gameClientMap[game]
	if !ok {
		return nil, fmt.Errorf("Failed to lookup client info for game %s", game)
	}

	roomClientInfo, ok := gameClientInfo.clientRoomRemoteAddrMap[room]
	if !ok {
		return nil, fmt.Errorf("Failed to lookup client info for room %s game %s", room, game)
	}

	// O(N) :thinking:
	// TODO: only returns first usrename that matches
	// but no uniqueness check. need to have unique usernames
	for _, client := range roomClientInfo {
		if client.ClientInfo.GetUsername() == username {
			return client, nil
		}
	}

	return nil, errors.New("No matching client found")
}

func (cm *ClientPubsubManager) SubscribeClientToRoom(serviceName string, conn gnet.Conn) (*ClientSockInfo, error) {
	// TODO: cleanup old sock addr in clientRemoteAddrMap

	// Split provided servicename into something we can use
	gameName, roomName, err := SplitServiceName(serviceName)
	if err != nil {
		return nil, err
	}

	serviceType, err := GetTypeFromRoomName(roomName)
	if err != nil {
		return nil, err
	}

	// Does the game client manager exist? if not, create
	cm.lock.Lock()
	gameServ, ok := cm.gameClientMap[gameName]
	if !ok {
		gameServ = GameClientInfo{
			clientRoomRemoteAddrMap: make(map[string][]*ClientSockInfo),
		}

		cm.gameClientMap[gameName] = gameServ
	}
	cm.lock.Unlock()

	// TODO: initialize second map

	// Does the client info already exist somewhere? If so, move it
	// Otherwise, just add it
	sockInfo := &ClientSockInfo{
		ClientInfo:  newClient(serviceType, conn),
		ServiceType: serviceType,
		GameName:    gameName,
		RoomName:    roomName,
		SyncObject:  NewSyncObject(),
	}

	// Deletion in old room will be handled in an async worker
	cm.lock.Lock()
	gameServ.clientRoomRemoteAddrMap[roomName] = append(gameServ.clientRoomRemoteAddrMap[roomName], sockInfo)
	cm.lock.Unlock()

	if serviceType == Game {
		cm.lock.RLock()
		players := gameServ.clientRoomRemoteAddrMap[roomName]
		cm.lock.RUnlock()
		// Send client all player room state
		for _, player := range players {
			syncChanges := player.SyncObject.GetAllChanges()
			syncPayload, err := json.Marshal(&syncChanges)
			if err != nil {
				log.Errorf("Could not get initial player state. Err: %s", err)
			}

			// Send the new client everyone's state
			err = sockInfo.ClientInfo.SendFromPlayer(syncPayload, player.ClientInfo.GetAddr())
			if err != nil {
				log.Errorf("Failed to send initial state from player. Err: %s", err)
			}

			// Send the client's state to everyone
			// syncChanges = sockInfo.SyncObject.GetAllChanges()
			// syncPayload, err = json.Marshal(&syncChanges)
			// if err != nil {
			// 	log.Errorf("Could not get initial player state. Err: %s", err)
			// }

			// err = player.ClientInfo.SendFromPlayer(syncPayload, sockInfo.ClientInfo.GetAddr())
			// if err != nil {
			// 	log.Errorf("Failed to send initial state from player. Err: %s", err)
			// }
		}
	}

	return sockInfo, nil
}

// Splits the service name into constituent parts
func SplitServiceName(serviceName string) (gameName, serviceType string, err error) {
	// ^([a-zA-Z\d]*)
	// THIS IS A TODO, just did a hacky fix here
	validID := regexp.MustCompile(`^(gchat|game\d*|chat\d*)\z`)
	rs := validID.FindStringSubmatch(serviceName)

	switch {
	case len(rs) > 0 && len(rs[0]) > 0:
		return "yumenikki", rs[0], nil

	default:
		return "", "", fmt.Errorf("invalid servicename pattern for %s", serviceName)
	}
}

// TODO: refactor here too
func (cm *ClientPubsubManager) Broadcast(payload interface{}, sockinfo *ClientSockInfo, fromServer bool) error {
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	cm.lock.RLock()
	gameClientInfo, ok := cm.gameClientMap[sockinfo.GameName]
	cm.lock.RUnlock()
	if !ok {
		return fmt.Errorf("Failed to broadcast, could not find game client info for game name %s", sockinfo.GameName)
	}

	cm.lock.RLock()
	clients, ok := gameClientInfo.clientRoomRemoteAddrMap[sockinfo.RoomName]
	cm.lock.RUnlock()
	if !ok {
		return fmt.Errorf("Failed to broadcast, could not find client list for room name %s", sockinfo.RoomName)
	}

	log.Debugf("Broadcasting for room %s, clients: %s", sockinfo.RoomName, len(clients))
	for _, client := range clients {
		switch fromServer {
		case true:
			err = client.ClientInfo.SendFromServer(payloadBytes)
			if err != nil {
				log.Errorf("Failed to send to client from server. Err: %s. Continuing...", err)
			}
		case false:
			c := sockinfo.ClientInfo
			if c == nil {
				// Skip? Disconnect? TODO
			}
			err = client.ClientInfo.SendFromPlayer(payloadBytes, c.GetAddr())
			if err != nil {
				log.Errorf("Failed to send to client. Err: %s. Continuing...", err)
			}
		}
	}

	return nil
}

type ServiceType string

const (
	GlobalChat ServiceType = "GlobalChat"
	Chat       ServiceType = "Chat"
	Game       ServiceType = "Game"
	List       ServiceType = "List"
)

func GetTypeFromRoomName(roomname string) (ServiceType, error) {
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
