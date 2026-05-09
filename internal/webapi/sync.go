package webapi

import (
	"context"
	"net/http"
)

type SyncPayload struct {
	InboxID       string           `json:"inboxId,omitempty"`
	Tasks         []map[string]any `json:"tasks,omitempty"`
	Projects      []map[string]any `json:"projects,omitempty"`
	ProjectGroups []map[string]any `json:"projectGroups,omitempty"`
	Tags          []map[string]any `json:"tags,omitempty"`
	Raw           map[string]any   `json:"-"`
}

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
