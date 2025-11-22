package tui

import "github.com/mattn/go-mastodon"

type TimelineItem interface {
	ID() mastodon.ID
}

type StatusItem struct {
	*mastodon.Status
}

func (s StatusItem) ID() mastodon.ID {
	return s.Status.ID
}

type AccountItem struct {
	*mastodon.Account
}

func (a AccountItem) ID() mastodon.ID {
	return a.Account.ID
}
