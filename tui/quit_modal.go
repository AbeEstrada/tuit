package tui

import (
	"git.sr.ht/~rockorager/vaxis"
	"git.sr.ht/~rockorager/vaxis/widgets/border"
)

type QuitModal struct{}

func CreateQuitModal() *QuitModal {
	return &QuitModal{}
}

func (m *QuitModal) Draw(win vaxis.Window) {
	width, height := win.Size()

	modalWidth := 40
	modalHeight := 5
	x := (width - modalWidth) / 2
	y := (height - modalHeight) / 2

	modalWin := win.New(x, y, modalWidth, modalHeight)
	modalWin.Clear()
	modalWin = border.All(modalWin, vaxis.Style{
		Foreground: vaxis.IndexColor(4),
		Attribute:  vaxis.AttrBold,
	})
	modalWin.Print(
		vaxis.Segment{
			Text: "Are you sure you want to quit?",
			Style: vaxis.Style{
				Attribute: vaxis.AttrBold,
			},
		},
	)
}

func (m *QuitModal) HandleKey(key vaxis.Key) string {
	if key.Matches('y') || key.Matches(vaxis.KeyEnter) {
		return "quit"
	} else if key.Matches('n') || key.Matches(vaxis.KeyEsc) || key.Matches('q') {
		return "close"
	}
	return ""
}
