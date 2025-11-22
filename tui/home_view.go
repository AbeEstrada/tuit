package tui

import (
	"context"
	"log"

	"git.sr.ht/~rockorager/vaxis"
	"github.com/mattn/go-mastodon"
)

type HomeView struct {
	app         *App
	left        *TimelineView
	right       *StatusView
	focusedView int
	isStreaming bool
}

func CreateHomeView() *HomeView {
	v := &HomeView{
		right:       CreateStatusView(),
		focusedView: 0,
	}
	leftView := CreateTimelineView()
	leftView.onLoadMore = v.loadMoreTimeline
	v.left = leftView

	return v
}

func (v *HomeView) SetApp(app *App) {
	v.app = app
	v.left.SetApp(app)
	v.right.SetApp(app)
}

func (v *HomeView) OnActivate() {
	go v.getHomeTimeline()
	go v.startStreaming()
	v.app.header.SetText("Home")
}

func (v *HomeView) getHomeTimeline() {
	v.app.SetLoading(true)

	statuses, err := v.app.client.GetTimelineHome(context.Background(), nil)
	if err == nil {
		items := make([]TimelineItem, len(statuses))
		for i, s := range statuses {
			items[i] = StatusItem{Status: s}
		}
		v.left.AddTimeline(items, nil, nil)
		v.app.vx.PostEvent(vaxis.Redraw{})
	}

	v.app.SetLoading(false)
}

func (v *HomeView) getStatusContext() {
	selectedItem := v.left.SelectedItem()

	item, ok := selectedItem.(StatusItem)
	if !ok || item.Status == nil {
		return
	}

	original := item.Status

	if original.Reblog != nil {
		original = original.Reblog
	}

	v.app.SetLoading(true)

	if original.ID != "" {
		ctx, err := v.app.client.GetStatusContext(context.Background(), original.ID)

		if err == nil {
			items := make([]TimelineItem, 0, len(ctx.Ancestors)+1+len(ctx.Descendants))

			for _, status := range ctx.Ancestors {
				items = append(items, StatusItem{Status: status})
			}

			items = append(items, StatusItem{Status: original})

			for _, status := range ctx.Descendants {
				items = append(items, StatusItem{Status: status})
			}

			v.left.AddTimeline(items, StatusItem{Status: original}, nil)
			v.app.vx.PostEvent(vaxis.Redraw{})
		}
	}

	v.app.SetLoading(false)
}

func (v *HomeView) reloadHomeTimeline() {
	v.app.SetLoading(true)

	index := 0 // Home
	if index >= len(v.left.timelines) {
		v.app.SetLoading(false)
		return
	}

	timeline := &v.left.timelines[index]
	items := timeline.Items

	if len(items) == 0 {
		v.app.SetLoading(false)
		return
	}

	var sinceID mastodon.ID
	if firstItem, ok := items[0].(StatusItem); ok && firstItem.Status != nil {
		sinceID = firstItem.Status.ID
	} else {
		v.app.SetLoading(false)
		return
	}

	newStatuses, err := v.app.client.GetTimelineHome(context.Background(), &mastodon.Pagination{
		SinceID: sinceID,
		Limit:   40,
	})

	if err == nil && len(newStatuses) > 0 {
		newItems := make([]TimelineItem, len(newStatuses))
		for i, s := range newStatuses {
			newItems[i] = StatusItem{Status: s}
		}
		v.left.PrependToTimeline(index, newItems)
		v.app.vx.PostEvent(vaxis.Redraw{})
	}
	v.app.SetLoading(false)
}

func (v *HomeView) loadMoreTimeline() {
	v.app.SetLoading(true)

	index := v.left.index
	if index >= len(v.left.timelines) {
		v.app.SetLoading(false)
		return
	}

	timeline := &v.left.timelines[index]
	items := timeline.Items

	if len(items) == 0 {
		v.app.SetLoading(false)
		return
	}

	var maxID mastodon.ID
	if lastItem, ok := items[len(items)-1].(StatusItem); ok && lastItem.Status != nil {
		maxID = lastItem.Status.ID
	} else {
		v.app.SetLoading(false)
		return
	}

	var newStatuses []*mastodon.Status
	var err error

	if index == 0 {
		newStatuses, err = v.app.client.GetTimelineHome(context.Background(), &mastodon.Pagination{
			MaxID: maxID,
			Limit: 20,
		})
	} else if timeline.Account != nil {
		newStatuses, err = v.app.client.GetAccountStatuses(context.Background(), timeline.Account.ID, &mastodon.Pagination{
			MaxID: maxID,
			Limit: 20,
		})
	}

	if err == nil && len(newStatuses) > 0 {
		newItems := make([]TimelineItem, len(newStatuses))
		for i, s := range newStatuses {
			newItems[i] = StatusItem{Status: s}
		}
		v.left.AppendToTimeline(index, newItems)
		v.app.vx.PostEvent(vaxis.Redraw{})
	}

	v.app.SetLoading(false)
}

