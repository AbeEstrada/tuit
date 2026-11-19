package tui

import (
	"strings"
	"testing"

	"github.com/mattn/go-mastodon"
)

func testItems(ids ...string) []TimelineItem {
	items := make([]TimelineItem, len(ids))
	for i, id := range ids {
		items[i] = StatusItem{Status: &mastodon.Status{ID: mastodon.ID(id)}}
	}
	return items
}

func itemIDs(items []TimelineItem) string {
	ids := make([]string, len(items))
	for i, item := range items {
		ids[i] = string(item.ID())
	}
	return strings.Join(ids, " ")
}

func TestTimelineInsert(t *testing.T) {
	tl := &Timeline{}
	if n := tl.Append(testItems("c", "d")); n != 2 || tl.Selected.ID() != "c" {
		t.Fatalf("Append added %d, selected %v", n, tl.Selected)
	}
	tl.scrollOffset = 1
	if n := tl.Prepend(testItems("a", "b", "c")); n != 2 {
		t.Fatalf("Prepend added %d, want 2 (c is a duplicate)", n)
	}
	if got := itemIDs(tl.Items); got != "a b c d" {
		t.Fatalf("items = %q", got)
	}
	if tl.scrollOffset != 3 {
		t.Fatalf("scrollOffset = %d, want 3 so the viewport stays on the same rows", tl.scrollOffset)
	}
	if n := tl.Append(testItems("d", "a")); n != 0 {
		t.Fatalf("Append of duplicates added %d", n)
	}
}

func TestTimelineDelete(t *testing.T) {
	cases := []struct {
		name, selected, remove, wantItems, wantSelected string
	}{
		{"selected middle moves to the next row", "b", "b", "a c", "c"},
		{"selected last moves to the previous row", "c", "c", "a b", "b"},
		{"other row keeps the selection", "a", "b", "a c", "a"},
		{"unknown id changes nothing", "a", "x", "a b c", "a"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			tl := &Timeline{Items: testItems("a", "b", "c")}
			tl.Selected = tl.Items[strings.Index("abc", tc.selected)]
			tl.Delete(mastodon.ID(tc.remove))
			if got := itemIDs(tl.Items); got != tc.wantItems || string(tl.Selected.ID()) != tc.wantSelected {
				t.Fatalf("items = %q selected = %q, want %q and %q", got, tl.Selected.ID(), tc.wantItems, tc.wantSelected)
			}
		})
	}

	tl := &Timeline{Items: testItems("a")}
	tl.Selected = tl.Items[0]
	if !tl.Delete("a") || tl.Selected != nil || len(tl.Items) != 0 {
		t.Fatalf("deleting the only item left items = %v selected = %v", tl.Items, tl.Selected)
	}
}

func TestJumpToUnread(t *testing.T) {
	cases := []struct {
		name string
		read string // ids already read, space separated
		want string // selected id afterwards
	}{
		{"new rows on top", "c d", "b"},
		{"unread only at the bottom is ignored", "a b", "c"},
		{"nothing read", "", "c"},
		{"everything read", "a b c d e", "c"},
		{"continues above the lowest read row", "c e", "d"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			v := CreateTimelineView()
			tl := &Timeline{Items: testItems("a", "b", "c", "d", "e")}
			tl.Selected = tl.Items[2]
			v.stack = []*Timeline{tl}
			for _, id := range strings.Fields(tc.read) {
				v.markRead(mastodon.ID(id))
			}
			v.jumpToUnread()
			if got := string(tl.Selected.ID()); got != tc.want {
				t.Fatalf("selected = %q, want %q", got, tc.want)
			}
		})
	}
}

// The timeline is read from the bottom up. This is the layout from a real
// session: the top four rows were seen on arrival, the bottom four were read on
// the way up, and twelve unread rows sit in between. Each g must land on the
// next unread row above the ones already read, however many read rows are on top.
func TestJumpToUnreadWalksUpFromTheBottom(t *testing.T) {
	ids := strings.Fields("r01 r02 r03 r04 r05 r06 r07 r08 r09 r10 r11 r12 r13 r14 r15 r16 r17 r18 r19 r20")
	v := CreateTimelineView()
	tl := &Timeline{Items: testItems(ids...)}
	tl.Selected = tl.Items[18] // r19
	v.stack = []*Timeline{tl}
	for _, id := range strings.Fields("r01 r02 r03 r04 r17 r18 r19 r20") {
		v.markRead(mastodon.ID(id))
	}

	var visited []string
	for range 13 {
		v.jumpToUnread()
		visited = append(visited, string(tl.Selected.ID()))
	}
	// Twelve unread rows from r16 up to r05, then nothing is left to jump to.
	want := "r16 r15 r14 r13 r12 r11 r10 r09 r08 r07 r06 r05 r05"
	if got := strings.Join(visited, " "); got != want {
		t.Fatalf("g visited %q, want %q", got, want)
	}
}
