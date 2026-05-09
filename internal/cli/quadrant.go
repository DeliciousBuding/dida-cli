package cli

import (
	"fmt"
	"io"
	"sort"
	"strings"
	"time"

	"github.com/DeliciousBuding/dida-cli/internal/model"
)

type quadrantBucket struct {
	ID          string       `json:"id"`
	Name        string       `json:"name"`
	Description string       `json:"description"`
	Tasks       []model.Task `json:"tasks"`
}

func runQuadrant(args []string, jsonOut bool, stdout io.Writer, stderr io.Writer) int {
	if len(args) == 0 || args[0] == "-h" || args[0] == "--help" {
		printQuadrantHelp(stdout)
		return 0
	}
	switch args[0] {
	case "list":
		return runQuadrantList("", jsonOut, stdout, stderr)
	case "view":
		if len(args) != 2 {
			return failTyped("quadrant view", "validation", "usage: dida quadrant view Q1|Q2|Q3|Q4", "run: dida quadrant --help", jsonOut, stdout, stderr)
		}
		return runQuadrantList(strings.ToUpper(args[1]), jsonOut, stdout, stderr)
	default:
		return fail("quadrant", fmt.Sprintf("unknown quadrant command %q", args[0]), jsonOut, stdout, stderr)
	}
}

func runQuadrantList(only string, jsonOut bool, stdout io.Writer, stderr io.Writer) int {
	if only != "" && !isQuadrantID(only) {
		return failTyped("quadrant view", "validation", "quadrant must be Q1, Q2, Q3, or Q4", "example: dida quadrant view Q2 --json", jsonOut, stdout, stderr)
	}
	view, err := loadSyncView()
	if err != nil {
		return failTyped("quadrant list", "auth", err.Error(), "run: dida auth login", jsonOut, stdout, stderr)
	}
	bucketMap := buildQuadrants(model.ActiveTasks(view.Tasks))
	buckets := orderedQuadrants(bucketMap)
	if only != "" {
		buckets = []quadrantBucket{bucketMap[only]}
	}
	total := 0
	for _, bucket := range buckets {
		total += len(bucket.Tasks)
	}
	data := map[string]any{"quadrants": buckets}
	meta := map[string]any{"count": total}
	command := "quadrant list"
	if only != "" {
		command = "quadrant view"
		meta["quadrant"] = only
	}
	if jsonOut {
		return writeJSON(stdout, envelope{OK: true, Command: command, Meta: meta, Data: data})
	}
	printQuadrants(stdout, buckets)
	return 0
}

func buildQuadrants(tasks []model.Task) map[string]quadrantBucket {
	buckets := map[string]quadrantBucket{
		"Q1": {ID: "Q1", Name: "Do First", Description: "important and urgent"},
		"Q2": {ID: "Q2", Name: "Schedule", Description: "important and not urgent"},
		"Q3": {ID: "Q3", Name: "Delegate", Description: "not important and urgent"},
		"Q4": {ID: "Q4", Name: "Eliminate", Description: "not important and not urgent"},
	}
	for _, task := range tasks {
		id := taskQuadrant(task)
		bucket := buckets[id]
		task.Raw = nil
		bucket.Tasks = append(bucket.Tasks, task)
		buckets[id] = bucket
	}
	for id, bucket := range buckets {
		sort.SliceStable(bucket.Tasks, func(i, j int) bool {
			a, b := bucket.Tasks[i], bucket.Tasks[j]
			if a.Overdue != b.Overdue {
				return a.Overdue
			}
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
		buckets[id] = bucket
	}
	return buckets
}

func orderedQuadrants(buckets map[string]quadrantBucket) []quadrantBucket {
	return []quadrantBucket{buckets["Q1"], buckets["Q2"], buckets["Q3"], buckets["Q4"]}
}

func taskQuadrant(task model.Task) string {
	if task.Priority == 5 {
		return "Q1"
	}
	if task.Priority == 3 && task.DueUnix != 0 {
		return "Q2"
	}
	if task.Priority == 1 && task.DueUnix != 0 {
		return "Q3"
	}
	return "Q4"
}

func isQuadrantID(value string) bool {
	switch strings.ToUpper(value) {
	case "Q1", "Q2", "Q3", "Q4":
		return true
	default:
		return false
	}
}

func printQuadrants(w io.Writer, buckets []quadrantBucket) {
	for _, bucket := range buckets {
		fmt.Fprintf(w, "\n%s %s - %d task(s)\n", bucket.ID, bucket.Name, len(bucket.Tasks))
		fmt.Fprintln(w, strings.Repeat("-", 48))
		if len(bucket.Tasks) == 0 {
			fmt.Fprintln(w, "No tasks found.")
			continue
		}
		for _, task := range bucket.Tasks {
			due := "-"
			if task.DueUnix != 0 {
				due = time.Unix(task.DueUnix, 0).Format("2006-01-02")
			}
			fmt.Fprintf(w, "%-28s  %-10s  %-8d  %s\n", task.ID, due, task.Priority, task.Title)
		}
	}
}

func printQuadrantHelp(w io.Writer) {
	fmt.Fprintln(w, strings.TrimSpace(`
Usage:
  dida quadrant list [--json]
  dida quadrant view Q1|Q2|Q3|Q4 [--json]

Quadrants:
  Q1  priority 5
  Q2  priority 3 with due date
  Q3  priority 1 with due date
  Q4  no priority, no due date, or lower-priority unscheduled work
`))
}
