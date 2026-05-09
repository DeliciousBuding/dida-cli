package cli

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/DeliciousBuding/dida-cli/internal/webapi"
)

type closedListOptions struct {
	ProjectIDs      []string
	Statuses        []int
	From            string
	To              string
	CompletedUserID string
	Limit           int
}

func runClosed(args []string, jsonOut bool, stdout io.Writer, stderr io.Writer) int {
	if len(args) == 0 || args[0] == "-h" || args[0] == "--help" {
		printClosedHelp(stdout)
		return 0
	}
	if args[0] != "list" {
		return fail("closed", fmt.Sprintf("unknown closed command %q", args[0]), jsonOut, stdout, stderr)
	}
	opts, err := parseClosedListFlags(args[1:], time.Now())
	if err != nil {
		return failTyped("closed list", "validation", err.Error(), "run: dida closed --help", jsonOut, stdout, stderr)
	}
	result, err := executeRead(func(ctx context.Context, client *webapi.Client) (any, error) {
		return client.ClosedItems(ctx, opts.ProjectIDs, opts.Statuses, opts.From, opts.To, opts.CompletedUserID)
	})
	if err != nil {
		return failTyped("closed list", "api", err.Error(), "", jsonOut, stdout, stderr)
	}
	items := result.([]map[string]any)
	total := len(items)
	items = limitMapItems(items, opts.Limit)
	data := map[string]any{
		"items":           items,
		"projectIds":      opts.ProjectIDs,
		"statuses":        opts.Statuses,
		"from":            opts.From,
		"to":              opts.To,
		"completedUserId": opts.CompletedUserID,
	}
	meta := map[string]any{"count": len(items), "total": total, "limit": opts.Limit}
	if jsonOut {
		return writeJSON(stdout, envelope{OK: true, Command: "closed list", Meta: meta, Data: data})
	}
	printMapList(stdout, items, "closed items")
	return 0
}

func parseClosedListFlags(args []string, now time.Time) (closedListOptions, error) {
	opts := closedListOptions{
		From:  startOfDay(now.AddDate(0, 0, -30)).Format("2006-01-02 15:04:05"),
		To:    endOfDay(now).Format("2006-01-02 15:04:05"),
		Limit: 50,
	}
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--project", "-p":
			if i+1 >= len(args) {
				return opts, fmt.Errorf("%s requires a project id", args[i])
			}
			opts.ProjectIDs = append(opts.ProjectIDs, args[i+1])
			i++
		case "--status":
			if i+1 >= len(args) {
				return opts, fmt.Errorf("--status requires a value")
			}
			var status int
			if _, err := fmt.Sscanf(args[i+1], "%d", &status); err != nil || status < 0 {
				return opts, fmt.Errorf("--status must be a non-negative integer")
			}
			opts.Statuses = append(opts.Statuses, status)
			i++
		case "--from":
			if i+1 >= len(args) {
				return opts, fmt.Errorf("--from requires YYYY-MM-DD")
			}
			parsed, err := time.ParseInLocation("2006-01-02", args[i+1], now.Location())
			if err != nil {
				return opts, fmt.Errorf("--from must be YYYY-MM-DD")
			}
			opts.From = startOfDay(parsed).Format("2006-01-02 15:04:05")
			i++
		case "--to":
			if i+1 >= len(args) {
				return opts, fmt.Errorf("--to requires YYYY-MM-DD")
			}
			parsed, err := time.ParseInLocation("2006-01-02", args[i+1], now.Location())
			if err != nil {
				return opts, fmt.Errorf("--to must be YYYY-MM-DD")
			}
			opts.To = endOfDay(parsed).Format("2006-01-02 15:04:05")
			i++
		case "--completed-user":
			if i+1 >= len(args) {
				return opts, fmt.Errorf("--completed-user requires a user id")
			}
			opts.CompletedUserID = args[i+1]
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
	return opts, nil
}
