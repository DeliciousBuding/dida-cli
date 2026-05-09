package model

import (
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
		now,
	)
	if view.Counts["tasks"] != 3 {
		t.Fatalf("task count = %d, want 3", view.Counts["tasks"])
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
