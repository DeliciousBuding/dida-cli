package cli

import (
	"context"
	"fmt"
	"io"

	"github.com/DeliciousBuding/dida-cli/internal/webapi"
)

func runAttachment(args []string, jsonOut bool, stdout io.Writer, stderr io.Writer) int {
	if len(args) == 0 || args[0] == "-h" || args[0] == "--help" {
		printAttachmentHelp(stdout)
		return 0
	}
	switch args[0] {
	case "quota":
		return runAttachmentQuota(jsonOut, stdout, stderr)
	default:
		return fail("attachment", fmt.Sprintf("unknown attachment command %q", args[0]), jsonOut, stdout, stderr)
	}
}

func runAttachmentQuota(jsonOut bool, stdout io.Writer, stderr io.Writer) int {
	result, err := executeRead(func(ctx context.Context, client *webapi.Client) (any, error) {
		return client.AttachmentQuota(ctx)
	})
	if err != nil {
		return failTyped("attachment quota", "api", err.Error(), "", jsonOut, stdout, stderr)
	}
	data := map[string]any{"quota": result}
	if jsonOut {
		return writeJSON(stdout, envelope{OK: true, Command: "attachment quota", Data: data})
	}
	quota := result.(map[string]any)
	fmt.Fprintf(stdout, "Under quota: %v\nDaily limit: %v\n", quota["underQuota"], quota["dailyLimit"])
	return 0
}

func runReminder(args []string, jsonOut bool, stdout io.Writer, stderr io.Writer) int {
	if len(args) == 0 || args[0] == "-h" || args[0] == "--help" {
		printReminderHelp(stdout)
		return 0
	}
	switch args[0] {
	case "daily", "preferences", "prefs":
		return runDailyReminder(jsonOut, stdout, stderr)
	default:
		return fail("reminder", fmt.Sprintf("unknown reminder command %q", args[0]), jsonOut, stdout, stderr)
	}
}

func runDailyReminder(jsonOut bool, stdout io.Writer, stderr io.Writer) int {
	result, err := executeRead(func(ctx context.Context, client *webapi.Client) (any, error) {
		return client.DailyReminderPreferences(ctx)
	})
	if err != nil {
		return failTyped("reminder daily", "api", err.Error(), "", jsonOut, stdout, stderr)
	}
	data := map[string]any{"preferences": result}
	if jsonOut {
		return writeJSON(stdout, envelope{OK: true, Command: "reminder daily", Data: data})
	}
	fmt.Fprintf(stdout, "Daily reminder preferences: %d keys\n", len(result.(map[string]any)))
	return 0
}

func runShare(args []string, jsonOut bool, stdout io.Writer, stderr io.Writer) int {
	if len(args) == 0 || args[0] == "-h" || args[0] == "--help" {
		printShareHelp(stdout)
		return 0
	}
	switch args[0] {
	case "contacts":
		return runShareContacts(jsonOut, stdout, stderr)
	case "recent-users":
		return runRecentProjectUsers(jsonOut, stdout, stderr)
	case "project":
		return runShareProject(args[1:], jsonOut, stdout, stderr)
	default:
		return fail("share", fmt.Sprintf("unknown share command %q", args[0]), jsonOut, stdout, stderr)
	}
}

func runShareContacts(jsonOut bool, stdout io.Writer, stderr io.Writer) int {
	result, err := executeRead(func(ctx context.Context, client *webapi.Client) (any, error) {
		return client.ShareContacts(ctx)
	})
	if err != nil {
		return failTyped("share contacts", "api", err.Error(), "", jsonOut, stdout, stderr)
	}
	data := map[string]any{"contacts": result}
	if jsonOut {
		return writeJSON(stdout, envelope{OK: true, Command: "share contacts", Data: data})
	}
	fmt.Fprintln(stdout, "Share contacts read.")
	return 0
}

func runRecentProjectUsers(jsonOut bool, stdout io.Writer, stderr io.Writer) int {
	result, err := executeRead(func(ctx context.Context, client *webapi.Client) (any, error) {
		return client.RecentProjectUsers(ctx)
	})
	if err != nil {
		return failTyped("share recent-users", "api", err.Error(), "", jsonOut, stdout, stderr)
	}
	items := result.([]map[string]any)
	meta := map[string]any{"count": len(items)}
	data := map[string]any{"recentUsers": items}
	if jsonOut {
		return writeJSON(stdout, envelope{OK: true, Command: "share recent-users", Meta: meta, Data: data})
	}
	printMapList(stdout, items, "recent project users")
	return 0
}

