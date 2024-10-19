package ui

import (
	"audio-player/audio"
	"audio-player/gtime"
	"audio-player/visu"
	"bytes"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
	"image"
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
	gtime.Start("ui.Run")
	// before start, clean up old stuff
	if u.a != nil {
		u.a.Stop()
	}

	u.audioFile = audioFile
	u.a = audio.New(u.audioFile)

	l := &LayoutMain{}

	gtime.Start("ui.Run.createAppAndWindow")
	didMakeApp := false
	if u.app == nil {
		u.app = app.New()
		didMakeApp = true
	}
	titleStr := path.Base(u.audioFile)
	if u.w == nil {
		u.w = u.app.NewWindow(titleStr)
		u.w.Resize(fyne.NewSize(800, 200))
	} else {
		u.w.SetTitle(titleStr)
	}
	gtime.End("ui.Run.createAppAndWindow")

	gtime.Start("ui.Run.CreateElements")
	var ct *fyne.Container
	cursor := canvas.NewRectangle(color.White)
	imageWidth, imageHeight := visu.GetSize()

	// TODO: would be nice if placeholder was a bit more fancy (loading animation? placeholder wave form?)
	placeholderImage := image.NewAlpha(image.Rect(0, 0, imageWidth, imageHeight))
	canvasImg := canvas.NewImageFromImage(placeholderImage)

	go (func() {

		imageBits, err := visu.GenerateImage(u.audioFile)
		if err != nil {
			log.Fatal("error generating image:", err)
		}
		decodedImage, _, err := image.Decode(bytes.NewReader(imageBits))
		if err != nil {
			log.Fatal("error decoding image:", err)
		}
		canvasImg.Image = decodedImage
		canvasImg.Refresh()
	})()

	//image := canvas.NewImageFromReader(bytes.NewReader(imageBits), "waveform_"+u.audioFile)
	cursorPositioner := NewClickableInvisible(func(event *fyne.PointEvent) {
		log.Println("clicked at", event.Position)
		// translate to duration
		percent := event.Position.X / float32(canvasImg.Size().Width)
		pos := u.a.Duration() * percent
		log.Println("seeking to", pos)

		playAudioSync(u.a, l, cursor, pos, u.a.Duration())
	})

	replayButton := widget.NewButton("Replay", func() {
		playAudioSync(u.a, l, cursor, 0, u.a.Duration())
	})

	closeButton := widget.NewButton("Close", func() {
		u.a.Stop()
		u.app.Quit()
	})

	ct = container.New(l, canvasImg, cursorPositioner, replayButton, closeButton, cursor)

	gtime.End("ui.Run.CreateElements") // 1ms

	go (func() {
		// u.a.Duration() is called before goroutine is created, so we can't just do "go playAudioSync(...)"
		playAudioSync(u.a, l, cursor, 0, u.a.Duration())
	})()

	u.w.SetContent(ct)

	if didMakeApp {
		gtime.End("ui.Run")
		gtime.End("main")
		u.w.ShowAndRun()
		log.Println("app exit")
		u.a.Stop()
	} else {
		// new song playing, so focus window
		u.w.RequestFocus()
	}

	return nil
}