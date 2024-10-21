package server

import (
	"audio-player/ui"
	"crypto/tls"
	"log"
	"net/rpc"
	"strconv"
)

type Client struct {
	u      *ui.UI
	client *rpc.Client
	s      *Shared
}

func NewClient() *Client {
	c := &Client{
		u: nil,
		s: newShared(),
	}
	return c
}

func (c *Client) TryConnect() bool {
	config := tls.Config{
		GetClientCertificate: func(*tls.CertificateRequestInfo) (*tls.Certificate, error) {
			cl := c.s.GetKeyPair(KeyPairModeClient)
			return &cl, nil
		},
		VerifyPeerCertificate: c.s.verifyCert(func() tls.Certificate {
			return c.s.GetKeyPair(KeyPairModeServer)
		}),
		InsecureSkipVerify: true, // still calls VerifyPeerCertificate (which is all we need)
	}
	conn, err := tls.Dial("tcp", "127.0.0.1:"+strconv.Itoa(c.s.GetPort()), &config)
	if err != nil {
		log.Println("error dialing:", err)
		return false
	}
	//defer conn.Close()
	//log.Println("client: connected to: ", conn.RemoteAddr())
	client := rpc.NewClient(conn)
	//client, err := rpc.DialHTTP("tcp", "127.0.0.1:"+strconv.Itoa(c.s.GetPort()))
	//if err != nil {
	//	return false
	//}

	c.client = client

	return true
}

func (c *Client) PlayAudio(audioFilePath string) error {
	var reply int
	if err := c.client.Call("RpcServer.PlayAudio", audioFilePath, &reply); err != nil {
		log.Fatal("error calling rpc:", err)
	}
	return nil
}