func runShareProject(args []string, jsonOut bool, stdout io.Writer, stderr io.Writer) int {
	if len(args) < 2 {
		return failTyped("share project", "validation", "usage: dida share project <shares|quota|invite-url> <project-id>", "run: dida share --help", jsonOut, stdout, stderr)
	}
	action, projectID := args[0], args[1]
	switch action {
	case "shares":
		return runProjectShares(projectID, jsonOut, stdout, stderr)
	case "quota":
		return runProjectShareQuota(projectID, jsonOut, stdout, stderr)
	case "invite-url":
		return runProjectInviteURL(projectID, jsonOut, stdout, stderr)
	default:
		return failTyped("share project", "validation", fmt.Sprintf("unknown project share action %q", action), "run: dida share --help", jsonOut, stdout, stderr)
	}
}

func runProjectShares(projectID string, jsonOut bool, stdout io.Writer, stderr io.Writer) int {
	result, err := executeRead(func(ctx context.Context, client *webapi.Client) (any, error) {
		return client.ProjectShares(ctx, projectID)
	})
	if err != nil {
		return failTyped("share project shares", "api", err.Error(), "", jsonOut, stdout, stderr)
	}
	items := result.([]map[string]any)
	meta := map[string]any{"count": len(items)}
	data := map[string]any{"projectId": projectID, "shares": items}
	if jsonOut {
		return writeJSON(stdout, envelope{OK: true, Command: "share project shares", Meta: meta, Data: data})
	}
	printMapList(stdout, items, "project shares")
	return 0
}

func runProjectShareQuota(projectID string, jsonOut bool, stdout io.Writer, stderr io.Writer) int {
	result, err := executeRead(func(ctx context.Context, client *webapi.Client) (any, error) {
		return client.ProjectShareQuota(ctx, projectID)
	})
	if err != nil {
		return failTyped("share project quota", "api", err.Error(), "", jsonOut, stdout, stderr)
	}
	data := map[string]any{"projectId": projectID, "quota": result}
	if jsonOut {
		return writeJSON(stdout, envelope{OK: true, Command: "share project quota", Data: data})
	}
	fmt.Fprintf(stdout, "Project share quota: %v\n", result)
	return 0
}

func runProjectInviteURL(projectID string, jsonOut bool, stdout io.Writer, stderr io.Writer) int {
	result, err := executeRead(func(ctx context.Context, client *webapi.Client) (any, error) {
		return client.ProjectInviteURL(ctx, projectID)
	})
	if err != nil {
		return failTyped("share project invite-url", "api", err.Error(), "", jsonOut, stdout, stderr)
	}
	data := map[string]any{"projectId": projectID, "invite": result}
	if jsonOut {
		return writeJSON(stdout, envelope{OK: true, Command: "share project invite-url", Data: data})
	}
	fmt.Fprintln(stdout, "Project invite URL read.")
	return 0
}

func runCalendar(args []string, jsonOut bool, stdout io.Writer, stderr io.Writer) int {
	if len(args) == 0 || args[0] == "-h" || args[0] == "--help" {
		printCalendarHelp(stdout)
		return 0
	}
	switch args[0] {
	case "subscriptions":
		return runCalendarSubscriptions(jsonOut, stdout, stderr)
	case "archived":
		return runCalendarArchivedEvents(jsonOut, stdout, stderr)
	case "third-accounts":
		return runCalendarThirdAccounts(jsonOut, stdout, stderr)
	default:
		return fail("calendar", fmt.Sprintf("unknown calendar command %q", args[0]), jsonOut, stdout, stderr)
	}
}

func runCalendarSubscriptions(jsonOut bool, stdout io.Writer, stderr io.Writer) int {
	result, err := executeRead(func(ctx context.Context, client *webapi.Client) (any, error) {
		return client.CalendarSubscriptions(ctx)
	})
	if err != nil {
		return failTyped("calendar subscriptions", "api", err.Error(), "", jsonOut, stdout, stderr)
	}
	items := result.([]map[string]any)
	meta := map[string]any{"count": len(items)}
	data := map[string]any{"subscriptions": items}
	if jsonOut {
		return writeJSON(stdout, envelope{OK: true, Command: "calendar subscriptions", Meta: meta, Data: data})
	}
	printMapList(stdout, items, "calendar subscriptions")
	return 0
}

