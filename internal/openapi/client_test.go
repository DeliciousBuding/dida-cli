package openapi

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestProjectsUsesExpectedEndpoint(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.Method + " " + r.URL.Path; got != "GET /project" {
			t.Fatalf("request = %s, want GET /project", got)
		}
		if got := r.Header.Get("Authorization"); got != "Bearer access-token" {
			t.Fatalf("authorization = %q", got)
		}
		_ = json.NewEncoder(w).Encode([]map[string]any{{"id": "p1", "name": "Project"}})
	}))
	defer server.Close()

	client := NewClient("access-token")
	client.BaseURL = server.URL
	projects, err := client.Projects(context.Background())
	if err != nil {
		t.Fatalf("Projects() error = %v", err)
	}
	if len(projects) != 1 {
		t.Fatalf("projects len = %d, want 1", len(projects))
	}
}

func TestProjectUsesExpectedEndpoint(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.Method + " " + r.URL.Path; got != "GET /project/p1" {
			t.Fatalf("request = %s, want GET /project/p1", got)
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"id": "p1"})
	}))
	defer server.Close()

	client := NewClient("access-token")
	client.BaseURL = server.URL
	project, err := client.Project(context.Background(), "p1")
	if err != nil {
		t.Fatalf("Project() error = %v", err)
	}
	if project["id"] != "p1" {
		t.Fatalf("project = %#v", project)
	}
}

func TestProjectDataUsesExpectedEndpoint(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.Method + " " + r.URL.Path; got != "GET /project/inbox/data" {
			t.Fatalf("request = %s, want GET /project/inbox/data", got)
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"tasks": []any{}})
	}))
	defer server.Close()

	client := NewClient("access-token")
	client.BaseURL = server.URL
	data, err := client.ProjectData(context.Background(), "inbox")
	if err != nil {
		t.Fatalf("ProjectData() error = %v", err)
	}
	if _, ok := data["tasks"]; !ok {
		t.Fatalf("data = %#v", data)
	}
}

func TestCreateTaskUsesExpectedEndpoint(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.Method + " " + r.URL.Path; got != "POST /task" {
			t.Fatalf("request = %s, want POST /task", got)
		}
		body, _ := io.ReadAll(r.Body)
		if string(body) != `{"projectId":"p1","title":"Task"}` {
			t.Fatalf("body = %s", body)
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"id": "t1"})
	}))
	defer server.Close()

	client := NewClient("access-token")
	client.BaseURL = server.URL
	task, err := client.CreateTask(context.Background(), map[string]any{"projectId": "p1", "title": "Task"})
	if err != nil {
		t.Fatalf("CreateTask() error = %v", err)
	}
	if task["id"] != "t1" {
		t.Fatalf("task = %#v", task)
	}
}

func TestCompleteAndDeleteTaskUseExpectedEndpoints(t *testing.T) {
	var requests []string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests = append(requests, r.Method+" "+r.URL.Path)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := NewClient("access-token")
	client.BaseURL = server.URL
	if err := client.CompleteTask(context.Background(), "p1", "t1"); err != nil {
		t.Fatalf("CompleteTask() error = %v", err)
	}
	if err := client.DeleteTask(context.Background(), "p1", "t1"); err != nil {
		t.Fatalf("DeleteTask() error = %v", err)
	}
	want := []string{"POST /project/p1/task/t1/complete", "DELETE /project/p1/task/t1"}
	for i := range want {
		if requests[i] != want[i] {
			t.Fatalf("request[%d] = %q, want %q", i, requests[i], want[i])
		}
	}
}
