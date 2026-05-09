package cli

import (
	"context"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/DeliciousBuding/dida-cli/internal/model"
	"github.com/DeliciousBuding/dida-cli/internal/webapi"
)

func runTask(args []string, jsonOut bool, stdout io.Writer, stderr io.Writer) int {
	if len(args) == 0 || args[0] == "-h" || args[0] == "--help" {
		printTaskHelp(stdout)
		return 0
	}
	switch args[0] {
	case "today":
		return runTaskList(append([]string{"list", "--filter", "today"}, args[1:]...), jsonOut, stdout, stderr)
	case "list":
		return runTaskList(args, jsonOut, stdout, stderr)
	case "search":
		return runTaskSearch(args[1:], jsonOut, stdout, stderr)
	case "upcoming":
		return runTaskUpcoming(args[1:], jsonOut, stdout, stderr)
	case "get":
		if len(args) != 2 {
			return failTyped("task get", "validation", "usage: dida task get <task-id>", "run: dida task --help", jsonOut, stdout, stderr)
		}
		return runTaskGet(args[1], jsonOut, stdout, stderr)
	case "create":
		return runTaskCreate(args[1:], jsonOut, stdout, stderr)
	case "update":
		return runTaskUpdate(args[1:], jsonOut, stdout, stderr)
	case "complete":
		return runTaskComplete(args[1:], jsonOut, stdout, stderr)
	case "delete":
		return runTaskDelete(args[1:], jsonOut, stdout, stderr)
	case "move":
		return runTaskMove(args[1:], jsonOut, stdout, stderr)
	case "parent":
		return runTaskParent(args[1:], jsonOut, stdout, stderr)
	default:
		return fail("task", fmt.Sprintf("unknown task command %q", args[0]), jsonOut, stdout, stderr)
	}
}

func runTaskList(args []string, jsonOut bool, stdout io.Writer, stderr io.Writer) int {
	filter, limit, compact, err := parseTaskListFlags(args[1:])
	if err != nil {
		return failTyped("task list", "validation", err.Error(), "run: dida task list --help", jsonOut, stdout, stderr)
	}
	view, err := loadSyncView()
	if err != nil {
		return failTyped("task list", "auth", err.Error(), "run: dida auth login", jsonOut, stdout, stderr)
	}
	now := time.Now()
	var tasks []model.Task
	switch filter {
	case "all":
		tasks = model.ActiveTasks(view.Tasks)
	case "today":
		tasks = model.TodayTasks(view.Tasks, now)
	default:
		return failTyped("task list", "validation", "unknown filter; supported filters: today, all", "run: dida task list --help", jsonOut, stdout, stderr)
	}
	total := len(tasks)
	if limit > 0 && len(tasks) > limit {
		tasks = tasks[:limit]
	}
	data := map[string]any{
		"filter":  filter,
		"compact": compact,
		"tasks":   taskOutput(tasks, compact),
	}
	meta := map[string]any{"count": len(tasks), "total": total}
	command := "task list"
	if filter == "today" && len(args) > 0 && args[0] != "list" {
		command = "task today"
	}
	if jsonOut {
		return writeJSON(stdout, envelope{OK: true, Command: command, Meta: meta, Data: data})
	}
	printTasks(stdout, tasks, total)
	return 0
}

func runTaskSearch(args []string, jsonOut bool, stdout io.Writer, stderr io.Writer) int {
	query, limit, compact, err := parseSearchFlags(args)
	if err != nil {
		return failTyped("task search", "validation", err.Error(), "run: dida task --help", jsonOut, stdout, stderr)
	}
	view, err := loadSyncView()
	if err != nil {
		return failTyped("task search", "auth", err.Error(), "run: dida auth login", jsonOut, stdout, stderr)
	}
	tasks := model.SearchTasks(model.ActiveTasks(view.Tasks), query)
	total := len(tasks)
	if limit > 0 && len(tasks) > limit {
		tasks = tasks[:limit]
	}
	data := map[string]any{"query": query, "compact": compact, "tasks": taskOutput(tasks, compact)}
	meta := map[string]any{"count": len(tasks), "total": total}
	if jsonOut {
		return writeJSON(stdout, envelope{OK: true, Command: "task search", Meta: meta, Data: data})
	}
	printTasks(stdout, tasks, total)
	return 0
}

