package ui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/widget"
	"log"
	"strconv"
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
	duration     float32
}

func (l *LayoutMain) Pause(cursor fyne.CanvasObject, label *widget.Label, pos float32, dur float32) {
	l.playStateMux.Lock()
	l.playState = l.playState + 1
	l.PlaybackPercent = pos / dur
	l.playStateMux.Unlock()

	cursorPosition := float32(float32(l.width) * l.PlaybackPercent)
	cursor.Move(fyne.NewPos(cursorPosition, 2))
	label.SetText(l.formatDuration(pos) + " / " + l.formatDuration(dur))
}

func (l *LayoutMain) formatDuration(duration float32) string {
	if duration <= 0 {
		return "0:00:00"
	}
	minute := int(duration / 60)
	second := int(duration) % 60
	ms := int((duration - float32(int(duration))) * 100)

	minuteStr := strconv.Itoa(minute)
	secondStr := strconv.Itoa(second)
	msStr := strconv.Itoa(ms)
	if len(secondStr) == 1 {
		secondStr = "0" + secondStr
	}
	if len(msStr) == 1 {
		msStr = "0" + msStr
	}
	return minuteStr + ":" + secondStr + ":" + msStr
}

func (l *LayoutMain) Play(cursor fyne.CanvasObject, label *widget.Label, pos float32, dur float32) {
	playStart := time.Now()
	l.playStateMux.Lock()
	l.playState++
	expect := l.playState
	l.playStateMux.Unlock()

	durationStr := l.formatDuration(dur)

	go (func() {
		origPos := pos
		for pos <= dur {
			l.playStateMux.Lock()
			if l.playState != expect {
				l.playStateMux.Unlock()
				log.Println("LayoutMain playback interrupted")
				return
			}
			l.PlaybackPercent = pos / dur
			percent := l.PlaybackPercent
			l.playStateMux.Unlock()

			amountToAdd := float32(time.Since(playStart).Milliseconds()) / 1000
			pos = origPos + amountToAdd

			cursorPosition := float32(float32(l.width) * percent)
			cursor.Move(fyne.NewPos(cursorPosition, 2))
			label.SetText(l.formatDuration(pos) + " / " + durationStr)

			time.Sleep(time.Millisecond * 10)
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
		panic("LayoutMain must have at least 5 objects")
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

	cursorPosition := float32(float32(size.Width) * l.PlaybackPercent)
	cursorHeight := size.Height - usedHeightForButtons
	cursor.Resize(fyne.NewSize(1, cursorHeight))
	cursor.Move(fyne.NewPos(cursorPosition, 2))
}
