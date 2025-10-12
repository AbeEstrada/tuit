package tui

import (
	_ "image/jpeg"
	_ "image/png"
	"log"

	"git.sr.ht/~rockorager/vaxis"
	"github.com/AbeEstrada/tuit/utils"
	"github.com/mattn/go-mastodon"
)

type Timeline struct {
	Statuses     []*mastodon.Status
	Selected     *mastodon.Status
	Account      *mastodon.Account
	scrollOffset int
}

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

func (v *TimelineView) AddTimeline(statuses []*mastodon.Status, selected *mastodon.Status, account *mastodon.Account) {
	if len(statuses) == 0 {
		if account == nil {
			return
		}
		return
	}

	selectedStatus := statuses[0]
	if selected != nil {
		for _, status := range statuses {
			if status.ID == selected.ID {
				selectedStatus = status
				break
			}
		}
	}

	t := Timeline{
		Statuses:     statuses,
		Selected:     selectedStatus,
		scrollOffset: 0,
	}

	if account != nil {
		t.Account = account
	}

	v.timelines = append(v.timelines, t)

	v.index = len(v.timelines) - 1

	v.setTitle()
}

func (v *TimelineView) RemoveLastTimeline() {
	if len(v.timelines) <= 1 {
		return
	}

	v.timelines = v.timelines[:len(v.timelines)-1]

	if v.index >= len(v.timelines) {
		v.index = len(v.timelines) - 1
	}

	v.setTitle()
}

func (v *TimelineView) setTitle() {
	timeline := &v.timelines[v.index]

	switch {
	case v.index == 0:
		v.app.header.SetText("Home")
	case timeline.Account != nil:
		v.app.header.SetText(timeline.Account.DisplayName)
	default: // v.index != 0 && timeline.Account == nil
		v.app.header.SetText("Thread")
	}
}

func (v *TimelineView) UpdateTimeline(index int, newStatuses []*mastodon.Status, prepend bool) {
	if len(newStatuses) == 0 || index < 0 || index >= len(v.timelines) {
		return
	}

	timeline := &v.timelines[index]
	statuses := timeline.Statuses
	selected := timeline.Selected

	existingIDs := make(map[mastodon.ID]struct{})
	for _, status := range statuses {
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
		statuses = append(freshStatuses, statuses...)
	} else {
		statuses = append(statuses, freshStatuses...)
	}

	v.timelines[index].Statuses = statuses

	if selected != nil {
		for _, status := range statuses {
			if status.ID == selected.ID {
				v.timelines[index].Selected = status
				break
			}
		}
	}
}

func (v *TimelineView) PrependToTimeline(index int, newStatuses []*mastodon.Status) {
	v.UpdateTimeline(index, newStatuses, true)
}

func (v *TimelineView) AppendToTimeline(index int, newStatuses []*mastodon.Status) {
	v.UpdateTimeline(index, newStatuses, false)
}

func (v *TimelineView) DeleteFromTimeline(index int, statusID mastodon.ID) {
	timeline := &v.timelines[index]
	statuses := timeline.Statuses
	selected := timeline.Selected

	if len(statuses) == 0 {
		return
	}

	var deleteIndex int = -1
	for i, status := range statuses {
		if status.ID == statusID {
			deleteIndex = i
			break
		}
	}

	if deleteIndex == -1 {
		return
	}

	if selected != nil && selected.ID == statusID {
		if len(statuses) == 1 {
			v.timelines[index].Selected = nil
		} else if deleteIndex == 0 {
			v.timelines[index].Selected = statuses[1]
		} else {
			v.timelines[index].Selected = statuses[deleteIndex-1]
		}
	}

	v.timelines[index].Statuses = append(statuses[:deleteIndex], statuses[deleteIndex+1:]...)
}

func (v *TimelineView) SelectedStatus() *mastodon.Status {
	if v.index >= len(v.timelines) {
		return nil
	}
	return v.timelines[v.index].Selected
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
	case key.Matches('o'):
		if selected != nil {
			var url string
			if selected.Reblog != nil && selected.Reblog.URL != "" {
				url = selected.Reblog.URL
			} else if selected.URL != "" {
				url = selected.URL
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
