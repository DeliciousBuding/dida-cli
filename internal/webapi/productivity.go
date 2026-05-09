package webapi

import (
	"context"
	"net/http"
	"net/url"
	"strconv"
)

func (c *Client) PomodoroPreferences(ctx context.Context) (map[string]any, error) {
	var out map[string]any
	if err := c.Do(ctx, http.MethodGet, "/user/preferences/pomodoro", nil, &out); err != nil {
		return nil, err
	}
	return out, nil
}

func (c *Client) Pomodoros(ctx context.Context, fromMillis int64, toMillis int64) ([]map[string]any, error) {
	return c.pomodoroRange(ctx, "/pomodoros", fromMillis, toMillis)
}

func (c *Client) PomodoroTimings(ctx context.Context, fromMillis int64, toMillis int64) ([]map[string]any, error) {
	return c.pomodoroRange(ctx, "/pomodoros/timing", fromMillis, toMillis)
}

func (c *Client) pomodoroRange(ctx context.Context, basePath string, fromMillis int64, toMillis int64) ([]map[string]any, error) {
	values := url.Values{}
	if fromMillis > 0 {
		values.Set("from", strconv.FormatInt(fromMillis, 10))
	} else {
		values.Set("from", "0")
	}
	if toMillis > 0 {
		values.Set("to", strconv.FormatInt(toMillis, 10))
	}
	path := basePath + "?" + values.Encode()
	var out []map[string]any
	if err := c.Do(ctx, http.MethodGet, path, nil, &out); err != nil {
		return nil, err
	}
	return out, nil
}

func (c *Client) HabitPreferences(ctx context.Context) (map[string]any, error) {
	var out map[string]any
	if err := c.Do(ctx, http.MethodGet, "/user/preferences/habit?platform=web", nil, &out); err != nil {
		return nil, err
	}
	return out, nil
}

func (c *Client) Habits(ctx context.Context) ([]map[string]any, error) {
	var out []map[string]any
	if err := c.Do(ctx, http.MethodGet, "/habits", nil, &out); err != nil {
		return nil, err
	}
	return out, nil
}

func (c *Client) HabitSections(ctx context.Context) ([]map[string]any, error) {
	var out []map[string]any
	if err := c.Do(ctx, http.MethodGet, "/habitSections", nil, &out); err != nil {
		return nil, err
	}
	return out, nil
}
