package ui

import (
	"audio-player/audio"
	"audio-player/visu"
	"bytes"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
	"image/color"
	"log"
	"path"
)

type UI struct {
	audioFile string

	a   *audio.Audio
	w   fyne.Window
	app fyne.App
}

func playAudioSync(a *audio.Audio, s *LayoutMain, bar fyne.CanvasObject, pos float32, dur float32) {
	// setup cursor render
	s.Play(bar, pos, dur)
	// play
	if err := a.Start(float64(pos)); err != nil {
		log.Println("error starting audio:", err)
	}
}

func New() *UI {
	u := &UI{}
	return u
}

func (u *UI) Run(audioFile string) error {
	// before start, clean up old stuff
	if u.a != nil {
		u.a.Stop()
	}

	u.audioFile = audioFile
	u.a = audio.New(u.audioFile)
	dur := u.a.Duration()

	imageBits, err := visu.GenerateImage(u.audioFile)
	if err != nil {
		log.Fatal("error generating image:", err)
	}

	l := &LayoutMain{}

	didMakeApp := false
	if u.app == nil {
		u.app = app.New()
		didMakeApp = true
	}
	if u.w == nil {
		u.w = u.app.NewWindow(path.Base(u.audioFile))
		u.w.Resize(fyne.NewSize(800, 200))
	}
	var ct *fyne.Container
	cursor := canvas.NewRectangle(color.White)

	image := canvas.NewImageFromReader(bytes.NewReader(imageBits), "waveform_"+u.audioFile)
	cursorPositioner := NewClickableInvisible(func(event *fyne.PointEvent) {
		log.Println("clicked at", event.Position)
		// translate to duration
		percent := event.Position.X / float32(image.Size().Width)
		pos := dur * percent
		log.Println("seeking to", pos)

		playAudioSync(u.a, l, cursor, pos, dur)
	})

	replayButton := widget.NewButton("Replay", func() {
		playAudioSync(u.a, l, cursor, 0, dur)
	})

	closeButton := widget.NewButton("Close", func() {
		u.a.Stop()
		u.app.Quit()
	})

	ct = container.New(l, image, cursorPositioner, replayButton, closeButton, cursor)

	playAudioSync(u.a, l, cursor, 0, dur)

	u.w.SetContent(ct)
	if didMakeApp {
		u.w.ShowAndRun()
		log.Println("app exit")
		u.a.Stop()
	} else {
		// new song playing, so focus window
		u.w.RequestFocus()
	}

	return nil
}
