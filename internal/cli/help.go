package cli

import (
	"fmt"
	"io"
	"strings"

	"github.com/DeliciousBuding/dida-cli/internal/model"
)

func printHelp(w io.Writer) {
	fmt.Fprintln(w, strings.TrimSpace(`
DidaCLI - Dida365 / TickTick command line client

Usage:
  dida <command> [options]

Commands:
  doctor       Check local config and auth status
  auth         Manage local cookie auth
  sync         Sync tasks/projects/tags
  settings     Read user preferences
  completed    Read completed task history
  quadrant     View active tasks by Eisenhower quadrant
  project      Project discovery and CRUD
  folder       Project folder CRUD
  tag          Tag discovery and CRUD
  column       Kanban column discovery and experimental create
  task         Task reads and writes
  raw          Raw read-only API escape hatch
  version      Print version
  +today       Shortcut for task today

Global options:
  -j, --json   Emit machine-readable JSON
  -h, --help   Show help
`))
}

func printAuthHelp(w io.Writer) {
	fmt.Fprintln(w, strings.TrimSpace(`
Usage:
  dida auth login [--json]
  dida auth status [--json]
  dida auth status --verify [--json]
  dida auth logout [--json]
  dida auth cookie set --token-stdin
  dida auth cookie set --token <token>
`))
}

func printAuthLoginHelp(w io.Writer) {
	fmt.Fprintln(w, strings.TrimSpace(`
Usage:
  dida auth login [--json]
  dida auth login --browser [--timeout 180] [--json]

This prints a browser login guide. Complete Dida365/WeChat/QR login in the browser,
then import only the resulting cookie named 't' with:
  dida auth cookie set --token-stdin

With --browser, the CLI opens a visible browser, waits for cookie 't', and saves it automatically.
`))
}

func printAuthCookieHelp(w io.Writer) {
	fmt.Fprintln(w, strings.TrimSpace(`
Usage:
  dida auth cookie set --token-stdin
  dida auth cookie set --token <token>

Prefer --token-stdin to avoid shell history.
`))
}

func printSyncHelp(w io.Writer) {
	fmt.Fprintln(w, strings.TrimSpace(`
Usage:
  dida sync all [--json]
  dida sync checkpoint <checkpoint> [--json]
`))
}

func printSettingsHelp(w io.Writer) {
	fmt.Fprintln(w, strings.TrimSpace(`
Usage:
  dida settings get [--json]
`))
}

func printCompletedHelp(w io.Writer) {
	fmt.Fprintln(w, strings.TrimSpace(`
Usage:
  dida completed today [--json]
  dida completed yesterday [--json]
  dida completed week [--json]
  dida completed list [--from YYYY-MM-DD] [--to YYYY-MM-DD] [--limit N] [--json]
`))
}

func printProjectHelp(w io.Writer) {
	fmt.Fprintln(w, strings.TrimSpace(`
Usage:
  dida project list [--json]
  dida project create --name <name> [--group <folder-id>] [--dry-run] [--json]
  dida project update <project-id> [--name <name>] [--group <folder-id>] [--dry-run] [--json]
  dida project delete <project-id> --yes [--dry-run] [--json]
  dida project tasks <project-id> [--json]
  dida project columns <project-id> [--json]
`))
}