func runTaskUpcoming(args []string, jsonOut bool, stdout io.Writer, stderr io.Writer) int {
	days, limit, compact, err := parseUpcomingFlags(args)
	if err != nil {
		return failTyped("task upcoming", "validation", err.Error(), "run: dida task --help", jsonOut, stdout, stderr)
	}
	view, err := loadSyncView()
	if err != nil {
		return failTyped("task upcoming", "auth", err.Error(), "run: dida auth login", jsonOut, stdout, stderr)
	}
	tasks := model.UpcomingTasks(view.Tasks, time.Now(), days)
	total := len(tasks)
	if limit > 0 && len(tasks) > limit {
		tasks = tasks[:limit]
	}
	data := map[string]any{"days": days, "compact": compact, "tasks": taskOutput(tasks, compact)}
	meta := map[string]any{"count": len(tasks), "total": total}
	if jsonOut {
		return writeJSON(stdout, envelope{OK: true, Command: "task upcoming", Meta: meta, Data: data})
	}
	printTasks(stdout, tasks, total)
	return 0
}

func runTaskGet(taskID string, jsonOut bool, stdout io.Writer, stderr io.Writer) int {
	view, err := loadSyncView()
	if err != nil {
		return failTyped("task get", "auth", err.Error(), "run: dida auth login", jsonOut, stdout, stderr)
	}
	task, ok := model.FindTask(view.Tasks, taskID)
	if !ok {
		return failTyped("task get", "not_found", "task not found", "run: dida task list --filter all --json", jsonOut, stdout, stderr)
	}
	if jsonOut {
		return writeJSON(stdout, envelope{OK: true, Command: "task get", Data: map[string]any{"task": stripSingleTaskRaw(task)}})
	}
	printTasks(stdout, []model.Task{task}, 1)
	return 0
}

func runTaskCreate(args []string, jsonOut bool, stdout io.Writer, stderr io.Writer) int {
	opts, err := parseTaskCreateFlags(args)
	if err != nil {
		return failTyped("task create", "validation", err.Error(), "run: dida task --help", jsonOut, stdout, stderr)
	}
	task := webapi.TaskMutation{
		ID:         opts.ID,
		ProjectID:  opts.ProjectID,
		Title:      opts.Title,
		Content:    opts.Content,
		Desc:       opts.Desc,
		AllDay:     opts.AllDay,
		StartDate:  opts.StartDate,
		DueDate:    opts.DueDate,
		Priority:   priorityPointer(opts.Priority),
		TimeZone:   opts.TimeZone,
		Reminders:  opts.Reminders,
		Repeat:     opts.Repeat,
		RepeatFrom: opts.RepeatFrom,
		RepeatFlag: opts.RepeatFlag,
		ColumnID:   opts.ColumnID,
		Tags:       opts.Tags,
		Items:      opts.Items,
		IsFloating: opts.IsFloating,
	}
	if task.ID == "" {
		task.ID = webapi.NewTaskID()
	}
	payload := map[string]any{"add": []webapi.TaskMutation{task}}
	if opts.DryRun {
		return writeMutationPreview("task create", payload, opts.Yes, jsonOut, stdout, stderr)
	}
	result, err := executeMutation(func(ctx context.Context, client *webapi.Client) (map[string]any, error) {
		return client.CreateTask(ctx, task)
	})
	if err != nil {
		return failTyped("task create", "api", err.Error(), "", jsonOut, stdout, stderr)
	}
	data := map[string]any{"taskId": task.ID, "projectId": task.ProjectID, "result": result}
	if jsonOut {
		return writeJSON(stdout, envelope{OK: true, Command: "task create", Data: data})
	}
	fmt.Fprintf(stdout, "Task created: %s\n", task.ID)
	return 0
}

