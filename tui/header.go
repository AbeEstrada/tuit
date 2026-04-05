package tui

import (
	"fmt"

	"git.sr.ht/~rockorager/vaxis"
	"github.com/AbeEstrada/tuit/constants"
)

type Header struct {
	text        string
	badge       int
	showBadge   bool
}

func CreateHeader() *Header {
	return &Header{showBadge: true}
}

func (h *Header) SetText(text string) {
	h.text = text
}

func (h *Header) SetBadge(n int) {
	h.badge = n
}

func (h *Header) SetBadgeVisible(visible bool) {
	h.showBadge = visible
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
	if h.badge > 0 && h.showBadge {
		segments = append(segments, vaxis.Segment{
			Text:  fmt.Sprintf(" [%d]", h.badge),
			Style: bold,
		})
	}
	win.Println(0, segments...)
}
