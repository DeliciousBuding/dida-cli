package webapi

import (
	"context"
	"encoding/json"
	"net/http"
	"net/url"
	"strconv"
)

type SyncPayload struct {
	InboxID       string           `json:"inboxId,omitempty"`
	CheckPoint    int64            `json:"checkPoint,omitempty"`
	Tasks         []map[string]any `json:"tasks,omitempty"`
	TaskAdds      []map[string]any `json:"taskAdds,omitempty"`
	TaskUpdates   []map[string]any `json:"taskUpdates,omitempty"`
	TaskDeletes   []map[string]any `json:"taskDeletes,omitempty"`
	Projects      []map[string]any `json:"projects,omitempty"`
	ProjectGroups []map[string]any `json:"projectGroups,omitempty"`
	Tags          []map[string]any `json:"tags,omitempty"`
	Checks        []map[string]any `json:"checks,omitempty"`
	Filters       []map[string]any `json:"filters,omitempty"`
	SyncOrder     any              `json:"syncOrderBean,omitempty"`
	SyncTaskOrder any              `json:"syncTaskOrderBean,omitempty"`
	Reminders     any              `json:"reminders,omitempty"`
	Raw           map[string]any   `json:"-"`
}

func (c *Client) FullSync(ctx context.Context) (*SyncPayload, error) {
	return c.SyncSince(ctx, 0)
}

func (c *Client) SyncSince(ctx context.Context, checkpoint int64) (*SyncPayload, error) {
	var raw map[string]any
	if err := c.Do(ctx, http.MethodGet, "/batch/check/"+strconv.FormatInt(checkpoint, 10), nil, &raw); err != nil {
		return nil, err
	}
	payload := &SyncPayload{Raw: raw}
	payload.InboxID, _ = raw["inboxId"].(string)
	payload.CheckPoint = int64ish(raw["checkPoint"])
	payload.Tasks = firstObjectSlice(raw, "tasks")
	if len(payload.Tasks) == 0 {
		payload.TaskAdds = nestedObjectSlice(raw, "syncTaskBean", "add")
		payload.TaskUpdates = nestedObjectSlice(raw, "syncTaskBean", "update")
		payload.TaskDeletes = nestedObjectSlice(raw, "syncTaskBean", "delete")
		payload.Tasks = append(payload.Tasks, payload.TaskAdds...)
		payload.Tasks = append(payload.Tasks, payload.TaskUpdates...)
	} else {
		payload.TaskAdds = payload.Tasks
	}
	payload.Projects = firstObjectSlice(raw, "projects", "projectProfiles")
	payload.ProjectGroups = objectSlice(raw["projectGroups"])
	payload.Tags = objectSlice(raw["tags"])
	payload.Checks = objectSlice(raw["checks"])
	payload.Filters = objectSlice(raw["filters"])
	payload.SyncOrder = raw["syncOrderBean"]
	payload.SyncTaskOrder = raw["syncTaskOrderBean"]
	payload.Reminders = firstPresent(raw, "reminders", "reminderChanges", "syncReminderBean")
	return payload, nil
}

func (c *Client) Settings(ctx context.Context) (map[string]any, error) {
	var out map[string]any
	if err := c.Do(ctx, http.MethodGet, "/user/preferences/settings", nil, &out); err != nil {
		return nil, err
	}
	return out, nil
}

func (c *Client) CompletedTasks(ctx context.Context, from string, to string, limit int) ([]map[string]any, error) {
	values := url.Values{}
	if from != "" {
		values.Set("from", from)
	}
	if to != "" {
		values.Set("to", to)
	}
	if limit > 0 {
		values.Set("limit", strconv.Itoa(limit))
	}
	path := "/project/all/completed"
	if encoded := values.Encode(); encoded != "" {
		path += "?" + encoded
	}
	var out []map[string]any
	if err := c.Do(ctx, http.MethodGet, path, nil, &out); err != nil {
		return nil, err
	}
	return out, nil
}

func firstObjectSlice(item map[string]any, keys ...string) []map[string]any {
	for _, key := range keys {
		if values := objectSlice(item[key]); len(values) > 0 {
			return values
		}
	}
	return nil
}

func nestedObjectSlice(item map[string]any, parent string, child string) []map[string]any {
	parentObj, ok := item[parent].(map[string]any)
	if !ok {
		return nil
	}
	return objectSlice(parentObj[child])
}

func firstPresent(item map[string]any, keys ...string) any {
	for _, key := range keys {
		if value, ok := item[key]; ok && value != nil {
			return value
		}
	}
	return nil
}

func objectSlice(value any) []map[string]any {
	items, ok := value.([]any)
	if !ok {
		return nil
	}
	out := make([]map[string]any, 0, len(items))
	for _, item := range items {
		if obj, ok := item.(map[string]any); ok {
			out = append(out, obj)
		}
	}
	return out
}

func int64ish(value any) int64 {
	switch typed := value.(type) {
	case int:
		return int64(typed)
	case int64:
		return typed
	case float64:
		return int64(typed)
	case json.Number:
		n, _ := typed.Int64()
		return n
	default:
		return 0
	}
}
