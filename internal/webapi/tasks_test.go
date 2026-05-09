package webapi

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
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
