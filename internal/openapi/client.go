package openapi

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

type Client struct {
	BaseURL    string
	Token      string
	HTTPClient *http.Client
}

func NewClient(token string) *Client {
	return &Client{
		BaseURL:    DefaultAPIBaseURL,
		Token:      token,
		HTTPClient: http.DefaultClient,
	}
}

func (c *Client) Do(ctx context.Context, method string, path string, body io.Reader, out any) error {
	if strings.TrimSpace(c.Token) == "" {
		return fmt.Errorf("missing openapi access token")
	}
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}
	req, err := http.NewRequestWithContext(ctx, method, strings.TrimRight(c.BaseURL, "/")+path, body)
	if err != nil {
		return fmt.Errorf("build request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+c.Token)
	req.Header.Set("Accept", "application/json")
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	httpClient := c.HTTPClient
	if httpClient == nil {
		httpClient = http.DefaultClient
	}
	resp, err := httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("openapi request failed: %w", err)
	}
	defer resp.Body.Close()
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("read response: %w", err)
	}
	if resp.StatusCode >= 400 {
		return fmt.Errorf("openapi returned HTTP %d: %s", resp.StatusCode, summarizeBody(string(data)))
	}
	if out == nil || len(data) == 0 {
		return nil
	}
	if err := json.Unmarshal(data, out); err != nil {
		return fmt.Errorf("decode response: %w", err)
	}
	return nil
}

func (c *Client) Projects(ctx context.Context) ([]map[string]any, error) {
	var out []map[string]any
	if err := c.Do(ctx, http.MethodGet, "/project", nil, &out); err != nil {
		return nil, err
	}
	return out, nil
}

func (c *Client) Project(ctx context.Context, projectID string) (map[string]any, error) {
	var out map[string]any
	if err := c.Do(ctx, http.MethodGet, "/project/"+url.PathEscape(projectID), nil, &out); err != nil {
		return nil, err
	}
	return out, nil
}

func (c *Client) ProjectData(ctx context.Context, projectID string) (map[string]any, error) {
	var out map[string]any
	if err := c.Do(ctx, http.MethodGet, "/project/"+url.PathEscape(projectID)+"/data", nil, &out); err != nil {
		return nil, err
	}
	return out, nil
}

func (c *Client) Task(ctx context.Context, projectID string, taskID string) (map[string]any, error) {
	var out map[string]any
	path := "/project/" + url.PathEscape(projectID) + "/task/" + url.PathEscape(taskID)
	if err := c.Do(ctx, http.MethodGet, path, nil, &out); err != nil {
		return nil, err
	}
	return out, nil
}

func (c *Client) CreateTask(ctx context.Context, payload map[string]any) (map[string]any, error) {
	var out map[string]any
	if err := c.doJSON(ctx, http.MethodPost, "/task", payload, &out); err != nil {
		return nil, err
	}
	return out, nil
}

func (c *Client) UpdateTask(ctx context.Context, taskID string, payload map[string]any) (map[string]any, error) {
	var out map[string]any
	if err := c.doJSON(ctx, http.MethodPost, "/task/"+url.PathEscape(taskID), payload, &out); err != nil {
		return nil, err
	}
	return out, nil
}

func (c *Client) CompleteTask(ctx context.Context, projectID string, taskID string) error {
	path := "/project/" + url.PathEscape(projectID) + "/task/" + url.PathEscape(taskID) + "/complete"
	return c.Do(ctx, http.MethodPost, path, nil, nil)
}

func (c *Client) DeleteTask(ctx context.Context, projectID string, taskID string) error {
	path := "/project/" + url.PathEscape(projectID) + "/task/" + url.PathEscape(taskID)
	return c.Do(ctx, http.MethodDelete, path, nil, nil)
}

func (c *Client) MoveTasks(ctx context.Context, payload any) (any, error) {
	var out any
	if err := c.doJSON(ctx, http.MethodPost, "/task/move", payload, &out); err != nil {
		return nil, err
	}
	return out, nil
}

func (c *Client) CompletedTasks(ctx context.Context, payload map[string]any) ([]map[string]any, error) {
	var out []map[string]any
	if err := c.doJSON(ctx, http.MethodPost, "/task/completed", payload, &out); err != nil {
		return nil, err
	}
	return out, nil
}

func (c *Client) FilterTasks(ctx context.Context, payload map[string]any) ([]map[string]any, error) {
	var out []map[string]any
	if err := c.doJSON(ctx, http.MethodPost, "/task/filter", payload, &out); err != nil {
		return nil, err
	}
	return out, nil
}

func (c *Client) doJSON(ctx context.Context, method string, path string, payload any, out any) error {
	data, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("encode request: %w", err)
	}
	return c.Do(ctx, method, path, bytes.NewReader(data), out)
}
