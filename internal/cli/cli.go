package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/DeliciousBuding/dida-cli/internal/auth"
	"github.com/DeliciousBuding/dida-cli/internal/config"
	"github.com/DeliciousBuding/dida-cli/internal/webapi"
)

type envelope struct {
	OK      bool      `json:"ok"`
	Command string    `json:"command"`
	Data    any       `json:"data,omitempty"`
	Error   *cliError `json:"error,omitempty"`
}

type cliError struct {
	Message string `json:"message"`
}

func Run(args []string, version string, stdout io.Writer, stderr io.Writer) int {
	if len(args) == 0 || args[0] == "-h" || args[0] == "--help" {
		printHelp(stdout)
		return 0
	}
	if args[0] == "--version" || args[0] == "version" {
		fmt.Fprintln(stdout, version)
		return 0
	}

	jsonOut, args := consumeJSONFlag(args)
	command := args[0]

	switch command {
	case "doctor":
		return runDoctor(args[1:], version, jsonOut, stdout, stderr)
	case "auth":
		return runAuth(args[1:], jsonOut, stdout, stderr)
	case "sync":
		return runSync(args[1:], jsonOut, stdout, stderr)
	case "raw":
		return runRaw(args[1:], jsonOut, stdout, stderr)
	case "project", "task", "report":
		return notImplemented(command, jsonOut, stdout, stderr)
	default:
		return fail(command, fmt.Sprintf("unknown command %q", command), jsonOut, stdout, stderr)
	}
}

func consumeJSONFlag(args []string) (bool, []string) {
	out := args[:0]
	jsonOut := false
	for _, arg := range args {
		if arg == "--json" || arg == "-j" {
			jsonOut = true
			continue
		}
		out = append(out, arg)
	}
	return jsonOut, out
}

func runDoctor(args []string, version string, jsonOut bool, stdout io.Writer, stderr io.Writer) int {
	if len(args) > 0 && (args[0] == "-h" || args[0] == "--help") {
		fmt.Fprintln(stdout, "Usage: dida doctor [--json]")
		return 0
	}

	cfgDir := config.DefaultDir()
	cookiePath := filepath.Join(cfgDir, "cookie.json")
	oauthPath := filepath.Join(cfgDir, "oauth.json")
	cookieExists := fileExists(cookiePath)
	oauthExists := fileExists(oauthPath)

	data := map[string]any{
		"version":       version,
		"goos":          runtime.GOOS,
		"goarch":        runtime.GOARCH,
		"config_dir":    cfgDir,
		"auth_sources":  map[string]bool{"cookie": cookieExists, "oauth": oauthExists},
		"cookie_status": auth.CookieStatus(),
		"network_check": "not_run",
	}

	if jsonOut {
		return writeJSON(stdout, envelope{OK: true, Command: "doctor", Data: data})
	}

	fmt.Fprintf(stdout, "DidaCLI %s\n", version)
	fmt.Fprintf(stdout, "Config: %s\n", cfgDir)
	fmt.Fprintf(stdout, "Cookie auth: %s\n", yesNo(cookieExists))
	fmt.Fprintf(stdout, "OAuth auth: %s\n", yesNo(oauthExists))
	fmt.Fprintln(stdout, "Network check: not run")
	return 0
}

func runAuth(args []string, jsonOut bool, stdout io.Writer, stderr io.Writer) int {
	if len(args) == 0 || args[0] == "-h" || args[0] == "--help" {
		printAuthHelp(stdout)
		return 0
	}
	switch args[0] {
	case "status":
		data := map[string]any{"cookie": auth.CookieStatus(), "oauth": map[string]any{"available": false, "message": "not implemented"}}
		if jsonOut {
			return writeJSON(stdout, envelope{OK: true, Command: "auth status", Data: data})
		}
		cookie := data["cookie"].(map[string]any)
		fmt.Fprintf(stdout, "Cookie auth: %v\n", cookie["available"])
		fmt.Fprintf(stdout, "Cookie path: %v\n", cookie["path"])
		if cookie["available"] == true {
			fmt.Fprintf(stdout, "Token: %v\n", cookie["token_preview"])
			fmt.Fprintf(stdout, "Saved at: %v\n", cookie["saved_at"])
		}
		return 0
	case "cookie":
		return runAuthCookie(args[1:], jsonOut, stdout, stderr)
	default:
		return fail("auth", fmt.Sprintf("unknown auth command %q", args[0]), jsonOut, stdout, stderr)
	}
}

func runAuthCookie(args []string, jsonOut bool, stdout io.Writer, stderr io.Writer) int {
	if len(args) == 0 || args[0] == "-h" || args[0] == "--help" {
		printAuthCookieHelp(stdout)
		return 0
	}
	if args[0] != "set" {
		return fail("auth cookie", fmt.Sprintf("unknown auth cookie command %q", args[0]), jsonOut, stdout, stderr)
	}
	token, err := parseTokenInput(args[1:])
	if err != nil {
		return fail("auth cookie set", err.Error(), jsonOut, stdout, stderr)
	}
	item, err := auth.SaveCookieToken(token)
	if err != nil {
		return fail("auth cookie set", err.Error(), jsonOut, stdout, stderr)
	}
	data := map[string]any{
		"path":          auth.CookiePath(),
		"saved_at":      time.UnixMilli(item.SavedAt).Format(time.RFC3339),
		"token_length":  len(item.Token),
		"token_preview": auth.RedactToken(item.Token),
	}
	if jsonOut {
		return writeJSON(stdout, envelope{OK: true, Command: "auth cookie set", Data: data})
	}
	fmt.Fprintf(stdout, "Cookie token saved: %s\n", auth.CookiePath())
	fmt.Fprintf(stdout, "Token: %s\n", auth.RedactToken(item.Token))
	return 0
}

