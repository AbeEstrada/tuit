package tui

import (
	"fmt"

	"git.sr.ht/~rockorager/vaxis"
)

type LinkItem struct {
	Label string
	URL   string
}

type LinksView struct {
	links    []LinkItem
	selected int
}

func CreateLinksView() *LinksView {
	return &LinksView{}
}

func (v *LinksView) SetLinks(links []LinkItem) {
	v.links = links
	v.selected = 0
}

func (v *LinksView) Draw(win vaxis.Window, focused bool) {
	width, height := win.Size()

	win.Println(0, vaxis.Segment{
		Text:  "Links",
		Style: vaxis.Style{Attribute: vaxis.AttrBold},
	})

	for i, item := range v.links {
		y := i + 2
		if y >= height-1 {
			break
		}

		label := item.Label
		if label == "" {
			label = item.URL
		}
		line := fmt.Sprintf(" %d. %s", i+1, label)
		if len(line) > width {
			line = line[:width-1] + "…"
		}

		var attr vaxis.AttributeMask
		if i == v.selected && focused {
			attr = vaxis.AttrReverse
		}
		win.Println(y, vaxis.Segment{
			Text:  line,
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
	default:
		if key.Keycode >= '1' && key.Keycode <= '9' {
			idx := int(key.Keycode - '1')
			if idx < len(v.links) {
				v.selected = idx
				return "open"
			}
		}
	}
	return ""
}
