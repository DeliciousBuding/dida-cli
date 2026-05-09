package model

import (
	"encoding/json"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"
)

type Project struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	Closed bool   `json:"closed,omitempty"`
	Color  string `json:"color,omitempty"`
	Raw    any    `json:"raw,omitempty"`
}

type Task struct {
	ID            string `json:"id"`
	ProjectID     string `json:"projectId"`
	ProjectName   string `json:"projectName,omitempty"`
	Title         string `json:"title"`
	Content       string `json:"content,omitempty"`
	DueDate       string `json:"dueDate,omitempty"`
	DueUnix       int64  `json:"dueUnix,omitempty"`
	StartDate     string `json:"startDate,omitempty"`
	StartUnix     int64  `json:"startUnix,omitempty"`
	CompletedAt   string `json:"completedTime,omitempty"`
	CompletedUnix int64  `json:"completedUnix,omitempty"`
	Priority      int    `json:"priority"`
	Status        int    `json:"status"`
	Deleted       int    `json:"deleted"`
	ColumnID      string `json:"columnId,omitempty"`
	Overdue       bool   `json:"overdue,omitempty"`
	Raw           any    `json:"raw,omitempty"`
}

type Column struct {
	ID        string `json:"id"`
	ProjectID string `json:"projectId,omitempty"`
	Name      string `json:"name,omitempty"`
	TaskCount int    `json:"taskCount"`
}

type SyncView struct {
	InboxID       string           `json:"inboxId,omitempty"`
	Projects      []Project        `json:"projects"`
	Tasks         []Task           `json:"tasks"`
	ProjectGroups []map[string]any `json:"projectGroups,omitempty"`
	Tags          []map[string]any `json:"tags,omitempty"`
	Counts        map[string]int   `json:"counts"`
}

func BuildSyncView(inboxID string, rawProjects []map[string]any, rawTasks []map[string]any, rawProjectGroups []map[string]any, rawTags []map[string]any, now time.Time) SyncView {
	projects := NormalizeProjects(rawProjects)
	projectNames := make(map[string]string, len(projects))
	for _, project := range projects {
		projectNames[project.ID] = project.Name
	}
	tasks := NormalizeTasks(rawTasks, projectNames, now)
	return SyncView{
		InboxID:       inboxID,
		Projects:      projects,
		Tasks:         tasks,
		ProjectGroups: rawProjectGroups,
		Tags:          rawTags,
		Counts: map[string]int{
			"projects":      len(projects),
			"tasks":         len(tasks),
			"projectGroups": len(rawProjectGroups),
			"tags":          len(rawTags),
		},
	}
}

func NormalizeProjects(items []map[string]any) []Project {
	out := make([]Project, 0, len(items))
	for _, item := range items {
		id := str(item["id"])
		if id == "" {
			continue
		}
		name := firstString(item, "name", "title", "projectName")
		if name == "" {
			name = id
		}
		out = append(out, Project{
			ID:     id,
			Name:   name,
			Closed: boolish(item["closed"]),
			Color:  str(item["color"]),
			Raw:    item,
		})
	}
	sort.SliceStable(out, func(i, j int) bool {
		return strings.ToLower(out[i].Name) < strings.ToLower(out[j].Name)
	})
	return out
}

func NormalizeTasks(items []map[string]any, projectNames map[string]string, now time.Time) []Task {
	out := make([]Task, 0, len(items))
	for _, item := range items {
		id := str(item["id"])
		title := strings.TrimSpace(str(item["title"]))
		if id == "" || title == "" {
			continue
		}
		projectID := str(item["projectId"])
		due := str(item["dueDate"])
		dueTime, dueOK := ParseDidaTime(due)
		start := str(item["startDate"])
		startTime, startOK := ParseDidaTime(start)
		completed := str(item["completedTime"])
		completedTime, completedOK := ParseDidaTime(completed)
		task := Task{
			ID:          id,
			ProjectID:   projectID,
			ProjectName: projectNames[projectID],
			Title:       title,
			Content:     str(item["content"]),
			DueDate:     due,
			StartDate:   start,
			CompletedAt: completed,
			Priority:    intish(item["priority"]),
			Status:      intish(item["status"]),
			Deleted:     intish(item["deleted"]),
			ColumnID:    str(item["columnId"]),
			Raw:         item,
		}
		if dueOK {
			task.DueUnix = dueTime.Unix()
			task.Overdue = dueTime.Before(now)
		}
		if startOK {
			task.StartUnix = startTime.Unix()
		}
		if completedOK {
			task.CompletedUnix = completedTime.Unix()
		}
		out = append(out, task)
	}
	sort.SliceStable(out, func(i, j int) bool {
		a, b := out[i], out[j]
		if a.DueUnix == 0 && b.DueUnix != 0 {
			return false
		}
		if a.DueUnix != 0 && b.DueUnix == 0 {
			return true
		}
		if a.DueUnix != b.DueUnix {
			return a.DueUnix < b.DueUnix
		}
		if a.Priority != b.Priority {
			return a.Priority > b.Priority
		}
		return strings.ToLower(a.Title) < strings.ToLower(b.Title)
	})
	return out
}