func runCalendarArchivedEvents(jsonOut bool, stdout io.Writer, stderr io.Writer) int {
	result, err := executeRead(func(ctx context.Context, client *webapi.Client) (any, error) {
		return client.CalendarArchivedEvents(ctx)
	})
	if err != nil {
		return failTyped("calendar archived", "api", err.Error(), "", jsonOut, stdout, stderr)
	}
	items := result.([]map[string]any)
	meta := map[string]any{"count": len(items)}
	data := map[string]any{"events": items}
	if jsonOut {
		return writeJSON(stdout, envelope{OK: true, Command: "calendar archived", Meta: meta, Data: data})
	}
	printMapList(stdout, items, "archived calendar events")
	return 0
}

func runCalendarThirdAccounts(jsonOut bool, stdout io.Writer, stderr io.Writer) int {
	result, err := executeRead(func(ctx context.Context, client *webapi.Client) (any, error) {
		return client.CalendarThirdAccounts(ctx)
	})
	if err != nil {
		return failTyped("calendar third-accounts", "api", err.Error(), "", jsonOut, stdout, stderr)
	}
	data := map[string]any{"accounts": result}
	if jsonOut {
		return writeJSON(stdout, envelope{OK: true, Command: "calendar third-accounts", Data: data})
	}
	fmt.Fprintln(stdout, "Calendar third-party accounts read.")
	return 0
}

func runStats(args []string, jsonOut bool, stdout io.Writer, stderr io.Writer) int {
	if len(args) == 0 || args[0] == "-h" || args[0] == "--help" {
		printStatsHelp(stdout)
		return 0
	}
	switch args[0] {
	case "general":
		return runStatsGeneral(jsonOut, stdout, stderr)
	default:
		return fail("stats", fmt.Sprintf("unknown stats command %q", args[0]), jsonOut, stdout, stderr)
	}
}

func runStatsGeneral(jsonOut bool, stdout io.Writer, stderr io.Writer) int {
	result, err := executeRead(func(ctx context.Context, client *webapi.Client) (any, error) {
		return client.StatisticsGeneral(ctx)
	})
	if err != nil {
		return failTyped("stats general", "api", err.Error(), "", jsonOut, stdout, stderr)
	}
	data := map[string]any{"statistics": result}
	if jsonOut {
		return writeJSON(stdout, envelope{OK: true, Command: "stats general", Data: data})
	}
	fmt.Fprintf(stdout, "General statistics: %d keys\n", len(result.(map[string]any)))
	return 0
}

func runTemplate(args []string, jsonOut bool, stdout io.Writer, stderr io.Writer) int {
	if len(args) == 0 || args[0] == "-h" || args[0] == "--help" {
		printTemplateHelp(stdout)
		return 0
	}
	if args[0] != "project" || len(args) < 2 || args[1] != "list" {
		return failTyped("template", "validation", "usage: dida template project list [--timestamp N] [--limit N]", "run: dida template --help", jsonOut, stdout, stderr)
	}
	timestamp, limit, err := parseTemplateListFlags(args[2:])
	if err != nil {
		return failTyped("template project list", "validation", err.Error(), "run: dida template --help", jsonOut, stdout, stderr)
	}
	result, err := executeRead(func(ctx context.Context, client *webapi.Client) (any, error) {
		return client.ProjectTemplates(ctx, timestamp)
	})
	if err != nil {
		return failTyped("template project list", "api", err.Error(), "", jsonOut, stdout, stderr)
	}
	payload := result.(map[string]any)
	templates, _ := payload["projectTemplates"].([]any)
	total := len(templates)
	if limit > 0 && len(templates) > limit {
		templates = templates[:limit]
		payload["projectTemplates"] = templates
	}
	meta := map[string]any{"count": len(templates), "total": total, "limit": limit}
	if jsonOut {
		return writeJSON(stdout, envelope{OK: true, Command: "template project list", Meta: meta, Data: payload})
	}
	fmt.Fprintf(stdout, "Project templates: %d of %d\n", len(templates), total)
	return 0
}

