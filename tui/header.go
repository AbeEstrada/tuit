package tui

import (
	"git.sr.ht/~rockorager/vaxis"
	"github.com/AbeEstrada/tuit/constants"
)

type Header struct {
	text string
}

func CreateHeader() *Header {
	return &Header{}
}

func (h *Header) SetText(text string) {
	h.text = text
}

func (h *Header) Draw(win vaxis.Window) {
	style := vaxis.Style{
		Attribute: vaxis.AttrBold,
	}
	win.Println(0, vaxis.Segment{
		Text:  constants.AppName,
		Style: style,
	}, vaxis.Segment{
		Text: " ",
	}, vaxis.Segment{
		Text: h.text,
	})
}
