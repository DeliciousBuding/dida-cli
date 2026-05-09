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
  official     Inspect the official dida365 MCP channel
  openapi      Use the official OAuth-based OpenAPI channel
  agent        Agent-oriented context pack
  auth         Manage local cookie auth
  sync         Sync tasks/projects/tags
  settings     Read user preferences
  completed    Read completed task history
  closed       Read closed-history items from the Web API
  attachment   Read attachment quota and upload limits
  reminder     Read reminder preferences
  share        Read sharing and collaboration metadata
  calendar     Read calendar subscription metadata
  stats        Read account statistics
  template     Read project templates
  search       Search across Web API indexed content
  user         Read account and session metadata
  pomo         Read Pomodoro preferences and records
  habit        Read habit preferences, habits, and sections
  quadrant     View active tasks by Eisenhower quadrant
  schema       List machine-readable command contracts
  project      Project discovery and CRUD
  folder       Project folder CRUD
  tag          Tag discovery and CRUD
  filter       Filter discovery
  column       Kanban column discovery and experimental create
  comment      Task comment reads and writes
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
  DIDA_ALLOW_TOKEN_ARG=1 dida auth cookie set --token <token>
`))
}

func printOfficialHelp(w io.Writer) {
	fmt.Fprintln(w, strings.TrimSpace(`
Usage:
  dida official doctor [--json]
  dida official tools [--limit N] [--full] [--json]
  dida official show <tool-name> [--json]
  dida official call <tool-name> [--args-json <json>] [--args-file <file>] [--json]
  dida official project <subcommand> [options] [--json]
  dida official habit <subcommand> [options] [--json]
  dida official focus <subcommand> [options] [--json]

These commands use the official dida365 MCP server and require DIDA365_TOKEN.
`))
}

func printOpenAPIHelp(w io.Writer) {
	fmt.Fprintln(w, strings.TrimSpace(`
Usage:
  dida openapi doctor [--json]
  dida openapi status [--json]
  dida openapi logout [--json]
  dida openapi login [--host HOST] [--port PORT] [--scope SCOPES] [--state VALUE] [--timeout SECONDS] [--no-open] [--json]
  dida openapi auth-url [--redirect-uri URL] [--scope SCOPES] [--state VALUE] [--json]
  dida openapi listen-callback [--host HOST] [--port PORT] [--json]
  dida openapi exchange-code --code CODE [--redirect-uri URL] [--scope SCOPES] [--json]
  dida openapi project list [--json]

These commands use the official OAuth-based OpenAPI channel.
They require DIDA365_OPENAPI_CLIENT_ID and DIDA365_OPENAPI_CLIENT_SECRET.
`))
}

func printAgentHelp(w io.Writer) {
	fmt.Fprintln(w, strings.TrimSpace(`
Usage:
  dida agent context [--days N] [--limit N] [--compact|--full] [--json]

Agent context performs one full sync and returns compact projects, folders,
tags, filters, today, upcoming, and quadrant views in a single JSON envelope.
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
  DIDA_ALLOW_TOKEN_ARG=1 dida auth cookie set --token <token>

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
  dida settings get [--include-web] [--json]
`))
}

func printCompletedHelp(w io.Writer) {
	fmt.Fprintln(w, strings.TrimSpace(`
Usage:
  dida completed today [--compact] [--json]
  dida completed yesterday [--compact] [--json]
  dida completed week [--compact] [--json]
  dida completed list [--from YYYY-MM-DD] [--to YYYY-MM-DD] [--limit N] [--compact] [--json]
`))
}

func printClosedHelp(w io.Writer) {
	fmt.Fprintln(w, strings.TrimSpace(`
Usage:
  dida closed list [--project <project-id>] [--status N] [--from YYYY-MM-DD] [--to YYYY-MM-DD] [--completed-user <user-id>] [--limit N] [--json]
`))
}

func printAttachmentHelp(w io.Writer) {
	fmt.Fprintln(w, strings.TrimSpace(`
Usage:
  dida attachment quota [--json]
`))
}

func printReminderHelp(w io.Writer) {
	fmt.Fprintln(w, strings.TrimSpace(`
Usage:
  dida reminder daily [--json]
`))
}

func printShareHelp(w io.Writer) {
	fmt.Fprintln(w, strings.TrimSpace(`
Usage:
  dida share contacts [--json]
  dida share recent-users [--json]
  dida share project shares <project-id> [--json]
  dida share project quota <project-id> [--json]
  dida share project invite-url <project-id> [--json]

These commands are read-only. Invite creation, deletion, and user invitation
writes are not exposed until collaboration payloads and rollback paths are verified.
`))
}

func printCalendarHelp(w io.Writer) {
	fmt.Fprintln(w, strings.TrimSpace(`
Usage:
  dida calendar subscriptions [--json]
  dida calendar archived [--json]
  dida calendar third-accounts [--json]
