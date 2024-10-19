package server

import (
	"audio-player/ui"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"sync"
	"time"
)

type RpcServer struct {
	u *ui.UI

	nonces map[string]bool // not directly related to crypto. this is to ensure messages aren't processed twice.
	shared *Shared
	m      sync.Mutex
}

type PlayAudioRequest struct {
	AudioFilePath string
	CreatedAt     string
	Nonce         string
}

func (r *RpcServer) PlayAudio(encryptedRequest *string, reply *int) error {
	r.m.Lock()
	defer r.m.Unlock()

	decodedBits, err := hex.DecodeString(*encryptedRequest)
	if err != nil {
		log.Println("error decoding request:", err)
		return errors.New("error decoding request")
	}

	jsonRequest, err := r.shared.Decrypt(decodedBits)
	if err != nil {
		log.Println("error decrypting request:", err)
		return errors.New("error decrypting request")
	}

	var request PlayAudioRequest
	if err := json.Unmarshal(jsonRequest, &request); err != nil {
		log.Println("error unmarshalling request:", err)
		return errors.New("error unmarshalling request")
	}

	t, err := time.Parse(time.RFC3339, request.CreatedAt)
	if err != nil {
		log.Println("error parsing time:", err)
		return errors.New("error decrypting request")
	}

	if time.Since(t) > time.Second {
		log.Println("request is too old: ", request.CreatedAt, time.Since(t))
		return errors.New("error decrypting request")
	}

	if _, exists := r.nonces[request.Nonce]; exists {
		log.Println("nonce already exists")
		return errors.New("error decrypting request")
	}
	r.nonces[request.Nonce] = true

	fmt.Println("playing audio", request.AudioFilePath)
	r.u.Run(request.AudioFilePath)
	return nil
}

func NewRpcServer(u *ui.UI) *RpcServer {
	return &RpcServer{
		u:      u,
		shared: newShared(),
		nonces: make(map[string]bool),
	}
}
