package cli

import (
	"context"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/DeliciousBuding/dida-cli/internal/auth"
	"github.com/DeliciousBuding/dida-cli/internal/webapi"
)

type resourceOptions struct {
	ID        string
	Name      string
	Color     string
	GroupID   string
	Parent    string
	Label     string
	ProjectID string
	DryRun    bool
	Yes       bool
}

type taskMoveOptions struct {
	TaskID        string
	FromProjectID string
	ToProjectID   string
	DryRun        bool
}

type taskParentOptions struct {
	TaskID    string
	ParentID  string
	ProjectID string
	DryRun    bool
}

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

func runFolder(args []string, jsonOut bool, stdout io.Writer, stderr io.Writer) int {
	if len(args) == 0 || args[0] == "-h" || args[0] == "--help" {
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

func executeMutation(fn func(context.Context, *webapi.Client) (map[string]any, error)) (map[string]any, error) {
	token, err := auth.LoadCookieToken()
	if err != nil {
		return nil, fmt.Errorf("missing cookie auth; run: dida auth login")
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

func parseProjectCreateFlags(args []string) (resourceOptions, error) {
	opts := resourceOptions{}
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--id":
			if i+1 >= len(args) {
				return opts, fmt.Errorf("--id requires a value")
			}
			opts.ID = args[i+1]
			i++
		case "--name", "-n":
			if i+1 >= len(args) {
				return opts, fmt.Errorf("%s requires a name", args[i])
			}
			opts.Name = args[i+1]
			i++
		case "--group":
			if i+1 >= len(args) {
				return opts, fmt.Errorf("--group requires a folder id")
			}
			opts.GroupID = args[i+1]
			i++
		case "--dry-run":
			opts.DryRun = true
		case "--yes":
			opts.Yes = true
		default:
			if opts.Name == "" {
				opts.Name = args[i]
				continue
			}
			return opts, fmt.Errorf("unknown flag %q", args[i])
		}
	}
	if strings.TrimSpace(opts.Name) == "" {
		return opts, fmt.Errorf("missing name; use --name <name>")
	}
	return opts, nil
}

func parseProjectUpdateFlags(args []string) (resourceOptions, error) {
	opts := resourceOptions{}
	if len(args) == 0 || strings.HasPrefix(args[0], "-") {
		return opts, fmt.Errorf("usage: dida project update <project-id> [--name <name>] [--group <folder-id>]")
	}
	opts.ID = args[0]
	for i := 1; i < len(args); i++ {
		switch args[i] {
		case "--name", "-n":
			if i+1 >= len(args) {
				return opts, fmt.Errorf("%s requires a name", args[i])
			}
			opts.Name = args[i+1]
			i++
		case "--group":
			if i+1 >= len(args) {
				return opts, fmt.Errorf("--group requires a folder id")
			}
			opts.GroupID = args[i+1]
			i++
		case "--dry-run":
			opts.DryRun = true
		case "--yes":
			opts.Yes = true
		default:
			return opts, fmt.Errorf("unknown flag %q", args[i])
		}
	}
	if opts.Name == "" && opts.GroupID == "" {
		return opts, fmt.Errorf("no updates provided")
	}
	return opts, nil
}

func parseNamedCreateFlags(args []string) (resourceOptions, error) {
	opts := resourceOptions{}
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--id":
			if i+1 >= len(args) {
				return opts, fmt.Errorf("--id requires a value")
			}
			opts.ID = args[i+1]
			i++
		case "--name", "-n":
			if i+1 >= len(args) {
				return opts, fmt.Errorf("%s requires a name", args[i])
			}
			opts.Name = args[i+1]
			i++
		case "--dry-run":
			opts.DryRun = true
		case "--yes":
			opts.Yes = true
		default:
			if opts.Name == "" {
				opts.Name = args[i]
				continue
			}
			return opts, fmt.Errorf("unknown flag %q", args[i])
		}
	}
	if strings.TrimSpace(opts.Name) == "" {
		return opts, fmt.Errorf("missing name; use --name <name>")
	}
	return opts, nil
}

