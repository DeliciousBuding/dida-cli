package cli

import (
	"context"
	"fmt"
	"io"

	"github.com/DeliciousBuding/dida-cli/internal/webapi"
)

func runTag(args []string, jsonOut bool, stdout io.Writer, stderr io.Writer) int {
	if len(args) == 0 || args[0] == "-h" || args[0] == "--help" {
		printTagHelp(stdout)
		return 0
	}
	switch args[0] {
	case "list":
		return runTagList(jsonOut, stdout, stderr)
	case "create":
		return runTagCreate(args[1:], jsonOut, stdout, stderr)
	case "update":
		return runTagUpdate(args[1:], jsonOut, stdout, stderr)
	case "rename":
		return runTagRename(args[1:], jsonOut, stdout, stderr)
	case "merge":
		return runTagMerge(args[1:], jsonOut, stdout, stderr)
	case "delete":
		return runTagDelete(args[1:], jsonOut, stdout, stderr)
	default:
		return fail("tag", fmt.Sprintf("unknown tag command %q", args[0]), jsonOut, stdout, stderr)
	}
}

func runTagList(jsonOut bool, stdout io.Writer, stderr io.Writer) int {
	view, err := loadSyncView()
	if err != nil {
		return failTyped("tag list", "auth", err.Error(), "run: dida auth login", jsonOut, stdout, stderr)
	}
	data := map[string]any{"tags": view.Tags}
	meta := map[string]any{"count": len(view.Tags)}
	if jsonOut {
		return writeJSON(stdout, envelope{OK: true, Command: "tag list", Meta: meta, Data: data})
	}
	printMapList(stdout, view.Tags, "tags")
	return 0
}

func runTagCreate(args []string, jsonOut bool, stdout io.Writer, stderr io.Writer) int {
	opts, err := parseTagCreateFlags(args)
	if err != nil {
		return failTyped("tag create", "validation", err.Error(), "run: dida tag --help", jsonOut, stdout, stderr)
	}
	tag := webapi.TagMutation{Name: opts.Name, Color: opts.Color, Parent: opts.Parent, Label: opts.Label}
	payload := map[string]any{"add": []webapi.TagMutation{tag}}
	if opts.DryRun {
		return writeMutationPreview("tag create", payload, opts.Yes, jsonOut, stdout, stderr)
	}
	result, err := executeMutation(func(ctx context.Context, client *webapi.Client) (map[string]any, error) {
		return client.CreateTag(ctx, tag)
	})
	if err != nil {
		return failTyped("tag create", "api", err.Error(), "", jsonOut, stdout, stderr)
	}
	return writeMutationResult("tag create", "Tag created", map[string]any{"name": tag.Name, "result": result}, tag.Name, jsonOut, stdout)
}

func runTagUpdate(args []string, jsonOut bool, stdout io.Writer, stderr io.Writer) int {
	opts, err := parseTagUpdateFlags(args)
	if err != nil {
		return failTyped("tag update", "validation", err.Error(), "run: dida tag --help", jsonOut, stdout, stderr)
	}
	tag := webapi.TagMutation{Name: opts.Name, Color: opts.Color, Parent: opts.Parent, Label: opts.Label}
	payload := map[string]any{"update": []webapi.TagMutation{tag}}
	if opts.DryRun {
		return writeMutationPreview("tag update", payload, opts.Yes, jsonOut, stdout, stderr)
	}
	result, err := executeMutation(func(ctx context.Context, client *webapi.Client) (map[string]any, error) {
		return client.UpdateTag(ctx, tag)
	})
	if err != nil {
		return failTyped("tag update", "api", err.Error(), "", jsonOut, stdout, stderr)
	}
	return writeMutationResult("tag update", "Tag updated", map[string]any{"name": tag.Name, "result": result}, tag.Name, jsonOut, stdout)
}

func runTagRename(args []string, jsonOut bool, stdout io.Writer, stderr io.Writer) int {
	opts, err := parseTwoNameFlags(args, "tag rename")
	if err != nil {
		return failTyped("tag rename", "validation", err.Error(), "run: dida tag --help", jsonOut, stdout, stderr)
	}
	payload := map[string]string{"name": opts.ID, "newName": opts.Name}
	if opts.DryRun {
		return writeMutationPreview("tag rename", payload, opts.Yes, jsonOut, stdout, stderr)
	}
	result, err := executeMutation(func(ctx context.Context, client *webapi.Client) (map[string]any, error) {
		return client.RenameTag(ctx, opts.ID, opts.Name)
	})
	if err != nil {
		return failTyped("tag rename", "api", err.Error(), "", jsonOut, stdout, stderr)
	}
	return writeMutationResult("tag rename", "Tag renamed", map[string]any{"oldName": opts.ID, "newName": opts.Name, "result": result}, opts.Name, jsonOut, stdout)
}

func runTagMerge(args []string, jsonOut bool, stdout io.Writer, stderr io.Writer) int {
	opts, err := parseTwoNameFlags(args, "tag merge")
	if err != nil {
		return failTyped("tag merge", "validation", err.Error(), "run: dida tag --help", jsonOut, stdout, stderr)
	}
	payload := map[string]string{"from": opts.ID, "to": opts.Name}
	if opts.DryRun {
		return writeMutationPreview("tag merge", payload, opts.Yes, jsonOut, stdout, stderr)
	}
	if !opts.Yes {
		return failTyped("tag merge", "confirmation_required", "tag merge requires --yes", "preview first with: dida tag merge <from> <to> --dry-run", jsonOut, stdout, stderr)
	}
	result, err := executeMutation(func(ctx context.Context, client *webapi.Client) (map[string]any, error) {
		return client.MergeTags(ctx, opts.ID, opts.Name)
	})
	if err != nil {
		return failTyped("tag merge", "api", err.Error(), "", jsonOut, stdout, stderr)
	}
	return writeMutationResult("tag merge", "Tag merged", map[string]any{"from": opts.ID, "to": opts.Name, "result": result}, opts.Name, jsonOut, stdout)
}

func runTagDelete(args []string, jsonOut bool, stdout io.Writer, stderr io.Writer) int {
	opts, err := parseDeleteIDFlags(args, "tag delete")
	if err != nil {
		return failTyped("tag delete", "validation", err.Error(), "run: dida tag --help", jsonOut, stdout, stderr)
	}
	payload := map[string]any{"delete": []string{opts.ID}}
	if opts.DryRun {
		return writeMutationPreview("tag delete", payload, opts.Yes, jsonOut, stdout, stderr)
	}
	if !opts.Yes {
		return failTyped("tag delete", "confirmation_required", "tag delete requires --yes", "preview first with: dida tag delete <name> --dry-run", jsonOut, stdout, stderr)
	}
	result, err := executeMutation(func(ctx context.Context, client *webapi.Client) (map[string]any, error) {
		return client.DeleteTag(ctx, opts.ID)
	})
	if err != nil {
		return failTyped("tag delete", "api", err.Error(), "", jsonOut, stdout, stderr)
	}
	return writeMutationResult("tag delete", "Tag deleted", map[string]any{"name": opts.ID, "result": result}, opts.ID, jsonOut, stdout)
}