func parseTokenInput(args []string) (string, error) {
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--token":
			if i+1 >= len(args) {
				return "", fmt.Errorf("--token requires a value")
			}
			return args[i+1], nil
		case "--token-stdin":
			data, err := io.ReadAll(os.Stdin)
			if err != nil {
				return "", fmt.Errorf("read token from stdin: %w", err)
			}
			return string(data), nil
		}
	}
	return "", fmt.Errorf("missing token; use --token-stdin to avoid shell history")
}

func runSync(args []string, jsonOut bool, stdout io.Writer, stderr io.Writer) int {
	if len(args) == 0 || args[0] == "-h" || args[0] == "--help" {
		printSyncHelp(stdout)
		return 0
	}
	if args[0] != "all" {
		return fail("sync", fmt.Sprintf("unknown sync command %q", args[0]), jsonOut, stdout, stderr)
	}
	token, err := auth.LoadCookieToken()
	if err != nil {
		return fail("sync all", "missing cookie auth; run: dida auth cookie set --token-stdin", jsonOut, stdout, stderr)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	payload, err := webapi.NewClient(token.Token).FullSync(ctx)
	if err != nil {
		return fail("sync all", err.Error(), jsonOut, stdout, stderr)
	}
	data := map[string]any{
		"inboxId":       payload.InboxID,
		"tasks":         payload.Tasks,
		"projects":      payload.Projects,
		"projectGroups": payload.ProjectGroups,
		"tags":          payload.Tags,
		"counts": map[string]int{
			"tasks":         len(payload.Tasks),
			"projects":      len(payload.Projects),
			"projectGroups": len(payload.ProjectGroups),
			"tags":          len(payload.Tags),
		},
	}
	if jsonOut {
		return writeJSON(stdout, envelope{OK: true, Command: "sync all", Data: data})
	}
	counts := data["counts"].(map[string]int)
	fmt.Fprintln(stdout, "Sync complete")
	fmt.Fprintf(stdout, "Tasks: %d\n", counts["tasks"])
	fmt.Fprintf(stdout, "Projects: %d\n", counts["projects"])
	fmt.Fprintf(stdout, "Project groups: %d\n", counts["projectGroups"])
	fmt.Fprintf(stdout, "Tags: %d\n", counts["tags"])
	return 0
}

func runRaw(args []string, jsonOut bool, stdout io.Writer, stderr io.Writer) int {
	if len(args) == 0 || args[0] == "-h" || args[0] == "--help" {
		printRawHelp(stdout)
		return 0
	}
	if args[0] != "get" || len(args) != 2 {
		return fail("raw", "usage: dida raw get <path>", jsonOut, stdout, stderr)
	}
	token, err := auth.LoadCookieToken()
	if err != nil {
		return fail("raw get", "missing cookie auth; run: dida auth cookie set --token-stdin", jsonOut, stdout, stderr)
	}
	var data any
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	if err := webapi.NewClient(token.Token).Do(ctx, "GET", args[1], nil, &data); err != nil {
		return fail("raw get", err.Error(), jsonOut, stdout, stderr)
	}
	if jsonOut {
		return writeJSON(stdout, envelope{OK: true, Command: "raw get", Data: data})
	}
	return writeJSON(stdout, data)
}

func notImplemented(command string, jsonOut bool, stdout io.Writer, stderr io.Writer) int {
	return fail(command, "command scaffolded but not implemented yet", jsonOut, stdout, stderr)
}

func fail(command string, message string, jsonOut bool, stdout io.Writer, stderr io.Writer) int {
	if jsonOut {
		_ = writeJSON(stdout, envelope{OK: false, Command: command, Error: &cliError{Message: message}})
		return 1
	}
	fmt.Fprintf(stderr, "dida: %s\n", message)
	return 1
}

func writeJSON(w io.Writer, value any) int {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	if err := enc.Encode(value); err != nil {
		return 1
	}
	return 0
}

func fileExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && !info.IsDir()
}

func yesNo(v bool) string {
	if v {
		return "yes"
	}
	return "no"
}

func printHelp(w io.Writer) {
	fmt.Fprintln(w, strings.TrimSpace(`
DidaCLI - Dida365 / TickTick command line client

Usage:
  dida <command> [options]

Commands:
  doctor       Check local config and auth status
  auth         Manage OAuth and cookie auth
  sync         Sync tasks/projects/tags
  project      Project discovery
  task         Task reads and writes
  report       Generate reports
  raw          Raw read-only API escape hatch
  version      Print version

Global options:
  -j, --json   Emit machine-readable JSON
  -h, --help   Show help
`))
}

func printAuthHelp(w io.Writer) {
	fmt.Fprintln(w, strings.TrimSpace(`
Usage:
  dida auth status [--json]
  dida auth cookie set --token-stdin
  dida auth cookie set --token <token>
`))
}

func printAuthCookieHelp(w io.Writer) {
	fmt.Fprintln(w, strings.TrimSpace(`
Usage:
  dida auth cookie set --token-stdin
  dida auth cookie set --token <token>

Prefer --token-stdin to avoid shell history.
`))
}

func printSyncHelp(w io.Writer) {
	fmt.Fprintln(w, strings.TrimSpace(`
Usage:
  dida sync all [--json]
`))
}

func printRawHelp(w io.Writer) {
	fmt.Fprintln(w, strings.TrimSpace(`
Usage:
  dida raw get <path> [--json]

Only GET is supported for raw calls.
`))
}
