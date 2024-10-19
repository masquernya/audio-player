package server

import (
	"audio-player/ui"
	"log"
	"net"
	"net/http"
	"net/rpc"
	"strconv"
)

const (
	DefaultPort = 19837
)

type Server struct {
	u      *ui.UI
	shared *Shared
}

func New(ui *ui.UI) *Server {
	s := &Server{
		u:      ui,
		shared: newShared(),
	}
	return s
}

func (s *Server) Start() error {
	rpcServer := NewRpcServer(s.u)
	if err := rpc.Register(rpcServer); err != nil {
		log.Fatal("error registering rpc:", err)
	}
	// start server
	go (func() {
		go rpc.HandleHTTP()
		l, err := net.Listen("tcp", "127.0.0.1:"+strconv.Itoa(s.shared.GetPort()))
		if err != nil {
			log.Fatal("listen error:", err)
		}
		if err := http.Serve(l, nil); err != nil {
			log.Fatal("error serving:", err)
		}
	})()

	return nil
}
