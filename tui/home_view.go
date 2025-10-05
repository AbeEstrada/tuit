package tui

import (
	"context"

	"git.sr.ht/~rockorager/vaxis"
	"github.com/mattn/go-mastodon"
)

type HomeView struct {
	app         *App
	left        *TimelineView
	right       *StatusView
	focusedView int
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
	go v.loadTimeline()
	v.app.header.SetText("Home")
}

func (v *HomeView) loadTimeline() {
	v.app.SetLoading(true)
	timeline, err := v.app.client.GetTimelineHome(context.Background(), nil)
	if err == nil {
		v.left.SetStatuses(timeline)
		v.app.vx.PostEvent(vaxis.Redraw{})
	}
	v.app.SetLoading(false)

}

func (v *HomeView) reloadTimeline() {
	v.app.SetLoading(true)
	var sinceID mastodon.ID
	if len(v.left.statuses) > 0 {
		sinceID = v.left.statuses[0].ID
	}

	timeline, err := v.app.client.GetTimelineHome(context.Background(), &mastodon.Pagination{
		SinceID: sinceID,
		Limit:   40,
	})

	if err == nil && len(timeline) > 0 {
		v.left.PrependStatuses(timeline)
		v.app.vx.PostEvent(vaxis.Redraw{})
	}
	v.app.SetLoading(false)
}

func (v *HomeView) loadMoreTimeline() {
	v.app.SetLoading(true)
	var maxID mastodon.ID
	if len(v.left.statuses) > 0 {
		lastStatus := v.left.statuses[len(v.left.statuses)-1]
		maxID = lastStatus.ID
	}

	timeline, err := v.app.client.GetTimelineHome(context.Background(), &mastodon.Pagination{
		MaxID: maxID,
		Limit: 20,
	})

	if err == nil && len(timeline) > 0 {
		v.left.AppendStatuses(timeline)
		v.app.vx.PostEvent(vaxis.Redraw{})
	}
	v.app.SetLoading(false)
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
	v.right.Draw(rightWin, v.focusedView == 1, v.left.SelectedStatus())

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
	} else if key.Matches('r') && !v.app.loading {
		go v.reloadTimeline()
	} else {
		if v.focusedView == 0 {
			v.left.HandleKey(key)
		} else {
			v.right.HandleKey(key)
		}
	}
}