func runSearch(args []string, jsonOut bool, stdout io.Writer, stderr io.Writer) int {
	if len(args) == 0 || args[0] == "-h" || args[0] == "--help" {
		printSearchHelp(stdout)
		return 0
	}
	if args[0] != "all" {
		return fail("search", fmt.Sprintf("unknown search command %q", args[0]), jsonOut, stdout, stderr)
	}
	query, limit, full, err := parseSearchAllFlags(args[1:])
	if err != nil {
		return failTyped("search all", "validation", err.Error(), "run: dida search --help", jsonOut, stdout, stderr)
	}
	result, err := executeRead(func(ctx context.Context, client *webapi.Client) (any, error) {
		return client.SearchAll(ctx, query)
	})
	if err != nil {
		return failTyped("search all", "api", err.Error(), "", jsonOut, stdout, stderr)
	}
	payload := result.(map[string]any)
	if !full {
		compactSearchPayload(payload)
	}
	total := limitSearchPayload(payload, limit)
	meta := map[string]any{"query": query, "limit": limit, "compact": !full}
	for key, value := range total {
		meta[key] = value
	}
	if jsonOut {
		return writeJSON(stdout, envelope{OK: true, Command: "search all", Meta: meta, Data: payload})
	}
	fmt.Fprintf(stdout, "Search results for %q: %v\n", query, total)
	return 0
}

func parseSearchAllFlags(args []string) (string, int, bool, error) {
	query := ""
	limit := 20
	full := false
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--query", "-q":
			if i+1 >= len(args) {
				return "", 0, false, fmt.Errorf("%s requires text", args[i])
			}
			query = args[i+1]
			i++
		case "--limit":
			if i+1 >= len(args) {
				return "", 0, false, fmt.Errorf("--limit requires a value")
			}
			if _, err := fmt.Sscanf(args[i+1], "%d", &limit); err != nil || limit < 0 {
				return "", 0, false, fmt.Errorf("--limit must be a non-negative integer")
			}
			i++
		case "--full":
			full = true
		default:
			if query == "" {
				query = args[i]
				continue
			}
			return "", 0, false, fmt.Errorf("unknown flag %q", args[i])
		}
	}
	if query == "" {
		return "", 0, false, fmt.Errorf("missing query; use --query <text>")
	}
	return query, limit, full, nil
}

func limitSearchPayload(payload map[string]any, limit int) map[string]int {
	counts := map[string]int{}
	for _, key := range []string{"hits", "tasks", "comments"} {
		items, ok := payload[key].([]any)
		if !ok {
			continue
		}
		counts[key+"Total"] = len(items)
		if limit > 0 && len(items) > limit {
			items = items[:limit]
			payload[key] = items
		}
		counts[key+"Count"] = len(items)
	}
	return counts
}

func compactSearchPayload(payload map[string]any) {
	compactList(payload, "hits", func(item map[string]any) map[string]any {
		out := pickKeys(item, "index", "id")
		if source, ok := item["source"].(map[string]any); ok {
			out["source"] = pickKeys(source, "title", "projectId", "modifiedTime")
		}
		return out
	})
	compactList(payload, "tasks", func(item map[string]any) map[string]any {
		return pickKeys(item, "id", "projectId", "title", "status", "priority", "startDate", "dueDate", "modifiedTime", "completedTime")
	})
	compactList(payload, "comments", func(item map[string]any) map[string]any {
		return pickKeys(item, "id", "projectId", "taskId", "title", "createdTime", "modifiedTime")
	})
}

func compactList(payload map[string]any, key string, compact func(map[string]any) map[string]any) {
	items, ok := payload[key].([]any)
	if !ok {
		return
	}
	out := make([]any, 0, len(items))
	for _, item := range items {
		if object, ok := item.(map[string]any); ok {
			out = append(out, compact(object))
		} else {
			out = append(out, item)
		}
	}
	payload[key] = out
}

func pickKeys(item map[string]any, keys ...string) map[string]any {
	out := map[string]any{}
	for _, key := range keys {
		if value, ok := item[key]; ok {
			out[key] = value
		}
	}
	return out
}

func parseTemplateListFlags(args []string) (int64, int, error) {
	var timestamp int64
	limit := 50
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--timestamp":
			if i+1 >= len(args) {
				return 0, 0, fmt.Errorf("--timestamp requires a value")
			}
			if _, err := fmt.Sscanf(args[i+1], "%d", &timestamp); err != nil || timestamp < 0 {
				return 0, 0, fmt.Errorf("--timestamp must be a non-negative integer")
			}
			i++
		case "--limit":
			if i+1 >= len(args) {
				return 0, 0, fmt.Errorf("--limit requires a value")
			}
			if _, err := fmt.Sscanf(args[i+1], "%d", &limit); err != nil || limit < 0 {
				return 0, 0, fmt.Errorf("--limit must be a non-negative integer")
			}
			i++
		default:
			return 0, 0, fmt.Errorf("unknown flag %q", args[i])
		}
	}
	return timestamp, limit, nil
}
