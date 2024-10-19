package ui

import (
	"fyne.io/fyne/v2"
	"log"
	"sync"
	"time"
)

const minImageHeight = 100
const buttonHeight = 30
const buttonPadding = 10

const usedHeightForButton = buttonHeight + buttonPadding

// LayoutMain is a fyne layout where the first component is full width and min 100px height, the second component is full width and 20px height, and the third is the cursor.
type LayoutMain struct {
	PlaybackPercent float32

	playState    int
	playStateMux sync.Mutex
	width        float32
}

func (l *LayoutMain) Pause(cursor fyne.CanvasObject) {
	l.playStateMux.Lock()
	l.playState = l.playState + 1
	l.playStateMux.Unlock()

	cursorPosition := float32(float32(l.width) * l.PlaybackPercent)
	cursor.Move(fyne.NewPos(cursorPosition, 2))
}

func (l *LayoutMain) Play(cursor fyne.CanvasObject, pos float32, dur float32) {
	l.playStateMux.Lock()
	l.playState = l.playState + 1
	expect := l.playState
	l.playStateMux.Unlock()

	go (func() {
		for pos <= dur {
			s := time.Now()
			l.playStateMux.Lock()
			if l.playState != expect {
				l.playStateMux.Unlock()
				log.Println("LayoutMain playback interrupted")
				return
			}
			l.PlaybackPercent = pos / dur
			l.playStateMux.Unlock()

			pos += 0.01
			//c.Refresh()

			cursorPosition := float32(float32(l.width) * l.PlaybackPercent)
			cursor.Move(fyne.NewPos(cursorPosition, 2))

			renderDur := time.Since(s)

			time.Sleep(time.Millisecond*10 - renderDur)
		}
		log.Println("LayoutMain playback finished")
	})()

}

func (l *LayoutMain) getButtonHeight(objects []fyne.CanvasObject) float32 {
	return usedHeightForButton*float32(len(objects)-3) + buttonPadding
}

func (l *LayoutMain) MinSize(objects []fyne.CanvasObject) fyne.Size {
	return fyne.NewSize(400, minImageHeight+l.getButtonHeight(objects))
}

func (l *LayoutMain) Layout(objects []fyne.CanvasObject, size fyne.Size) {
	l.width = size.Width // kinda hacky
	if len(objects) < 5 {
		panic("LayoutMain must have 3 objects (image, imageClicker, replayButton, closeButton, cursor)")
	}

	image := objects[0]
	clicker := objects[1]
	cursor := objects[2]

	buttons := objects[3:]
	usedHeightForButtons := l.getButtonHeight(objects)

	imageHeight := size.Height - usedHeightForButtons

	image.Resize(fyne.NewSize(size.Width, imageHeight))
	image.Move(fyne.NewPos(0, 0))

	// move clicker to exact same size/pos as image
	clicker.Resize(fyne.NewSize(size.Width, imageHeight))
	clicker.Move(fyne.NewPos(0, 0))

	// buttons
	y := imageHeight + buttonPadding
	for _, btn := range buttons {
		btn.Resize(fyne.NewSize(size.Width, buttonHeight))
		btn.Move(fyne.NewPos(0, y))

		y += buttonHeight + buttonPadding

		//closeButton.Resize(fyne.NewSize(size.Width, buttonHeight))
		//closeButton.Move(fyne.NewPos(0, y))
	}

	//y += buttonHeight

	//cursorPosition := float32(float32(size.Width) * l.PlaybackPercent)
	//log.Println("render with pos", cursorPosition)
	cursorHeight := size.Height - usedHeightForButtons
	cursor.Resize(fyne.NewSize(1, cursorHeight))
	//cursor.Move(fyne.NewPos(cursorPosition, 0))
}
