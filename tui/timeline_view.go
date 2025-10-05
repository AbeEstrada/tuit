package tui

import (
	_ "image/jpeg"
	_ "image/png"
	"log"

	"git.sr.ht/~rockorager/vaxis"
	"github.com/AbeEstrada/mastty/utils"
	"github.com/mattn/go-mastodon"
)

type TimelineView struct {
	app               *App
	statuses          []*mastodon.Status
	selectedStatus    *mastodon.Status
	scrollOffset      int
	onSelectionChange func(status *mastodon.Status)
	onLoadMore        func()
}

func CreateTimelineView() *TimelineView {
	return &TimelineView{
		statuses: make([]*mastodon.Status, 0),
	}
}

func (v *TimelineView) SetApp(app *App) {
	v.app = app
}

func (v *TimelineView) SetStatuses(statuses []*mastodon.Status) {
	v.statuses = statuses

	if len(v.statuses) == 0 {
		v.selectedStatus = nil
		return
	}

	if v.selectedStatus != nil {
		for _, s := range v.statuses {
			if s.ID == v.selectedStatus.ID {
				v.selectedStatus = s
				return
			}
		}
	}

	v.selectedStatus = v.statuses[0]
}

func (v *TimelineView) AddStatuses(newStatuses []*mastodon.Status, prepend bool) {
	if len(newStatuses) == 0 {
		return
	}

	existingIDs := make(map[mastodon.ID]struct{})
	for _, status := range v.statuses {
		existingIDs[status.ID] = struct{}{}
	}

	var freshStatuses []*mastodon.Status
	for _, status := range newStatuses {
		if _, exists := existingIDs[status.ID]; !exists {
			freshStatuses = append(freshStatuses, status)
		}
	}

	if len(freshStatuses) == 0 {
		return
	}

	if prepend {
		v.statuses = append(freshStatuses, v.statuses...)
	} else {
		v.statuses = append(v.statuses, freshStatuses...)
	}

	if v.selectedStatus != nil {
		for _, s := range v.statuses {
			if s.ID == v.selectedStatus.ID {
				v.selectedStatus = s
				return
			}
		}
	}
}

func (v *TimelineView) PrependStatuses(newStatuses []*mastodon.Status) {
	v.AddStatuses(newStatuses, true)
}

func (v *TimelineView) AppendStatuses(newStatuses []*mastodon.Status) {
	v.AddStatuses(newStatuses, false)
}

func (v *TimelineView) SelectedStatus() *mastodon.Status {
	return v.selectedStatus
}

func (v *TimelineView) Draw(win vaxis.Window, focused bool) {
	width, height := win.Size()

	if len(v.statuses) == 0 {
		win.Println(0, vaxis.Segment{Text: "Loading..."})
		return
	}

	y := 0
	for i := v.scrollOffset; i < len(v.statuses) && y < height-1; i++ {
		status := v.statuses[i]
		user := "@" + status.Account.Acct
		createdAt := status.CreatedAt.Local()
		timestamp := createdAt.Format("2006-01-02 15:04")
		if width < 60 {
			timestamp = createdAt.Format("15:04")
		}

		selectedStyle := vaxis.Style{}
		if status == v.selectedStatus && focused {
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
	if len(v.statuses) == 0 {
		return
	}

	currentIndex := -1
	if v.selectedStatus != nil {
		for i, s := range v.statuses {
			if s == v.selectedStatus || s.ID == v.selectedStatus.ID {
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
			newIndex = len(v.statuses) - 1
		} else {
			newIndex--
		}
	case key.Matches('g'):
		newIndex = 0
	case key.Matches('G'):
		newIndex = len(v.statuses) - 1
	case key.Matches('o'):
		if v.selectedStatus != nil {
			var url string
			if v.selectedStatus.Reblog.URL != "" {
				url = v.selectedStatus.Reblog.URL
			} else if v.selectedStatus.URL != "" {
				url = v.selectedStatus.URL
			}
			if err := utils.OpenBrowser(url); err != nil {
				log.Printf("Failed to open URL: %v", err)
			}
		}
		return
	default:
		return
	}

	if newIndex < 0 {
		newIndex = 0
	}
	if newIndex >= len(v.statuses) {
		if v.onLoadMore != nil && !v.app.loading {
			go v.onLoadMore()
		}
		return
	}

	_, height := v.app.vx.Window().Size()
	if newIndex >= v.scrollOffset+height-3 {
		v.scrollOffset = newIndex - (height - 4)
	}
	if newIndex < v.scrollOffset {
		v.scrollOffset = newIndex
	}

	newStatus := v.statuses[newIndex]
	if v.selectedStatus == nil || newStatus.ID != v.selectedStatus.ID {
		v.selectedStatus = newStatus
	}
}
