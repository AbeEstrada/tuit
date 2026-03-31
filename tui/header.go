package tui

import (
	"fmt"

	"git.sr.ht/~rockorager/vaxis"
	"github.com/AbeEstrada/tuit/constants"
)

type Header struct {
	text  string
	badge int
}

func CreateHeader() *Header {
	return &Header{}
}

func (h *Header) SetText(text string) {
	h.text = text
}

func (h *Header) SetBadge(n int) {
	h.badge = n
}

func (h *Header) Draw(win vaxis.Window) {
	bold := vaxis.Style{
		Attribute: vaxis.AttrBold,
	}
	segments := []vaxis.Segment{
		{Text: constants.AppName, Style: bold},
		{Text: " "},
		{Text: h.text},
	}
	if h.badge > 0 {
		segments = append(segments, vaxis.Segment{
			Text:  fmt.Sprintf(" [%d]", h.badge),
			Style: bold,
		})
	}
	win.Println(0, segments...)
}
