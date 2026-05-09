package cli

import (
	"context"
	"fmt"
	"io"

	"github.com/DeliciousBuding/dida-cli/internal/webapi"
)

type trashListOptions struct {
	Cursor  int
	Limit   int
	Compact bool
}

type compactTrashTask struct {
	ID          string `json:"id"`
	ProjectID   string `json:"projectId,omitempty"`
	Title       string `json:"title"`
	Kind        string `json:"kind,omitempty"`
	Status      int    `json:"status"`
	Priority    int    `json:"priority"`
	DeletedTime int64  `json:"deletedTime,omitempty"`
}

func runTrash(args []string, jsonOut bool, stdout io.Writer, stderr io.Writer) int {
	if len(args) == 0 || args[0] == "-h" || args[0] == "--help" {
		printTrashHelp(stdout)
		return 0
	}
	if args[0] != "list" {
		return fail("trash", fmt.Sprintf("unknown trash command %q", args[0]), jsonOut, stdout, stderr)
	}
	opts, err := parseTrashListFlags(args[1:])
	if err != nil {
		return failTyped("trash list", "validation", err.Error(), "run: dida trash --help", jsonOut, stdout, stderr)
	}
	result, err := executeRead(func(ctx context.Context, client *webapi.Client) (any, error) {
		return client.TrashPage(ctx, opts.Cursor)
	})
	if err != nil {
		return failTyped("trash list", "api", err.Error(), "", jsonOut, stdout, stderr)
	}
	page := result.(map[string]any)
	tasks := mapSlice(page["tasks"])
	total := len(tasks)
	tasks = limitMapItems(tasks, opts.Limit)
	data := map[string]any{
		"cursor":  opts.Cursor,
		"next":    page["next"],
		"compact": opts.Compact,
		"tasks":   trashTaskOutput(tasks, opts.Compact),
	}
	meta := map[string]any{"count": len(tasks), "total": total, "limit": opts.Limit, "next": page["next"]}
	if jsonOut {
		return writeJSON(stdout, envelope{OK: true, Command: "trash list", Meta: meta, Data: data})
	}
	printMapList(stdout, tasks, "trash tasks")
	return 0
}

func parseTrashListFlags(args []string) (trashListOptions, error) {
	opts := trashListOptions{Limit: 20, Compact: true}
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--cursor", "--from":
			if i+1 >= len(args) {
				return opts, fmt.Errorf("%s requires a non-negative integer", args[i])
			}
			if _, err := fmt.Sscanf(args[i+1], "%d", &opts.Cursor); err != nil || opts.Cursor < 0 {
				return opts, fmt.Errorf("%s must be a non-negative integer", args[i])
			}
			i++
		case "--limit":
			if i+1 >= len(args) {
				return opts, fmt.Errorf("--limit requires a positive integer")
			}
			if _, err := fmt.Sscanf(args[i+1], "%d", &opts.Limit); err != nil || opts.Limit <= 0 {
				return opts, fmt.Errorf("--limit must be a positive integer")
			}
			i++
		case "--compact", "--brief":
			opts.Compact = true
		case "--full":
			opts.Compact = false
		default:
			return opts, fmt.Errorf("unknown flag %q", args[i])
		}
	}
	return opts, nil
}

func trashTaskOutput(tasks []map[string]any, compact bool) any {
	if !compact {
		return tasks
	}
	out := make([]compactTrashTask, 0, len(tasks))
	for _, task := range tasks {
		out = append(out, compactTrashTask{
			ID:          stringField(task, "id"),
			ProjectID:   stringField(task, "projectId"),
			Title:       stringField(task, "title"),
			Kind:        stringField(task, "kind"),
			Status:      intField(task, "status"),
			Priority:    intField(task, "priority"),
			DeletedTime: int64Field(task, "deletedTime"),
		})
	}
	return out
}

func mapSlice(value any) []map[string]any {
	switch typed := value.(type) {
	case []map[string]any:
		return typed
	case []any:
		out := make([]map[string]any, 0, len(typed))
		for _, item := range typed {
			if mapped, ok := item.(map[string]any); ok {
				out = append(out, mapped)
			}
		}
		return out
	default:
		return nil
	}
}

func stringField(values map[string]any, key string) string {
	if value, ok := values[key].(string); ok {
		return value
	}
	return ""
}

func intField(values map[string]any, key string) int {
	switch value := values[key].(type) {
	case int:
		return value
	case int64:
		return int(value)
	case float64:
		return int(value)
	default:
		return 0
	}
}

func int64Field(values map[string]any, key string) int64 {
	switch value := values[key].(type) {
	case int64:
		return value
	case int:
		return int64(value)
	case float64:
		return int64(value)
	default:
		return 0
	}
}
