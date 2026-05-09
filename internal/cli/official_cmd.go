package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/DeliciousBuding/dida-cli/internal/officialmcp"
)

func runOfficial(args []string, jsonOut bool, stdout io.Writer, stderr io.Writer) int {
	if len(args) == 0 || args[0] == "-h" || args[0] == "--help" {
		printOfficialHelp(stdout)
		return 0
	}
	switch args[0] {
	case "doctor":
		return runOfficialDoctor(jsonOut, stdout, stderr)
	case "tools":
		return runOfficialTools(args[1:], jsonOut, stdout, stderr)
	case "show":
		return runOfficialShow(args[1:], jsonOut, stdout, stderr)
	case "call":
		return runOfficialCall(args[1:], jsonOut, stdout, stderr)
	case "habit":
		return runOfficialHabit(args[1:], jsonOut, stdout, stderr)
	case "focus":
		return runOfficialFocus(args[1:], jsonOut, stdout, stderr)
	default:
		return fail("official", fmt.Sprintf("unknown official command %q", args[0]), jsonOut, stdout, stderr)
	}
}

func runOfficialDoctor(jsonOut bool, stdout io.Writer, stderr io.Writer) int {
	token, err := officialmcp.ResolveToken("")
	if err != nil {
		return failTyped("official doctor", "auth", err.Error(), "set DIDA365_TOKEN to the official dida365 API token", jsonOut, stdout, stderr)
	}
	client := officialmcp.NewClient(token)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	if err := client.Initialize(ctx, "dida-cli", "0.1.0"); err != nil {
		return failTyped("official doctor", "api", err.Error(), "", jsonOut, stdout, stderr)
	}
	tools, err := client.Tools(ctx)
	if err != nil {
		return failTyped("official doctor", "api", err.Error(), "", jsonOut, stdout, stderr)
	}
	names := make([]string, 0, len(tools))
	for i, tool := range tools {
		if i >= 8 {
			break
		}
		names = append(names, tool.Name)
	}
	data := map[string]any{
		"url":             client.URL,
		"sessionId":       client.SessionID,
		"toolCount":       len(tools),
		"sampleToolNames": names,
	}
	if jsonOut {
		return writeJSON(stdout, envelope{OK: true, Command: "official doctor", Data: data})
	}
	fmt.Fprintf(stdout, "Official MCP URL: %s\n", client.URL)
	fmt.Fprintf(stdout, "Tool count: %d\n", len(tools))
	return 0
}

func runOfficialTools(args []string, jsonOut bool, stdout io.Writer, stderr io.Writer) int {
	limit, full, err := parseOfficialToolsFlags(args)
	if err != nil {
		return failTyped("official tools", "validation", err.Error(), "run: dida official --help", jsonOut, stdout, stderr)
	}
	token, err := officialmcp.ResolveToken("")
	if err != nil {
		return failTyped("official tools", "auth", err.Error(), "set DIDA365_TOKEN to the official dida365 API token", jsonOut, stdout, stderr)
	}
	client := officialmcp.NewClient(token)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	if err := client.Initialize(ctx, "dida-cli", "0.1.0"); err != nil {
		return failTyped("official tools", "api", err.Error(), "", jsonOut, stdout, stderr)
	}
	tools, err := client.Tools(ctx)
	if err != nil {
		return failTyped("official tools", "api", err.Error(), "", jsonOut, stdout, stderr)
	}
	total := len(tools)
	if !full {
		tools = compactOfficialTools(tools)
	}
	if limit > 0 && len(tools) > limit {
		tools = tools[:limit]
	}
	meta := map[string]any{"count": len(tools), "total": total, "limit": limit, "compact": !full}
	data := map[string]any{"tools": tools}
	if jsonOut {
		return writeJSON(stdout, envelope{OK: true, Command: "official tools", Meta: meta, Data: data})
	}
	fmt.Fprintf(stdout, "Official tools: %d of %d\n", len(tools), total)
	return 0
}

func runOfficialShow(args []string, jsonOut bool, stdout io.Writer, stderr io.Writer) int {
	if len(args) != 1 {
		return failTyped("official show", "validation", "usage: dida official show <tool-name>", "run: dida official --help", jsonOut, stdout, stderr)
	}
	token, err := officialmcp.ResolveToken("")
	if err != nil {
		return failTyped("official show", "auth", err.Error(), "set DIDA365_TOKEN to the official dida365 API token", jsonOut, stdout, stderr)
	}
	client := officialmcp.NewClient(token)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	if err := client.Initialize(ctx, "dida-cli", "0.1.0"); err != nil {
		return failTyped("official show", "api", err.Error(), "", jsonOut, stdout, stderr)
	}
	tool, err := client.ToolSchema(ctx, args[0])
	if err != nil {
		return failTyped("official show", "not_found", err.Error(), "", jsonOut, stdout, stderr)
	}
	if jsonOut {
		return writeJSON(stdout, envelope{OK: true, Command: "official show", Data: map[string]any{"tool": tool}})
	}
	fmt.Fprintf(stdout, "Official tool: %s\n", tool.Name)
	return 0
}

