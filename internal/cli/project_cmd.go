package cli

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/DeliciousBuding/dida-cli/internal/model"
	"github.com/DeliciousBuding/dida-cli/internal/webapi"
)

func runProject(args []string, jsonOut bool, stdout io.Writer, stderr io.Writer) int {
	if len(args) == 0 || args[0] == "-h" || args[0] == "--help" {
		printProjectHelp(stdout)
		return 0
	}
	switch args[0] {
	case "list":
		return runProjectList(jsonOut, stdout, stderr)
	case "create":
		return runProjectCreate(args[1:], jsonOut, stdout, stderr)
	case "update":
		return runProjectUpdate(args[1:], jsonOut, stdout, stderr)
	case "delete":
		return runProjectDelete(args[1:], jsonOut, stdout, stderr)
	case "tasks":
		projectID, compact, err := parseProjectTasksArgs(args[1:])
		if err != nil {
			return failTyped("project tasks", "validation", err.Error(), "run: dida project --help", jsonOut, stdout, stderr)
		}
		return runProjectTasks(projectID, compact, jsonOut, stdout, stderr)
	case "columns":
		if len(args) != 2 {
			return failTyped("project columns", "validation", "usage: dida project columns <project-id>", "run: dida project --help", jsonOut, stdout, stderr)
		}
		return runProjectColumns(args[1], jsonOut, stdout, stderr)
	default:
		return fail("project", fmt.Sprintf("unknown project command %q", args[0]), jsonOut, stdout, stderr)
	}
}

func runProjectList(jsonOut bool, stdout io.Writer, stderr io.Writer) int {
	view, err := loadSyncView()
	if err != nil {
		return failTyped("project list", "auth", err.Error(), "run: dida auth login", jsonOut, stdout, stderr)
	}
	projects := stripProjectRaw(view.Projects)
	data := map[string]any{"projects": projects}
	meta := map[string]any{"count": len(projects)}
	if jsonOut {
		return writeJSON(stdout, envelope{OK: true, Command: "project list", Meta: meta, Data: data})
	}
	printProjects(stdout, view.Projects)
	return 0
}

func runProjectTasks(projectID string, compact bool, jsonOut bool, stdout io.Writer, stderr io.Writer) int {
	result, err := executeRead(func(ctx context.Context, client *webapi.Client) (any, error) {
		return client.ProjectTasks(ctx, projectID)
	})
	if err != nil {
		return failTyped("project tasks", "api", err.Error(), "", jsonOut, stdout, stderr)
	}
	rawTasks := result.([]map[string]any)
	projectNames := map[string]string{}
	if view, err := loadSyncView(); err == nil {
		for _, project := range view.Projects {
			projectNames[project.ID] = project.Name
		}
	}
	tasks := model.NormalizeTasks(rawTasks, projectNames, time.Now())
	tasks = model.ActiveTasks(tasks)
	data := map[string]any{"projectId": projectID, "compact": compact, "tasks": taskOutput(tasks, compact)}
	meta := map[string]any{"count": len(tasks), "source": "project_tasks_endpoint"}
	if jsonOut {
		return writeJSON(stdout, envelope{OK: true, Command: "project tasks", Meta: meta, Data: data})
	}
	printTasks(stdout, tasks, len(tasks))
	return 0
}

func parseProjectTasksArgs(args []string) (string, bool, error) {
	projectID := ""
	compact := false
	for _, arg := range args {
		switch arg {
		case "--compact", "--brief":
			compact = true
		default:
			if projectID == "" {
				projectID = arg
				continue
			}
			return "", false, fmt.Errorf("unknown flag %q", arg)
		}
	}
	if projectID == "" {
		return "", false, fmt.Errorf("usage: dida project tasks <project-id> [--compact]")
	}
	return projectID, compact, nil
}

func runProjectColumns(projectID string, jsonOut bool, stdout io.Writer, stderr io.Writer) int {
	result, err := executeRead(func(ctx context.Context, client *webapi.Client) (any, error) {
		columns, err := client.ProjectColumns(ctx, projectID)
		return map[string]any{"projectId": projectID, "columns": columns}, err
	})
	if err != nil {
		return failTyped("project columns", "api", err.Error(), "", jsonOut, stdout, stderr)
	}
	data := result.(map[string]any)
	columns := data["columns"].([]map[string]any)
	meta := map[string]any{"count": len(columns), "source": "column_project_endpoint"}
	if jsonOut {
		return writeJSON(stdout, envelope{OK: true, Command: "project columns", Meta: meta, Data: data})
	}
	if len(columns) == 0 {
		fmt.Fprintln(stdout, "No columns found.")
		return 0
	}
	fmt.Fprintf(stdout, "%-28s  %-12s  %s\n", "ID", "SORT", "NAME")
	for _, column := range columns {
		fmt.Fprintf(stdout, "%-28v  %-12v  %v\n", column["id"], column["sortOrder"], column["name"])
	}
	return 0
}
