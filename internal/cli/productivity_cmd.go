package cli

import (
	"context"
	"fmt"
	"io"
	"strconv"
	"time"

	"github.com/DeliciousBuding/dida-cli/internal/webapi"
)

type rangeListOptions struct {
	From  time.Time
	To    time.Time
	Limit int
}

func runPomo(args []string, jsonOut bool, stdout io.Writer, stderr io.Writer) int {
	if len(args) == 0 || args[0] == "-h" || args[0] == "--help" {
		printPomoHelp(stdout)
		return 0
	}
	switch args[0] {
	case "preferences", "prefs":
		return runPomoPreferences(jsonOut, stdout, stderr)
	case "list":
		return runPomoList(args[1:], false, jsonOut, stdout, stderr)
	case "timing":
		return runPomoList(args[1:], true, jsonOut, stdout, stderr)
	case "task":
		return runPomoTask(args[1:], jsonOut, stdout, stderr)
	default:
		return fail("pomo", fmt.Sprintf("unknown pomo command %q", args[0]), jsonOut, stdout, stderr)
	}
}

func runPomoPreferences(jsonOut bool, stdout io.Writer, stderr io.Writer) int {
	result, err := executeRead(func(ctx context.Context, client *webapi.Client) (any, error) {
		return client.PomodoroPreferences(ctx)
	})
	if err != nil {
		return failTyped("pomo preferences", "api", err.Error(), "", jsonOut, stdout, stderr)
	}
	data := map[string]any{"preferences": result}
	if jsonOut {
		return writeJSON(stdout, envelope{OK: true, Command: "pomo preferences", Data: data})
	}
	fmt.Fprintf(stdout, "Pomodoro preferences: %d keys\n", len(result.(map[string]any)))
	return 0
}

func runPomoTask(args []string, jsonOut bool, stdout io.Writer, stderr io.Writer) int {
	opts, err := parseCommentTargetFlags(args, "pomo task", false)
	if err != nil {
		return failTyped("pomo task", "validation", err.Error(), "run: dida pomo --help", jsonOut, stdout, stderr)
	}
	result, err := executeRead(func(ctx context.Context, client *webapi.Client) (any, error) {
		return client.TaskPomodoros(ctx, opts.ProjectID, opts.TaskID)
	})
	if err != nil {
		return failTyped("pomo task", "api", err.Error(), "", jsonOut, stdout, stderr)
	}
	items := result.([]map[string]any)
	meta := map[string]any{"count": len(items)}
	data := map[string]any{"projectId": opts.ProjectID, "taskId": opts.TaskID, "items": items}
	if jsonOut {
		return writeJSON(stdout, envelope{OK: true, Command: "pomo task", Meta: meta, Data: data})
	}
	printMapList(stdout, items, "task pomodoros")
	return 0
}

func runPomoList(args []string, timing bool, jsonOut bool, stdout io.Writer, stderr io.Writer) int {
	opts, err := parseRangeListFlags(args, time.Now(), 30)
	command := "pomo list"
	if timing {
		command = "pomo timing"
	}
	if err != nil {
		return failTyped(command, "validation", err.Error(), "run: dida pomo --help", jsonOut, stdout, stderr)
	}
	result, err := executeRead(func(ctx context.Context, client *webapi.Client) (any, error) {
		fromMillis := opts.From.UnixMilli()
		toMillis := opts.To.UnixMilli()
		if timing {
			return client.PomodoroTimings(ctx, fromMillis, toMillis)
		}
		return client.Pomodoros(ctx, fromMillis, toMillis)
	})
	if err != nil {
		return failTyped(command, "api", err.Error(), "", jsonOut, stdout, stderr)
	}
	items := result.([]map[string]any)
	total := len(items)
	items = limitMapItems(items, opts.Limit)
	data := map[string]any{
		"from":  opts.From.Format("2006-01-02"),
		"to":    opts.To.Format("2006-01-02"),
		"items": items,
	}
	meta := map[string]any{"count": len(items), "total": total, "limit": opts.Limit}
	if jsonOut {
		return writeJSON(stdout, envelope{OK: true, Command: command, Meta: meta, Data: data})
	}
	printMapList(stdout, items, "pomodoros")
	return 0
}

func runHabit(args []string, jsonOut bool, stdout io.Writer, stderr io.Writer) int {
	if len(args) == 0 || args[0] == "-h" || args[0] == "--help" {
		printHabitHelp(stdout)
		return 0
	}
	switch args[0] {
	case "preferences", "prefs":
		return runHabitPreferences(jsonOut, stdout, stderr)
	case "list":
		return runHabitList(jsonOut, stdout, stderr)
	case "sections":
		return runHabitSections(jsonOut, stdout, stderr)
	case "checkins":
		return runHabitCheckins(args[1:], jsonOut, stdout, stderr)
	default:
		return fail("habit", fmt.Sprintf("unknown habit command %q", args[0]), jsonOut, stdout, stderr)
	}
}

func runHabitPreferences(jsonOut bool, stdout io.Writer, stderr io.Writer) int {
	result, err := executeRead(func(ctx context.Context, client *webapi.Client) (any, error) {
		return client.HabitPreferences(ctx)
	})
	if err != nil {
		return failTyped("habit preferences", "api", err.Error(), "", jsonOut, stdout, stderr)
	}
	data := map[string]any{"preferences": result}
	if jsonOut {
		return writeJSON(stdout, envelope{OK: true, Command: "habit preferences", Data: data})
	}
	fmt.Fprintf(stdout, "Habit preferences: %d keys\n", len(result.(map[string]any)))
	return 0
}

