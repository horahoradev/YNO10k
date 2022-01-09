package main

import (
	"github.com/horahoradev/YNO10k/internal/client"
	"github.com/horahoradev/YNO10k/internal/msghandler"
	"github.com/panjf2000/gnet"
	"github.com/panjf2000/gnet/pkg/pool/goroutine"

	"io"

	log "github.com/sirupsen/logrus"

	"github.com/gobwas/ws"
)

type AsyncWS struct {
	gnet.Conn
}

func (ws AsyncWS) Read(p []byte) (n int, err error) {
	p = ws.Conn.Read()
	return len(p), nil
}

func newAsyncWS(c gnet.Conn) io.ReadWriter {
	return AsyncWS{
		Conn: c,
	}
}

func (ws AsyncWS) Write(p []byte) (n int, err error) {
	return len(p), ws.AsyncWrite(p)
}

type messageServer struct {
	*gnet.EventServer
	pool       *goroutine.Pool
	serviceMux msghandler.Handler
}

func newMessageServer(pool *goroutine.Pool) messageServer {
	ps := client.NewClientPubsubManager()

	lh := msghandler.NewListHandler()
	ch := msghandler.NewChatHandler(ps)
	gh := msghandler.NewGameHandler(ps)
	sMux := msghandler.NewServiceMux(gh, ch, lh, ps)
	return messageServer{
		serviceMux:  sMux,
		pool:        pool,
		EventServer: &gnet.EventServer{},
	}
}

func (es messageServer) OnOpened(c gnet.Conn) (out []byte, action gnet.Action) {
	asyncWS := newAsyncWS(c)
	_, err := ws.Upgrade(asyncWS)
	if err != nil {
		log.Errorf("Failed to upgrade websocket. Err: %s", err)
		return
	}

	return
}

// Shouldn't be called until after we upgrade the websocket, so this is safe
func (ms messageServer) React(frame []byte, c gnet.Conn) (out []byte, action gnet.Action) {
	asyncWS := newAsyncWS(c)

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

		if header.OpCode == ws.OpClose {
			c.Close()
			return
		}

		err = ms.serviceMux.HandleMessage(clientPayload, c, nil)
		if err != nil {
			log.Errorf("Could not handle client message. Err: %s", err)
		}
	})

	return
}

func main() {
	p := goroutine.Default()
	defer p.Release()

	mServ := newMessageServer(p)
	log.Fatal(gnet.Serve(mServ, "tcp://:9000", gnet.WithMulticore(true)))
}
