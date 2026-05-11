package model

import (
	"encoding/json"
	"testing"
	"time"
)

func TestBuildSyncViewAndTodayTasks(t *testing.T) {
	now := time.Date(2026, 5, 9, 12, 0, 0, 0, time.FixedZone("CST", 8*3600))
	view := BuildSyncView(
		"inbox",
		[]map[string]any{{"id": "p1", "name": "Work"}},
		[]map[string]any{
			{"id": "t1", "projectId": "p1", "title": "Today", "dueDate": "2026-05-09T09:00:00.000+0800", "priority": float64(5), "status": float64(0)},
			{"id": "t2", "projectId": "p1", "title": "Tomorrow", "dueDate": "2026-05-10T09:00:00.000+0800", "priority": float64(3), "status": float64(0)},
			{"id": "t3", "projectId": "p1", "title": "Done", "dueDate": "2026-05-09T10:00:00.000+0800", "status": float64(2)},
		},
		nil,
		nil,
		[]map[string]any{{"id": "f1", "name": "Focus"}},
		now,
	)
	if view.Counts["tasks"] != 3 {
		t.Fatalf("task count = %d, want 3", view.Counts["tasks"])
	}
	if view.Counts["filters"] != 1 || len(view.Filters) != 1 {
		t.Fatalf("filter count = %d len=%d, want 1", view.Counts["filters"], len(view.Filters))
	}
	today := TodayTasks(view.Tasks, now)
	if len(today) != 1 {
		t.Fatalf("today len = %d, want 1: %#v", len(today), today)
	}
	if today[0].ProjectName != "Work" {
		t.Fatalf("project name = %q, want Work", today[0].ProjectName)
	}
	if !today[0].Overdue {
		t.Fatalf("today task should be overdue at noon")
	}
}

func TestSearchAndUpcomingTasks(t *testing.T) {
	now := time.Date(2026, 5, 9, 12, 0, 0, 0, time.FixedZone("CST", 8*3600))
	view := BuildSyncView(
		"inbox",
		[]map[string]any{{"id": "p1", "name": "School"}},
		[]map[string]any{
			{"id": "t1", "projectId": "p1", "title": "Computer Systems Lab", "dueDate": "2026-05-10T09:00:00+08:00", "status": 0},
			{"id": "t2", "projectId": "p1", "title": "Far Future", "dueDate": "2026-06-10T09:00:00+08:00", "status": 0},
			{"id": "t3", "projectId": "p1", "title": "Done Computer", "dueDate": "2026-05-10T09:00:00+08:00", "status": 2},
		},
		nil,
		nil,
		nil,
		now,
	)
	found := SearchTasks(ActiveTasks(view.Tasks), "computer")
	if len(found) != 1 || found[0].ID != "t1" {
		t.Fatalf("search result = %#v, want active t1 only", found)
	}
	upcoming := UpcomingTasks(view.Tasks, now, 7)
	if len(upcoming) != 1 || upcoming[0].ID != "t1" {
		t.Fatalf("upcoming result = %#v, want t1 only", upcoming)
	}
}

func TestFindTask(t *testing.T) {
	tasks := []Task{
		{ID: "t1", Title: "First"},
		{ID: "t2", Title: "Second"},
	}
	task, found := FindTask(tasks, "t2")
	if !found || task.ID != "t2" || task.Title != "Second" {
		t.Fatalf("FindTask(t2) = %#v, found=%v", task, found)
	}
	_, found = FindTask(tasks, "missing")
	if found {
		t.Fatalf("FindTask(missing) found = true, want false")
	}
	_, found = FindTask(nil, "t1")
	if found {
		t.Fatalf("FindTask(nil) found = true, want false")
	}
}

func TestProjectTasks(t *testing.T) {
	tasks := []Task{
		{ID: "t1", ProjectID: "p1"},
		{ID: "t2", ProjectID: "p2"},
		{ID: "t3", ProjectID: "p1"},
	}
	result := ProjectTasks(tasks, "p1")
	if len(result) != 2 || result[0].ID != "t1" || result[1].ID != "t3" {
		t.Fatalf("ProjectTasks(p1) = %#v", result)
	}
	result = ProjectTasks(tasks, "missing")
	if len(result) != 0 {
		t.Fatalf("ProjectTasks(missing) = %#v", result)
	}
}

func TestInferColumns(t *testing.T) {
	tasks := []Task{
		{ID: "t1", ProjectID: "p1", ColumnID: "c1"},
		{ID: "t2", ProjectID: "p1", ColumnID: "c1"},
		{ID: "t3", ProjectID: "p1", ColumnID: "c2"},
		{ID: "t4", ProjectID: "p2", ColumnID: "c1"},    // different project
		{ID: "t5", ProjectID: "p1", ColumnID: ""},       // no column
	}
	columns := InferColumns("p1", tasks)
	if len(columns) != 2 {
		t.Fatalf("InferColumns() len = %d, want 2", len(columns))
	}
	// sorted by ID
	if columns[0].ID != "c1" || columns[0].TaskCount != 2 {
		t.Fatalf("column[0] = %+v", columns[0])
	}
	if columns[1].ID != "c2" || columns[1].TaskCount != 1 {
		t.Fatalf("column[1] = %+v", columns[1])
	}
}

