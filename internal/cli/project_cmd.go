package cli

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/DeliciousBuding/dida-cli/internal/auth"
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
		if len(args) != 2 {
			return failTyped("project tasks", "validation", "usage: dida project tasks <project-id>", "run: dida project --help", jsonOut, stdout, stderr)
		}
		return runProjectTasks(args[1], jsonOut, stdout, stderr)
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

func runProjectTasks(projectID string, jsonOut bool, stdout io.Writer, stderr io.Writer) int {
	token, err := auth.LoadCookieToken()
	if err != nil {
		return missingAuth("project tasks", jsonOut, stdout, stderr)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	rawTasks, err := webapi.NewClient(token.Token).ProjectTasks(ctx, projectID)
	if err != nil {
		return failTyped("project tasks", "api", err.Error(), "", jsonOut, stdout, stderr)
	}
	projectNames := map[string]string{}
	if view, err := loadSyncView(); err == nil {
		for _, project := range view.Projects {
			projectNames[project.ID] = project.Name
		}
	}
	tasks := model.NormalizeTasks(rawTasks, projectNames, time.Now())
	tasks = model.ActiveTasks(tasks)
	data := map[string]any{"projectId": projectID, "tasks": stripTaskRaw(tasks)}
	meta := map[string]any{"count": len(tasks), "source": "project_tasks_endpoint"}
	if jsonOut {
		return writeJSON(stdout, envelope{OK: true, Command: "project tasks", Meta: meta, Data: data})
	}
	printTasks(stdout, tasks, len(tasks))
	return 0
}

func runProjectColumns(projectID string, jsonOut bool, stdout io.Writer, stderr io.Writer) int {
	view, err := loadSyncView()
	if err != nil {
		return failTyped("project columns", "auth", err.Error(), "run: dida auth login", jsonOut, stdout, stderr)
	}
	columns := model.InferColumns(projectID, view.Tasks)
	data := map[string]any{"projectId": projectID, "columns": columns}
	meta := map[string]any{"count": len(columns), "source": "inferred_from_tasks"}
	if jsonOut {
		return writeJSON(stdout, envelope{OK: true, Command: "project columns", Meta: meta, Data: data})
	}
	if len(columns) == 0 {
		fmt.Fprintln(stdout, "No columns found.")
		return 0
	}
	fmt.Fprintf(stdout, "%-28s  %-8s\n", "ID", "TASKS")
	for _, column := range columns {
		fmt.Fprintf(stdout, "%-28s  %-8d\n", column.ID, column.TaskCount)
	}
	return 0
}
