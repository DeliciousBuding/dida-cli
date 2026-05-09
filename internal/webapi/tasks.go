package webapi

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"net/http"
	"time"
)

type TaskMutation struct {
	ID        string `json:"id,omitempty"`
	ProjectID string `json:"projectId"`
	Title     string `json:"title,omitempty"`
	Content   string `json:"content,omitempty"`
	DueDate   string `json:"dueDate,omitempty"`
	Priority  int    `json:"priority,omitempty"`
	Status    int    `json:"status,omitempty"`
	TimeZone  string `json:"timeZone,omitempty"`
}

func NewTaskID() string {
	var buf [12]byte
	if _, err := rand.Read(buf[:]); err != nil {
		return hex.EncodeToString([]byte(time.Now().Format("060102150405")))[:24]
	}
	return hex.EncodeToString(buf[:])
}

func (c *Client) CreateTask(ctx context.Context, task TaskMutation) (map[string]any, error) {
	if task.ID == "" {
		task.ID = NewTaskID()
	}
	if task.TimeZone == "" {
		task.TimeZone = "Asia/Shanghai"
	}
	return c.batchTask(ctx, map[string]any{"add": []TaskMutation{task}})
}

func (c *Client) UpdateTask(ctx context.Context, task TaskMutation) (map[string]any, error) {
	return c.batchTask(ctx, map[string]any{"update": []TaskMutation{task}})
}

func (c *Client) CompleteTask(ctx context.Context, taskID string, projectID string) (map[string]any, error) {
	return c.batchTask(ctx, map[string]any{"update": []TaskMutation{{ID: taskID, ProjectID: projectID, Status: 2}}})
}

func (c *Client) DeleteTask(ctx context.Context, taskID string, projectID string) (map[string]any, error) {
	payload := map[string]any{"delete": []map[string]string{{"taskId": taskID, "projectId": projectID}}}
	return c.batchTask(ctx, payload)
}

func (c *Client) batchTask(ctx context.Context, payload map[string]any) (map[string]any, error) {
	var out map[string]any
	if err := c.Do(ctx, http.MethodPost, "/batch/task", payload, &out); err != nil {
		return nil, err
	}
	return out, nil
}