func printTaskHelp(w io.Writer) {
	fmt.Fprintln(w, strings.TrimSpace(`
Usage:
  dida task today [--json] [--limit N]
  dida task list [--json] [--filter today|all] [--limit N]
  dida task search --query <text> [--limit N] [--json]
  dida task upcoming [--days N] [--limit N] [--json]
  dida task get <task-id> [--json]
  dida task create --project <project-id> --title <title> [task fields...] [--dry-run] [--json]
  dida task update <task-id> --project <project-id> [task fields...] [--dry-run] [--json]
  dida task complete <task-id> --project <project-id> [--dry-run] [--json]
  dida task delete <task-id> --project <project-id> --yes [--dry-run] [--json]
  dida task move <task-id> --from <project-id> --to <project-id> [--dry-run] [--json]
  dida task parent <task-id> --parent <task-id> --project <project-id> [--dry-run] [--json]
  dida +today [--json] [--limit N]

Task fields:
  --content <text>        Task content
  --desc <markdown>       Rich description field
  --start <time>          Start date/time
  --due <time>            Due date/time
  --timezone <zone>       IANA timezone, e.g. Asia/Shanghai
  --priority 0|1|3|5      None, low, medium, high
  --tag <name>            Add a tag; repeatable
  --tags a,b              Add comma-separated tags
  --item <title>          Add a checklist item; repeatable
  --column <id>           Kanban column id
  --reminder <value>      Reminder value; repeatable
  --repeat <rule>         Repeat rule from Web API
  --repeat-from <value>   Repeat base
  --repeat-flag <value>   Repeat flag
  --all-day | --not-all-day
  --floating | --not-floating
`))
}

func printFolderHelp(w io.Writer) {
	fmt.Fprintln(w, strings.TrimSpace(`
Usage:
  dida folder list [--json]
  dida folder create --name <name> [--dry-run] [--json]
  dida folder update <folder-id> --name <name> [--dry-run] [--json]
  dida folder delete <folder-id> --yes [--dry-run] [--json]
`))
}

func printTagHelp(w io.Writer) {
	fmt.Fprintln(w, strings.TrimSpace(`
Usage:
  dida tag list [--json]
  dida tag create <name> [--color <color>] [--parent <name>] [--dry-run] [--json]
  dida tag update <name> [--color <color>] [--parent <name>] [--label <label>] [--dry-run] [--json]
  dida tag rename <old-name> <new-name> [--dry-run] [--json]
  dida tag merge <from-name> <to-name> --yes [--dry-run] [--json]
  dida tag delete <name> --yes [--dry-run] [--json]
`))
}

func printColumnHelp(w io.Writer) {
	fmt.Fprintln(w, strings.TrimSpace(`
Usage:
  dida column list <project-id> [--json]
  dida column create --project <project-id> --name <name> [--dry-run] [--json]

Column create uses an experimental private Web API endpoint. Update/delete are not exposed until verified.
`))
}

func printRawHelp(w io.Writer) {
	fmt.Fprintln(w, strings.TrimSpace(`
Usage:
  dida raw get <path> [--json]

Only GET is supported for raw calls.
`))
}

func printProjects(w io.Writer, projects []model.Project) {
	if len(projects) == 0 {
		fmt.Fprintln(w, "No projects found.")
		return
	}
	fmt.Fprintf(w, "%-28s  %s\n", "ID", "NAME")
	for _, project := range projects {
		fmt.Fprintf(w, "%-28s  %s\n", project.ID, project.Name)
	}
}

func printTasks(w io.Writer, tasks []model.Task, total int) {
	if len(tasks) == 0 {
		fmt.Fprintln(w, "No tasks found.")
		return
	}
	fmt.Fprintf(w, "Showing %d of %d task(s)\n", len(tasks), total)
	fmt.Fprintf(w, "%-28s  %-10s  %-8s  %-16s  %s\n", "ID", "PROJECT", "PRIORITY", "DUE", "TITLE")
	for _, task := range tasks {
		due := "-"
		if task.DueDate != "" {
			due = task.DueDate
		}
		project := task.ProjectName
		if project == "" {
			project = task.ProjectID
		}
		if len(project) > 10 {
			project = project[:10]
		}
		fmt.Fprintf(w, "%-28s  %-10s  %-8d  %-16s  %s\n", task.ID, project, task.Priority, due, task.Title)
	}
}

func stripProjectRaw(projects []model.Project) []model.Project {
	out := make([]model.Project, len(projects))
	copy(out, projects)
	for i := range out {
		out[i].Raw = nil
	}
	return out
}

func stripTaskRaw(tasks []model.Task) []model.Task {
	out := make([]model.Task, len(tasks))
	copy(out, tasks)
	for i := range out {
		out[i].Raw = nil
	}
	return out
}

func stripSingleTaskRaw(task model.Task) model.Task {
	task.Raw = nil
	return task
}