func SearchTasks(tasks []Task, query string) []Task {
	query = strings.ToLower(strings.TrimSpace(query))
	if query == "" {
		return nil
	}
	out := make([]Task, 0)
	for _, task := range tasks {
		if strings.Contains(strings.ToLower(task.Title), query) ||
			strings.Contains(strings.ToLower(task.Content), query) ||
			strings.Contains(strings.ToLower(task.ProjectName), query) {
			out = append(out, task)
		}
	}
	return out
}

func UpcomingTasks(tasks []Task, now time.Time, days int) []Task {
	if days <= 0 {
		days = 7
	}
	end := time.Date(now.Year(), now.Month(), now.Day(), 23, 59, 59, int(time.Second-time.Nanosecond), now.Location()).AddDate(0, 0, days)
	out := make([]Task, 0)
	for _, task := range ActiveTasks(tasks) {
		if task.DueUnix == 0 {
			continue
		}
		due := time.Unix(task.DueUnix, 0).In(now.Location())
		if !due.Before(now) && !due.After(end) {
			out = append(out, task)
		}
	}
	return out
}

func ActiveTasks(tasks []Task) []Task {
	out := make([]Task, 0, len(tasks))
	for _, task := range tasks {
		if task.Deleted != 0 {
			continue
		}
		if task.Status != 0 {
			continue
		}
		out = append(out, task)
	}
	return out
}

func TodayTasks(tasks []Task, now time.Time) []Task {
	end := time.Date(now.Year(), now.Month(), now.Day(), 23, 59, 59, int(time.Second-time.Nanosecond), now.Location())
	out := make([]Task, 0, len(tasks))
	for _, task := range ActiveTasks(tasks) {
		if task.DueUnix == 0 {
			continue
		}
		due := time.Unix(task.DueUnix, 0).In(now.Location())
		if !due.After(end) {
			out = append(out, task)
		}
	}
	return out
}

func FindTask(tasks []Task, id string) (Task, bool) {
	for _, task := range tasks {
		if task.ID == id {
			return task, true
		}
	}
	return Task{}, false
}

func ProjectTasks(tasks []Task, projectID string) []Task {
	out := make([]Task, 0)
	for _, task := range tasks {
		if task.ProjectID == projectID {
			out = append(out, task)
		}
	}
	return out
}

func InferColumns(projectID string, tasks []Task) []Column {
	columns := make(map[string]*Column)
	for _, task := range tasks {
		if task.ProjectID != projectID || task.ColumnID == "" {
			continue
		}
		column := columns[task.ColumnID]
		if column == nil {
			column = &Column{ID: task.ColumnID, ProjectID: projectID}
			columns[task.ColumnID] = column
		}
		column.TaskCount++
	}
	out := make([]Column, 0, len(columns))
	for _, column := range columns {
		out = append(out, *column)
	}
	sort.SliceStable(out, func(i, j int) bool {
		return out[i].ID < out[j].ID
	})
	return out
}

func ParseDidaTime(value string) (time.Time, bool) {
	value = strings.TrimSpace(value)
	if value == "" {
		return time.Time{}, false
	}
	layouts := []string{
		"2006-01-02T15:04:05.000-0700",
		time.RFC3339Nano,
		time.RFC3339,
		"2006-01-02",
	}
	for _, layout := range layouts {
		if t, err := time.Parse(layout, value); err == nil {
			return t, true
		}
	}
	return time.Time{}, false
}

func firstString(item map[string]any, keys ...string) string {
	for _, key := range keys {
		if value := str(item[key]); value != "" {
			return value
		}
	}
	return ""
}

func str(value any) string {
	switch typed := value.(type) {
	case string:
		return typed
	case fmt.Stringer:
		return typed.String()
	case nil:
		return ""
	default:
		return fmt.Sprint(typed)
	}
}

func intish(value any) int {
	switch typed := value.(type) {
	case int:
		return typed
	case int64:
		return int(typed)
	case float64:
		return int(typed)
	case json.Number:
		n, _ := strconv.Atoi(typed.String())
		return n
	case string:
		n, _ := strconv.Atoi(typed)
		return n
	default:
		return 0
	}
}

func boolish(value any) bool {
	switch typed := value.(type) {
	case bool:
		return typed
	case string:
		return typed == "true" || typed == "1"
	case float64:
		return typed != 0
	default:
		return false
	}
}
