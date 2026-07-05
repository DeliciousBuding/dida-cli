package cli

import (
	"context"
	"fmt"
	"io"

	"github.com/DeliciousBuding/dida-cli/internal/webapi"
)

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
			parsed, err := parseIntStrict(args[i+1])
			if err != nil || parsed < 0 {
				return "", 0, false, fmt.Errorf("--limit must be a non-negative integer")
			}
			limit = parsed
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

// compactList applies a compact function to each item in a list within a payload map.
// Shared by search and user commands.
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

// pickKeys returns a new map with only the specified keys from the input map.
// Shared by search and user commands.
func pickKeys(item map[string]any, keys ...string) map[string]any {
	out := map[string]any{}
	for _, key := range keys {
		if value, ok := item[key]; ok {
			out[key] = value
		}
	}
	return out
}