func runOfficialCall(args []string, jsonOut bool, stdout io.Writer, stderr io.Writer) int {
	tool, payload, err := parseOfficialCallFlags(args)
	if err != nil {
		return failTyped("official call", "validation", err.Error(), "run: dida official --help", jsonOut, stdout, stderr)
	}
	token, err := officialmcp.ResolveToken("")
	if err != nil {
		return failTyped("official call", "auth", err.Error(), "set DIDA365_TOKEN to the official dida365 API token", jsonOut, stdout, stderr)
	}
	client := officialmcp.NewClient(token)
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	if err := client.Initialize(ctx, "dida-cli", "0.1.0"); err != nil {
		return failTyped("official call", "api", err.Error(), "", jsonOut, stdout, stderr)
	}
	result, err := client.CallTool(ctx, tool, payload)
	if err != nil {
		return failTyped("official call", "api", err.Error(), "", jsonOut, stdout, stderr)
	}
	if jsonOut {
		return writeJSON(stdout, envelope{OK: true, Command: "official call", Data: map[string]any{"tool": tool, "result": result}})
	}
	return writeJSON(stdout, result)
}

func parseOfficialToolsFlags(args []string) (int, bool, error) {
	limit := 100
	full := false
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--limit":
			if i+1 >= len(args) {
				return 0, false, fmt.Errorf("--limit requires a value")
			}
			if _, err := fmt.Sscanf(args[i+1], "%d", &limit); err != nil || limit < 0 {
				return 0, false, fmt.Errorf("--limit must be a non-negative integer")
			}
			i++
		case "--full":
			full = true
		default:
			return 0, false, fmt.Errorf("unknown flag %q", args[i])
		}
	}
	return limit, full, nil
}

func parseOfficialCallFlags(args []string) (string, map[string]any, error) {
	if len(args) == 0 {
		return "", nil, fmt.Errorf("usage: dida official call <tool-name> [--args-json <json>] [--args-file <file>]")
	}
	tool := args[0]
	payload, err := parseOfficialPayloadFlags(args[1:])
	if err != nil {
		return "", nil, err
	}
	return tool, payload, nil
}

func parseOfficialPayloadFlags(args []string) (map[string]any, error) {
	payload := map[string]any{}
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--args-json":
			if i+1 >= len(args) {
				return nil, fmt.Errorf("--args-json requires a value")
			}
			if err := json.Unmarshal([]byte(args[i+1]), &payload); err != nil {
				return nil, fmt.Errorf("decode --args-json: %w", err)
			}
			i++
		case "--args-file":
			if i+1 >= len(args) {
				return nil, fmt.Errorf("--args-file requires a path")
			}
			data, err := os.ReadFile(args[i+1])
			if err != nil {
				return nil, fmt.Errorf("read args file: %w", err)
			}
			if err := json.Unmarshal(data, &payload); err != nil {
				return nil, fmt.Errorf("decode args file: %w", err)
			}
			i++
		default:
			return nil, fmt.Errorf("unknown flag %q", args[i])
		}
	}
	return payload, nil
}

func compactOfficialTools(tools []officialmcp.Tool) []officialmcp.Tool {
	out := make([]officialmcp.Tool, 0, len(tools))
	for _, tool := range tools {
		compact := officialmcp.Tool{
			Name:        tool.Name,
			Description: tool.Description,
		}
		if len(tool.Annotations) > 0 {
			compact.Annotations = tool.Annotations
		}
		out = append(out, compact)
	}
	return out
}

