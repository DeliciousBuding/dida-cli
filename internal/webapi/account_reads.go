package webapi

import (
	"context"
	"net/http"
	"net/url"
)

func (c *Client) AttachmentQuota(ctx context.Context) (map[string]any, error) {
	var underQuota bool
	if err := c.DoV1(ctx, http.MethodGet, "/attachment/isUnderQuota", nil, &underQuota); err != nil {
		return nil, err
	}
	var dailyLimit int64
	if err := c.DoV1(ctx, http.MethodGet, "/attachment/dailyLimit", nil, &dailyLimit); err != nil {
		return nil, err
	}
	return map[string]any{
		"underQuota": underQuota,
		"dailyLimit": dailyLimit,
	}, nil
}

func (c *Client) DailyReminderPreferences(ctx context.Context) (map[string]any, error) {
	var out map[string]any
	if err := c.Do(ctx, http.MethodGet, "/user/preferences/dailyReminder", nil, &out); err != nil {
		return nil, err
	}
	return out, nil
}

func (c *Client) ShareContacts(ctx context.Context) (map[string]any, error) {
	var out map[string]any
	if err := c.Do(ctx, http.MethodGet, "/share/shareContacts", nil, &out); err != nil {
		return nil, err
	}
	return out, nil
}

func (c *Client) RecentProjectUsers(ctx context.Context) ([]map[string]any, error) {
	var out []map[string]any
	if err := c.Do(ctx, http.MethodGet, "/project/share/recentProjectUsers", nil, &out); err != nil {
		return nil, err
	}
	return out, nil
}

func (c *Client) ProjectShares(ctx context.Context, projectID string) ([]map[string]any, error) {
	var out []map[string]any
	path := "/project/" + url.PathEscape(projectID) + "/shares"
	if err := c.Do(ctx, http.MethodGet, path, nil, &out); err != nil {
		return nil, err
	}
	return out, nil
}

func (c *Client) ProjectShareQuota(ctx context.Context, projectID string) (int64, error) {
	var out int64
	path := "/project/" + url.PathEscape(projectID) + "/share/check-quota"
	if err := c.Do(ctx, http.MethodGet, path, nil, &out); err != nil {
		return 0, err
	}
	return out, nil
}

func (c *Client) ProjectInviteURL(ctx context.Context, projectID string) (map[string]any, error) {
	var out map[string]any
	path := "/project/" + url.PathEscape(projectID) + "/collaboration/invite-url"
	if err := c.Do(ctx, http.MethodGet, path, nil, &out); err != nil {
		return nil, err
	}
	return out, nil
}

func (c *Client) CalendarSubscriptions(ctx context.Context) ([]map[string]any, error) {
	var out []map[string]any
	if err := c.Do(ctx, http.MethodGet, "/calendar/subscription", nil, &out); err != nil {
		return nil, err
	}
	return out, nil
}
