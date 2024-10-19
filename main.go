package main

import (
	"audio-player/gtime"
	"audio-player/server"
	"audio-player/ui"
	"audio-player/visu"
	"log"
	"os"
)

func main() {
	gtime.Start("main")
	audioFile := os.Args[1]

	gtime.Start("main.BeforeUI")
	client := server.NewClient()
	if client.TryConnect() {
		gtime.End("main.BeforeUI")
		if err := client.PlayAudio(audioFile); err != nil {
			log.Fatal("error playing audio:", err)
		}
	} else {
		// fail, so setup server.
		u := ui.New()
		s := server.New(u)
		if err := s.Start(); err != nil {
			log.Fatal("error starting server:", err)
		}
		// do this on server only (so two processes aren't trying to clear cache at once)
		go visu.ClearCache()

		gtime.End("main.BeforeUI") // under 1ms
		u.Run(audioFile)
	}
}
