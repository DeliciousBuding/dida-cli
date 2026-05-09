package cli

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/DeliciousBuding/DidaCLI/internal/config"
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
	case "auth", "sync", "project", "task", "report", "raw":
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
