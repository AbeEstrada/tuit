package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/mattn/go-mastodon"
)

// Client extends go-mastodon with endpoints it does not expose.
type Client struct {
	*mastodon.Client
}

func NewClient(client *mastodon.Client) *Client {
	return &Client{Client: client}
}

type idempotencyKey struct{}

// WithIdempotencyKey returns a context whose requests carry key in the
// Idempotency-Key header. Mastodon answers a repeated key with the status the
// first request created (for one hour), so retrying a post cannot post twice.
func WithIdempotencyKey(ctx context.Context, key string) context.Context {
	return context.WithValue(ctx, idempotencyKey{}, key)
}

// IdempotencyTransport sends the key set with WithIdempotencyKey. go-mastodon
// cannot set request headers, but it passes the context and sends through its
// exported http.Client, so the header is added here.
type IdempotencyTransport struct{}

func (IdempotencyTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	if key, ok := req.Context().Value(idempotencyKey{}).(string); ok {
		req = req.Clone(req.Context()) // a RoundTripper must not modify the caller's request
		req.Header.Set("Idempotency-Key", key)
	}
	return http.DefaultTransport.RoundTrip(req)
}

type UnreadCount struct {
	Count int `json:"count"`
}

func (c *Client) newRequest(ctx context.Context, method, apiPath string, body string) (*http.Request, error) {
	u, err := url.Parse(c.Config.Server)
	if err != nil {
		return nil, err
	}
	u = u.JoinPath(apiPath)

	req, err := http.NewRequestWithContext(ctx, method, u.String(), strings.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+c.Config.AccessToken)
	if body != "" {
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	}
	return req, nil
}

// GetNotificationsUnreadCount returns how many notifications are newer than
// the user's notification read marker (Mastodon 4.3 and later).
func (c *Client) GetNotificationsUnreadCount(ctx context.Context) (int, error) {
	req, err := c.newRequest(ctx, http.MethodGet, "api/v1/notifications/unread_count", "")
	if err != nil {
		return 0, err
	}

	resp, err := c.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var result UnreadCount
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return 0, err
	}

	return result.Count, nil
}

// SetNotificationsMarker records the newest notification the user has seen so
// the server's unread count resets.
func (c *Client) SetNotificationsMarker(ctx context.Context, lastReadID mastodon.ID) error {
	form := url.Values{}
	form.Set("notifications[last_read_id]", string(lastReadID))
	req, err := c.newRequest(ctx, http.MethodPost, "api/v1/markers", form.Encode())
	if err != nil {
		return err
	}

	resp, err := c.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}
	return nil
}
