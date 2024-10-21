package server

import (
	"audio-player/ui"
	"fmt"
	"strings"
	"sync"
)

type RpcServer struct {
	u *ui.UI
	m sync.Mutex
}

func (r *RpcServer) PlayAudio(audioFilePath *string, reply *int) error {
	r.m.Lock()
	defer r.m.Unlock()

	fmt.Println("playing audio", *audioFilePath)
	r.u.Run(strings.Clone(*audioFilePath))
	return nil
}

func NewRpcServer(u *ui.UI) *RpcServer {
	return &RpcServer{
		u: u,
	}
}
