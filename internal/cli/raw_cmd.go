package cli

import (
	"context"
	"io"

	"github.com/DeliciousBuding/dida-cli/internal/webapi"
)

func runRaw(args []string, jsonOut bool, stdout io.Writer, stderr io.Writer) int {
	if len(args) == 0 || args[0] == "-h" || args[0] == "--help" {
		printRawHelp(stdout)
		return 0
	}
	if args[0] != "get" || len(args) != 2 {
		return fail("raw", "usage: dida raw get <path>", jsonOut, stdout, stderr)
	}
	var data any
	_, err := executeRead(func(ctx context.Context, client *webapi.Client) (any, error) {
		return nil, client.Do(ctx, "GET", args[1], nil, &data)
	})
	if err != nil {
		return fail("raw get", err.Error(), jsonOut, stdout, stderr)
	}
	if jsonOut {
		return writeJSON(stdout, envelope{OK: true, Command: "raw get", Data: data})
	}
	return writeJSON(stdout, data)
}
