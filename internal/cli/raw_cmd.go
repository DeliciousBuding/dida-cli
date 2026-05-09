package cli

import (
	"context"
	"io"
	"time"

	"github.com/DeliciousBuding/dida-cli/internal/auth"
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
	token, err := auth.LoadCookieToken()
	if err != nil {
		return missingAuth("raw get", jsonOut, stdout, stderr)
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