func runOfficialHabit(args []string, jsonOut bool, stdout io.Writer, stderr io.Writer) int {
	if len(args) == 0 || args[0] == "-h" || args[0] == "--help" {
		printOfficialHabitHelp(stdout)
		return 0
	}

	token, err := officialmcp.ResolveToken("")
	if err != nil {
		return failTyped("official habit", "auth", err.Error(), "set DIDA365_TOKEN to the official dida365 API token", jsonOut, stdout, stderr)
	}
	client := officialmcp.NewClient(token)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	if err := client.Initialize(ctx, "dida-cli", "0.1.0"); err != nil {
		return failTyped("official habit", "api", err.Error(), "", jsonOut, stdout, stderr)
	}

	switch args[0] {
	case "get":
		if len(args) < 2 {
			return failTyped("official habit get", "validation", "usage: dida official habit get <habit-id>", "run: dida official habit --help", jsonOut, stdout, stderr)
		}
		habitID := args[1]
		result, err := client.CallTool(ctx, "get_habit", map[string]any{"habitId": habitID})
		if err != nil {
			return failTyped("official habit get", "api", err.Error(), "", jsonOut, stdout, stderr)
		}
		if jsonOut {
			return writeJSON(stdout, envelope{OK: true, Command: "official habit get", Data: result})
		}
		return writeJSON(stdout, result)

	case "create":
		payload, err := parseHabitCreateArgs(args[1:])
		if err != nil {
			return failTyped("official habit create", "validation", err.Error(), "run: dida official habit --help", jsonOut, stdout, stderr)
		}
		result, err := client.CallTool(ctx, "create_habit", payload)
		if err != nil {
			return failTyped("official habit create", "api", err.Error(), "", jsonOut, stdout, stderr)
		}
		if jsonOut {
			return writeJSON(stdout, envelope{OK: true, Command: "official habit create", Data: result})
		}
		return writeJSON(stdout, result)

	case "update":
		if len(args) < 2 {
			return failTyped("official habit update", "validation", "usage: dida official habit update <habit-id> [--args-json <json>]", "run: dida official habit --help", jsonOut, stdout, stderr)
		}
		habitID := args[1]
		payload, err := parseOfficialPayloadFlags(args[2:])
		if err != nil {
			return failTyped("official habit update", "validation", err.Error(), "", jsonOut, stdout, stderr)
		}
		payload["habitId"] = habitID
		result, err := client.CallTool(ctx, "update_habit", payload)
		if err != nil {
			return failTyped("official habit update", "api", err.Error(), "", jsonOut, stdout, stderr)
		}
		if jsonOut {
			return writeJSON(stdout, envelope{OK: true, Command: "official habit update", Data: result})
		}
		return writeJSON(stdout, result)

	case "checkin":
		if len(args) < 2 {
			return failTyped("official habit checkin", "validation", "usage: dida official habit checkin <habit-id> --date YYYY-MM-DD --value <number>", "run: dida official habit --help", jsonOut, stdout, stderr)
		}
		habitID := args[1]
		payload, err := parseHabitCheckinArgs(args[2:])
		if err != nil {
			return failTyped("official habit checkin", "validation", err.Error(), "", jsonOut, stdout, stderr)
		}
		payload["habitId"] = habitID
		result, err := client.CallTool(ctx, "upsert_habit_checkins", map[string]any{
			"checkins": []any{payload},
		})
		if err != nil {
			return failTyped("official habit checkin", "api", err.Error(), "", jsonOut, stdout, stderr)
		}
		if jsonOut {
			return writeJSON(stdout, envelope{OK: true, Command: "official habit checkin", Data: result})
		}
		return writeJSON(stdout, result)

	default:
		return fail("official habit", fmt.Sprintf("unknown habit subcommand %q", args[0]), jsonOut, stdout, stderr)
	}
}

func printOfficialHabitHelp(stdout io.Writer) {
	fmt.Fprintln(stdout, "Usage:")
	fmt.Fprintln(stdout, "  dida official habit get <habit-id> [--json]")
	fmt.Fprintln(stdout, "  dida official habit create [--args-json <json>] [--json]")
	fmt.Fprintln(stdout, "  dida official habit update <habit-id> [--args-json <json>] [--json]")
	fmt.Fprintln(stdout, "  dida official habit checkin <habit-id> --date YYYY-MM-DD --value <number> [--json]")
	fmt.Fprintln(stdout, "")
	fmt.Fprintln(stdout, "Manage habits using the official MCP API.")
}

func parseHabitCreateArgs(args []string) (map[string]any, error) {
	return parseOfficialPayloadFlags(args)
}

func parseHabitCheckinArgs(args []string) (map[string]any, error) {
	payload := map[string]any{}
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--date":
			if i+1 >= len(args) {
				return nil, fmt.Errorf("--date requires a value in YYYY-MM-DD format")
			}
			payload["date"] = args[i+1]
			i++
		case "--value":
			if i+1 >= len(args) {
				return nil, fmt.Errorf("--value requires a numeric value")
			}
			var value float64
			if _, err := fmt.Sscanf(args[i+1], "%f", &value); err != nil {
				return nil, fmt.Errorf("--value must be a number")
			}
			payload["value"] = value
			i++
		default:
			return nil, fmt.Errorf("unknown flag %q", args[i])
		}
	}
	if payload["date"] == nil {
		return nil, fmt.Errorf("missing required --date flag")
	}
	if payload["value"] == nil {
		return nil, fmt.Errorf("missing required --value flag")
	}
	return payload, nil
}