func runTaskUpdate(args []string, jsonOut bool, stdout io.Writer, stderr io.Writer) int {
	opts, err := parseTaskUpdateFlags(args)
	if err != nil {
		return failTyped("task update", "validation", err.Error(), "run: dida task --help", jsonOut, stdout, stderr)
	}
	task := webapi.TaskMutation{
		ID:         opts.TaskID,
		ProjectID:  opts.ProjectID,
		Title:      opts.Title,
		Content:    opts.Content,
		Desc:       opts.Desc,
		AllDay:     opts.AllDay,
		StartDate:  opts.StartDate,
		DueDate:    opts.DueDate,
		Priority:   priorityPointer(opts.Priority),
		TimeZone:   opts.TimeZone,
		Reminders:  opts.Reminders,
		Repeat:     opts.Repeat,
		RepeatFrom: opts.RepeatFrom,
		RepeatFlag: opts.RepeatFlag,
		ColumnID:   opts.ColumnID,
		Tags:       opts.Tags,
		Items:      opts.Items,
		IsFloating: opts.IsFloating,
	}
	payload := map[string]any{"update": []webapi.TaskMutation{task}}
	if opts.DryRun {
		return writeMutationPreview("task update", payload, opts.Yes, jsonOut, stdout, stderr)
	}
	result, err := executeMutation(func(ctx context.Context, client *webapi.Client) (map[string]any, error) {
		return client.UpdateTask(ctx, task)
	})
	if err != nil {
		return failTyped("task update", "api", err.Error(), "", jsonOut, stdout, stderr)
	}
	data := map[string]any{"taskId": task.ID, "projectId": task.ProjectID, "result": result}
	if jsonOut {
		return writeJSON(stdout, envelope{OK: true, Command: "task update", Data: data})
	}
	fmt.Fprintf(stdout, "Task updated: %s\n", task.ID)
	return 0
}

func runTaskComplete(args []string, jsonOut bool, stdout io.Writer, stderr io.Writer) int {
	opts, err := parseTaskIDProjectFlags(args, "complete")
	if err != nil {
		return failTyped("task complete", "validation", err.Error(), "run: dida task --help", jsonOut, stdout, stderr)
	}
	status := 2
	payload := map[string]any{"update": []webapi.TaskMutation{{ID: opts.TaskID, ProjectID: opts.ProjectID, Status: &status}}}
	if opts.DryRun {
		return writeMutationPreview("task complete", payload, opts.Yes, jsonOut, stdout, stderr)
	}
	result, err := executeMutation(func(ctx context.Context, client *webapi.Client) (map[string]any, error) {
		return client.CompleteTask(ctx, opts.TaskID, opts.ProjectID)
	})
	if err != nil {
		return failTyped("task complete", "api", err.Error(), "", jsonOut, stdout, stderr)
	}
	data := map[string]any{"taskId": opts.TaskID, "projectId": opts.ProjectID, "result": result}
	if jsonOut {
		return writeJSON(stdout, envelope{OK: true, Command: "task complete", Data: data})
	}
	fmt.Fprintf(stdout, "Task completed: %s\n", opts.TaskID)
	return 0
}

func runTaskDelete(args []string, jsonOut bool, stdout io.Writer, stderr io.Writer) int {
	opts, err := parseTaskIDProjectFlags(args, "delete")
	if err != nil {
		return failTyped("task delete", "validation", err.Error(), "run: dida task --help", jsonOut, stdout, stderr)
	}
	payload := map[string]any{"delete": []map[string]string{{"taskId": opts.TaskID, "projectId": opts.ProjectID}}}
	if opts.DryRun {
		return writeMutationPreview("task delete", payload, opts.Yes, jsonOut, stdout, stderr)
	}
	if !opts.Yes {
		return failTyped("task delete", "confirmation_required", "task delete requires --yes", "preview first with: dida task delete <task-id> --project <project-id> --dry-run", jsonOut, stdout, stderr)
	}
	result, err := executeMutation(func(ctx context.Context, client *webapi.Client) (map[string]any, error) {
		return client.DeleteTask(ctx, opts.TaskID, opts.ProjectID)
	})
	if err != nil {
		return failTyped("task delete", "api", err.Error(), "", jsonOut, stdout, stderr)
	}
	data := map[string]any{"taskId": opts.TaskID, "projectId": opts.ProjectID, "result": result}
	if jsonOut {
		return writeJSON(stdout, envelope{OK: true, Command: "task delete", Data: data})
	}
	fmt.Fprintf(stdout, "Task deleted: %s\n", opts.TaskID)
	return 0
}

