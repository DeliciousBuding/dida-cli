package webapi

import (
	"context"
	"net/http"
)

type SyncPayload struct {
	InboxID       string         `json:"inboxId,omitempty"`
	Tasks         []jsonObject   `json:"tasks,omitempty"`
	Projects      []jsonObject   `json:"projects,omitempty"`
	ProjectGroups []jsonObject   `json:"projectGroups,omitempty"`
	Tags          []jsonObject   `json:"tags,omitempty"`
	Raw           map[string]any `json:"-"`
}

type jsonObject map[string]any

func (c *Client) FullSync(ctx context.Context) (*SyncPayload, error) {
	var raw map[string]any
	if err := c.Do(ctx, http.MethodGet, "/batch/check/0", nil, &raw); err != nil {
		return nil, err
	}
	payload := &SyncPayload{Raw: raw}
	payload.InboxID, _ = raw["inboxId"].(string)
	payload.Tasks = objectSlice(raw["tasks"])
	payload.Projects = objectSlice(raw["projects"])
	payload.ProjectGroups = objectSlice(raw["projectGroups"])
	payload.Tags = objectSlice(raw["tags"])
	return payload, nil
}

func objectSlice(value any) []jsonObject {
	items, ok := value.([]any)
	if !ok {
		return nil
	}
	out := make([]jsonObject, 0, len(items))
	for _, item := range items {
		if obj, ok := item.(map[string]any); ok {
			out = append(out, obj)
		}
	}
	return out
}
