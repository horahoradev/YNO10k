package main

import (
	"github.com/panjf2000/gnet"
	"github.com/panjf2000/gnet/pkg/pool/goroutine"

	"io"
	"log"
	"net"
	"time"

	"github.com/gobwas/ws"
	"github.com/gobwas/ws/wsutil"
)

type AsyncWS struct {
	gnet.Conn
}

func (ws AsyncWS) Read(p []byte) (n int, err error) {
	return ws.Read(p)
}

func newAsyncWS(c gnet.Conn) io.ReadWriter {
	return AsyncWS{
		Conn: c,
	}
}

func (ws AsyncWS) Write(p []byte) (n int, err error) {
	return len(p), ws.AsyncWrite(p)
}

type chatServer struct {
	*gnet.EventServer
	pool *goroutine.Pool
}

func (es *chatServer) React(frame []byte, c gnet.Conn) (out []byte, action gnet.Action) {
	asyncWS := newAsyncWS(c)

	_, err := ws.Upgrade(asyncWS)
	if err != nil {
		log.Errorf("Failed to upgrade websocket. Err: %s", err)
		return
	}

	header, err := ws.ReadHeader(asyncWS)
	if err != nil {
		log.Errorf("Failed to upgrade websocket. Err: %s", err)
		return
	}

	data := append([]byte{}, frame...)

	// Use ants pool to unblock the event-loop.
	_ = es.pool.Submit(func() {
		time.Sleep(1 * time.Second)
		c.AsyncWrite(data)
	})

	return
}

func main() {
	p := goroutine.Default()
	defer p.Release()

	echo := &chatServer{pool: p}
	log.Fatal(gnet.Serve(echo, "tcp://:9000", gnet.WithMulticore(true)))
}

/*
func main() {
	http.ListenAndServe(":8080", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Upgrade to a websocket
		conn, _, _, err := ws.UpgradeHTTP(r, w)
		if err != nil {
			log.Errorf("Failed to upgrade to websocket. Err: %s", err)
			return
		}
		go func() {
			defer conn.Close()

			for {
				msg, op, err := wsutil.ReadClientData(conn)
				if err != nil {
					// handle error
				}
				err = wsutil.WriteServerMessage(conn, op, msg)
				if err != nil {
					// handle error
				}
			}
		}()
	}))
}

func muxMessage(msg)

*/
