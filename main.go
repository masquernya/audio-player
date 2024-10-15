package main

import (
	"audio-player/ui"
	"fmt"
	"log"
	"net"
	"net/http"
	"net/rpc"
	"os"
)

type RpcServer struct {
	u *ui.UI
}

func (r *RpcServer) PlayAudio(audioFile *string, reply *int) error {
	fmt.Println("playing audio", audioFile)
	r.u.Run(*audioFile)
	return nil
}

func main() {
	audioFile := os.Args[1]
	fmt.Println("stating with", audioFile)
	portStr := "19837"

	client, err := rpc.DialHTTP("tcp", "127.0.0.1:"+portStr)
	if err != nil {
		// fail, so setup server.
		u := ui.New()
		rpcServer := &RpcServer{u: u}
		if err := rpc.Register(rpcServer); err != nil {
			log.Fatal("error registering rpc:", err)
		}
		// start server
		go (func() {
			go rpc.HandleHTTP()
			l, err := net.Listen("tcp", "127.0.0.1:"+portStr)
			if err != nil {
				log.Fatal("listen error:", err)
			}
			if err := http.Serve(l, nil); err != nil {
				log.Fatal("error serving:", err)
			}
		})()

		u.Run(audioFile)
	} else {
		var reply int
		if err := client.Call("RpcServer.PlayAudio", &audioFile, &reply); err != nil {
			log.Fatal("error calling rpc:", err)
		}
		log.Println("reply is", reply)
	}
}
