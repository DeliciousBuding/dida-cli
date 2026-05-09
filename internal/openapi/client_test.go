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

func TestFocusEndpoints(t *testing.T) {
	var requests []string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests = append(requests, r.Method+" "+r.URL.RequestURI())
		switch r.Method {
		case http.MethodGet:
			if r.URL.Path == "/focus" {
				_ = json.NewEncoder(w).Encode([]map[string]any{{"id": "f1"}})
				return
			}
			_ = json.NewEncoder(w).Encode(map[string]any{"id": "f1"})
		case http.MethodDelete:
			_ = json.NewEncoder(w).Encode(map[string]any{"id": "f1", "type": 0})
		}
	}))
	defer server.Close()

	client := NewClient("access-token")
	client.BaseURL = server.URL
	if _, err := client.Focus(context.Background(), "f1", "0"); err != nil {
		t.Fatalf("Focus() error = %v", err)
	}
	if _, err := client.Focuses(context.Background(), "2026-04-01T00:00:00+0800", "2026-04-02T00:00:00+0800", "1"); err != nil {
		t.Fatalf("Focuses() error = %v", err)
	}
	if _, err := client.DeleteFocus(context.Background(), "f1", "0"); err != nil {
		t.Fatalf("DeleteFocus() error = %v", err)
	}
	want := []string{
		"GET /focus/f1?type=0",
		"GET /focus?from=2026-04-01T00%3A00%3A00%2B0800&to=2026-04-02T00%3A00%3A00%2B0800&type=1",
		"DELETE /focus/f1?type=0",
	}
	for i := range want {
		if requests[i] != want[i] {
			t.Fatalf("request[%d] = %q, want %q", i, requests[i], want[i])
		}
	}
}

func TestHabitEndpoints(t *testing.T) {
	var requests []string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests = append(requests, r.Method+" "+r.URL.RequestURI())
		if r.URL.Path == "/habit" && r.Method == http.MethodGet {
			_ = json.NewEncoder(w).Encode([]map[string]any{{"id": "h1"}})
			return
		}
		if r.URL.Path == "/habit/checkins" {
			_ = json.NewEncoder(w).Encode([]map[string]any{{"habitId": "h1"}})
			return
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"id": "h1"})
	}))
	defer server.Close()

	client := NewClient("access-token")
	client.BaseURL = server.URL
	if _, err := client.Habits(context.Background()); err != nil {
		t.Fatalf("Habits() error = %v", err)
	}
	if _, err := client.Habit(context.Background(), "h1"); err != nil {
		t.Fatalf("Habit() error = %v", err)
	}
	if _, err := client.CreateHabit(context.Background(), map[string]any{"name": "Read"}); err != nil {
		t.Fatalf("CreateHabit() error = %v", err)
	}
	if _, err := client.UpdateHabit(context.Background(), "h1", map[string]any{"name": "Read more"}); err != nil {
		t.Fatalf("UpdateHabit() error = %v", err)
	}
	if _, err := client.UpsertHabitCheckin(context.Background(), "h1", map[string]any{"stamp": 20260407}); err != nil {
		t.Fatalf("UpsertHabitCheckin() error = %v", err)
	}
	if _, err := client.HabitCheckins(context.Background(), "h1,h2", "20260401", "20260407"); err != nil {
		t.Fatalf("HabitCheckins() error = %v", err)
	}
	want := []string{
		"GET /habit",
		"GET /habit/h1",
		"POST /habit",
		"POST /habit/h1",
		"POST /habit/h1/checkin",
		"GET /habit/checkins?from=20260401&habitIds=h1%2Ch2&to=20260407",
	}
	for i := range want {
		if requests[i] != want[i] {
			t.Fatalf("request[%d] = %q, want %q", i, requests[i], want[i])
		}
	}
}
