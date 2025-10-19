package tui

import "github.com/mattn/go-mastodon"

type Timeline struct {
	Statuses     []*mastodon.Status
	Selected     *mastodon.Status
	Account      *mastodon.Account
	scrollOffset int
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

func (v *TimelineView) UpdateEdit(index int, status *mastodon.Status) {
	if status == nil || index < 0 || index >= len(v.timelines) {
		return
	}

	timeline := &v.timelines[index]
	statuses := timeline.Statuses
	selected := timeline.Selected

	for i, s := range statuses {
		if s.ID == status.ID {
			v.timelines[index].Statuses[i] = status

			if selected != nil && selected.ID == status.ID {
				v.timelines[index].Selected = status
			}

			break
		}
	}
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
