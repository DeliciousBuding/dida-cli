package cli

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/DeliciousBuding/dida-cli/internal/auth"
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
		parsedFrom, parsedTo, parsedLimit, err := parseCompletedListFlags(args[1:], now)
		if err != nil {
			return failTyped("completed list", "validation", err.Error(), "run: dida completed --help", jsonOut, stdout, stderr)
		}
		from, to, limit = parsedFrom, parsedTo, parsedLimit
	default:
		return fail("completed", fmt.Sprintf("unknown completed command %q", args[0]), jsonOut, stdout, stderr)
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
		"from":  formatDidaQueryTime(from),
		"to":    formatDidaQueryTime(to),
		"tasks": stripTaskRaw(normalized),
	}
	meta := map[string]any{"count": len(normalized), "limit": limit}
	if jsonOut {
		return writeJSON(stdout, envelope{OK: true, Command: command, Meta: meta, Data: data})
	}
	printTasks(stdout, normalized, len(normalized))
	return 0
}

func parseCompletedListFlags(args []string, now time.Time) (time.Time, time.Time, int, error) {
	from, to := dayRange(now)
	limit := 100
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--from":
			if i+1 >= len(args) {
				return time.Time{}, time.Time{}, 0, fmt.Errorf("--from requires YYYY-MM-DD")
			}
			parsed, err := time.ParseInLocation("2006-01-02", args[i+1], now.Location())
			if err != nil {
				return time.Time{}, time.Time{}, 0, fmt.Errorf("--from must be YYYY-MM-DD")
			}
			from = parsed
			i++
		case "--to":
			if i+1 >= len(args) {
				return time.Time{}, time.Time{}, 0, fmt.Errorf("--to requires YYYY-MM-DD")
			}
			parsed, err := time.ParseInLocation("2006-01-02", args[i+1], now.Location())
			if err != nil {
				return time.Time{}, time.Time{}, 0, fmt.Errorf("--to must be YYYY-MM-DD")
			}
			to = parsed.AddDate(0, 0, 1).Add(-time.Second)
			i++
		case "--limit":
			if i+1 >= len(args) {
				return time.Time{}, time.Time{}, 0, fmt.Errorf("--limit requires a value")
			}
			if _, err := fmt.Sscanf(args[i+1], "%d", &limit); err != nil || limit <= 0 {
				return time.Time{}, time.Time{}, 0, fmt.Errorf("--limit must be a positive integer")
			}
			i++
		default:
			return time.Time{}, time.Time{}, 0, fmt.Errorf("unknown flag %q", args[i])
		}
	}
	if from.After(to) {
		return time.Time{}, time.Time{}, 0, fmt.Errorf("--from must be before or equal to --to")
	}
	return from, to, limit, nil
}

func loadCompletedTasks(from time.Time, to time.Time, limit int) ([]map[string]any, error) {
	token, err := auth.LoadCookieToken()
	if err != nil {
		return nil, fmt.Errorf("missing cookie auth; run: dida auth cookie set --token-stdin")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	return webapi.NewClient(token.Token).CompletedTasks(ctx, formatDidaQueryTime(from), formatDidaQueryTime(to), limit)
}

func dayRange(t time.Time) (time.Time, time.Time) {
	start := time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
	return start, start.AddDate(0, 0, 1).Add(-time.Second)
}

func formatDidaQueryTime(t time.Time) string {
	return t.Format("2006-01-02 15:04:05")
}
