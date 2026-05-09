package cli

import (
	"fmt"
	"io"
	"strings"
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