func parseTaskListFlags(args []string) (string, int, bool, error) {
	filter := "all"
	limit := 50
	compact := false
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--filter":
			if i+1 >= len(args) {
				return "", 0, false, fmt.Errorf("--filter requires a value")
			}
			filter = args[i+1]
			i++
		case "--limit":
			if i+1 >= len(args) {
				return "", 0, false, fmt.Errorf("--limit requires a value")
			}
			var parsed int
			if _, err := fmt.Sscanf(args[i+1], "%d", &parsed); err != nil {
				return "", 0, false, fmt.Errorf("--limit must be an integer")
			}
			limit = parsed
			i++
		case "--compact", "--brief":
			compact = true
		default:
			return "", 0, false, fmt.Errorf("unknown flag %q", args[i])
		}
	}
	return filter, limit, compact, nil
}

func parseSearchFlags(args []string) (string, int, bool, error) {
	query := ""
	limit := 50
	compact := false
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--query", "-q":
			if i+1 >= len(args) {
				return "", 0, false, fmt.Errorf("%s requires a value", args[i])
			}
			query = args[i+1]
			i++
		case "--limit":
			if i+1 >= len(args) {
				return "", 0, false, fmt.Errorf("--limit requires a value")
			}
			if _, err := fmt.Sscanf(args[i+1], "%d", &limit); err != nil || limit < 0 {
				return "", 0, false, fmt.Errorf("--limit must be a non-negative integer")
			}
			i++
		case "--compact", "--brief":
			compact = true
		default:
			if query == "" {
				query = args[i]
				continue
			}
			return "", 0, false, fmt.Errorf("unknown flag %q", args[i])
		}
	}
	if strings.TrimSpace(query) == "" {
		return "", 0, false, fmt.Errorf("missing query; use: dida task search --query <text>")
	}
	return query, limit, compact, nil
}

func parseUpcomingFlags(args []string) (int, int, bool, error) {
	days := 7
	limit := 50
	compact := false
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--days":
			if i+1 >= len(args) {
				return 0, 0, false, fmt.Errorf("--days requires a value")
			}
			if _, err := fmt.Sscanf(args[i+1], "%d", &days); err != nil || days <= 0 {
				return 0, 0, false, fmt.Errorf("--days must be a positive integer")
			}
			i++
		case "--limit":
			if i+1 >= len(args) {
				return 0, 0, false, fmt.Errorf("--limit requires a value")
			}
			if _, err := fmt.Sscanf(args[i+1], "%d", &limit); err != nil || limit < 0 {
				return 0, 0, false, fmt.Errorf("--limit must be a non-negative integer")
			}
			i++
		case "--compact", "--brief":
			compact = true
		default:
			return 0, 0, false, fmt.Errorf("unknown flag %q", args[i])
		}
	}
	return days, limit, compact, nil
}

type taskCreateOptions struct {
	ID         string
	ProjectID  string
	Title      string
	Content    string
	Desc       string
	AllDay     *bool
	StartDate  string
	DueDate    string
	TimeZone   string
	Reminders  []string
	Repeat     string
	RepeatFrom string
	RepeatFlag string
	Priority   int
	ColumnID   string
	Tags       []string
	Items      []webapi.SubTaskItem
	IsFloating *bool
	DryRun     bool
	Yes        bool
}

