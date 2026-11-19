package tui

import (
	"errors"
	"strconv"
	"testing"

	"github.com/mattn/go-mastodon"
)

// fakeFeed stands in for a timeline endpoint holding statuses 1..newest. Like
// Mastodon, min_id returns the page right above an ID, newest first.
type fakeFeed struct {
	newest      int
	maxLimit    int  // the server's own cap on the page size
	oldestFirst bool // return pages in ascending order
	ignoreMinID bool // always answer with the newest page
	fail        error
	requests    int
}

func (f *fakeFeed) fetch(pg *mastodon.Pagination) ([]TimelineItem, error) {
	f.requests++
	if f.fail != nil {
		return nil, f.fail
	}
	limit := min(int(pg.Limit), f.maxLimit)
	from, _ := strconv.Atoi(string(pg.MinID))
	if f.ignoreMinID {
		from = max(0, f.newest-limit)
	}
	to := min(f.newest, from+limit)
	var ids []string
	for id := to; id > from; id-- {
		ids = append(ids, strconv.Itoa(id))
	}
	if f.oldestFirst {
		for i, j := 0, len(ids)-1; i < j; i, j = i+1, j-1 {
			ids[i], ids[j] = ids[j], ids[i]
		}
	}
	return testItems(ids...), nil
}

func TestNewerPages(t *testing.T) {
	cases := []struct {
		name         string
		feed         fakeFeed
		since        int
		wantOK       bool
		wantCount    int // items returned, all of since+1..newest when ok
		wantRequests int
	}{
		{"nothing new", fakeFeed{newest: 95, maxLimit: 40}, 95, true, 0, 1},
		// 95 to 220 crosses from two to three digits: longer IDs are newer.
		{"125 missed arrive in four pages", fakeFeed{newest: 220, maxLimit: 40}, 95, true, 125, 5},
		{"a server that caps pages at 30 is still complete", fakeFeed{newest: 220, maxLimit: 30}, 120, true, 100, 5},
		{"pages in ascending order are still complete", fakeFeed{newest: 220, maxLimit: 40, oldestFirst: true}, 120, true, 100, 4},
		{"exactly the cap is complete", fakeFeed{newest: 900, maxLimit: 40}, 100, true, 800, 21},
		{"one past the cap gives up", fakeFeed{newest: 901, maxLimit: 40}, 100, false, 0, 21},
		{"a server that ignores min_id gives up quickly", fakeFeed{newest: 500, maxLimit: 40, ignoreMinID: true}, 100, false, 0, 2},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			feed := tc.feed
			items, ok, err := newerPages(mastodon.ID(strconv.Itoa(tc.since)), feed.fetch)
			if err != nil || ok != tc.wantOK || len(items) != tc.wantCount || feed.requests != tc.wantRequests {
				t.Fatalf("got %d items, ok=%v, err=%v, %d requests; want %d items, ok=%v, %d requests",
					len(items), ok, err, feed.requests, tc.wantCount, tc.wantOK, tc.wantRequests)
			}
			// Newest first with no hole: newest, newest-1, ... down to since+1.
			for i, item := range items {
				if want := strconv.Itoa(feed.newest - i); string(item.ID()) != want {
					t.Fatalf("item %d is %q, want %q", i, item.ID(), want)
				}
			}
		})
	}
}

func TestNewerPagesReportsErrors(t *testing.T) {
	feed := fakeFeed{newest: 200, maxLimit: 40, fail: errors.New("boom")}
	items, ok, err := newerPages("100", feed.fetch)
	// ok stays true so the caller reports the error instead of refetching from the top.
	if err == nil || !ok || items != nil {
		t.Fatalf("got items=%v ok=%v err=%v, want the error with ok=true", items, ok, err)
	}
}