`))
}

func printStatsHelp(w io.Writer) {
	fmt.Fprintln(w, strings.TrimSpace(`
Usage:
  dida stats general [--json]
`))
}

func printTemplateHelp(w io.Writer) {
	fmt.Fprintln(w, strings.TrimSpace(`
Usage:
  dida template project list [--timestamp N] [--limit N] [--json]
`))
}

func printSearchHelp(w io.Writer) {
	fmt.Fprintln(w, strings.TrimSpace(`
Usage:
  dida search all --query <text> [--limit N] [--full] [--json]
`))
}

func printUserHelp(w io.Writer) {
	fmt.Fprintln(w, strings.TrimSpace(`
Usage:
  dida user status [--full] [--json]
  dida user profile [--full] [--json]
  dida user sessions [--lang <locale>] [--limit N] [--full] [--json]

Default output is compact and only keeps the fields that are usually useful in CLI/agent flows.
Use --full when you actually need the raw Web API response.
`))
}

func printPomoHelp(w io.Writer) {
	fmt.Fprintln(w, strings.TrimSpace(`
Usage:
  dida pomo preferences [--json]
  dida pomo list [--from YYYY-MM-DD] [--to YYYY-MM-DD] [--limit N] [--json]
  dida pomo timing [--from YYYY-MM-DD] [--to YYYY-MM-DD] [--limit N] [--json]
  dida pomo task --project <project-id> --task <task-id> [--json]
  dida pomo stats [--json]
  dida pomo timeline [--to <cursor>] [--limit N] [--json]
`))
}

func printHabitHelp(w io.Writer) {
	fmt.Fprintln(w, strings.TrimSpace(`
Usage:
  dida habit preferences [--json]
  dida habit list [--json]
  dida habit sections [--json]
  dida habit checkins [--habit <habit-id>] [--after-stamp <millis>] [--json]
`))
}

func printProjectHelp(w io.Writer) {
	fmt.Fprintln(w, strings.TrimSpace(`
Usage:
  dida project list [--json]
  dida project create --name <name> [--group <folder-id>] [--dry-run] [--json]
  dida project update <project-id> [--name <name>] [--group <folder-id>] [--dry-run] [--json]
  dida project delete <project-id> --yes [--dry-run] [--json]
  dida project tasks <project-id> [--limit N] [--compact] [--json]
  dida project columns <project-id> [--json]
`))
}

func printTaskHelp(w io.Writer) {
	fmt.Fprintln(w, strings.TrimSpace(`
Usage:
  dida task today [--json] [--limit N] [--compact]
  dida task list [--json] [--filter today|all] [--limit N] [--compact]
  dida task search --query <text> [--limit N] [--compact] [--json]
  dida task upcoming [--days N] [--limit N] [--compact] [--json]
  dida task due-counts [--json]
  dida task get <task-id> [--json]
  dida task create --project <project-id> --title <title> [task fields...] [--dry-run] [--json]
  dida task update <task-id> --project <project-id> [task fields...] [--dry-run] [--json]
  dida task complete <task-id> --project <project-id> [--dry-run] [--json]
  dida task delete <task-id> --project <project-id> --yes [--dry-run] [--json]
  dida task move <task-id> --from <project-id> --to <project-id> [--dry-run] [--json]
  dida task parent <task-id> --parent <task-id> --project <project-id> [--dry-run] [--json]
  dida +today [--json] [--limit N] [--compact]

Use --compact (or --brief) for agent reads that should omit large text, checklist,
reminder, and raw fields.

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

func printFilterHelp(w io.Writer) {
	fmt.Fprintln(w, strings.TrimSpace(`
Usage:
  dida filter list [--json]

Filters are read from the sync payload. Filter writes are not exposed until the
private /batch/filter payload shape is verified.
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

func printCommentHelp(w io.Writer) {
	fmt.Fprintln(w, strings.TrimSpace(`
Usage:
  dida comment list --project <project-id> --task <task-id> [--json]
  dida comment create --project <project-id> --task <task-id> --text <text> [--dry-run] [--json]
  dida comment update --project <project-id> --task <task-id> --comment <comment-id> --text <text> [--dry-run] [--json]
  dida comment delete --project <project-id> --task <task-id> --comment <comment-id> --yes [--dry-run] [--json]

Comment attachments are not exposed until the multipart upload flow is verified.
`))
}

func printRawHelp(w io.Writer) {
	fmt.Fprintln(w, strings.TrimSpace(`
Usage:
  dida raw get <path> [--api-version v1|v2] [--json]

Only GET is supported for raw calls.
`))
}

func printSchemaHelp(w io.Writer) {
	fmt.Fprintln(w, strings.TrimSpace(`
Usage:
  dida schema list [--json]
  dida schema show <schema-id> [--json]

Schema output is local and does not require auth. Use it to discover command
contracts, safety flags, endpoint coverage, and compact-output support.
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