type taskUpdateOptions struct {
	TaskID     string
	ProjectID  string
	Title      string
	Content    string
	Desc       string
	AllDay     *bool
	StartDate  string
	DueDate    string
	TimeZone   string
	Reminders  []string
	Repeat     string
	RepeatFrom string
	RepeatFlag string
	Priority   int
	ColumnID   string
	Tags       []string
	Items      []webapi.SubTaskItem
	IsFloating *bool
	DryRun     bool
	Yes        bool
}

type taskIDProjectOptions struct {
	TaskID    string
	ProjectID string
	DryRun    bool
	Yes       bool
}

func parseTaskCreateFlags(args []string) (taskCreateOptions, error) {
	opts := taskCreateOptions{Priority: -1}
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--id":
			if i+1 >= len(args) {
				return opts, fmt.Errorf("--id requires a value")
			}
			opts.ID = args[i+1]
			i++
		case "--project", "-p":
			if i+1 >= len(args) {
				return opts, fmt.Errorf("%s requires a project id", args[i])
			}
			opts.ProjectID = args[i+1]
			i++
		case "--title", "-t":
			if i+1 >= len(args) {
				return opts, fmt.Errorf("%s requires a title", args[i])
			}
			opts.Title = args[i+1]
			i++
		case "--content", "-c":
			if i+1 >= len(args) {
				return opts, fmt.Errorf("%s requires content", args[i])
			}
			opts.Content = args[i+1]
			i++
		case "--desc":
			if i+1 >= len(args) {
				return opts, fmt.Errorf("--desc requires a value")
			}
			opts.Desc = args[i+1]
			i++
		case "--all-day":
			value := true
			opts.AllDay = &value
		case "--not-all-day":
			value := false
			opts.AllDay = &value
		case "--start":
			if i+1 >= len(args) {
				return opts, fmt.Errorf("--start requires a date")
			}
			opts.StartDate = args[i+1]
			i++
		case "--due", "-d":
			if i+1 >= len(args) {
				return opts, fmt.Errorf("%s requires a date", args[i])
			}
			opts.DueDate = args[i+1]
			i++
		case "--timezone", "--time-zone":
			if i+1 >= len(args) {
				return opts, fmt.Errorf("%s requires a timezone", args[i])
			}
			opts.TimeZone = args[i+1]
			i++
		case "--reminder":
			if i+1 >= len(args) {
				return opts, fmt.Errorf("--reminder requires a value")
			}
			opts.Reminders = append(opts.Reminders, args[i+1])
			i++
		case "--repeat":
			if i+1 >= len(args) {
				return opts, fmt.Errorf("--repeat requires a value")
			}
			opts.Repeat = args[i+1]
			i++
		case "--repeat-from":
			if i+1 >= len(args) {
				return opts, fmt.Errorf("--repeat-from requires a value")
			}
			opts.RepeatFrom = args[i+1]
			i++
		case "--repeat-flag":
			if i+1 >= len(args) {
				return opts, fmt.Errorf("--repeat-flag requires a value")
			}
			opts.RepeatFlag = args[i+1]
			i++
		case "--priority":
			if i+1 >= len(args) {
				return opts, fmt.Errorf("--priority requires a value")
			}
			priority, err := parsePriority(args[i+1])
			if err != nil {
				return opts, err
			}
			opts.Priority = priority
			i++
		case "--column":
			if i+1 >= len(args) {
				return opts, fmt.Errorf("--column requires a column id")
			}
			opts.ColumnID = args[i+1]
			i++
		case "--tag":
			if i+1 >= len(args) {
				return opts, fmt.Errorf("--tag requires a tag name")
			}
			opts.Tags = append(opts.Tags, args[i+1])
			i++
		case "--tags":
			if i+1 >= len(args) {
				return opts, fmt.Errorf("--tags requires comma-separated tag names")
			}
			opts.Tags = append(opts.Tags, splitCSV(args[i+1])...)
			i++
		case "--item":
			if i+1 >= len(args) {
				return opts, fmt.Errorf("--item requires a checklist item title")
			}
			opts.Items = append(opts.Items, webapi.SubTaskItem{Title: args[i+1]})
			i++
		case "--floating":
			value := true
			opts.IsFloating = &value
		case "--not-floating":
			value := false
			opts.IsFloating = &value
		case "--dry-run":
			opts.DryRun = true
		case "--yes":
			opts.Yes = true
		default:
			if opts.Title == "" {
				opts.Title = args[i]
				continue
			}
			return opts, fmt.Errorf("unknown flag %q", args[i])
		}
	}
	if strings.TrimSpace(opts.ProjectID) == "" {
		return opts, fmt.Errorf("missing project id; use --project <project-id>")
	}
	if strings.TrimSpace(opts.Title) == "" {
		return opts, fmt.Errorf("missing title; use --title <title>")
	}
	return opts, nil
}

