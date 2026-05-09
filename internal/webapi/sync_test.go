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
			},
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
}
