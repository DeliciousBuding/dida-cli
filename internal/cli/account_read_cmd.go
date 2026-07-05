package cli

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/DeliciousBuding/dida-cli/internal/webapi"
)

func runAttachment(args []string, jsonOut bool, stdout io.Writer, stderr io.Writer) int {
	if len(args) == 0 || args[0] == "-h" || args[0] == "--help" {
		printAttachmentHelp(stdout)
		return 0
	}
	switch args[0] {
	case "quota":
		return runAttachmentQuota(jsonOut, stdout, stderr)
	case "download":
		return runAttachmentDownload(args[1:], jsonOut, stdout, stderr)
	default:
		return fail("attachment", fmt.Sprintf("unknown attachment command %q", args[0]), jsonOut, stdout, stderr)
	}
}

func runAttachmentQuota(jsonOut bool, stdout io.Writer, stderr io.Writer) int {
	result, err := executeRead(func(ctx context.Context, client *webapi.Client) (any, error) {
		return client.AttachmentQuota(ctx)
	})
	if err != nil {
		return failTyped("attachment quota", "api", err.Error(), "", jsonOut, stdout, stderr)
	}
	data := map[string]any{"quota": result}
	if jsonOut {
		return writeJSON(stdout, envelope{OK: true, Command: "attachment quota", Data: data})
	}
	quota := result.(map[string]any)
	fmt.Fprintf(stdout, "Under quota: %v\nDaily limit: %v\n", quota["underQuota"], quota["dailyLimit"])
	return 0
}

type attachmentDownloadOptions struct {
	ProjectID    string
	TaskID       string
	AttachmentID string
	Output       string
	Force        bool
}

func runAttachmentDownload(args []string, jsonOut bool, stdout io.Writer, stderr io.Writer) int {
	opts, err := parseAttachmentDownloadFlags(args)
	if err != nil {
		return failTyped("attachment download", "validation", err.Error(), "run: dida attachment --help", jsonOut, stdout, stderr)
	}
	output, err := filepath.Abs(opts.Output)
	if err != nil {
		return failTyped("attachment download", "validation", fmt.Sprintf("resolve output path: %v", err), "", jsonOut, stdout, stderr)
	}
	if !opts.Force && fileExists(output) {
		return failTyped("attachment download", "file_exists", "output file already exists", "pass --force to overwrite it", jsonOut, stdout, stderr)
	}
	if err := os.MkdirAll(filepath.Dir(output), 0o755); err != nil {
		return failTyped("attachment download", "filesystem", fmt.Sprintf("create output directory: %v", err), "", jsonOut, stdout, stderr)
	}
	tmp, err := os.CreateTemp(filepath.Dir(output), ".dida-download-*")
	if err != nil {
		return failTyped("attachment download", "filesystem", fmt.Sprintf("create temporary file: %v", err), "", jsonOut, stdout, stderr)
	}
	tmpPath := tmp.Name()
	defer os.Remove(tmpPath)

	var bytesWritten int64
	var contentType string
	_, err = executeRead(func(ctx context.Context, client *webapi.Client) (any, error) {
		var downloadErr error
		bytesWritten, contentType, downloadErr = client.DownloadTaskAttachment(ctx, opts.ProjectID, opts.TaskID, opts.AttachmentID, tmp)
		return nil, downloadErr
	})
	closeErr := tmp.Close()
	if err != nil {
		return failTyped("attachment download", "api", err.Error(), "", jsonOut, stdout, stderr)
	}
	if closeErr != nil {
		return failTyped("attachment download", "filesystem", fmt.Sprintf("close temporary file: %v", closeErr), "", jsonOut, stdout, stderr)
	}
	if opts.Force {
		if err := os.Remove(output); err != nil && !os.IsNotExist(err) {
			return failTyped("attachment download", "filesystem", fmt.Sprintf("remove existing output: %v", err), "", jsonOut, stdout, stderr)
		}
	}
	if err := os.Rename(tmpPath, output); err != nil {
		return failTyped("attachment download", "filesystem", fmt.Sprintf("move temporary file into place: %v", err), "", jsonOut, stdout, stderr)
	}
	data := map[string]any{
		"projectId":    opts.ProjectID,
		"taskId":       opts.TaskID,
		"attachmentId": opts.AttachmentID,
		"output":       output,
		"bytes":        bytesWritten,
	}
	if contentType != "" {
		data["contentType"] = contentType
	}
	if jsonOut {
		return writeJSON(stdout, envelope{OK: true, Command: "attachment download", Data: data})
	}
	fmt.Fprintf(stdout, "Attachment downloaded: %s (%d bytes)\n", output, bytesWritten)
	return 0
}

func parseAttachmentDownloadFlags(args []string) (attachmentDownloadOptions, error) {
	opts := attachmentDownloadOptions{}
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--project", "-p":
			if i+1 >= len(args) {
				return opts, fmt.Errorf("%s requires a project id", args[i])
			}
			opts.ProjectID = args[i+1]
			i++
		case "--task", "-t":
			if i+1 >= len(args) {
				return opts, fmt.Errorf("%s requires a task id", args[i])
			}
			opts.TaskID = args[i+1]
			i++
		case "--attachment", "-a":
			if i+1 >= len(args) {
				return opts, fmt.Errorf("%s requires an attachment id", args[i])
			}
			opts.AttachmentID = args[i+1]
			i++
		case "--output", "-o":
			if i+1 >= len(args) {
				return opts, fmt.Errorf("%s requires a file path", args[i])
			}
			opts.Output = args[i+1]
			i++
		case "--force":
			opts.Force = true
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
	if strings.TrimSpace(opts.AttachmentID) == "" {
		return opts, fmt.Errorf("missing attachment id; use --attachment <attachment-id>")
	}
	if strings.TrimSpace(opts.Output) == "" {
		return opts, fmt.Errorf("missing output path; use --output <file>")
	}
	return opts, nil
}

func runReminder(args []string, jsonOut bool, stdout io.Writer, stderr io.Writer) int {
	if len(args) == 0 || args[0] == "-h" || args[0] == "--help" {
		printReminderHelp(stdout)
		return 0
	}
	switch args[0] {
	case "daily", "preferences", "prefs":
		return runDailyReminder(jsonOut, stdout, stderr)
	default:
		return fail("reminder", fmt.Sprintf("unknown reminder command %q", args[0]), jsonOut, stdout, stderr)
	}
}

func runDailyReminder(jsonOut bool, stdout io.Writer, stderr io.Writer) int {
	result, err := executeRead(func(ctx context.Context, client *webapi.Client) (any, error) {
		return client.DailyReminderPreferences(ctx)
	})
	if err != nil {
		return failTyped("reminder daily", "api", err.Error(), "", jsonOut, stdout, stderr)
	}
	data := map[string]any{"preferences": result}
	if jsonOut {
		return writeJSON(stdout, envelope{OK: true, Command: "reminder daily", Data: data})
	}
	fmt.Fprintf(stdout, "Daily reminder preferences: %d keys\n", len(result.(map[string]any)))
	return 0
}