func parseTaskUpdateFlags(args []string) (taskUpdateOptions, error) {
	opts := taskUpdateOptions{Priority: -1}
	if len(args) == 0 || strings.HasPrefix(args[0], "-") {
		return opts, fmt.Errorf("usage: dida task update <task-id> --project <project-id> [--title ...]")
	}
	opts.TaskID = args[0]
	for i := 1; i < len(args); i++ {
		switch args[i] {
		case "--project", "-p":
			if i+1 >= len(args) {
				return opts, fmt.Errorf("%s requires a project id", args[i])
			}
			opts.ProjectID = args[i+1]
			i++
		case "--title", "-t":
			if i+1 >= len(args) {
				return opts, fmt.Errorf("%s requires a title", args[i])
			}
			opts.Title = args[i+1]
			i++
		case "--content", "-c":
			if i+1 >= len(args) {
				return opts, fmt.Errorf("%s requires content", args[i])
			}
			opts.Content = args[i+1]
			i++
		case "--desc":
			if i+1 >= len(args) {
				return opts, fmt.Errorf("--desc requires a value")
			}
			opts.Desc = args[i+1]
			i++
		case "--all-day":
			value := true
			opts.AllDay = &value
		case "--not-all-day":
			value := false
			opts.AllDay = &value
		case "--start":
			if i+1 >= len(args) {
				return opts, fmt.Errorf("--start requires a date")
			}
			opts.StartDate = args[i+1]
			i++
		case "--due", "-d":
			if i+1 >= len(args) {
				return opts, fmt.Errorf("%s requires a date", args[i])
			}
			opts.DueDate = args[i+1]
			i++
		case "--timezone", "--time-zone":
			if i+1 >= len(args) {
				return opts, fmt.Errorf("%s requires a timezone", args[i])
			}
			opts.TimeZone = args[i+1]
			i++
		case "--reminder":
			if i+1 >= len(args) {
				return opts, fmt.Errorf("--reminder requires a value")
			}
			opts.Reminders = append(opts.Reminders, args[i+1])
			i++
		case "--repeat":
			if i+1 >= len(args) {
				return opts, fmt.Errorf("--repeat requires a value")
			}
			opts.Repeat = args[i+1]
			i++
		case "--repeat-from":
			if i+1 >= len(args) {
				return opts, fmt.Errorf("--repeat-from requires a value")
			}
			opts.RepeatFrom = args[i+1]
			i++
		case "--repeat-flag":
			if i+1 >= len(args) {
				return opts, fmt.Errorf("--repeat-flag requires a value")
			}
			opts.RepeatFlag = args[i+1]
			i++
		case "--priority":
			if i+1 >= len(args) {
				return opts, fmt.Errorf("--priority requires a value")
			}
			priority, err := parsePriority(args[i+1])
			if err != nil {
				return opts, err
			}
			opts.Priority = priority
			i++
		case "--column":
			if i+1 >= len(args) {
				return opts, fmt.Errorf("--column requires a column id")
			}
			opts.ColumnID = args[i+1]
			i++
		case "--tag":
			if i+1 >= len(args) {
				return opts, fmt.Errorf("--tag requires a tag name")
			}
			opts.Tags = append(opts.Tags, args[i+1])
			i++
		case "--tags":
			if i+1 >= len(args) {
				return opts, fmt.Errorf("--tags requires comma-separated tag names")
			}
			opts.Tags = append(opts.Tags, splitCSV(args[i+1])...)
			i++
		case "--item":
			if i+1 >= len(args) {
				return opts, fmt.Errorf("--item requires a checklist item title")
			}
			opts.Items = append(opts.Items, webapi.SubTaskItem{Title: args[i+1]})
			i++
		case "--floating":
			value := true
			opts.IsFloating = &value
		case "--not-floating":
			value := false
			opts.IsFloating = &value
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
	if opts.Title == "" && opts.Content == "" && opts.Desc == "" && opts.StartDate == "" && opts.DueDate == "" && opts.TimeZone == "" && len(opts.Reminders) == 0 && opts.Repeat == "" && opts.RepeatFrom == "" && opts.RepeatFlag == "" && opts.Priority < 0 && opts.ColumnID == "" && len(opts.Tags) == 0 && len(opts.Items) == 0 && opts.AllDay == nil && opts.IsFloating == nil {
		return opts, fmt.Errorf("no updates provided")
	}
	return opts, nil
}

func parseTaskIDProjectFlags(args []string, action string) (taskIDProjectOptions, error) {
	opts := taskIDProjectOptions{}
	if len(args) == 0 || strings.HasPrefix(args[0], "-") {
		return opts, fmt.Errorf("usage: dida task %s <task-id> --project <project-id>", action)
	}
	opts.TaskID = args[0]
	for i := 1; i < len(args); i++ {
		switch args[i] {
		case "--project", "-p":
			if i+1 >= len(args) {
				return opts, fmt.Errorf("%s requires a project id", args[i])
			}
			opts.ProjectID = args[i+1]
			i++
		case "--dry-run":
			opts.DryRun = true
		case "--yes":
			opts.Yes = true
		default:
			if opts.ProjectID == "" && !strings.HasPrefix(args[i], "-") {
				opts.ProjectID = args[i]
				continue
			}
			return opts, fmt.Errorf("unknown flag %q", args[i])
		}
	}
	if strings.TrimSpace(opts.ProjectID) == "" {
		return opts, fmt.Errorf("missing project id; use --project <project-id>")
	}
	return opts, nil
}

func parsePriority(value string) (int, error) {
	var priority int
	if _, err := fmt.Sscanf(value, "%d", &priority); err != nil {
		return 0, fmt.Errorf("--priority must be one of 0, 1, 3, 5")
	}
	switch priority {
	case 0, 1, 3, 5:
		return priority, nil
	default:
		return 0, fmt.Errorf("--priority must be one of 0, 1, 3, 5")
	}
}

func priorityPointer(value int) *int {
	if value < 0 {
		return nil
	}
	return &value
}

func splitCSV(value string) []string {
	parts := strings.Split(value, ",")
	out := make([]string, 0, len(parts))
	for _, part := range parts {
		if item := strings.TrimSpace(part); item != "" {
			out = append(out, item)
		}
	}
	return out
}

func writeMutationPreview(command string, payload any, yes bool, jsonOut bool, stdout io.Writer, stderr io.Writer) int {
	data := map[string]any{
		"dryRun":  true,
		"payload": payload,
		"hint":    "remove --dry-run to execute this write",
	}
	if strings.Contains(command, "delete") && !yes {
		data["hint"] = "remove --dry-run and add --yes to execute this write"
	}
	if jsonOut {
		return writeJSON(stdout, envelope{OK: true, Command: command, Data: data})
	}
	fmt.Fprintf(stdout, "%s dry run. Add --yes to execute.\n", command)
	return writeJSON(stdout, payload)
}

func hasFlag(args []string, flag string) bool {
	for _, arg := range args {
		if arg == flag {
			return true
		}
	}
	return false
}