func runOfficialFocus(args []string, jsonOut bool, stdout io.Writer, stderr io.Writer) int {
	if len(args) == 0 || args[0] == "-h" || args[0] == "--help" {
		printOfficialFocusHelp(stdout)
		return 0
	}

	token, err := officialmcp.ResolveToken("")
	if err != nil {
		return failTyped("official focus", "auth", err.Error(), "set DIDA365_TOKEN to the official dida365 API token", jsonOut, stdout, stderr)
	}
	client := officialmcp.NewClient(token)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	if err := client.Initialize(ctx, "dida-cli", "0.1.0"); err != nil {
		return failTyped("official focus", "api", err.Error(), "", jsonOut, stdout, stderr)
	}

	switch args[0] {
	case "get":
		if len(args) < 2 {
			return failTyped("official focus get", "validation", "usage: dida official focus get <focus-id>", "run: dida official focus --help", jsonOut, stdout, stderr)
		}
		focusID := args[1]
		result, err := client.CallTool(ctx, "get_focus", map[string]any{"focusId": focusID})
		if err != nil {
			return failTyped("official focus get", "api", err.Error(), "", jsonOut, stdout, stderr)
		}
		if jsonOut {
			return writeJSON(stdout, envelope{OK: true, Command: "official focus get", Data: result})
		}
		return writeJSON(stdout, result)

	case "list":
		startTime, endTime, err := parseFocusListArgs(args[1:])
		if err != nil {
			return failTyped("official focus list", "validation", err.Error(), "run: dida official focus --help", jsonOut, stdout, stderr)
		}
		payload := map[string]any{}
		if startTime != "" {
			payload["startTime"] = startTime
		}
		if endTime != "" {
			payload["endTime"] = endTime
		}
		result, err := client.CallTool(ctx, "get_focuses_by_time", payload)
		if err != nil {
			return failTyped("official focus list", "api", err.Error(), "", jsonOut, stdout, stderr)
		}
		if jsonOut {
			return writeJSON(stdout, envelope{OK: true, Command: "official focus list", Data: result})
		}
		return writeJSON(stdout, result)

	case "delete":
		if len(args) < 2 {
			return failTyped("official focus delete", "validation", "usage: dida official focus delete <focus-id> --yes", "run: dida official focus --help", jsonOut, stdout, stderr)
		}
		focusID := args[1]
		confirmed, err := parseOfficialYesFlag(args[2:])
		if err != nil {
			return failTyped("official focus delete", "validation", err.Error(), "", jsonOut, stdout, stderr)
		}
		if !confirmed {
			return failTyped("official focus delete", "confirmation_required", "official focus delete requires --yes", "only delete known disposable focus records", jsonOut, stdout, stderr)
		}
		result, err := client.CallTool(ctx, "delete_focus", map[string]any{"focusId": focusID})
		if err != nil {
			return failTyped("official focus delete", "api", err.Error(), "", jsonOut, stdout, stderr)
		}
		if jsonOut {
			return writeJSON(stdout, envelope{OK: true, Command: "official focus delete", Data: result})
		}
		return writeJSON(stdout, result)

	default:
		return fail("official focus", fmt.Sprintf("unknown focus subcommand %q", args[0]), jsonOut, stdout, stderr)
	}
}

func printOfficialFocusHelp(stdout io.Writer) {
	fmt.Fprintln(stdout, "Usage:")
	fmt.Fprintln(stdout, "  dida official focus get <focus-id> [--json]")
	fmt.Fprintln(stdout, "  dida official focus list [--start-time RFC3339] [--end-time RFC3339] [--json]")
	fmt.Fprintln(stdout, "  dida official focus delete <focus-id> --yes [--json]")
	fmt.Fprintln(stdout, "")
	fmt.Fprintln(stdout, "Manage focus sessions using the official MCP API.")
}

func parseOfficialYesFlag(args []string) (bool, error) {
	confirmed := false
	for _, arg := range args {
		switch arg {
		case "--yes":
			confirmed = true
		default:
			return false, fmt.Errorf("unknown flag %q", arg)
		}
	}
	return confirmed, nil
}

func parseFocusListArgs(args []string) (string, string, error) {
	startTime := ""
	endTime := ""
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--start-time":
			if i+1 >= len(args) {
				return "", "", fmt.Errorf("--start-time requires an RFC3339 timestamp value")
			}
			startTime = args[i+1]
			i++
		case "--end-time":
			if i+1 >= len(args) {
				return "", "", fmt.Errorf("--end-time requires an RFC3339 timestamp value")
			}
			endTime = args[i+1]
			i++
		default:
			return "", "", fmt.Errorf("unknown flag %q", args[i])
		}
	}
	return startTime, endTime, nil
}
