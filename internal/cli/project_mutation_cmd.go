package cli

import (
	"context"
	"io"

	"github.com/DeliciousBuding/dida-cli/internal/webapi"
)

func runProjectCreate(args []string, jsonOut bool, stdout io.Writer, stderr io.Writer) int {
	opts, err := parseProjectCreateFlags(args)
	if err != nil {
		return failTyped("project create", "validation", err.Error(), "run: dida project --help", jsonOut, stdout, stderr)
	}
	project := webapi.ProjectMutation{ID: opts.ID, Name: opts.Name, GroupID: opts.GroupID, ViewMode: "list", Kind: "TASK"}
	if project.ID == "" {
		project.ID = webapi.NewTaskID()
	}
	payload := map[string]any{"add": []webapi.ProjectMutation{project}}
	if opts.DryRun {
		return writeMutationPreview("project create", payload, opts.Yes, jsonOut, stdout, stderr)
	}
	result, err := executeMutation(func(ctx context.Context, client *webapi.Client) (map[string]any, error) {
		return client.CreateProject(ctx, project)
	})
	if err != nil {
		return failTyped("project create", "api", err.Error(), "", jsonOut, stdout, stderr)
	}
	return writeMutationResult("project create", "Project created", map[string]any{"projectId": project.ID, "result": result}, project.ID, jsonOut, stdout)
}

func runProjectUpdate(args []string, jsonOut bool, stdout io.Writer, stderr io.Writer) int {
	opts, err := parseProjectUpdateFlags(args)
	if err != nil {
		return failTyped("project update", "validation", err.Error(), "run: dida project --help", jsonOut, stdout, stderr)
	}
	project := webapi.ProjectMutation{ID: opts.ID, Name: opts.Name, GroupID: opts.GroupID}
	payload := map[string]any{"update": []webapi.ProjectMutation{project}}
	if opts.DryRun {
		return writeMutationPreview("project update", payload, opts.Yes, jsonOut, stdout, stderr)
	}
	result, err := executeMutation(func(ctx context.Context, client *webapi.Client) (map[string]any, error) {
		return client.UpdateProject(ctx, project)
	})
	if err != nil {
		return failTyped("project update", "api", err.Error(), "", jsonOut, stdout, stderr)
	}
	return writeMutationResult("project update", "Project updated", map[string]any{"projectId": project.ID, "result": result}, project.ID, jsonOut, stdout)
}

func runProjectDelete(args []string, jsonOut bool, stdout io.Writer, stderr io.Writer) int {
	opts, err := parseDeleteIDFlags(args, "project delete")
	if err != nil {
		return failTyped("project delete", "validation", err.Error(), "run: dida project --help", jsonOut, stdout, stderr)
	}
	payload := map[string]any{"delete": []string{opts.ID}}
	if opts.DryRun {
		return writeMutationPreview("project delete", payload, opts.Yes, jsonOut, stdout, stderr)
	}
	if !opts.Yes {
		return failTyped("project delete", "confirmation_required", "project delete requires --yes", "preview first with: dida project delete <project-id> --dry-run", jsonOut, stdout, stderr)
	}
	result, err := executeMutation(func(ctx context.Context, client *webapi.Client) (map[string]any, error) {
		return client.DeleteProject(ctx, opts.ID)
	})
	if err != nil {
		return failTyped("project delete", "api", err.Error(), "", jsonOut, stdout, stderr)
	}
	return writeMutationResult("project delete", "Project deleted", map[string]any{"projectId": opts.ID, "result": result}, opts.ID, jsonOut, stdout)
}