func parseNamedUpdateFlags(args []string, command string) (resourceOptions, error) {
	opts := resourceOptions{}
	if len(args) == 0 || strings.HasPrefix(args[0], "-") {
		return opts, fmt.Errorf("usage: dida %s <id> --name <name>", command)
	}
	opts.ID = args[0]
	for i := 1; i < len(args); i++ {
		switch args[i] {
		case "--name", "-n":
			if i+1 >= len(args) {
				return opts, fmt.Errorf("%s requires a name", args[i])
			}
			opts.Name = args[i+1]
			i++
		case "--dry-run":
			opts.DryRun = true
		case "--yes":
			opts.Yes = true
		default:
			return opts, fmt.Errorf("unknown flag %q", args[i])
		}
	}
	if strings.TrimSpace(opts.Name) == "" {
		return opts, fmt.Errorf("missing name; use --name <name>")
	}
	return opts, nil
}

func parseDeleteIDFlags(args []string, command string) (resourceOptions, error) {
	opts := resourceOptions{}
	if len(args) == 0 || strings.HasPrefix(args[0], "-") {
		return opts, fmt.Errorf("usage: dida %s <id> --yes", command)
	}
	opts.ID = args[0]
	for i := 1; i < len(args); i++ {
		switch args[i] {
		case "--dry-run":
			opts.DryRun = true
		case "--yes":
			opts.Yes = true
		default:
			return opts, fmt.Errorf("unknown flag %q", args[i])
		}
	}
	return opts, nil
}

func parseTagCreateFlags(args []string) (resourceOptions, error) {
	opts := resourceOptions{}
	if len(args) > 0 && !strings.HasPrefix(args[0], "-") {
		opts.Name = args[0]
		args = args[1:]
	}
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--name", "-n":
			if i+1 >= len(args) {
				return opts, fmt.Errorf("%s requires a name", args[i])
			}
			opts.Name = args[i+1]
			i++
		case "--color", "-c":
			if i+1 >= len(args) {
				return opts, fmt.Errorf("%s requires a color", args[i])
			}
			opts.Color = args[i+1]
			i++
		case "--parent", "-p":
			if i+1 >= len(args) {
				return opts, fmt.Errorf("%s requires a parent tag name", args[i])
			}
			opts.Parent = args[i+1]
			i++
		case "--label":
			if i+1 >= len(args) {
				return opts, fmt.Errorf("--label requires a value")
			}
			opts.Label = args[i+1]
			i++
		case "--dry-run":
			opts.DryRun = true
		case "--yes":
			opts.Yes = true
		default:
			return opts, fmt.Errorf("unknown flag %q", args[i])
		}
	}
	if strings.TrimSpace(opts.Name) == "" {
		return opts, fmt.Errorf("missing tag name")
	}
	return opts, nil
}

func parseTagUpdateFlags(args []string) (resourceOptions, error) {
	opts, err := parseTagCreateFlags(args)
	if err != nil {
		return opts, err
	}
	if opts.Color == "" && opts.Parent == "" && opts.Label == "" {
		return opts, fmt.Errorf("no updates provided")
	}
	return opts, nil
}

func parseTwoNameFlags(args []string, command string) (resourceOptions, error) {
	opts := resourceOptions{}
	names := make([]string, 0, 2)
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--dry-run":
			opts.DryRun = true
		case "--yes":
			opts.Yes = true
		default:
			if strings.HasPrefix(args[i], "-") {
				return opts, fmt.Errorf("unknown flag %q", args[i])
			}
			names = append(names, args[i])
		}
	}
	if len(names) != 2 {
		return opts, fmt.Errorf("usage: dida %s <name> <new-name>", command)
	}
	opts.ID = names[0]
	opts.Name = names[1]
	return opts, nil
}