func runHabitList(jsonOut bool, stdout io.Writer, stderr io.Writer) int {
	result, err := executeRead(func(ctx context.Context, client *webapi.Client) (any, error) {
		return client.Habits(ctx)
	})
	if err != nil {
		return failTyped("habit list", "api", err.Error(), "", jsonOut, stdout, stderr)
	}
	items := result.([]map[string]any)
	meta := map[string]any{"count": len(items)}
	data := map[string]any{"habits": items}
	if jsonOut {
		return writeJSON(stdout, envelope{OK: true, Command: "habit list", Meta: meta, Data: data})
	}
	printMapList(stdout, items, "habits")
	return 0
}

func runHabitSections(jsonOut bool, stdout io.Writer, stderr io.Writer) int {
	result, err := executeRead(func(ctx context.Context, client *webapi.Client) (any, error) {
		return client.HabitSections(ctx)
	})
	if err != nil {
		return failTyped("habit sections", "api", err.Error(), "", jsonOut, stdout, stderr)
	}
	items := result.([]map[string]any)
	meta := map[string]any{"count": len(items)}
	data := map[string]any{"sections": items}
	if jsonOut {
		return writeJSON(stdout, envelope{OK: true, Command: "habit sections", Meta: meta, Data: data})
	}
	printMapList(stdout, items, "habit sections")
	return 0
}

func runHabitCheckins(args []string, jsonOut bool, stdout io.Writer, stderr io.Writer) int {
	habitIDs, afterStamp, err := parseHabitCheckinFlags(args)
	if err != nil {
		return failTyped("habit checkins", "validation", err.Error(), "run: dida habit --help", jsonOut, stdout, stderr)
	}
	result, err := executeRead(func(ctx context.Context, client *webapi.Client) (any, error) {
		return client.HabitCheckins(ctx, habitIDs, afterStamp)
	})
	if err != nil {
		return failTyped("habit checkins", "api", err.Error(), "", jsonOut, stdout, stderr)
	}
	data := map[string]any{"habitIds": habitIDs, "afterStamp": afterStamp, "checkins": result}
	meta := map[string]any{"habitCount": len(habitIDs)}
	if jsonOut {
		return writeJSON(stdout, envelope{OK: true, Command: "habit checkins", Meta: meta, Data: data})
	}
	fmt.Fprintf(stdout, "Habit checkins read for %d habit(s).\n", len(habitIDs))
	return 0
}

func parseHabitCheckinFlags(args []string) ([]string, int64, error) {
	habitIDs := []string{}
	var afterStamp int64
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--habit", "--id":
			if i+1 >= len(args) {
				return nil, 0, fmt.Errorf("%s requires a habit id", args[i])
			}
			habitIDs = append(habitIDs, args[i+1])
			i++
		case "--after-stamp":
			if i+1 >= len(args) {
				return nil, 0, fmt.Errorf("--after-stamp requires a millisecond timestamp")
			}
			value, err := strconv.ParseInt(args[i+1], 10, 64)
			if err != nil || value < 0 {
				return nil, 0, fmt.Errorf("--after-stamp must be a non-negative integer")
			}
			afterStamp = value
			i++
		default:
			return nil, 0, fmt.Errorf("unknown flag %q", args[i])
		}
	}
	return habitIDs, afterStamp, nil
}

func parseRangeListFlags(args []string, now time.Time, defaultDays int) (rangeListOptions, error) {
	to := endOfDay(now)
	from := startOfDay(now.AddDate(0, 0, -defaultDays))
	opts := rangeListOptions{From: from, To: to, Limit: 50}
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--from":
			if i+1 >= len(args) {
				return opts, fmt.Errorf("--from requires YYYY-MM-DD")
			}
			parsed, err := time.ParseInLocation("2006-01-02", args[i+1], now.Location())
			if err != nil {
				return opts, fmt.Errorf("--from must be YYYY-MM-DD")
			}
			opts.From = startOfDay(parsed)
			i++
		case "--to":
			if i+1 >= len(args) {
				return opts, fmt.Errorf("--to requires YYYY-MM-DD")
			}
			parsed, err := time.ParseInLocation("2006-01-02", args[i+1], now.Location())
			if err != nil {
				return opts, fmt.Errorf("--to must be YYYY-MM-DD")
			}
			opts.To = endOfDay(parsed)
			i++
		case "--limit":
			if i+1 >= len(args) {
				return opts, fmt.Errorf("--limit requires a value")
			}
			if _, err := fmt.Sscanf(args[i+1], "%d", &opts.Limit); err != nil || opts.Limit < 0 {
				return opts, fmt.Errorf("--limit must be a non-negative integer")
			}
			i++
		default:
			return opts, fmt.Errorf("unknown flag %q", args[i])
		}
	}
	if opts.From.After(opts.To) {
		return opts, fmt.Errorf("--from must be before or equal to --to")
	}
	return opts, nil
}

func startOfDay(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
}

func endOfDay(t time.Time) time.Time {
	return startOfDay(t).AddDate(0, 0, 1).Add(-time.Millisecond)
}

func limitMapItems(items []map[string]any, limit int) []map[string]any {
	if limit > 0 && len(items) > limit {
		return items[:limit]
	}
	return items
}
