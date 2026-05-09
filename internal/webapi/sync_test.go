package webapi

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestFullSyncSupportsBatchCheckShape(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/batch/check/0" {
			t.Fatalf("path = %s, want /batch/check/0", r.URL.Path)
		}
		if got := r.Header.Get("Cookie"); got != "t=test-token" {
			t.Fatalf("cookie = %q", got)
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"inboxId": "inbox1",
			"projectProfiles": []map[string]any{
				{"id": "p1", "name": "Work"},
			},
			"syncTaskBean": map[string]any{
				"add": []map[string]any{
					{"id": "t1", "title": "Added", "projectId": "p1"},
				},
				"update": []map[string]any{
					{"id": "t2", "title": "Updated", "projectId": "p1"},
				},
				"delete": []map[string]any{
					{"taskId": "t3", "projectId": "p1"},
				},
			},
			"syncOrderBean":     map[string]any{"update": []any{"order"}},
			"syncTaskOrderBean": map[string]any{"update": []any{"taskOrder"}},
			"syncReminderBean":  map[string]any{"delete": []any{"reminder"}},
		})
	}))
	defer server.Close()

	client := NewClient("test-token")
	client.BaseURL = server.URL
	payload, err := client.FullSync(context.Background())
	if err != nil {
		t.Fatalf("FullSync() error = %v", err)
	}
	if len(payload.Projects) != 1 {
		t.Fatalf("projects len = %d, want 1", len(payload.Projects))
	}
	if len(payload.Tasks) != 2 {
		t.Fatalf("tasks len = %d, want 2", len(payload.Tasks))
	}
	if len(payload.TaskAdds) != 1 || len(payload.TaskUpdates) != 1 || len(payload.TaskDeletes) != 1 {
		t.Fatalf("delta counts add/update/delete = %d/%d/%d, want 1/1/1", len(payload.TaskAdds), len(payload.TaskUpdates), len(payload.TaskDeletes))
	}
	if payload.SyncOrder == nil || payload.SyncTaskOrder == nil || payload.Reminders == nil {
		t.Fatalf("expected order and reminder deltas to be preserved")
	}
}

func TestCompletedTasksBuildsWebQuery(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/project/all/completed" {
			t.Fatalf("path = %s, want /project/all/completed", r.URL.Path)
		}
		if got := r.URL.Query().Get("from"); got != "2026-05-09 00:00:00" {
			t.Fatalf("from = %q", got)
		}
		if got := r.URL.Query().Get("to"); got != "2026-05-09 23:59:59" {
			t.Fatalf("to = %q", got)
		}
		if got := r.URL.Query().Get("limit"); got != "100" {
			t.Fatalf("limit = %q", got)
		}
		_ = json.NewEncoder(w).Encode([]map[string]any{
			{"id": "t1", "projectId": "p1", "title": "Done", "status": 2},
		})
	}))
	defer server.Close()

	client := NewClient("test-token")
	client.BaseURL = server.URL
	tasks, err := client.CompletedTasks(context.Background(), "2026-05-09 00:00:00", "2026-05-09 23:59:59", 100)
	if err != nil {
		t.Fatalf("CompletedTasks() error = %v", err)
	}
	if len(tasks) != 1 {
		t.Fatalf("tasks len = %d, want 1", len(tasks))
	}
}
