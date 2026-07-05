package cli

import (
	"context"
	"fmt"
	"io"

	"github.com/DeliciousBuding/dida-cli/internal/webapi"
)

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
