package server

import (
	"audio-player/ui"
	crypto_rand "crypto/rand"
	"encoding/hex"
	"encoding/json"
	"log"
	"net/rpc"
	"strconv"
	"time"
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
	client, err := rpc.DialHTTP("tcp", "127.0.0.1:"+strconv.Itoa(c.s.GetPort()))
	if err != nil {
		return false
	}

	c.client = client

	return true
}

func (c *Client) PlayAudio(audioFilePath string) error {
	nonce := make([]byte, 32)
	if _, err := crypto_rand.Read(nonce); err != nil {
		log.Fatal("error generating nonce:", err)
	}

	request := &PlayAudioRequest{
		AudioFilePath: audioFilePath,
		CreatedAt:     time.Now().Format(time.RFC3339),
		Nonce:         hex.EncodeToString(nonce),
	}
	bits, err := json.Marshal(request)
	if err != nil {
		log.Fatal("error marshalling request:", err)
	}

	requestStr := hex.EncodeToString(c.s.Encrypt(bits))
	var reply int
	if err := c.client.Call("RpcServer.PlayAudio", &requestStr, &reply); err != nil {
		log.Fatal("error calling rpc:", err)
	}
	return nil
}
