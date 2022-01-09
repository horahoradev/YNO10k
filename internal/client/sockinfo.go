package client

type ClientSockInfo struct {
	ServiceType ServiceType
	GameName    string
	RoomName    string
	ClientInfo  Client
	SyncObject  *SyncObject
}
