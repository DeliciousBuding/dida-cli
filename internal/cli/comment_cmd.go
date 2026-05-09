package cli

import (
	"context"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/DeliciousBuding/dida-cli/internal/webapi"
)

type commentOptions struct {
	ProjectID string
	TaskID    string
	CommentID string
	Title     string
	DryRun    bool
	Yes       bool
}

func runComment(args []string, jsonOut bool, stdout io.Writer, stderr io.Writer) int {
	if len(args) == 0 || args[0] == "-h" || args[0] == "--help" {
		printCommentHelp(stdout)
		return 0
	}
	switch args[0] {
	case "list":
		return runCommentList(args[1:], jsonOut, stdout, stderr)
	case "create":
		return runCommentCreate(args[1:], jsonOut, stdout, stderr)
	case "update":
		return runCommentUpdate(args[1:], jsonOut, stdout, stderr)
	case "delete":
		return runCommentDelete(args[1:], jsonOut, stdout, stderr)
	default:
		return fail("comment", fmt.Sprintf("unknown comment command %q", args[0]), jsonOut, stdout, stderr)
	}
}

func runCommentList(args []string, jsonOut bool, stdout io.Writer, stderr io.Writer) int {
	opts, err := parseCommentTargetFlags(args, "comment list", false)
	if err != nil {
		return failTyped("comment list", "validation", err.Error(), "run: dida comment --help", jsonOut, stdout, stderr)
	}
	var comments []map[string]any
	result, err := executeRead(func(ctx context.Context, client *webapi.Client) (any, error) {
		out, err := client.TaskComments(ctx, opts.ProjectID, opts.TaskID)
		comments = out
		return map[string]any{"comments": out}, err
	})
	if err != nil {
		return failTyped("comment list", "api", err.Error(), "", jsonOut, stdout, stderr)
	}
	meta := map[string]any{"count": len(comments)}
	if jsonOut {
		return writeJSON(stdout, envelope{OK: true, Command: "comment list", Meta: meta, Data: result})
	}
	printMapList(stdout, comments, "comments")
	return 0
}

func runCommentCreate(args []string, jsonOut bool, stdout io.Writer, stderr io.Writer) int {
	opts, err := parseCommentWriteFlags(args, "comment create", false)
	if err != nil {
		return failTyped("comment create", "validation", err.Error(), "run: dida comment --help", jsonOut, stdout, stderr)
	}
	comment := buildCommentCreatePayload(opts)
	payload := map[string]any{"projectId": opts.ProjectID, "taskId": opts.TaskID, "comment": comment}
	if opts.DryRun {
		return writeMutationPreview("comment create", payload, opts.Yes, jsonOut, stdout, stderr)
	}
	result, err := executeMutation(func(ctx context.Context, client *webapi.Client) (map[string]any, error) {
		return client.CreateComment(ctx, opts.ProjectID, opts.TaskID, comment)
	})
	if err != nil {
		return failTyped("comment create", "api", err.Error(), "", jsonOut, stdout, stderr)
	}
	return writeMutationResult("comment create", "Comment created", map[string]any{"projectId": opts.ProjectID, "taskId": opts.TaskID, "commentId": comment.ID, "result": result}, comment.ID, jsonOut, stdout)
}

func runCommentUpdate(args []string, jsonOut bool, stdout io.Writer, stderr io.Writer) int {
	opts, err := parseCommentWriteFlags(args, "comment update", true)
	if err != nil {
		return failTyped("comment update", "validation", err.Error(), "run: dida comment --help", jsonOut, stdout, stderr)
	}
	comment := webapi.CommentMutation{Title: opts.Title}
	payload := map[string]any{"projectId": opts.ProjectID, "taskId": opts.TaskID, "commentId": opts.CommentID, "comment": comment}
	if opts.DryRun {
		return writeMutationPreview("comment update", payload, opts.Yes, jsonOut, stdout, stderr)
	}
	result, err := executeMutation(func(ctx context.Context, client *webapi.Client) (map[string]any, error) {
		return client.UpdateComment(ctx, opts.ProjectID, opts.TaskID, opts.CommentID, comment)
	})
	if err != nil {
		return failTyped("comment update", "api", err.Error(), "", jsonOut, stdout, stderr)
	}
	return writeMutationResult("comment update", "Comment updated", map[string]any{"projectId": opts.ProjectID, "taskId": opts.TaskID, "commentId": opts.CommentID, "result": result}, opts.CommentID, jsonOut, stdout)
}

