package cli

import (
	"context"
	"io"

	"github.com/DeliciousBuding/dida-cli/internal/webapi"
)

func runTaskMove(args []string, jsonOut bool, stdout io.Writer, stderr io.Writer) int {
	opts, err := parseTaskMoveFlags(args)
	if err != nil {
		return failTyped("task move", "validation", err.Error(), "run: dida task --help", jsonOut, stdout, stderr)
	}
	payload := []webapi.TaskMovePayload{{TaskID: opts.TaskID, FromProjectID: opts.FromProjectID, ToProjectID: opts.ToProjectID}}
	if opts.DryRun {
		return writeMutationPreview("task move", payload, false, jsonOut, stdout, stderr)
	}
	result, err := executeMutation(func(ctx context.Context, client *webapi.Client) (map[string]any, error) {
		return client.MoveTask(ctx, opts.TaskID, opts.FromProjectID, opts.ToProjectID)
	})
	if err != nil {
		return failTyped("task move", "api", err.Error(), "", jsonOut, stdout, stderr)
	}
	return writeMutationResult("task move", "Task moved", map[string]any{"taskId": opts.TaskID, "fromProjectId": opts.FromProjectID, "toProjectId": opts.ToProjectID, "result": result}, opts.TaskID, jsonOut, stdout)
}

func runTaskParent(args []string, jsonOut bool, stdout io.Writer, stderr io.Writer) int {
	opts, err := parseTaskParentFlags(args)
	if err != nil {
		return failTyped("task parent", "validation", err.Error(), "run: dida task --help", jsonOut, stdout, stderr)
	}
	payload := []webapi.TaskParentPayload{{TaskID: opts.TaskID, ParentID: opts.ParentID, ProjectID: opts.ProjectID}}
	if opts.DryRun {
		return writeMutationPreview("task parent", payload, false, jsonOut, stdout, stderr)
	}
	result, err := executeMutation(func(ctx context.Context, client *webapi.Client) (map[string]any, error) {
		return client.SetTaskParent(ctx, opts.TaskID, opts.ParentID, opts.ProjectID)
	})
	if err != nil {
		return failTyped("task parent", "api", err.Error(), "", jsonOut, stdout, stderr)
	}
	return writeMutationResult("task parent", "Task parent set", map[string]any{"taskId": opts.TaskID, "parentId": opts.ParentID, "projectId": opts.ProjectID, "result": result}, opts.TaskID, jsonOut, stdout)
}
