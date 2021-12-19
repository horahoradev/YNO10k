package main

import (
	"github.com/panjf2000/gnet"
	"github.com/panjf2000/gnet/pkg/pool/goroutine"

	log "github.com/sirupsen/logrus"
	"io"

	"github.com/gobwas/ws"
	guuid "github.com/google/uuid"
)

type AsyncWS struct {
	gnet.Conn
	uuid guuid.UUID
}

func (ws AsyncWS) Read(p []byte) (n int, err error) {
	return ws.Read(p)
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

type chatServer struct {
	*gnet.EventServer
	pool *goroutine.Pool
}

func (es *chatServer) OnOpened(c gnet.Conn) (out []byte, action gnet.Action) {
	asyncWS := newAsyncWS(c)
	_, err := ws.Upgrade(asyncWS)
	if err != nil {
		log.Errorf("Failed to upgrade websocket. Err: %s", err)
		return
	}

	return
}

// Shouldn't be called until after we upgrade the websocket, so this is safe
func (es *chatServer) React(frame []byte, c gnet.Conn) (out []byte, action gnet.Action) {
	asyncWS := newAsyncWS(c)

	// Use ants pool to unblock the event-loop.
	// This is a blocking thread pool, we don't want to loop infinitely and consume all workers
	_ = es.pool.Submit(func() {

		header, err := ws.ReadHeader(asyncWS)
		if err != nil {
			log.Errorf("Failed to upgrade websocket. Err: %s", err)
			return
		}

		payload := make([]byte, header.Length)
		_, err = io.ReadFull(asyncWS, payload)
		if err != nil {
			log.Errorf("Failed to read full message. Err: %s", err)
			return
		}

		// We're using the input header, FIXME
		if err := ws.WriteHeader(asyncWS, header); err != nil {
			log.Errorf("Failed to write response header. Err: %s", err)
			return
		}

		if _, err := asyncWS.Write(payload); err != nil {
			log.Errorf("Failed to write response payload. Err: %s", err)
			return
		}

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

	echo := &chatServer{
		pool: p,
	}
	log.Fatal(gnet.Serve(echo, "tcp://:9000", gnet.WithMulticore(true)))
}
