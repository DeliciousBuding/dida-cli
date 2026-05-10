package cli

import (
	"context"
	"fmt"
	"io"

	"github.com/DeliciousBuding/dida-cli/internal/webapi"
)

func runFolder(args []string, jsonOut bool, stdout io.Writer, stderr io.Writer) int {
	if len(args) == 0 || hasHelpFlag(args) {
		printFolderHelp(stdout)
		return 0
	}
	switch args[0] {
	case "list":
		return runFolderList(jsonOut, stdout, stderr)
	case "create":
		return runFolderCreate(args[1:], jsonOut, stdout, stderr)
	case "update":
		return runFolderUpdate(args[1:], jsonOut, stdout, stderr)
	case "delete":
		return runFolderDelete(args[1:], jsonOut, stdout, stderr)
	default:
		return fail("folder", fmt.Sprintf("unknown folder command %q", args[0]), jsonOut, stdout, stderr)
	}
}

func runFolderList(jsonOut bool, stdout io.Writer, stderr io.Writer) int {
	view, err := loadSyncView()
	if err != nil {
		return failTyped("folder list", "auth", err.Error(), "run: dida auth login", jsonOut, stdout, stderr)
	}
	data := map[string]any{"folders": view.ProjectGroups}
	meta := map[string]any{"count": len(view.ProjectGroups)}
	if jsonOut {
		return writeJSON(stdout, envelope{OK: true, Command: "folder list", Meta: meta, Data: data})
	}
	printMapList(stdout, view.ProjectGroups, "folders")
	return 0
}

func runFolderCreate(args []string, jsonOut bool, stdout io.Writer, stderr io.Writer) int {
	opts, err := parseNamedCreateFlags(args)
	if err != nil {
		return failTyped("folder create", "validation", err.Error(), "run: dida folder --help", jsonOut, stdout, stderr)
	}
	group := webapi.ProjectGroupMutation{ID: opts.ID, Name: opts.Name}
	if group.ID == "" {
		group.ID = webapi.NewTaskID()
	}
	payload := map[string]any{"add": []webapi.ProjectGroupMutation{group}}
	if opts.DryRun {
		return writeMutationPreview("folder create", payload, opts.Yes, jsonOut, stdout, stderr)
	}
	result, err := executeMutation(func(ctx context.Context, client *webapi.Client) (map[string]any, error) {
		return client.CreateProjectGroup(ctx, group)
	})
	if err != nil {
		return failTyped("folder create", "api", err.Error(), "", jsonOut, stdout, stderr)
	}
	return writeMutationResult("folder create", "Folder created", map[string]any{"folderId": group.ID, "result": result}, group.ID, jsonOut, stdout)
}

func runFolderUpdate(args []string, jsonOut bool, stdout io.Writer, stderr io.Writer) int {
	opts, err := parseNamedUpdateFlags(args, "folder update")
	if err != nil {
		return failTyped("folder update", "validation", err.Error(), "run: dida folder --help", jsonOut, stdout, stderr)
	}
	group := webapi.ProjectGroupMutation{ID: opts.ID, Name: opts.Name}
	payload := map[string]any{"update": []webapi.ProjectGroupMutation{group}}
	if opts.DryRun {
		return writeMutationPreview("folder update", payload, opts.Yes, jsonOut, stdout, stderr)
	}
	result, err := executeMutation(func(ctx context.Context, client *webapi.Client) (map[string]any, error) {
		return client.UpdateProjectGroup(ctx, group)
	})
	if err != nil {
		return failTyped("folder update", "api", err.Error(), "", jsonOut, stdout, stderr)
	}
	return writeMutationResult("folder update", "Folder updated", map[string]any{"folderId": group.ID, "result": result}, group.ID, jsonOut, stdout)
}

func runFolderDelete(args []string, jsonOut bool, stdout io.Writer, stderr io.Writer) int {
	opts, err := parseDeleteIDFlags(args, "folder delete")
	if err != nil {
		return failTyped("folder delete", "validation", err.Error(), "run: dida folder --help", jsonOut, stdout, stderr)
	}
	payload := map[string]any{"delete": []string{opts.ID}}
	if opts.DryRun {
		return writeMutationPreview("folder delete", payload, opts.Yes, jsonOut, stdout, stderr)
	}
	if !opts.Yes {
		return failTyped("folder delete", "confirmation_required", "folder delete requires --yes", "preview first with: dida folder delete <folder-id> --dry-run", jsonOut, stdout, stderr)
	}
	result, err := executeMutation(func(ctx context.Context, client *webapi.Client) (map[string]any, error) {
		return client.DeleteProjectGroup(ctx, opts.ID)
	})
	if err != nil {
		return failTyped("folder delete", "api", err.Error(), "", jsonOut, stdout, stderr)
	}
	return writeMutationResult("folder delete", "Folder deleted", map[string]any{"folderId": opts.ID, "result": result}, opts.ID, jsonOut, stdout)
}
