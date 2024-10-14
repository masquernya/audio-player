package ui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/widget"
)

type ClickableInvisible struct {
	widget.BaseWidget
	onTap func(event *fyne.PointEvent)
}

func NewClickableInvisible(onTap func(event *fyne.PointEvent)) *ClickableInvisible {
	c := &ClickableInvisible{
		onTap: onTap,
	}
	// c.ExtendBaseWidget(c)
	return c
}

func (c *ClickableInvisible) Tapped(p *fyne.PointEvent) {
	c.onTap(p)
}
