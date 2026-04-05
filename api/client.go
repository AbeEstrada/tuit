package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	"github.com/mattn/go-mastodon"
)

type Client struct {
	*mastodon.Client
}

func NewClient(client *mastodon.Client) *Client {
	return &Client{Client: client}
}

type UnreadCount struct {
	Count int `json:"count"`
}

func (c *Client) GetNotificationsUnreadCount(ctx context.Context) (int, error) {
	u, err := url.Parse(c.Config.Server)
	if err != nil {
		return 0, err
	}
	u = u.JoinPath("api/v1/notifications/unread_count")

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return 0, err
	}
	req.Header.Set("Authorization", "Bearer "+c.Config.AccessToken)

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
