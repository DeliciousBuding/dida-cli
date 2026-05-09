package cli

import (
	"context"
	"errors"
	"fmt"
	"io"

	"github.com/DeliciousBuding/dida-cli/internal/webapi"
)

func runRaw(args []string, jsonOut bool, stdout io.Writer, stderr io.Writer) int {
	if len(args) == 0 || args[0] == "-h" || args[0] == "--help" {
		printRawHelp(stdout)
		return 0
	}
	if args[0] != "get" {
		return fail("raw", "usage: dida raw get <path> [--api-version v1|v2]", jsonOut, stdout, stderr)
	}
	path, version, err := parseRawGetArgs(args[1:])
	if err != nil {
		return failTyped("raw get", "validation", err.Error(), "run: dida raw --help", jsonOut, stdout, stderr)
	}
	var data any
	_, err = executeRead(func(ctx context.Context, client *webapi.Client) (any, error) {
		if version == "v1" {
			return nil, client.DoV1(ctx, "GET", path, nil, &data)
		}
		return nil, client.Do(ctx, "GET", path, nil, &data)
	})
	if err != nil {
		if jsonOut {
			return failRawAPIError(err, version, stdout)
		}
		return fail("raw get", err.Error(), jsonOut, stdout, stderr)
	}
	if jsonOut {
		return writeJSON(stdout, envelope{OK: true, Command: "raw get", Meta: map[string]any{"apiVersion": version}, Data: data})
	}
	return writeJSON(stdout, data)
}

func failRawAPIError(err error, version string, stdout io.Writer) int {
	var apiErr *webapi.APIError
	if errors.As(err, &apiErr) {
		details := map[string]any{
			"apiVersion":  version,
			"method":      apiErr.Method,
			"path":        apiErr.Path,
			"statusCode":  apiErr.StatusCode,
			"bodySnippet": apiErr.BodySnippet,
		}
		_ = writeJSON(stdout, envelope{
			OK:      false,
			Command: "raw get",
			Meta:    map[string]any{"apiVersion": version},
			Error:   &cliError{Type: "api", Message: apiErr.Error(), Details: details},
		})
		return 1
	}
	_ = writeJSON(stdout, envelope{
		OK:      false,
		Command: "raw get",
		Meta:    map[string]any{"apiVersion": version},
		Error:   &cliError{Type: "api", Message: err.Error()},
	})
	return 1
}

func parseRawGetArgs(args []string) (string, string, error) {
	path := ""
	version := "v2"
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--api-version":
			if i+1 >= len(args) {
				return "", "", fmt.Errorf("--api-version requires v1 or v2")
			}
			version = args[i+1]
			i++
		case "--v1":
			version = "v1"
		case "--v2":
			version = "v2"
		default:
			if path == "" {
				path = args[i]
				continue
			}
			return "", "", fmt.Errorf("unknown flag %q", args[i])
		}
	}
	if path == "" {
		return "", "", fmt.Errorf("missing path")
	}
	if version != "v1" && version != "v2" {
		return "", "", fmt.Errorf("--api-version must be v1 or v2")
	}
	return path, version, nil
}
