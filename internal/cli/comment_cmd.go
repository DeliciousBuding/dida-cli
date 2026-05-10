package cli

import (
	"context"
	"fmt"
	"io"
	"mime"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/DeliciousBuding/dida-cli/internal/webapi"
)

type commentOptions struct {
	ProjectID string
	TaskID    string
	CommentID string
	Title     string
	Files     []string
	DryRun    bool
	Yes       bool
}

func runComment(args []string, jsonOut bool, stdout io.Writer, stderr io.Writer) int {
	if len(args) == 0 || hasHelpFlag(args) {
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
	if len(opts.Files) > 0 {
		payload["files"] = commentFilePlan(opts.Files)
		payload["upload"] = map[string]any{
			"method": "POST",
			"path":   "/api/v1/attachment/upload/comment/{projectId}/{taskId}",
			"field":  "file",
		}
	}
	if opts.DryRun {
		return writeMutationPreview("comment create", payload, opts.Yes, jsonOut, stdout, stderr)
	}
	result, err := executeMutation(func(ctx context.Context, client *webapi.Client) (map[string]any, error) {
		if len(opts.Files) > 0 {
			quota, err := client.AttachmentQuota(ctx)
			if err != nil {
				return nil, fmt.Errorf("check attachment quota: %w", err)
			}
			if underQuota, _ := quota["underQuota"].(bool); !underQuota {
				return nil, fmt.Errorf("attachment quota exceeded; run: dida attachment quota --json")
			}
		}
		for _, path := range opts.Files {
			attachment, err := uploadCommentFile(ctx, client, opts.ProjectID, opts.TaskID, path)
			if err != nil {
				return nil, err
			}
			id, ok := attachment["id"].(string)
			if !ok || strings.TrimSpace(id) == "" {
				return nil, fmt.Errorf("upload %s did not return attachment id", filepath.Base(path))
			}
			comment.Attachments = append(comment.Attachments, webapi.CommentAttach{ID: id})
		}
		return client.CreateComment(ctx, opts.ProjectID, opts.TaskID, comment)
	})
	if err != nil {
		return failTyped("comment create", "api", err.Error(), "", jsonOut, stdout, stderr)
	}
	return writeMutationResult("comment create", "Comment created", map[string]any{"projectId": opts.ProjectID, "taskId": opts.TaskID, "commentId": comment.ID, "attachmentCount": len(comment.Attachments), "result": result}, comment.ID, jsonOut, stdout)
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
		case "--file", "--attachment":
			if i+1 >= len(args) {
				return opts, fmt.Errorf("%s requires a file path", args[i])
			}
			opts.Files = append(opts.Files, args[i+1])
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
		case "--file", "--attachment":
			if i+1 >= len(args) {
				return opts, fmt.Errorf("%s requires a file path", args[i])
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

func uploadCommentFile(ctx context.Context, client *webapi.Client, projectID string, taskID string, path string) (map[string]any, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open attachment %s: %w", filepath.Base(path), err)
	}
	defer file.Close()
	info, err := file.Stat()
	if err != nil {
		return nil, fmt.Errorf("stat attachment %s: %w", filepath.Base(path), err)
	}
	if info.IsDir() {
		return nil, fmt.Errorf("attachment %s is a directory", filepath.Base(path))
	}
	contentType := mime.TypeByExtension(strings.ToLower(filepath.Ext(path)))
	if contentType == "" {
		contentType = "application/octet-stream"
	}
	return client.UploadCommentAttachment(ctx, projectID, taskID, filepath.Base(path), contentType, file)
}

func commentFilePlan(paths []string) []map[string]any {
	files := make([]map[string]any, 0, len(paths))
	for _, path := range paths {
		files = append(files, map[string]any{"name": filepath.Base(path)})
	}
	return files
}
