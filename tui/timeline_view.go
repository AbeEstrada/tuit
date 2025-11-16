package tui

import (
	"fmt"
	_ "image/jpeg"
	_ "image/png"
	"log"

	"git.sr.ht/~rockorager/vaxis"
	"github.com/AbeEstrada/tuit/utils"
	"github.com/mattn/go-mastodon"
)

type TimelineView struct {
	app               *App
	timelines         []Timeline
	index             int
	onSelectionChange func(status *mastodon.Status)
	onLoadMore        func()
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

	if v.index >= len(v.timelines) || len(v.timelines[v.index].Statuses) == 0 {
		win.Println(0, vaxis.Segment{Text: "Loading..."})
		return
	}

	timeline := &v.timelines[v.index]
	statuses := timeline.Statuses
	selected := timeline.Selected
	scrollOffset := timeline.scrollOffset

	y := 0
	for i := scrollOffset; i < len(statuses) && y < height-1; i++ {
		status := statuses[i]
		user := "@" + status.Account.Acct
		createdAt := status.CreatedAt.Local()
		timestamp := createdAt.Format("2006-01-02 15:04")
		if width < 60 {
			timestamp = createdAt.Format("15:04")
		}

		selectedStyle := vaxis.Style{}
		if status == selected && focused {
			selectedStyle = vaxis.Style{
				Attribute: vaxis.AttrReverse,
			}
		}

		statusType := " "
		if status.Reblog != nil {
			statusType = "♺"
		} else if status.InReplyToID != nil {
			statusType = "↩"
		}

		win.Println(y, vaxis.Segment{
			Text:  timestamp + " " + statusType + " " + user,
			Style: selectedStyle,
		})
		y++
	}
}

func (v *TimelineView) HandleKey(key vaxis.Key) {
	if v.timelines == nil || len(v.timelines[v.index].Statuses) == 0 {
		return
	}

	timeline := &v.timelines[v.index]
	statuses := timeline.Statuses
	selected := timeline.Selected
	scrollOffset := timeline.scrollOffset

	currentIndex := -1
	if selected != nil {
		for i, status := range statuses {
			if status == selected || status.ID == selected.ID {
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
			newIndex = len(statuses) - 1
		} else {
			newIndex--
		}
	case key.Matches('g'):
		newIndex = 0
	case key.Matches('G'):
		newIndex = len(statuses) - 1
	case key.Matches('O'):
		if selected.URL != "" {
			if err := utils.OpenBrowser(selected.URL); err != nil {
				log.Printf("Failed to open URL: %v", err)
			}
		}

	case key.Matches('o'):
		if selected != nil {
			var url string
			if selected.Reblog != nil && selected.Reblog.URL != "" {
				url = selected.Reblog.URL
			} else if selected.URL != "" {
				url = fmt.Sprintf("%s/@%s/%s", v.app.config.Auth.Server, selected.Account.Acct, selected.ID)
			}
			if url != "" {
				if err := utils.OpenBrowser(url); err != nil {
					log.Printf("Failed to open URL: %v", err)
				}
			}
		}
		return
	case key.Matches('v'):
		if selected != nil {
			var url string
			if selected.Reblog != nil && selected.Reblog.Card != nil && selected.Reblog.Card.URL != "" {
				url = selected.Reblog.Card.URL
			} else if selected.Card != nil && selected.Card.URL != "" {
				url = selected.Card.URL
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
	if newIndex >= len(statuses) {
		if v.onLoadMore != nil && !v.app.loading {
			go v.onLoadMore()
		}
		return
	}

	_, height := v.app.vx.Window().Size()
	if newIndex >= scrollOffset+height-3 {
		v.timelines[v.index].scrollOffset = newIndex - (height - 4)
	}
	if newIndex < scrollOffset {
		v.timelines[v.index].scrollOffset = newIndex
	}

	newSelected := statuses[newIndex]
	if selected == nil || newSelected.ID != selected.ID {
		v.timelines[v.index].Selected = newSelected
	}
}
