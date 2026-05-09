package webapi

import (
	"context"
	"net/http"
	"net/url"
	"time"
)

type CommentMutation struct {
	ID             string           `json:"id,omitempty"`
	CreatedTime    string           `json:"createdTime,omitempty"`
	TaskID         string           `json:"taskId,omitempty"`
	ProjectID      string           `json:"projectId,omitempty"`
	Title          string           `json:"title,omitempty"`
	ReplyCommentID string           `json:"replyCommentId,omitempty"`
	ReplyUser      map[string]any   `json:"replyUserProfile,omitempty"`
	UserProfile    map[string]any   `json:"userProfile,omitempty"`
	Attachments    []CommentAttach  `json:"attachments,omitempty"`
	Mentions       []map[string]any `json:"mentions,omitempty"`
	IsNew          bool             `json:"isNew,omitempty"`
}

type CommentAttach struct {
	ID string `json:"id"`
}

func (c *Client) TaskComments(ctx context.Context, projectID string, taskID string) ([]map[string]any, error) {
	var out []map[string]any
	path := "/project/" + url.PathEscape(projectID) + "/task/" + url.PathEscape(taskID) + "/comments"
	if err := c.Do(ctx, http.MethodGet, path, nil, &out); err != nil {
		return nil, err
	}
	return out, nil
}

func (c *Client) CreateComment(ctx context.Context, projectID string, taskID string, comment CommentMutation) (map[string]any, error) {
	var out map[string]any
	comment = normalizeCommentForCreate(projectID, taskID, comment)
	path := "/project/" + url.PathEscape(projectID) + "/task/" + url.PathEscape(taskID) + "/comment"
	if err := c.Do(ctx, http.MethodPost, path, comment, &out); err != nil {
		return nil, err
	}
	return out, nil
}

func normalizeCommentForCreate(projectID string, taskID string, comment CommentMutation) CommentMutation {
	if comment.ID == "" {
		comment.ID = NewTaskID()
	}
	if comment.CreatedTime == "" {
		comment.CreatedTime = time.Now().UTC().Format("2006-01-02T15:04:05.000+0000")
	}
	if comment.TaskID == "" {
		comment.TaskID = taskID
	}
	if comment.ProjectID == "" {
		comment.ProjectID = projectID
	}
	if comment.UserProfile == nil {
		comment.UserProfile = map[string]any{"isMyself": true}
	}
	if comment.Mentions == nil {
		comment.Mentions = []map[string]any{}
	}
	comment.IsNew = true
	return comment
}

func (c *Client) UpdateComment(ctx context.Context, projectID string, taskID string, commentID string, comment CommentMutation) (map[string]any, error) {
	var out map[string]any
	path := "/project/" + url.PathEscape(projectID) + "/task/" + url.PathEscape(taskID) + "/comment/" + url.PathEscape(commentID)
	if err := c.Do(ctx, http.MethodPut, path, comment, &out); err != nil {
		return nil, err
	}
	return out, nil
}

func (c *Client) DeleteComment(ctx context.Context, projectID string, taskID string, commentID string) (map[string]any, error) {
	var out map[string]any
	path := "/project/" + url.PathEscape(projectID) + "/task/" + url.PathEscape(taskID) + "/comment/" + url.PathEscape(commentID)
	if err := c.Do(ctx, http.MethodDelete, path, nil, &out); err != nil {
		return nil, err
	}
	return out, nil
}
