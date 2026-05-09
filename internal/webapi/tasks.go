package webapi

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"net/http"
	"net/url"
	"time"
)

type TaskMutation struct {
	ID         string        `json:"id,omitempty"`
	ProjectID  string        `json:"projectId"`
	Title      string        `json:"title,omitempty"`
	Content    string        `json:"content,omitempty"`
	Desc       string        `json:"desc,omitempty"`
	AllDay     *bool         `json:"allDay,omitempty"`
	StartDate  string        `json:"startDate,omitempty"`
	DueDate    string        `json:"dueDate,omitempty"`
	TimeZone   string        `json:"timeZone,omitempty"`
	Reminders  []string      `json:"reminders,omitempty"`
	Repeat     string        `json:"repeat,omitempty"`
	RepeatFrom string        `json:"repeatFrom,omitempty"`
	RepeatFlag string        `json:"repeatFlag,omitempty"`
	Priority   *int          `json:"priority,omitempty"`
	Status     *int          `json:"status,omitempty"`
	ColumnID   string        `json:"columnId,omitempty"`
	Tags       []string      `json:"tags,omitempty"`
	Items      []SubTaskItem `json:"items,omitempty"`
	IsFloating *bool         `json:"isFloating,omitempty"`
}

type SubTaskItem struct {
	ID     string `json:"id,omitempty"`
	Title  string `json:"title"`
	Status int    `json:"status,omitempty"`
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
	status := 2
	return c.batchTask(ctx, map[string]any{"update": []TaskMutation{{ID: taskID, ProjectID: projectID, Status: &status}}})
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

func (c *Client) ProjectTasks(ctx context.Context, projectID string) ([]map[string]any, error) {
	var out []map[string]any
	path := "/project/" + url.PathEscape(projectID) + "/tasks"
	if err := c.Do(ctx, http.MethodGet, path, nil, &out); err != nil {
		return nil, err
	}
	return out, nil
}
