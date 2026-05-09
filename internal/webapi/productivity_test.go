package webapi

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestProductivityReadsUseExpectedEndpoints(t *testing.T) {
	var seen []string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		seen = append(seen, r.Method+" "+r.URL.RequestURI())
		switch r.URL.Path {
		case "/user/preferences/pomodoro", "/user/preferences/habit":
			_ = json.NewEncoder(w).Encode(map[string]any{"enabled": true})
		default:
			_ = json.NewEncoder(w).Encode([]map[string]any{{"id": "x1", "name": "item"}})
		}
	}))
	defer server.Close()

	client := NewClient("test-token")
	client.BaseURL = server.URL
	ctx := context.Background()

	calls := []func() error{
		func() error { _, err := client.PomodoroPreferences(ctx); return err },
		func() error { _, err := client.Pomodoros(ctx, 1000, 2000); return err },
		func() error { _, err := client.PomodoroTimings(ctx, 1000, 2000); return err },
		func() error { _, err := client.TaskPomodoros(ctx, "p1", "t1"); return err },
		func() error { _, err := client.HabitPreferences(ctx); return err },
		func() error { _, err := client.Habits(ctx); return err },
		func() error { _, err := client.HabitSections(ctx); return err },
	}
	for _, call := range calls {
		if err := call(); err != nil {
			t.Fatalf("call failed: %v", err)
		}
	}

	want := []string{
		"GET /user/preferences/pomodoro",
		"GET /pomodoros?from=1000&to=2000",
		"GET /pomodoros/timing?from=1000&to=2000",
		"GET /pomodoros/task?projectId=p1&taskId=t1",
		"GET /user/preferences/habit?platform=web",
		"GET /habits",
		"GET /habitSections",
	}
	if strings.Join(seen, "\n") != strings.Join(want, "\n") {
		t.Fatalf("seen endpoints:\n%s\nwant:\n%s", strings.Join(seen, "\n"), strings.Join(want, "\n"))
	}
}
