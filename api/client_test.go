package api

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	"github.com/mattn/go-mastodon"
)

// A stand-in for Mastodon's status endpoint: a repeated Idempotency-Key gets
// the status the first request created, anything else creates a new one.
type fakeStatuses struct {
	mu      sync.Mutex
	created int
	byKey   map[string]int
	keys    []string // header of every request, "" when absent
}

func (f *fakeStatuses) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	f.mu.Lock()
	defer f.mu.Unlock()
	key := r.Header.Get("Idempotency-Key")
	f.keys = append(f.keys, key)
	id, seen := f.byKey[key]
	if key == "" || !seen {
		f.created++
		id = f.created
		if key != "" {
			f.byKey[key] = id
		}
	}
	fmt.Fprintf(w, `{"id":"%d"}`, id)
}

func TestIdempotencyKeyReachesServer(t *testing.T) {
	fake := &fakeStatuses{byKey: map[string]int{}}
	srv := httptest.NewServer(fake)
	defer srv.Close()

	// The real go-mastodon client, so the test fails if it ever stops passing
	// the context or sending through its embedded http.Client.
	client := mastodon.NewClient(&mastodon.Config{Server: srv.URL, AccessToken: "token"})
	client.Transport = IdempotencyTransport{}
	post := func(ctx context.Context) mastodon.ID {
		t.Helper()
		status, err := client.PostStatus(ctx, &mastodon.Toot{Status: "hello"})
		if err != nil {
			t.Fatalf("PostStatus: %v", err)
		}
		return status.ID
	}

	keyed := WithIdempotencyKey(context.Background(), "draft-1")
	first, retry := post(keyed), post(keyed)
	if first != retry || fake.created != 1 {
		t.Fatalf("retry with the same key: ids %q and %q, %d statuses created, want one", first, retry, fake.created)
	}
	if other := post(context.Background()); other == first || fake.created != 2 {
		t.Fatalf("post without a key: id %q, %d statuses created, want a second status", other, fake.created)
	}
	if want := []string{"draft-1", "draft-1", ""}; fmt.Sprint(fake.keys) != fmt.Sprint(want) {
		t.Fatalf("Idempotency-Key headers = %q, want %q", fake.keys, want)
	}
}

func TestIdempotencyTransportLeavesRequestAlone(t *testing.T) {
	srv := httptest.NewServer(&fakeStatuses{byKey: map[string]int{}})
	defer srv.Close()

	req, err := http.NewRequestWithContext(WithIdempotencyKey(context.Background(), "k"), http.MethodPost, srv.URL, nil)
	if err != nil {
		t.Fatal(err)
	}
	resp, err := IdempotencyTransport{}.RoundTrip(req)
	if err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()
	if got := req.Header.Get("Idempotency-Key"); got != "" {
		t.Fatalf("caller's request was modified: Idempotency-Key = %q", got)
	}
}