func parseColumnCreateFlags(args []string) (resourceOptions, error) {
	opts := resourceOptions{}
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--project", "-p":
			if i+1 >= len(args) {
				return opts, fmt.Errorf("%s requires a project id", args[i])
			}
			opts.ProjectID = args[i+1]
			i++
		case "--name", "-n":
			if i+1 >= len(args) {
				return opts, fmt.Errorf("%s requires a name", args[i])
			}
			opts.Name = args[i+1]
			i++
		case "--dry-run":
			opts.DryRun = true
		case "--yes":
			opts.Yes = true
		default:
			if opts.Name == "" {
				opts.Name = args[i]
				continue
			}
			return opts, fmt.Errorf("unknown flag %q", args[i])
		}
	}
	if strings.TrimSpace(opts.ProjectID) == "" {
		return opts, fmt.Errorf("missing project id; use --project <project-id>")
	}
	if strings.TrimSpace(opts.Name) == "" {
		return opts, fmt.Errorf("missing name; use --name <name>")
	}
	return opts, nil
}

func parseTaskMoveFlags(args []string) (taskMoveOptions, error) {
	opts := taskMoveOptions{}
	if len(args) == 0 || strings.HasPrefix(args[0], "-") {
		return opts, fmt.Errorf("usage: dida task move <task-id> --from <project-id> --to <project-id>")
	}
	opts.TaskID = args[0]
	for i := 1; i < len(args); i++ {
		switch args[i] {
		case "--from":
			if i+1 >= len(args) {
				return opts, fmt.Errorf("--from requires a project id")
			}
			opts.FromProjectID = args[i+1]
			i++
		case "--to":
			if i+1 >= len(args) {
				return opts, fmt.Errorf("--to requires a project id")
			}
			opts.ToProjectID = args[i+1]
			i++
		case "--dry-run":
			opts.DryRun = true
		default:
			return opts, fmt.Errorf("unknown flag %q", args[i])
		}
	}
	if opts.FromProjectID == "" || opts.ToProjectID == "" {
		return opts, fmt.Errorf("missing --from or --to project id")
	}
	return opts, nil
}

func parseTaskParentFlags(args []string) (taskParentOptions, error) {
	opts := taskParentOptions{}
	if len(args) == 0 || strings.HasPrefix(args[0], "-") {
		return opts, fmt.Errorf("usage: dida task parent <task-id> --parent <task-id> --project <project-id>")
	}
	opts.TaskID = args[0]
	for i := 1; i < len(args); i++ {
		switch args[i] {
		case "--parent":
			if i+1 >= len(args) {
				return opts, fmt.Errorf("--parent requires a task id")
			}
			opts.ParentID = args[i+1]
			i++
		case "--project", "-p":
			if i+1 >= len(args) {
				return opts, fmt.Errorf("%s requires a project id", args[i])
			}
			opts.ProjectID = args[i+1]
			i++
		case "--dry-run":
			opts.DryRun = true
		default:
			return opts, fmt.Errorf("unknown flag %q", args[i])
		}
	}
	if opts.ParentID == "" || opts.ProjectID == "" {
		return opts, fmt.Errorf("missing --parent or --project")
	}
	return opts, nil
}

func printMapList(w io.Writer, items []map[string]any, label string) {
	if len(items) == 0 {
		fmt.Fprintf(w, "No %s found.\n", label)
		return
	}
	fmt.Fprintf(w, "%-28s  %s\n", "ID/NAME", "LABEL")
	for _, item := range items {
		id := fmt.Sprint(item["id"])
		if id == "" || id == "<nil>" {
			id = fmt.Sprint(item["name"])
		}
		name := fmt.Sprint(item["name"])
		if name == "<nil>" {
			name = fmt.Sprint(item["label"])
		}
		fmt.Fprintf(w, "%-28s  %s\n", id, name)
	}
}
