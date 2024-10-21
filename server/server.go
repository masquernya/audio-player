package server

import (
	"audio-player/ui"
	"crypto/tls"
	"errors"
	"log"
	"net/rpc"
	"os"
	"strconv"
	"time"
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
	serverCert := s.shared.GetKeyPair(KeyPairModeServer)

	// this is hacky, but I don't think we can dynamically edit TLS certificates.
	timeUntilExpire := serverCert.Leaf.NotAfter.Sub(serverCert.Leaf.NotBefore)
	go (func() {
		time.Sleep(timeUntilExpire)
		// TODO: ui alert would be nice
		log.Println("server cert expired, request exit so a new one can be generated")
		os.Exit(0)
	})()

	go (func() {
		config := tls.Config{
			Certificates: []tls.Certificate{
				serverCert,
			},
			InsecureSkipVerify: true, // still calls VerifyPeerCertificate (which is all we need)
			VerifyPeerCertificate: s.shared.verifyCert(func() tls.Certificate {
				return s.shared.GetKeyPair(KeyPairModeClient)
			}),
			VerifyConnection: func(state tls.ConnectionState) error {
				// If there was a cert, it was already verified in VerifyPeerCertificate.
				// Here, we just make sure at least one cert was sent.
				// If no certs were sent, VerifyPeerCertificate is not called, so that's why we do this here.
				if len(state.PeerCertificates) == 0 {
					return errors.New("no certificate specified")
				}
				return nil
			},
			ClientAuth: tls.RequireAnyClientCert,
		}
		listener, err := tls.Listen("tcp", "127.0.0.1:"+strconv.Itoa(s.shared.GetPort()), &config)
		if err != nil {
			log.Fatalf("server: listen: %s", err)
		}

		for {
			conn, err := listener.Accept()
			if err != nil {
				log.Println("error accepting connection", err)
				break
			}
			// no need for multithreading (yet...)
			rpc.ServeConn(conn)
			conn.Close()
		}

	})()

	return nil
}
