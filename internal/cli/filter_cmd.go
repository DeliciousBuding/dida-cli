package cli

import (
	"fmt"
	"io"
)

func runFilter(args []string, jsonOut bool, stdout io.Writer, stderr io.Writer) int {
	if len(args) == 0 || args[0] == "-h" || args[0] == "--help" {
		printFilterHelp(stdout)
		return 0
	}
	switch args[0] {
	case "list":
		return runFilterList(jsonOut, stdout, stderr)
	default:
		return fail("filter", fmt.Sprintf("unknown filter command %q", args[0]), jsonOut, stdout, stderr)
	}
}

func runFilterList(jsonOut bool, stdout io.Writer, stderr io.Writer) int {
	view, err := loadSyncView()
	if err != nil {
		return failTyped("filter list", "auth", err.Error(), "run: dida auth login --browser --json", jsonOut, stdout, stderr)
	}
	filters := view.Filters
	if filters == nil {
		filters = []map[string]any{}
	}
	data := map[string]any{"filters": filters}
	meta := map[string]any{"count": len(filters), "source": "sync_filters"}
	if jsonOut {
		return writeJSON(stdout, envelope{OK: true, Command: "filter list", Meta: meta, Data: data})
	}
	printMapList(stdout, view.Filters, "filters")
	return 0
}
