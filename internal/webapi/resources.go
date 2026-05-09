package webapi

import (
	"context"
	"net/http"
	"net/url"
)

type ProjectMutation struct {
	ID         string `json:"id,omitempty"`
	Name       string `json:"name,omitempty"`
	Color      string `json:"color,omitempty"`
	SortOrder  int64  `json:"sortOrder,omitempty"`
	Closed     bool   `json:"closed,omitempty"`
	GroupID    string `json:"groupId,omitempty"`
	ViewMode   string `json:"viewMode,omitempty"`
	Permission string `json:"permission,omitempty"`
	Kind       string `json:"kind,omitempty"`
}

type ProjectGroupMutation struct {
	ID        string `json:"id,omitempty"`
	Name      string `json:"name,omitempty"`
	SortOrder int64  `json:"sortOrder,omitempty"`
	ShowAll   bool   `json:"showAll,omitempty"`
}

type TagMutation struct {
	Name      string `json:"name"`
	Label     string `json:"label,omitempty"`
	SortOrder int64  `json:"sortOrder,omitempty"`
	SortType  string `json:"sortType,omitempty"`
	Color     string `json:"color,omitempty"`
	Parent    string `json:"parent,omitempty"`
	RawName   string `json:"rawName,omitempty"`
}

type TaskMovePayload struct {
	TaskID        string `json:"taskId"`
	FromProjectID string `json:"fromProjectId"`
	ToProjectID   string `json:"toProjectId"`
}

type TaskParentPayload struct {
	TaskID    string `json:"taskId"`
	ParentID  string `json:"parentId"`
	ProjectID string `json:"projectId"`
}

func (c *Client) CreateProject(ctx context.Context, project ProjectMutation) (map[string]any, error) {
	if project.ID == "" {
		project.ID = NewTaskID()
	}
	return c.batchProject(ctx, map[string]any{"add": []ProjectMutation{project}})
}

func (c *Client) UpdateProject(ctx context.Context, project ProjectMutation) (map[string]any, error) {
	return c.batchProject(ctx, map[string]any{"update": []ProjectMutation{project}})
}

func (c *Client) DeleteProject(ctx context.Context, projectID string) (map[string]any, error) {
	return c.batchProject(ctx, map[string]any{"delete": []string{projectID}})
}

func (c *Client) batchProject(ctx context.Context, payload map[string]any) (map[string]any, error) {
	var out map[string]any
	if err := c.Do(ctx, http.MethodPost, "/batch/project", payload, &out); err != nil {
		return nil, err
	}
	return out, nil
}

func (c *Client) CreateProjectGroup(ctx context.Context, group ProjectGroupMutation) (map[string]any, error) {
	if group.ID == "" {
		group.ID = NewTaskID()
	}
	return c.batchProjectGroup(ctx, map[string]any{"add": []ProjectGroupMutation{group}})
}

func (c *Client) UpdateProjectGroup(ctx context.Context, group ProjectGroupMutation) (map[string]any, error) {
	return c.batchProjectGroup(ctx, map[string]any{"update": []ProjectGroupMutation{group}})
}

func (c *Client) DeleteProjectGroup(ctx context.Context, groupID string) (map[string]any, error) {
	return c.batchProjectGroup(ctx, map[string]any{"delete": []string{groupID}})
}

func (c *Client) batchProjectGroup(ctx context.Context, payload map[string]any) (map[string]any, error) {
	var out map[string]any
	if err := c.Do(ctx, http.MethodPost, "/batch/projectGroup", payload, &out); err != nil {
		return nil, err
	}
	return out, nil
}

func (c *Client) CreateTag(ctx context.Context, tag TagMutation) (map[string]any, error) {
	return c.batchTag(ctx, map[string]any{"add": []TagMutation{tag}})
}

func (c *Client) UpdateTag(ctx context.Context, tag TagMutation) (map[string]any, error) {
	return c.batchTag(ctx, map[string]any{"update": []TagMutation{tag}})
}

func (c *Client) DeleteTag(ctx context.Context, name string) (map[string]any, error) {
	var out map[string]any
	path := "/tag?name=" + url.QueryEscape(name)
	if err := c.Do(ctx, http.MethodDelete, path, nil, &out); err != nil {
		return nil, err
	}
	return out, nil
}

func (c *Client) RenameTag(ctx context.Context, oldName string, newName string) (map[string]any, error) {
	var out map[string]any
	payload := map[string]string{"name": oldName, "newName": newName}
	if err := c.Do(ctx, http.MethodPut, "/tag/rename", payload, &out); err != nil {
		return nil, err
	}
	return out, nil
}

func (c *Client) MergeTags(ctx context.Context, fromName string, toName string) (map[string]any, error) {
	var out map[string]any
	payload := map[string]string{"from": fromName, "to": toName}
	if err := c.Do(ctx, http.MethodPut, "/tag/merge", payload, &out); err != nil {
		return nil, err
	}
	return out, nil
}

func (c *Client) batchTag(ctx context.Context, payload map[string]any) (map[string]any, error) {
	var out map[string]any
	if err := c.Do(ctx, http.MethodPost, "/batch/tag", payload, &out); err != nil {
		return nil, err
	}
	return out, nil
}

func (c *Client) CreateColumn(ctx context.Context, projectID string, name string) (map[string]any, error) {
	var out map[string]any
	payload := map[string]string{"projectId": projectID, "name": name}
	if err := c.Do(ctx, http.MethodPost, "/column", payload, &out); err != nil {
		return nil, err
	}
	return out, nil
}

func (c *Client) MoveTask(ctx context.Context, taskID string, fromProjectID string, toProjectID string) (map[string]any, error) {
	var out map[string]any
	payload := []TaskMovePayload{{TaskID: taskID, FromProjectID: fromProjectID, ToProjectID: toProjectID}}
	if err := c.Do(ctx, http.MethodPost, "/batch/taskProject", payload, &out); err != nil {
		return nil, err
	}
	return out, nil
}

func (c *Client) SetTaskParent(ctx context.Context, taskID string, parentID string, projectID string) (map[string]any, error) {
	var out map[string]any
	payload := []TaskParentPayload{{TaskID: taskID, ParentID: parentID, ProjectID: projectID}}
	if err := c.Do(ctx, http.MethodPost, "/batch/taskParent", payload, &out); err != nil {
		return nil, err
	}
	return out, nil
}