func runCommentDelete(args []string, jsonOut bool, stdout io.Writer, stderr io.Writer) int {
	opts, err := parseCommentTargetFlags(args, "comment delete", true)
	if err != nil {
		return failTyped("comment delete", "validation", err.Error(), "run: dida comment --help", jsonOut, stdout, stderr)
	}
	payload := map[string]any{"projectId": opts.ProjectID, "taskId": opts.TaskID, "commentId": opts.CommentID}
	if opts.DryRun {
		return writeMutationPreview("comment delete", payload, opts.Yes, jsonOut, stdout, stderr)
	}
	if !opts.Yes {
		return failTyped("comment delete", "confirmation_required", "comment delete requires --yes", "preview first with: dida comment delete --project <project-id> --task <task-id> --comment <comment-id> --dry-run", jsonOut, stdout, stderr)
	}
	result, err := executeMutation(func(ctx context.Context, client *webapi.Client) (map[string]any, error) {
		return client.DeleteComment(ctx, opts.ProjectID, opts.TaskID, opts.CommentID)
	})
	if err != nil {
		return failTyped("comment delete", "api", err.Error(), "", jsonOut, stdout, stderr)
	}
	return writeMutationResult("comment delete", "Comment deleted", map[string]any{"projectId": opts.ProjectID, "taskId": opts.TaskID, "commentId": opts.CommentID, "result": result}, opts.CommentID, jsonOut, stdout)
}

func parseCommentWriteFlags(args []string, command string, requireComment bool) (commentOptions, error) {
	opts, err := parseCommentTargetFlags(args, command, requireComment)
	if err != nil {
		return opts, err
	}
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--title", "--text", "-t":
			if i+1 >= len(args) {
				return opts, fmt.Errorf("%s requires text", args[i])
			}
			opts.Title = args[i+1]
			i++
		}
	}
	if strings.TrimSpace(opts.Title) == "" {
		return opts, fmt.Errorf("missing comment text; use --text <text>")
	}
	return opts, nil
}

func buildCommentCreatePayload(opts commentOptions) webapi.CommentMutation {
	return webapi.CommentMutation{
		ID:          webapi.NewTaskID(),
		CreatedTime: time.Now().UTC().Format("2006-01-02T15:04:05.000+0000"),
		TaskID:      opts.TaskID,
		ProjectID:   opts.ProjectID,
		Title:       opts.Title,
		UserProfile: map[string]any{"isMyself": true},
		Mentions:    []map[string]any{},
		IsNew:       true,
	}
}

func parseCommentTargetFlags(args []string, command string, requireComment bool) (commentOptions, error) {
	opts := commentOptions{}
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--project", "-p":
			if i+1 >= len(args) {
				return opts, fmt.Errorf("%s requires a project id", args[i])
			}
			opts.ProjectID = args[i+1]
			i++
		case "--task":
			if i+1 >= len(args) {
				return opts, fmt.Errorf("--task requires a task id")
			}
			opts.TaskID = args[i+1]
			i++
		case "--comment", "--id":
			if i+1 >= len(args) {
				return opts, fmt.Errorf("%s requires a comment id", args[i])
			}
			opts.CommentID = args[i+1]
			i++
		case "--title", "--text", "-t":
			if i+1 >= len(args) {
				return opts, fmt.Errorf("%s requires text", args[i])
			}
			i++
		case "--dry-run":
			opts.DryRun = true
		case "--yes":
			opts.Yes = true
		default:
			return opts, fmt.Errorf("unknown flag %q", args[i])
		}
	}
	if strings.TrimSpace(opts.ProjectID) == "" {
		return opts, fmt.Errorf("missing project id; use --project <project-id>")
	}
	if strings.TrimSpace(opts.TaskID) == "" {
		return opts, fmt.Errorf("missing task id; use --task <task-id>")
	}
	if requireComment && strings.TrimSpace(opts.CommentID) == "" {
		return opts, fmt.Errorf("missing comment id; use --comment <comment-id>")
	}
	return opts, nil
}
