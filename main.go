package main

import (
	"audio-player/audio"
	"audio-player/ui"
	"audio-player/visu"
	"bytes"
	"fmt"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
	"image/color"
	"log"
	"os"
	"path"
)

func playAudioSync(a *audio.Audio, s *ui.LayoutMain, bar fyne.CanvasObject, pos float32, dur float32) {
	// setup cursor render
	s.Play(bar, pos, dur)
	// play
	if err := a.Start(float64(pos)); err != nil {
		log.Println("error starting audio:", err)
	}
}

func main() {
	audioFile := os.Args[1]
	fmt.Println("stating with", audioFile)

	l := &ui.LayoutMain{
		PlaybackPercent: 0,
	}

	a := audio.New(audioFile)
	dur := a.Duration()

	imageBits, err := visu.GenerateImage(audioFile)
	if err != nil {
		log.Fatal("error generating image:", err)
	}

	app := app.New()
	w := app.NewWindow(path.Base(audioFile))
	var ct *fyne.Container
	cursor := canvas.NewRectangle(color.White)

	image := canvas.NewImageFromReader(bytes.NewReader(imageBits), "waveform")
	cursorPositioner := ui.NewClickableInvisible(func(event *fyne.PointEvent) {
		log.Println("clicked at", event.Position)
		// translate to duration
		percent := event.Position.X / float32(image.Size().Width)
		pos := dur * percent
		log.Println("seeking to", pos)

		playAudioSync(a, l, cursor, pos, dur)
	})

	replayButton := widget.NewButton("Replay", func() {
		playAudioSync(a, l, cursor, 0, dur)
	})

	closeButton := widget.NewButton("Close", func() {
		a.Stop()
		app.Quit()
	})

	ct = container.New(l, image, cursorPositioner, replayButton, closeButton, cursor)

	playAudioSync(a, l, cursor, 0, dur)

	w.SetContent(ct)
	w.ShowAndRun()

	log.Println("app exit")
	a.Stop()
}