func (v *HomeView) getAccountAndTimeline() {
	selectedItem := v.left.SelectedItem()

	item, ok := selectedItem.(StatusItem)
	if !ok || item.Status == nil {
		return
	}
	original := item.Status

	if original.Reblog != nil {
		original = original.Reblog
	}

	if original.ID == "" {
		return
	}

	v.app.SetLoading(true)

	account, err := v.app.client.GetAccount(context.Background(), original.Account.ID)
	if err == nil {
		statuses, err := v.app.client.GetAccountStatuses(context.Background(), account.ID, &mastodon.Pagination{})
		if err == nil {
			items := make([]TimelineItem, len(statuses))
			for i, s := range statuses {
				items[i] = StatusItem{Status: s}
			}
			v.left.AddTimeline(items, StatusItem{Status: original}, account)
		}
	}

	v.app.SetLoading(false)
}

func (v *HomeView) startStreaming() {
	ctx := context.Background()

	events, err := v.app.client.StreamingUser(ctx)
	if err != nil {
		log.Printf("Failed to start streaming: %v", err)
		return
	}

	v.isStreaming = true

	for {
		select {
		case event := <-events:
			v.handleStreamingEvent(event)
		case <-ctx.Done():
			v.isStreaming = false
			return
		}
	}
}

func (v *HomeView) handleStreamingEvent(event mastodon.Event) {
	switch e := event.(type) {
	case *mastodon.UpdateEvent:
		v.left.PrependToTimeline(0, []TimelineItem{StatusItem{Status: e.Status}})
		v.app.vx.PostEvent(vaxis.Redraw{})

	case *mastodon.UpdateEditEvent:
		v.left.UpdateEdit(0, StatusItem{Status: e.Status})
		v.app.vx.PostEvent(vaxis.Redraw{})

	case *mastodon.NotificationEvent:
		log.Printf("New Notification [%s] from @%s\n", e.Notification.Type, e.Notification.Account.Acct)

	case *mastodon.DeleteEvent:
		v.left.DeleteFromTimeline(0, e.ID)
		v.app.vx.PostEvent(vaxis.Redraw{})

	case *mastodon.ErrorEvent:
		log.Printf("Error %v\n", e.Error())

	default:
		log.Printf("Streaming unhandled event type\n")
	}
}

func (v *HomeView) Draw(win vaxis.Window) {
	var (
		leftRatio  = 2
		rightRatio = 3
	)
	if leftRatio <= 0 {
		leftRatio = 1
	}
	if rightRatio <= 0 {
		leftRatio = 1
	}

	width, height := win.Size()
	separatorStyle := vaxis.Style{
		Foreground: vaxis.IndexColor(0),
	}

	total := leftRatio + rightRatio
	split := width * leftRatio / total

	leftWin := win.New(0, 1, split, height)
	rightWidth := max(0, width-split-2)
	rightWin := win.New(split+2, 1, rightWidth, height)

	v.left.Draw(leftWin, v.focusedView == 0)

	var selectedStatus *mastodon.Status
	if item := v.left.SelectedItem(); item != nil {
		if item, ok := item.(StatusItem); ok {
			selectedStatus = item.Status
		}
	}

	v.right.Draw(rightWin, v.focusedView == 1, selectedStatus)

	for row := 0; row < height-2; row++ {
		win.SetCell(split, row+1, vaxis.Cell{
			Character: vaxis.Character{
				Grapheme: "│",
			},
			Style: separatorStyle,
		})
	}
}

func (v *HomeView) HandleKey(key vaxis.Key) {
	if key.Matches(vaxis.KeyTab) {
		v.focusedView = (v.focusedView + 1) % 2
	} else if key.Matches('h') {
		v.focusedView = 0
	} else if key.Matches('l') {
		v.focusedView = 1
	} else if key.Matches('r') && !v.app.loading && !v.isStreaming {
		go v.reloadHomeTimeline()
	} else if key.Matches('t') && !v.app.loading {
		go v.getStatusContext()
	} else if key.Matches('u') && !v.app.loading {
		go v.getAccountAndTimeline()
	} else if key.Matches('q') {
		if len(v.left.timelines) <= 1 {
			v.app.RequestQuit()
		} else {
			v.left.RemoveLastTimeline()
		}
		return
	} else {
		if v.focusedView == 0 {
			v.left.HandleKey(key)
		} else {
			v.right.HandleKey(key)
		}
	}
}
