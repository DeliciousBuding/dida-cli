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
	payload := map[string]any{}
	for i := 1; i < len(args); i++ {
		switch args[i] {
		case "--args-json":
			if i+1 >= len(args) {
				return "", nil, fmt.Errorf("--args-json requires a value")
			}
			if err := json.Unmarshal([]byte(args[i+1]), &payload); err != nil {
				return "", nil, fmt.Errorf("decode --args-json: %w", err)
			}
			i++
		case "--args-file":
			if i+1 >= len(args) {
				return "", nil, fmt.Errorf("--args-file requires a path")
			}
			data, err := os.ReadFile(args[i+1])
			if err != nil {
				return "", nil, fmt.Errorf("read args file: %w", err)
			}
			if err := json.Unmarshal(data, &payload); err != nil {
				return "", nil, fmt.Errorf("decode args file: %w", err)
			}
			i++
		default:
			return "", nil, fmt.Errorf("unknown flag %q", args[i])
		}
	}
	return tool, payload, nil
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
