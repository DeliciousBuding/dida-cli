package openapi

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
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

func TestCreateUpdateDeleteProject(t *testing.T) {
	var requests []string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests = append(requests, r.Method+" "+r.URL.Path)
		switch r.Method {
		case http.MethodPost:
			if r.URL.Path == "/project" {
				_ = json.NewEncoder(w).Encode(map[string]any{"id": "p-new"})
				return
			}
			_ = json.NewEncoder(w).Encode(map[string]any{"id": "p1"})
		case http.MethodDelete:
			w.WriteHeader(http.StatusOK)
		}
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL

	proj, err := client.CreateProject(context.Background(), map[string]any{"name": "New"})
	if err != nil {
		t.Fatalf("CreateProject() error = %v", err)
	}
	if proj["id"] != "p-new" {
		t.Fatalf("CreateProject() = %#v", proj)
	}

	proj, err = client.UpdateProject(context.Background(), "p1", map[string]any{"name": "Renamed"})
	if err != nil {
		t.Fatalf("UpdateProject() error = %v", err)
	}
	if proj["id"] != "p1" {
		t.Fatalf("UpdateProject() = %#v", proj)
	}

	if err := client.DeleteProject(context.Background(), "p1"); err != nil {
		t.Fatalf("DeleteProject() error = %v", err)
	}

	want := []string{"POST /project", "POST /project/p1", "DELETE /project/p1"}
	for i := range want {
		if requests[i] != want[i] {
			t.Fatalf("request[%d] = %q, want %q", i, requests[i], want[i])
		}
	}
}

func TestGetTask(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.Method + " " + r.URL.Path; got != "GET /project/p1/task/t1" {
			t.Fatalf("request = %s, want GET /project/p1/task/t1", got)
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"id": "t1", "title": "Task"})
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	task, err := client.Task(context.Background(), "p1", "t1")
	if err != nil {
		t.Fatalf("Task() error = %v", err)
	}
	if task["id"] != "t1" || task["title"] != "Task" {
		t.Fatalf("task = %#v", task)
	}
}

func TestUpdateTask(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.Method + " " + r.URL.Path; got != "POST /task/t1" {
			t.Fatalf("request = %s, want POST /task/t1", got)
		}
		body, _ := io.ReadAll(r.Body)
		var payload map[string]any
		_ = json.Unmarshal(body, &payload)
		if payload["title"] != "Updated" {
			t.Fatalf("body title = %v", payload["title"])
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"id": "t1"})
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	task, err := client.UpdateTask(context.Background(), "t1", map[string]any{"title": "Updated"})
	if err != nil {
		t.Fatalf("UpdateTask() error = %v", err)
	}
	if task["id"] != "t1" {
		t.Fatalf("task = %#v", task)
	}
}

func TestMoveTasks(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.Method + " " + r.URL.Path; got != "POST /task/move" {
			t.Fatalf("request = %s, want POST /task/move", got)
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"success": true})
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	result, err := client.MoveTasks(context.Background(), map[string]any{"taskIds": []string{"t1"}, "toProject": "p2"})
	if err != nil {
		t.Fatalf("MoveTasks() error = %v", err)
	}
	resultMap, _ := result.(map[string]any)
	if resultMap["success"] != true {
		t.Fatalf("result = %#v", result)
	}
}

func TestCompletedTasks(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.Method + " " + r.URL.Path; got != "POST /task/completed" {
			t.Fatalf("request = %s, want POST /task/completed", got)
		}
		_ = json.NewEncoder(w).Encode([]map[string]any{{"id": "t-done"}})
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	tasks, err := client.CompletedTasks(context.Background(), map[string]any{"projectId": "p1"})
	if err != nil {
		t.Fatalf("CompletedTasks() error = %v", err)
	}
	if len(tasks) != 1 || tasks[0]["id"] != "t-done" {
		t.Fatalf("tasks = %#v", tasks)
	}
}

func TestFilterTasks(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.Method + " " + r.URL.Path; got != "POST /task/filter" {
			t.Fatalf("request = %s, want POST /task/filter", got)
		}
		_ = json.NewEncoder(w).Encode([]map[string]any{{"id": "t-filtered"}})
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	tasks, err := client.FilterTasks(context.Background(), map[string]any{"query": "today"})
	if err != nil {
		t.Fatalf("FilterTasks() error = %v", err)
	}
	if len(tasks) != 1 || tasks[0]["id"] != "t-filtered" {
		t.Fatalf("tasks = %#v", tasks)
	}
}

func TestDoRequiresToken(t *testing.T) {
	client := NewClient("")
	err := client.Do(context.Background(), "GET", "/test", nil, nil)
	if err == nil {
		t.Fatalf("Do() with empty token: error = nil")
	}
	if !strings.Contains(err.Error(), "missing openapi access token") {
		t.Fatalf("error = %v", err)
	}
}

func TestDoHandlesHTTPError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"error":"not found"}`))
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	err := client.Do(context.Background(), "GET", "/missing", nil, nil)
	if err == nil {
		t.Fatalf("Do() 404: error = nil")
	}
	if !strings.Contains(err.Error(), "404") {
		t.Fatalf("error = %v, want 404", err)
	}
}

func TestDoHandlesNilOutput(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	if err := client.Do(context.Background(), "GET", "/test", nil, nil); err != nil {
		t.Fatalf("Do() nil output: error = %v", err)
	}
}
