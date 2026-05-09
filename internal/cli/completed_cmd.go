package cli

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/DeliciousBuding/dida-cli/internal/model"
	"github.com/DeliciousBuding/dida-cli/internal/webapi"
)

func runCompleted(args []string, jsonOut bool, stdout io.Writer, stderr io.Writer) int {
	if len(args) == 0 || args[0] == "-h" || args[0] == "--help" {
		printCompletedHelp(stdout)
		return 0
	}
	now := time.Now()
	limit := 100
	compact := false
	var from, to time.Time
	command := "completed " + args[0]
	switch args[0] {
	case "today":
		from, to = dayRange(now)
	case "yesterday":
		from, to = dayRange(now.AddDate(0, 0, -1))
	case "week":
		start := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location()).AddDate(0, 0, -int(now.Weekday()))
		from = start
		to = start.AddDate(0, 0, 7).Add(-time.Second)
	case "list":
		parsedFrom, parsedTo, parsedLimit, parsedCompact, err := parseCompletedListFlags(args[1:], now)
		if err != nil {
			return failTyped("completed list", "validation", err.Error(), "run: dida completed --help", jsonOut, stdout, stderr)
		}
		from, to, limit, compact = parsedFrom, parsedTo, parsedLimit, parsedCompact
	default:
		return fail("completed", fmt.Sprintf("unknown completed command %q", args[0]), jsonOut, stdout, stderr)
	}
	if args[0] != "list" {
		parsedCompact, err := parseCompactOnlyFlags(args[1:])
		if err != nil {
			return failTyped(command, "validation", err.Error(), "run: dida completed --help", jsonOut, stdout, stderr)
		}
		compact = parsedCompact
	}
	tasks, err := loadCompletedTasks(from, to, limit)
	if err != nil {
		return failTyped(command, "auth", err.Error(), "run: dida auth login", jsonOut, stdout, stderr)
	}
	view, _ := loadSyncView()
	projectNames := map[string]string{}
	for _, project := range view.Projects {
		projectNames[project.ID] = project.Name
	}
	normalized := model.NormalizeTasks(tasks, projectNames, now)
	data := map[string]any{
		"from":    formatDidaQueryTime(from),
		"to":      formatDidaQueryTime(to),
		"compact": compact,
		"tasks":   taskOutput(normalized, compact),
	}
	meta := map[string]any{"count": len(normalized), "limit": limit}
	if jsonOut {
		return writeJSON(stdout, envelope{OK: true, Command: command, Meta: meta, Data: data})
	}
	printTasks(stdout, normalized, len(normalized))
	return 0
}

func parseCompletedListFlags(args []string, now time.Time) (time.Time, time.Time, int, bool, error) {
	from, to := dayRange(now)
	limit := 100
	compact := false
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--from":
			if i+1 >= len(args) {
				return time.Time{}, time.Time{}, 0, false, fmt.Errorf("--from requires YYYY-MM-DD")
			}
			parsed, err := time.ParseInLocation("2006-01-02", args[i+1], now.Location())
			if err != nil {
				return time.Time{}, time.Time{}, 0, false, fmt.Errorf("--from must be YYYY-MM-DD")
			}
			from = parsed
			i++
		case "--to":
			if i+1 >= len(args) {
				return time.Time{}, time.Time{}, 0, false, fmt.Errorf("--to requires YYYY-MM-DD")
			}
			parsed, err := time.ParseInLocation("2006-01-02", args[i+1], now.Location())
			if err != nil {
				return time.Time{}, time.Time{}, 0, false, fmt.Errorf("--to must be YYYY-MM-DD")
			}
			to = parsed.AddDate(0, 0, 1).Add(-time.Second)
			i++
		case "--limit":
			if i+1 >= len(args) {
				return time.Time{}, time.Time{}, 0, false, fmt.Errorf("--limit requires a value")
			}
			if _, err := fmt.Sscanf(args[i+1], "%d", &limit); err != nil || limit <= 0 {
				return time.Time{}, time.Time{}, 0, false, fmt.Errorf("--limit must be a positive integer")
			}
			i++
		case "--compact", "--brief":
			compact = true
		default:
			return time.Time{}, time.Time{}, 0, false, fmt.Errorf("unknown flag %q", args[i])
		}
	}
	if from.After(to) {
		return time.Time{}, time.Time{}, 0, false, fmt.Errorf("--from must be before or equal to --to")
	}
	return from, to, limit, compact, nil
}

func parseCompactOnlyFlags(args []string) (bool, error) {
	compact := false
	for _, arg := range args {
		switch arg {
		case "--compact", "--brief":
			compact = true
		default:
			return false, fmt.Errorf("unknown flag %q", arg)
		}
	}
	return compact, nil
}

func loadCompletedTasks(from time.Time, to time.Time, limit int) ([]map[string]any, error) {
	result, err := executeRead(func(ctx context.Context, client *webapi.Client) (any, error) {
		return client.CompletedTasks(ctx, formatDidaQueryTime(from), formatDidaQueryTime(to), limit)
	})
	if err != nil {
		return nil, err
	}
	return result.([]map[string]any), nil
}

func dayRange(t time.Time) (time.Time, time.Time) {
	start := time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
	return start, start.AddDate(0, 0, 1).Add(-time.Second)
}

func formatDidaQueryTime(t time.Time) string {
	return t.Format("2006-01-02 15:04:05")
}
