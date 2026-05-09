package webapi

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestCommentEndpointsAndBodies(t *testing.T) {
	var seen []string
	var bodies []any
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		seen = append(seen, r.Method+" "+r.URL.RequestURI())
		var payload any
		if r.Body != nil && r.ContentLength != 0 {
			if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
				t.Fatalf("decode request for %s: %v", r.URL.Path, err)
			}
		}
		bodies = append(bodies, payload)
		if r.Method == http.MethodGet {
			_ = json.NewEncoder(w).Encode([]map[string]any{{"id": "c1", "title": "hello"}})
			return
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"id": "c1", "ok": true})
	}))
	defer server.Close()

	client := NewClient("test-token")
	client.BaseURL = server.URL
	ctx := context.Background()

	if _, err := client.TaskComments(ctx, "p 1", "t/1"); err != nil {
		t.Fatalf("TaskComments() error = %v", err)
	}
	if _, err := client.CreateComment(ctx, "p 1", "t/1", CommentMutation{Title: "hello"}); err != nil {
		t.Fatalf("CreateComment() error = %v", err)
	}
	if _, err := client.UpdateComment(ctx, "p 1", "t/1", "c 1", CommentMutation{Title: "updated"}); err != nil {
		t.Fatalf("UpdateComment() error = %v", err)
	}
	if _, err := client.DeleteComment(ctx, "p 1", "t/1", "c 1"); err != nil {
		t.Fatalf("DeleteComment() error = %v", err)
	}

	want := []string{
		"GET /project/p%201/task/t%2F1/comments",
		"POST /project/p%201/task/t%2F1/comment",
		"PUT /project/p%201/task/t%2F1/comment/c%201",
		"DELETE /project/p%201/task/t%2F1/comment/c%201",
	}
	if strings.Join(seen, "\n") != strings.Join(want, "\n") {
		t.Fatalf("seen endpoints:\n%s\nwant:\n%s", strings.Join(seen, "\n"), strings.Join(want, "\n"))
	}
	createBody := bodies[1].(map[string]any)
	if createBody["title"] != "hello" {
		t.Fatalf("create body = %#v", createBody)
	}
	if createBody["id"] == "" || createBody["createdTime"] == "" || createBody["taskId"] != "t/1" || createBody["projectId"] != "p 1" {
		t.Fatalf("create body missing webapp-generated fields: %#v", createBody)
	}
	userProfile := createBody["userProfile"].(map[string]any)
	if userProfile["isMyself"] != true {
		t.Fatalf("userProfile = %#v", userProfile)
	}
	updateBody := bodies[2].(map[string]any)
	if updateBody["title"] != "updated" {
		t.Fatalf("update body = %#v", updateBody)
	}
}
