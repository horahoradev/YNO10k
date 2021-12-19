package main

import (
	"github.com/horahoradev/YNO10k/internal/client"
	"github.com/horahoradev/YNO10k/internal/msghandler"
	"github.com/panjf2000/gnet"
	"github.com/panjf2000/gnet/pkg/pool/goroutine"

	log "github.com/sirupsen/logrus"
	"io"

	"github.com/gobwas/ws"
	guuid "github.com/google/uuid"
)

type AsyncWS struct {
	gnet.Conn
}

func (ws AsyncWS) Read(p []byte) (n int, err error) {
	return ws.ReadN(p)
}

func newAsyncWS(c gnet.Conn) io.ReadWriter {
	return AsyncWS{
		Conn: c,
		uuid: guuid.New(),
	}
}

func (ws AsyncWS) Write(p []byte) (n int, err error) {
	return len(p), ws.AsyncWrite(p)
}

type messageServer struct {
	*gnet.EventServer
	pool *goroutine.Pool
	cm   *client.ClientManager

	chatHandler *msghandler.ChatHandler
	gameHandler *msghandler.GameHandler
	listHandler *msghandler.ListHandler
}

func (es *messageServer) OnOpened(c gnet.Conn) (out []byte, action gnet.Action) {
	asyncWS := newAsyncWS(c)
	_, err := ws.Upgrade(asyncWS)
	if err != nil {
		log.Errorf("Failed to upgrade websocket. Err: %s", err)
		return
	}

	return
}

// Shouldn't be called until after we upgrade the websocket, so this is safe
func (ms *messageServer) React(frame []byte, c gnet.Conn) (out []byte, action gnet.Action) {
	asyncWS := newAsyncWS(c)

	var clientInfo *client.ClientSockInfo
	ctx := c.Context()
	if ctx != nil {
		var ok bool
		clientInfo, ok = ctx.(*client.ClientSockInfo)
		if !ok {
			log.Errorf("Failed to typecast context to client")
			c.Close()
			return
		}
	}

	// Use ants pool to unblock the event-loop.
	// This is a blocking thread pool, we don't want to loop infinitely and consume all workers
	// It's assumed that all messages will arrive in a single tcp packet, but that's required by the websocket protocol
	_ = ms.pool.Submit(func() {
		header, err := ws.ReadHeader(asyncWS)
		if err != nil {
			log.Errorf("Failed to upgrade websocket. Err: %s", err)
			return
		}

		clientPayload := make([]byte, header.Length)
		_, err = io.ReadFull(asyncWS, clientPayload)
		if err != nil {
			log.Errorf("Failed to read full message. Err: %s", err)
			return
		}

		switch {
		case clientInfo == nil:
			// This is the servicename packet, use it to initialize the client info
			clientInfo, err := ms.cm.AddClientForRoom(string(clientPayload), c)
			if err != nil {
				log.Errorf("Failed to add client for room. Err: %s", err)
			}

			// Store the client info with the connection
			c.SetContext(clientInfo)

		default:
			// We've already received the service packet, so this is regular message
			switch clientInfo.ServiceType {
			case client.GlobalChat, client.Chat:
				err = ms.chatHandler.HandleMessage(clientPayload)
				if err != nil {
					log.Errorf("chat handler failed to handle message: %s", err)
					return
				}
			case client.Game:
				err = ms.gameHandler.HandleMessage(clientPayload)
				if err != nil {
					log.Errorf("game handler failed to handle message: %s", err)
					return
				}
			case client.List:
				err = ms.listHandler.HandleMessage(clientPayload)
				if err != nil {
					log.Errorf("list handler failed to handle message: %s", err)
					return
				}
			default:
				log.Errorf("Could not handle message, client socket servicetype was not set")
			}
		}

		//// We're using the input header, FIXME
		//if err := ws.WriteHeader(asyncWS, header); err != nil {
		//	log.Errorf("Failed to write response header. Err: %s", err)
		//	return
		//}

		//if _, err := asyncWS.Write(payload); err != nil {
		//	log.Errorf("Failed to write response payload. Err: %s", err)
		//	return
		//}

		if header.OpCode == ws.OpClose {
			c.Close()
			return
		}
	})

	return
}

func main() {
	p := goroutine.Default()
	defer p.Release()

	echo := &messageServer{
		pool: p,
	}
	log.Fatal(gnet.Serve(echo, "tcp://:9000", gnet.WithMulticore(true)))
}