func TestInferColumnsEmpty(t *testing.T) {
	columns := InferColumns("p1", nil)
	if len(columns) != 0 {
		t.Fatalf("InferColumns(nil) = %#v", columns)
	}
}

func TestIntish(t *testing.T) {
	cases := []struct {
		input any
		want  int
	}{
		{42, 42},
		{int64(100), 100},
		{float64(3.14), 3},
		{json.Number("7"), 7},
		{"13", 13},
		{"abc", 0},
		{nil, 0},
		{true, 0},
	}
	for _, tc := range cases {
		got := intish(tc.input)
		if got != tc.want {
			t.Errorf("intish(%#v) = %d, want %d", tc.input, got, tc.want)
		}
	}
}

func TestBoolish(t *testing.T) {
	cases := []struct {
		input any
		want  bool
	}{
		{true, true},
		{false, false},
		{"true", true},
		{"1", true},
		{"false", false},
		{"0", false},
		{float64(1), true},
		{float64(0), false},
		{nil, false},
		{42, false}, // non-float64 int
	}
	for _, tc := range cases {
		got := boolish(tc.input)
		if got != tc.want {
			t.Errorf("boolish(%#v) = %v, want %v", tc.input, got, tc.want)
		}
	}
}

func TestStringSlice(t *testing.T) {
	got := stringSlice([]any{"a", " b ", nil, "c"})
	if len(got) != 3 || got[0] != "a" || got[1] != "b" || got[2] != "c" {
		t.Fatalf("stringSlice = %#v", got)
	}
	if got := stringSlice("not-array"); got != nil {
		t.Fatalf("stringSlice(string) = %#v, want nil", got)
	}
	if got := stringSlice(nil); got != nil {
		t.Fatalf("stringSlice(nil) = %#v, want nil", got)
	}
}

func TestObjectSlice(t *testing.T) {
	input := []any{
		map[string]any{"id": "1"},
		"string",
		map[string]any{"id": "2"},
		42,
	}
	got := objectSlice(input)
	if len(got) != 2 || got[0]["id"] != "1" || got[1]["id"] != "2" {
		t.Fatalf("objectSlice = %#v", got)
	}
	if got := objectSlice("not-array"); got != nil {
		t.Fatalf("objectSlice(string) = %#v", got)
	}
}

func TestAnySlice(t *testing.T) {
	input := []any{"a", 1, nil}
	got := anySlice(input)
	if len(got) != 3 {
		t.Fatalf("anySlice = %#v", got)
	}
	if got := anySlice("not"); got != nil {
		t.Fatalf("anySlice(string) = %#v", got)
	}
}

func TestFirstPresent(t *testing.T) {
	item := map[string]any{"a": 1, "b": nil, "c": "hello"}
	if got := firstPresent(item, "b", "c", "a"); got != "hello" {
		t.Fatalf("firstPresent(b,c,a) = %#v, want hello", got)
	}
	if got := firstPresent(item, "missing", "also-missing"); got != nil {
		t.Fatalf("firstPresent(missing) = %#v, want nil", got)
	}
}

func TestFirstString(t *testing.T) {
	item := map[string]any{"a": "", "b": "hello", "c": 42}
	if got := firstString(item, "a", "b", "c"); got != "hello" {
		t.Fatalf("firstString = %q, want hello", got)
	}
	if got := firstString(item, "missing"); got != "" {
		t.Fatalf("firstString(missing) = %q, want empty", got)
	}
}

func TestStr(t *testing.T) {
	if got := str("hello"); got != "hello" {
		t.Fatalf("str(string) = %q", got)
	}
	if got := str(nil); got != "" {
		t.Fatalf("str(nil) = %q, want empty", got)
	}
	if got := str(42); got != "42" {
		t.Fatalf("str(42) = %q", got)
	}
}

func TestParseDidaTimeEdgeCases(t *testing.T) {
	// empty
	_, ok := ParseDidaTime("")
	if ok {
		t.Fatalf("ParseDidaTime(\"\") ok = true")
	}

	// whitespace only
	_, ok = ParseDidaTime("  ")
	if ok {
		t.Fatalf("ParseDidaTime(\"  \") ok = true")
	}

	// date only
	tm, ok := ParseDidaTime("2026-05-09")
	if !ok || tm.Year() != 2026 || tm.Month() != 5 || tm.Day() != 9 {
		t.Fatalf("ParseDidaTime(date only) = %v, ok=%v", tm, ok)
	}

	// RFC3339
	tm, ok = ParseDidaTime("2026-05-09T12:00:00+08:00")
	if !ok || tm.Hour() != 12 {
		t.Fatalf("ParseDidaTime(RFC3339) = %v, ok=%v", tm, ok)
	}

	// dida format
	tm, ok = ParseDidaTime("2026-05-09T09:00:00.000+0800")
	if !ok || tm.Hour() != 9 {
		t.Fatalf("ParseDidaTime(dida) = %v, ok=%v", tm, ok)
	}

	// garbage
	_, ok = ParseDidaTime("not-a-date")
	if ok {
		t.Fatalf("ParseDidaTime(garbage) ok = true")
	}
}
