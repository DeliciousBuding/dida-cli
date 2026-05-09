package cli

import (
	"context"
	"fmt"
	"io"

	"github.com/DeliciousBuding/dida-cli/internal/webapi"
)

func runColumn(args []string, jsonOut bool, stdout io.Writer, stderr io.Writer) int {
	if len(args) == 0 || args[0] == "-h" || args[0] == "--help" {
		printColumnHelp(stdout)
		return 0
	}
	switch args[0] {
	case "list":
		if len(args) != 2 {
			return failTyped("column list", "validation", "usage: dida column list <project-id>", "run: dida column --help", jsonOut, stdout, stderr)
		}
		return runProjectColumns(args[1], jsonOut, stdout, stderr)
	case "create":
		return runColumnCreate(args[1:], jsonOut, stdout, stderr)
	default:
		return fail("column", fmt.Sprintf("unknown column command %q", args[0]), jsonOut, stdout, stderr)
	}
}

func runColumnCreate(args []string, jsonOut bool, stdout io.Writer, stderr io.Writer) int {
	opts, err := parseColumnCreateFlags(args)
	if err != nil {
		return failTyped("column create", "validation", err.Error(), "run: dida column --help", jsonOut, stdout, stderr)
	}
	payload := map[string]string{"projectId": opts.ProjectID, "name": opts.Name}
	if opts.DryRun {
		return writeMutationPreview("column create", payload, opts.Yes, jsonOut, stdout, stderr)
	}
	result, err := executeMutation(func(ctx context.Context, client *webapi.Client) (map[string]any, error) {
		return client.CreateColumn(ctx, opts.ProjectID, opts.Name)
	})
	if err != nil {
		return failTyped("column create", "api", err.Error(), "", jsonOut, stdout, stderr)
	}
	return writeMutationResult("column create", "Column created", map[string]any{"projectId": opts.ProjectID, "name": opts.Name, "result": result}, opts.Name, jsonOut, stdout)
}
