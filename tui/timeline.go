package tui

import "github.com/mattn/go-mastodon"

type Timeline struct {
	Items        []TimelineItem
	Selected     TimelineItem
	Account      *mastodon.Account
	scrollOffset int
}

func (v *TimelineView) AddTimeline(items []TimelineItem, selected TimelineItem, account *mastodon.Account) {
	if len(items) == 0 {
		if account == nil {
			return
		}
	}

	var selectedItem TimelineItem
	if len(items) > 0 {
		selectedItem = items[0]
	}

	if selected != nil {
		targetID := selected.ID()
		for _, item := range items {
			if item.ID() == targetID {
				selectedItem = item
				break
			}
		}
	}

	t := Timeline{
		Items:        items,
		Selected:     selectedItem,
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

func (v *TimelineView) UpdateTimeline(index int, newItems []TimelineItem, prepend bool) {
	if len(newItems) == 0 || index < 0 || index >= len(v.timelines) {
		return
	}

	timeline := &v.timelines[index]
	items := timeline.Items
	selected := timeline.Selected

	// Create a set of existing IDs for deduplication
	existingIDs := make(map[mastodon.ID]struct{})
	for _, item := range items {
		existingIDs[item.ID()] = struct{}{}
	}

	var freshItems []TimelineItem
	for _, item := range newItems {
		if _, exists := existingIDs[item.ID()]; !exists {
			freshItems = append(freshItems, item)
		}
	}

	if len(freshItems) == 0 {
		return
	}

	if prepend {
		items = append(freshItems, items...)
	} else {
		items = append(items, freshItems...)
	}

	v.timelines[index].Items = items

	if selected != nil {
		targetID := selected.ID()
		for _, item := range items {
			if item.ID() == targetID {
				v.timelines[index].Selected = item
				break
			}
		}
	}
}

func (v *TimelineView) PrependToTimeline(index int, newItems []TimelineItem) {
	v.UpdateTimeline(index, newItems, true)
}

func (v *TimelineView) AppendToTimeline(index int, newItems []TimelineItem) {
	v.UpdateTimeline(index, newItems, false)
}

func (v *TimelineView) UpdateEdit(index int, newItem TimelineItem) {
	if newItem == nil || index < 0 || index >= len(v.timelines) {
		return
	}

	timeline := &v.timelines[index]
	items := timeline.Items
	selected := timeline.Selected
	targetID := newItem.ID()

	for i, item := range items {
		if item.ID() == targetID {
			v.timelines[index].Items[i] = newItem

			if selected != nil && selected.ID() == targetID {
				v.timelines[index].Selected = newItem
			}
			break
		}
	}
}

func (v *TimelineView) DeleteFromTimeline(index int, targetID mastodon.ID) {
	if index < 0 || index >= len(v.timelines) {
		return
	}

	timeline := &v.timelines[index]
	items := timeline.Items
	selected := timeline.Selected

	if len(items) == 0 {
		return
	}

	var deleteIndex int = -1
	for i, item := range items {
		if item.ID() == targetID {
			deleteIndex = i
			break
		}
	}

	if deleteIndex == -1 {
		return
	}

	if selected != nil && selected.ID() == targetID {
		if len(items) == 1 {
			v.timelines[index].Selected = nil
		} else if deleteIndex == 0 {
			v.timelines[index].Selected = items[1]
		} else {
			v.timelines[index].Selected = items[deleteIndex-1]
		}
	}

	v.timelines[index].Items = append(items[:deleteIndex], items[deleteIndex+1:]...)
}

func (v *TimelineView) SelectedItem() TimelineItem {
	if v.index >= len(v.timelines) {
		return nil
	}
	return v.timelines[v.index].Selected
}
