package ui

import (
	"git.sr.ht/~rockorager/vaxis"
)

type Footer struct {
	text string
}

func CreateFooter(vx *vaxis.Vaxis) *Footer {
	return &Footer{
		text: "",
	}
}

func (f *Footer) SetText(text string) {
	f.text = text
}

func (f *Footer) Draw(win vaxis.Window) {
	_, height := win.Size()
	y := height - 1
	win.Println(y, vaxis.Segment{Text: f.text})
}
