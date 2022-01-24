package main

import (
	"bytes"
	"context"
	"net/http"
	"strings"
	"sync"

	"github.com/horahoradev/YNO10k/internal/client"
	ynmetrics "github.com/horahoradev/YNO10k/internal/metrics"
	"github.com/horahoradev/YNO10k/internal/msghandler"
	"github.com/panjf2000/gnet"
	"github.com/panjf2000/gnet/pkg/pool/goroutine"

	"io"

	log "github.com/sirupsen/logrus"

	"github.com/gobwas/ws"
)

type AsyncWriterWS struct {
	Conn gnet.Conn
}

func newAyncWriter(c gnet.Conn) io.Writer {
	return AsyncWS{
		Conn: c,
	}
}

func (ws AsyncWriterWS) Write(p []byte) (n int, err error) {
	return len(p), ws.Conn.AsyncWrite(p)
}

type AsyncWS struct {
	Conn gnet.Conn
	Buf  io.Reader
}

func (ws AsyncWS) Read(p []byte) (n int, err error) {
	return ws.Buf.Read(p)
}

func newAsyncWS(c gnet.Conn, buf io.Reader) io.ReadWriter {
	return AsyncWS{
		Conn: c,
		Buf:  buf,
	}
}

func (ws AsyncWS) Write(p []byte) (n int, err error) {
	return len(p), ws.Conn.AsyncWrite(p)
}

type messageServer struct {
	*gnet.EventServer
	pool       *goroutine.Pool
	serviceMux msghandler.Handler
	m          *sync.Mutex
}

func newMessageServer(pool *goroutine.Pool) messageServer {
	ps := client.NewClientPubsubManager()

	lh := msghandler.NewListHandler()
	ch := msghandler.NewChatHandler(ps)
	gh := msghandler.NewGameHandler(ps)
	sMux := msghandler.NewServiceMux(gh, ch, lh, ps)
	return messageServer{
		serviceMux:  &sMux,
		pool:        pool,
		EventServer: &gnet.EventServer{},
		m:           &sync.Mutex{},
	}
}

type gnetWrapper struct {
	gnet.Conn
}

func (g *gnetWrapper) AsyncWrite(buf []byte) error {
	respFrame := ws.NewFrame(ws.OpText, true, buf)
	asyncWriter := newAyncWriter(g.Conn)
	return ws.WriteFrame(asyncWriter, respFrame)
}

// Shouldn't be called until after we upgrade the websocket, so this is safe
func (ms messageServer) React(frame []byte, c gnet.Conn) (out []byte, action gnet.Action) {
	asyncWS := newAsyncWS(c, bytes.NewReader(frame))
	log.Debugf("Reacting to new packet: %s", string(frame))

	// OK this is super lame but whatever
	if strings.Contains(string(frame), "Upgrade") {
		_, err := ws.Upgrade(asyncWS)
		if err != nil {
			log.Errorf("Failed to upgrade websocket. Err: %s", err)
			return
		}
		return
	}

	// Use ants pool to unblock the event-loop.
	// This is a blocking thread pool, we don't want to loop infinitely and consume all workers
	// It's assumed that all messages will arrive in a single tcp packet, but that's required by the websocket protocol
	err1 := ms.pool.Submit(func() {
		wsFrame, err := ws.ReadFrame(asyncWS)
		if err != nil {
			log.Errorf("Failed to read frame. Err: %s", err)
			return
		}

		if wsFrame.Header.Masked {
			wsFrame = ws.UnmaskFrame(wsFrame)
		}

		if wsFrame.Header.OpCode == ws.OpClose {
			log.Print("Received close opcode, closing connection")
			c.Close()
			return
		}

		err = ms.serviceMux.HandleMessage(wsFrame.Payload, &gnetWrapper{Conn: c}, nil)
		if err != nil {
			log.Errorf("Could not handle client message. Err: %s", err)
		}
	})
	if err1 != nil {
		log.Errorf("Failed to submit job to worker pool. Err: %s", err1)
	}

	return nil, gnet.None
}

func (ms messageServer) OnClosed(c gnet.Conn, err error) (action gnet.Action) {
	log.Errorf("Closing connection, err: %s. State: %s. ", err)
	// Send a disconnect broadcast LOL
	err1 := ms.serviceMux.HandleMessage([]byte("DC"), &gnetWrapper{Conn: c}, nil)
	if err1 != nil {
		log.Errorf("Failed to handle disconnect message. Err: %s", err)
	}
	return
}

func main() {
	p := goroutine.Default()
	defer p.Release()

	go func() {
		http.Handle("/", http.FileServer(http.Dir("public/")))
		log.Fatal(http.ListenAndServe("0.0.0.0:8085", nil))
	}()

	ynmetrics.StartExporter(context.Background())

	mServ := newMessageServer(p)
	log.Print("Listening on 443")
	log.Fatal(gnet.Serve(mServ, "0.0.0.0:443", gnet.WithNumEventLoop(1), gnet.WithMulticore(false)))
}
