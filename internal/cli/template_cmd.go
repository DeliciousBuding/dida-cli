package cli

import (
	"context"
	"fmt"
	"io"

	"github.com/DeliciousBuding/dida-cli/internal/webapi"
)

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

func parseTemplateListFlags(args []string) (int64, int, error) {
	var timestamp int64
	limit := 50
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--timestamp":
			if i+1 >= len(args) {
				return 0, 0, fmt.Errorf("--timestamp requires a value")
			}
			parsed, err := parseInt64Strict(args[i+1])
			if err != nil || parsed < 0 {
				return 0, 0, fmt.Errorf("--timestamp must be a non-negative integer")
			}
			timestamp = parsed
			i++
		case "--limit":
			if i+1 >= len(args) {
				return 0, 0, fmt.Errorf("--limit requires a value")
			}
			parsed, err := parseIntStrict(args[i+1])
			if err != nil || parsed < 0 {
				return 0, 0, fmt.Errorf("--limit must be a non-negative integer")
			}
			limit = parsed
			i++
		default:
			return 0, 0, fmt.Errorf("unknown flag %q", args[i])
		}
	}
	return timestamp, limit, nil
}
