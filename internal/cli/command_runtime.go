package cli

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/DeliciousBuding/dida-cli/internal/auth"
	"github.com/DeliciousBuding/dida-cli/internal/webapi"
)

func executeMutation(fn func(context.Context, *webapi.Client) (map[string]any, error)) (map[string]any, error) {
	token, err := auth.LoadCookieToken()
	if err != nil {
		return nil, fmt.Errorf("missing cookie auth; run: dida auth login --browser --json")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	return fn(ctx, webapi.NewClient(token.Token))
}

func executeRead(fn func(context.Context, *webapi.Client) (any, error)) (any, error) {
	token, err := auth.LoadCookieToken()
	if err != nil {
		return nil, fmt.Errorf("missing cookie auth; run: dida auth login --browser --json")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	return fn(ctx, webapi.NewClient(token.Token))
}

func writeMutationResult(command string, label string, data map[string]any, id string, jsonOut bool, stdout io.Writer) int {
	if jsonOut {
		return writeJSON(stdout, envelope{OK: true, Command: command, Data: data})
	}
	fmt.Fprintf(stdout, "%s: %s\n", label, id)
	return 0
}
