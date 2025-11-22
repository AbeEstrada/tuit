package tui

import (
	"fmt"
	_ "image/jpeg"
	_ "image/png"
	"log"

	"git.sr.ht/~rockorager/vaxis"
	"github.com/AbeEstrada/tuit/utils"
)

type TimelineView struct {
	app        *App
	timelines  []Timeline
	index      int
	onLoadMore func()
}

func CreateTimelineView() *TimelineView {
	return &TimelineView{
		timelines: []Timeline{},
		index:     0,
	}
}

func (v *TimelineView) SetApp(app *App) {
	v.app = app
}

func (v *TimelineView) setTitle() {
	if v.index >= len(v.timelines) {
		return
	}
	timeline := &v.timelines[v.index]
	switch {
	case v.index == 0:
		v.app.header.SetText("Home")
	case timeline.Account != nil:
		v.app.header.SetText("Home → " + timeline.Account.DisplayName)
	default: // v.index != 0 && timeline.Account == nil
		v.app.header.SetText("Home → Thread")
	}
}

func (v *TimelineView) Draw(win vaxis.Window, focused bool) {
	width, height := win.Size()

	if v.index >= len(v.timelines) || len(v.timelines[v.index].Items) == 0 {
		win.Println(0, vaxis.Segment{Text: "Loading..."})
		return
	}

	timeline := &v.timelines[v.index]
	items := timeline.Items
	selected := timeline.Selected
	selectedID := selected.ID()
	scrollOffset := timeline.scrollOffset

	y := 0
	for i := scrollOffset; i < len(items) && y < height-2; i++ {
		item := items[i]

		var displayText string

		switch t := item.(type) {
		case StatusItem:
			createdAt := t.CreatedAt.Local()
			timestamp := createdAt.Format("2006-01-02 15:04")
			if width < 60 {
				timestamp = createdAt.Format("15:04")
			}

			statusType := " "
			if t.Reblog != nil {
				statusType = "♺"
			} else if t.InReplyToID != nil {
				statusType = "↩"
			}
			displayText = fmt.Sprintf("%s %s @%s", timestamp, statusType, t.Account.Acct)

		case AccountItem:
			displayText = fmt.Sprintf("@%s", t.Acct)

		default:
			continue
		}

		selectedStyle := vaxis.Style{}
		if item.ID() == selectedID && focused {
			selectedStyle = vaxis.Style{
				Attribute: vaxis.AttrReverse,
			}
		}

		win.Println(y, vaxis.Segment{
			Text:  displayText,
			Style: selectedStyle,
		})
		y++
	}
}

func (v *TimelineView) HandleKey(key vaxis.Key) {
	if v.timelines == nil || v.index >= len(v.timelines) || len(v.timelines[v.index].Items) == 0 {
		return
	}

	timeline := &v.timelines[v.index]
	items := timeline.Items
	selected := timeline.Selected
	selectedID := selected.ID()
	scrollOffset := timeline.scrollOffset

	currentIndex := -1
	if selected != nil {
		for i, item := range items {
			if item.ID() == selectedID {
				currentIndex = i
				break
			}
		}
	}

	newIndex := currentIndex
	switch {
	case key.Matches('j'):
		if currentIndex == -1 {
			newIndex = 0
		} else {
			newIndex++
		}
	case key.Matches('k'):
		if currentIndex == -1 {
			newIndex = len(items) - 1
		} else {
			newIndex--
		}
	case key.Matches('g'):
		newIndex = 0
	case key.Matches('G'):
		newIndex = len(items) - 1
	case key.Matches('O'):
		if status, ok := selected.(StatusItem); ok {
			var url string
			if status.Reblog != nil && status.Reblog.URL != "" {
				url = status.Reblog.URL
			} else if status.URL != "" {
				url = status.URL
			}
			if url != "" {
				if err := utils.OpenBrowser(url); err != nil {
					log.Printf("Failed to open URL: %v", err)
				}
			}
		} else if account, ok := selected.(AccountItem); ok {
			if account.URL != "" {
				if err := utils.OpenBrowser(account.URL); err != nil {
					log.Printf("Failed to open URL: %v", err)
				}
			}
		}

	case key.Matches('o'):
		if status, ok := selected.(StatusItem); ok {
			var url string
			if status.Reblog != nil && status.Reblog.URL != "" {
				url = fmt.Sprintf("%s/@%s/%s", v.app.config.Auth.Server, status.Reblog.Account.Acct, status.Reblog.ID)
			} else if status.URL != "" {
				statusID := status.ID()
				url = fmt.Sprintf("%s/@%s/%s", v.app.config.Auth.Server, status.Account.Acct, statusID)
			}
			if url != "" {
				if err := utils.OpenBrowser(url); err != nil {
					log.Printf("Failed to open URL: %v", err)
				}
			}
		}
		return
	case key.Matches('v'):
		if status, ok := selected.(StatusItem); ok {
			var url string
			if status.Reblog != nil && status.Reblog.Card != nil && status.Reblog.Card.URL != "" {
				url = status.Reblog.Card.URL
			} else if status.Card != nil && status.Card.URL != "" {
				url = status.Card.URL
			}
			if url != "" {
				if err := utils.OpenBrowser(url); err != nil {
					log.Printf("Failed to open URL: %v", err)
				}
			} else {
				log.Printf("No URL available to open")
			}
		}
		return
	default:
		return
	}

	if newIndex < 0 {
		newIndex = 0
	}
	if newIndex >= len(items) {
		if v.onLoadMore != nil && !v.app.loading {
			go v.onLoadMore()
		}
		return
	}

	_, height := v.app.vx.Window().Size()
	if newIndex >= scrollOffset+height-4 {
		v.timelines[v.index].scrollOffset = newIndex - (height - 5)
	}
	if newIndex < scrollOffset {
		v.timelines[v.index].scrollOffset = newIndex
	}

	newSelected := items[newIndex]
	if selected == nil || newSelected.ID() != selected.ID() {
		v.timelines[v.index].Selected = newSelected
	}
}
