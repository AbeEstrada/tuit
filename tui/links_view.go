package tui

import (
	"fmt"

	"git.sr.ht/~rockorager/vaxis"
)

type LinksView struct {
	links    []string
	selected int
}

func CreateLinksView() *LinksView {
	return &LinksView{}
}

func (v *LinksView) SetLinks(links []string) {
	v.links = links
	v.selected = 0
}

func (v *LinksView) Draw(win vaxis.Window, focused bool) {
	width, height := win.Size()

	win.Println(0, vaxis.Segment{
		Text:  "Links",
		Style: vaxis.Style{Attribute: vaxis.AttrBold},
	})

	for i, link := range v.links {
		y := i + 2
		if y >= height-1 {
			break
		}

		label := fmt.Sprintf(" %d. %s", i+1, link)
		if len(label) > width {
			label = label[:width-1] + "…"
		}

		var attr vaxis.AttributeMask
		if i == v.selected && focused {
			attr = vaxis.AttrReverse
		}
		win.Println(y, vaxis.Segment{
			Text:  label,
			Style: vaxis.Style{Attribute: attr},
		})
	}
}

func (v *LinksView) HandleKey(key vaxis.Key) string {
	switch {
	case key.Matches('j'):
		if v.selected < len(v.links)-1 {
			v.selected++
		}
	case key.Matches('k'):
		if v.selected > 0 {
			v.selected--
		}
	case key.Matches(vaxis.KeyEnter):
		return "open"
	case key.Matches('q'):
		return "close"
	}
	return ""
}
