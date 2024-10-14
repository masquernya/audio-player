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

const usedHeightForButtons = (buttonHeight * 2) + (buttonPadding * 2)

// LayoutMain is a fyne layout where the first component is full width and min 100px height, the second component is full width and 20px height, and the third is the cursor.
type LayoutMain struct {
	PlaybackPercent float32

	playState    int
	playStateMux sync.Mutex
	width        float32
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

func (l *LayoutMain) MinSize(objects []fyne.CanvasObject) fyne.Size {
	return fyne.NewSize(400, minImageHeight+usedHeightForButtons)
}

func (l *LayoutMain) Layout(objects []fyne.CanvasObject, size fyne.Size) {
	l.width = size.Width // kinda hacky
	if len(objects) < 5 {
		panic("LayoutMain must have 3 objects (image, imageClicker, replayButton, closeButton, cursor)")
	}

	image := objects[0]
	clicker := objects[1]
	button := objects[2]
	closeButton := objects[3]
	cursor := objects[4]

	imageHeight := size.Height - usedHeightForButtons

	image.Resize(fyne.NewSize(size.Width, imageHeight))
	image.Move(fyne.NewPos(0, 0))

	// move clicker to exact same size/pos as image
	clicker.Resize(fyne.NewSize(size.Width, imageHeight))
	clicker.Move(fyne.NewPos(0, 0))

	// buttons
	y := imageHeight + buttonPadding
	button.Resize(fyne.NewSize(size.Width, buttonHeight))
	button.Move(fyne.NewPos(0, y))

	y += buttonHeight + buttonPadding

	closeButton.Resize(fyne.NewSize(size.Width, buttonHeight))
	closeButton.Move(fyne.NewPos(0, y))

	y += buttonHeight

	//cursorPosition := float32(float32(size.Width) * l.PlaybackPercent)
	//log.Println("render with pos", cursorPosition)
	cursorHeight := size.Height - usedHeightForButtons
	cursor.Resize(fyne.NewSize(1, cursorHeight))
	//cursor.Move(fyne.NewPos(cursorPosition, 0))
}
