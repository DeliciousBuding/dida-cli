package webapi

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestTaskMutationsUseBatchTaskEndpoint(t *testing.T) {
	var requests []map[string]any
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/batch/task" {
			t.Fatalf("path = %s, want /batch/task", r.URL.Path)
		}
		var payload map[string]any
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			t.Fatalf("decode request: %v", err)
		}
		requests = append(requests, payload)
		_ = json.NewEncoder(w).Encode(map[string]any{"ok": true})
	}))
	defer server.Close()

	client := NewClient("test-token")
	client.BaseURL = server.URL
	ctx := context.Background()

	if _, err := client.CreateTask(ctx, TaskMutation{ID: "t1", ProjectID: "p1", Title: "Create"}); err != nil {
		t.Fatalf("CreateTask() error = %v", err)
	}
	if _, err := client.UpdateTask(ctx, TaskMutation{ID: "t1", ProjectID: "p1", Title: "Update"}); err != nil {
		t.Fatalf("UpdateTask() error = %v", err)
	}
	if _, err := client.CompleteTask(ctx, "t1", "p1"); err != nil {
		t.Fatalf("CompleteTask() error = %v", err)
	}
	if _, err := client.DeleteTask(ctx, "t1", "p1"); err != nil {
		t.Fatalf("DeleteTask() error = %v", err)
	}

	if len(requests) != 4 {
		t.Fatalf("request count = %d, want 4", len(requests))
	}
	if _, ok := requests[0]["add"]; !ok {
		t.Fatalf("create payload missing add: %#v", requests[0])
	}
	if _, ok := requests[1]["update"]; !ok {
		t.Fatalf("update payload missing update: %#v", requests[1])
	}
	if _, ok := requests[2]["update"]; !ok {
		t.Fatalf("complete payload missing update: %#v", requests[2])
	}
	if _, ok := requests[3]["delete"]; !ok {
		t.Fatalf("delete payload missing delete: %#v", requests[3])
	}
}

func TestTaskUpdateCanSendPriorityZero(t *testing.T) {
	var rawBody string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		data, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatalf("read request: %v", err)
		}
		rawBody = string(data)
		_ = json.NewEncoder(w).Encode(map[string]any{"ok": true})
	}))
	defer server.Close()

	client := NewClient("test-token")
	client.BaseURL = server.URL
	priority := 0
	if _, err := client.UpdateTask(context.Background(), TaskMutation{ID: "t1", ProjectID: "p1", Priority: &priority}); err != nil {
		t.Fatalf("UpdateTask() error = %v", err)
	}
	if !strings.Contains(rawBody, `"priority":0`) {
		t.Fatalf("request body missing priority zero: %s", rawBody)
	}
}

func TestProjectTasksUsesProjectEndpoint(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.Method + " " + r.URL.Path; got != "GET /project/p1/tasks" {
			t.Fatalf("request = %s, want GET /project/p1/tasks", got)
		}
		_ = json.NewEncoder(w).Encode([]map[string]any{{"id": "t1", "projectId": "p1", "title": "Task"}})
	}))
	defer server.Close()

	client := NewClient("test-token")
	client.BaseURL = server.URL
	tasks, err := client.ProjectTasks(context.Background(), "p1")
	if err != nil {
		t.Fatalf("ProjectTasks() error = %v", err)
	}
	if len(tasks) != 1 {
		t.Fatalf("tasks len = %d, want 1", len(tasks))
	}
}

func TestTaskDueActivityCountsUsesExpectedEndpoint(t *testing.T) {
	var payload map[string]string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.Method + " " + r.URL.Path; got != "POST /task/activity/count/all" {
			t.Fatalf("request = %s, want POST /task/activity/count/all", got)
		}
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			t.Fatalf("decode request: %v", err)
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"t1": 2})
	}))
	defer server.Close()

	client := NewClient("test-token")
	client.BaseURL = server.URL
	counts, err := client.TaskDueActivityCounts(context.Background())
	if err != nil {
		t.Fatalf("TaskDueActivityCounts() error = %v", err)
	}
	if payload["action"] != "T_DUE" {
		t.Fatalf("action = %q, want T_DUE", payload["action"])
	}
	if counts["t1"] != float64(2) {
		t.Fatalf("counts = %#v, want t1=2", counts)
	}
}
