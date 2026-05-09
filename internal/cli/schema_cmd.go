package cli

import (
	"fmt"
	"io"
)

type commandSchema struct {
	ID                   string   `json:"id"`
	Title                string   `json:"title"`
	Resource             string   `json:"resource"`
	Operation            string   `json:"operation"`
	Command              string   `json:"command"`
	Aliases              []string `json:"aliases,omitempty"`
	HTTP                 []string `json:"http,omitempty"`
	Status               string   `json:"status"`
	AuthRequired         bool     `json:"authRequired"`
	DryRun               bool     `json:"dryRun"`
	ConfirmationRequired bool     `json:"confirmationRequired"`
	Compact              bool     `json:"compact"`
	Notes                string   `json:"notes,omitempty"`
}

func runSchema(args []string, jsonOut bool, stdout io.Writer, stderr io.Writer) int {
	if len(args) == 0 || args[0] == "-h" || args[0] == "--help" {
		printSchemaHelp(stdout)
		return 0
	}
	switch args[0] {
	case "list":
		return runSchemaList(jsonOut, stdout)
	case "show":
		if len(args) < 2 {
			return fail("schema show", "missing schema id", jsonOut, stdout, stderr)
		}
		return runSchemaShow(args[1], jsonOut, stdout, stderr)
	default:
		return fail("schema", fmt.Sprintf("unknown schema command %q", args[0]), jsonOut, stdout, stderr)
	}
}

func runSchemaList(jsonOut bool, stdout io.Writer) int {
	schemas := didaCommandSchemas()
	meta := map[string]any{"count": len(schemas)}
	data := map[string]any{"schemas": schemas}
	if jsonOut {
		return writeJSON(stdout, envelope{OK: true, Command: "schema list", Meta: meta, Data: data})
	}
	fmt.Fprintf(stdout, "%-24s  %-9s  %-10s  %s\n", "ID", "STATUS", "RESOURCE", "COMMAND")
	for _, schema := range schemas {
		fmt.Fprintf(stdout, "%-24s  %-9s  %-10s  %s\n", schema.ID, schema.Status, schema.Resource, schema.Command)
	}
	return 0
}

func runSchemaShow(id string, jsonOut bool, stdout io.Writer, stderr io.Writer) int {
	for _, schema := range didaCommandSchemas() {
		if schema.ID == id {
			if jsonOut {
				return writeJSON(stdout, envelope{OK: true, Command: "schema show", Data: map[string]any{"schema": schema}})
			}
			fmt.Fprintf(stdout, "ID: %s\nTitle: %s\nResource: %s\nOperation: %s\nCommand: %s\nStatus: %s\n", schema.ID, schema.Title, schema.Resource, schema.Operation, schema.Command, schema.Status)
			if len(schema.HTTP) > 0 {
				fmt.Fprintf(stdout, "HTTP: %v\n", schema.HTTP)
			}
			return 0
		}
	}
	return failTyped("schema show", "not_found", fmt.Sprintf("unknown schema id %q", id), "run: dida schema list --json", jsonOut, stdout, stderr)
}

func didaCommandSchemas() []commandSchema {
	return []commandSchema{
		{ID: "auth.login", Title: "Browser or manual cookie login", Resource: "auth", Operation: "auth", Command: "dida auth login --browser --json", Status: "stable", AuthRequired: false, Notes: "Stores only the Dida365 t cookie under the local config directory."},
		{ID: "auth.status", Title: "Check local auth", Resource: "auth", Operation: "read", Command: "dida auth status --verify --json", HTTP: []string{"GET /batch/check/0"}, Status: "stable", AuthRequired: false},
		{ID: "doctor", Title: "Check CLI, config, auth, and endpoint health", Resource: "system", Operation: "read", Command: "dida doctor --json", Status: "stable", AuthRequired: false},
		{ID: "official.doctor", Title: "Check official dida365 MCP channel", Resource: "official", Operation: "read", Command: "dida official doctor --json", HTTP: []string{"POST https://mcp.dida365.com initialize", "POST https://mcp.dida365.com tools/list"}, Status: "stable", AuthRequired: false, Notes: "Requires DIDA365_TOKEN in the environment."},
		{ID: "official.tools", Title: "List official dida365 MCP tools", Resource: "official", Operation: "read", Command: "dida official tools --limit 100 --json", HTTP: []string{"POST https://mcp.dida365.com tools/list"}, Status: "stable", AuthRequired: false, Compact: true, Notes: "Requires DIDA365_TOKEN in the environment. Default output is compact; use --full for raw schemas."},
		{ID: "official.show", Title: "Show official dida365 MCP tool schema", Resource: "official", Operation: "read", Command: "dida official show <tool-name> --json", HTTP: []string{"POST https://mcp.dida365.com tools/list"}, Status: "stable", AuthRequired: false, Notes: "Requires DIDA365_TOKEN in the environment."},
		{ID: "official.call", Title: "Call an official dida365 MCP tool", Resource: "official", Operation: "write", Command: "dida official call <tool-name> --args-json '{...}' --json", HTTP: []string{"POST https://mcp.dida365.com tools/call"}, Status: "stable", AuthRequired: false, Notes: "Requires DIDA365_TOKEN in the environment. This is the generic entrypoint for all 32 official MCP tools."},
		{ID: "agent.context", Title: "One-call compact context pack", Resource: "agent", Operation: "read", Command: "dida agent context --json", HTTP: []string{"GET /batch/check/0"}, Status: "stable", AuthRequired: true, Compact: true, Notes: "Returns projects, folders, tags, filters, today, upcoming, and quadrants from one sync."},
		{ID: "sync.all", Title: "Full account sync", Resource: "sync", Operation: "read", Command: "dida sync all --json", HTTP: []string{"GET /batch/check/0"}, Status: "stable", AuthRequired: true},
		{ID: "sync.checkpoint", Title: "Incremental account sync", Resource: "sync", Operation: "read", Command: "dida sync checkpoint <checkpoint> --json", HTTP: []string{"GET /batch/check/{checkpoint}"}, Status: "stable", AuthRequired: true, Notes: "Preserves add/update/delete/order/reminder deltas."},
		{ID: "settings.get", Title: "Read user settings", Resource: "settings", Operation: "read", Command: "dida settings get --json", HTTP: []string{"GET /user/preferences/settings", "GET /user/preferences/settings?includeWeb=true"}, Status: "stable", AuthRequired: true, Notes: "Use --include-web to include Web-side preference fields such as smartProjects and shortcut config."},
		{ID: "project.list", Title: "List projects", Resource: "project", Operation: "read", Command: "dida project list --json", HTTP: []string{"GET /batch/check/0"}, Status: "stable", AuthRequired: true},
		{ID: "project.tasks", Title: "List tasks in a project", Resource: "project", Operation: "read", Command: "dida project tasks <project-id> --limit 50 --compact --json", HTTP: []string{"GET /project/{projectId}/tasks"}, Status: "stable", AuthRequired: true, Compact: true},
		{ID: "project.create", Title: "Create project", Resource: "project", Operation: "write", Command: "dida project create --name <name> --dry-run --json", HTTP: []string{"POST /batch/project"}, Status: "stable", AuthRequired: true, DryRun: true},
		{ID: "project.update", Title: "Update project", Resource: "project", Operation: "write", Command: "dida project update <project-id> --name <name> --dry-run --json", HTTP: []string{"POST /batch/project"}, Status: "stable", AuthRequired: true, DryRun: true},
		{ID: "project.delete", Title: "Delete project", Resource: "project", Operation: "delete", Command: "dida project delete <project-id> --yes --json", HTTP: []string{"POST /batch/project"}, Status: "stable", AuthRequired: true, DryRun: true, ConfirmationRequired: true},
		{ID: "folder.list", Title: "List project folders", Resource: "folder", Operation: "read", Command: "dida folder list --json", HTTP: []string{"GET /batch/check/0"}, Status: "stable", AuthRequired: true},
		{ID: "folder.create", Title: "Create project folder", Resource: "folder", Operation: "write", Command: "dida folder create --name <name> --dry-run --json", HTTP: []string{"POST /batch/projectGroup"}, Status: "stable", AuthRequired: true, DryRun: true},
		{ID: "folder.update", Title: "Update project folder", Resource: "folder", Operation: "write", Command: "dida folder update <folder-id> --name <name> --dry-run --json", HTTP: []string{"POST /batch/projectGroup"}, Status: "stable", AuthRequired: true, DryRun: true},
		{ID: "folder.delete", Title: "Delete project folder", Resource: "folder", Operation: "delete", Command: "dida folder delete <folder-id> --yes --json", HTTP: []string{"POST /batch/projectGroup"}, Status: "stable", AuthRequired: true, DryRun: true, ConfirmationRequired: true},
		{ID: "tag.list", Title: "List tags", Resource: "tag", Operation: "read", Command: "dida tag list --json", HTTP: []string{"GET /batch/check/0"}, Status: "stable", AuthRequired: true},
		{ID: "tag.create", Title: "Create tag", Resource: "tag", Operation: "write", Command: "dida tag create <name> --dry-run --json", HTTP: []string{"POST /batch/tag"}, Status: "stable", AuthRequired: true, DryRun: true},
		{ID: "tag.update", Title: "Update tag metadata", Resource: "tag", Operation: "write", Command: "dida tag update <name> --dry-run --json", HTTP: []string{"POST /batch/tag"}, Status: "stable", AuthRequired: true, DryRun: true},
		{ID: "tag.rename", Title: "Rename tag", Resource: "tag", Operation: "write", Command: "dida tag rename <old-name> <new-name> --dry-run --json", HTTP: []string{"PUT /tag/rename"}, Status: "stable", AuthRequired: true, DryRun: true},
		{ID: "tag.merge", Title: "Merge tag associations", Resource: "tag", Operation: "write", Command: "dida tag merge <from-name> <to-name> --yes --json", HTTP: []string{"PUT /tag/merge"}, Status: "stable", AuthRequired: true, DryRun: true, ConfirmationRequired: true, Notes: "Observed endpoint may leave the source tag object present; list tags after merge."},
		{ID: "tag.delete", Title: "Delete tag", Resource: "tag", Operation: "delete", Command: "dida tag delete <name> --yes --json", HTTP: []string{"DELETE /tag?name=..."}, Status: "stable", AuthRequired: true, DryRun: true, ConfirmationRequired: true},
		{ID: "filter.list", Title: "List filters from sync payload", Resource: "filter", Operation: "read", Command: "dida filter list --json", HTTP: []string{"GET /batch/check/0"}, Status: "stable", AuthRequired: true},
		{ID: "column.list", Title: "List kanban columns", Resource: "column", Operation: "read", Command: "dida column list <project-id> --json", Aliases: []string{"dida project columns <project-id> --json"}, HTTP: []string{"GET /column/project/{projectId}"}, Status: "stable", AuthRequired: true},
		{ID: "column.create", Title: "Create kanban column", Resource: "column", Operation: "write", Command: "dida column create --project <project-id> --name <name> --dry-run --json", HTTP: []string{"POST /column"}, Status: "experimental", AuthRequired: true, DryRun: true, Notes: "Update/delete/order remain blocked until /batch/columnProject payloads are verified."},
		{ID: "task.today", Title: "List today's tasks", Resource: "task", Operation: "read", Command: "dida task today --compact --json", Aliases: []string{"dida +today --compact --json"}, HTTP: []string{"GET /batch/check/0"}, Status: "stable", AuthRequired: true, Compact: true},
		{ID: "task.list", Title: "List active tasks", Resource: "task", Operation: "read", Command: "dida task list --filter all --limit 50 --compact --json", HTTP: []string{"GET /batch/check/0"}, Status: "stable", AuthRequired: true, Compact: true},
		{ID: "task.search", Title: "Search tasks locally from sync", Resource: "task", Operation: "read", Command: "dida task search --query <text> --limit 20 --compact --json", HTTP: []string{"GET /batch/check/0"}, Status: "stable", AuthRequired: true, Compact: true},
		{ID: "task.upcoming", Title: "List upcoming tasks", Resource: "task", Operation: "read", Command: "dida task upcoming --days 14 --limit 50 --compact --json", HTTP: []string{"GET /batch/check/0"}, Status: "stable", AuthRequired: true, Compact: true},
		{ID: "task.dueCounts", Title: "Read due-date activity counts", Resource: "task", Operation: "read", Command: "dida task due-counts --json", HTTP: []string{"POST /task/activity/count/all"}, Status: "stable", AuthRequired: true, Notes: "POST read observed in the webapp bundle with action T_DUE; does not mutate tasks."},
		{ID: "task.create", Title: "Create task", Resource: "task", Operation: "write", Command: "dida task create --project <project-id> --title <title> --dry-run --json", HTTP: []string{"POST /batch/task"}, Status: "stable", AuthRequired: true, DryRun: true},
		{ID: "task.update", Title: "Update task", Resource: "task", Operation: "write", Command: "dida task update <task-id> --project <project-id> --dry-run --json", HTTP: []string{"POST /batch/task"}, Status: "stable", AuthRequired: true, DryRun: true},
		{ID: "task.complete", Title: "Complete task", Resource: "task", Operation: "write", Command: "dida task complete <task-id> --project <project-id> --dry-run --json", HTTP: []string{"POST /batch/task"}, Status: "stable", AuthRequired: true, DryRun: true},
		{ID: "task.delete", Title: "Delete task", Resource: "task", Operation: "delete", Command: "dida task delete <task-id> --project <project-id> --yes --json", HTTP: []string{"POST /batch/task"}, Status: "stable", AuthRequired: true, DryRun: true, ConfirmationRequired: true},
		{ID: "task.move", Title: "Move task between projects", Resource: "task", Operation: "write", Command: "dida task move <task-id> --from <project-id> --to <project-id> --dry-run --json", HTTP: []string{"POST /batch/taskProject"}, Status: "stable", AuthRequired: true, DryRun: true},
		{ID: "task.parent", Title: "Set task parent", Resource: "task", Operation: "write", Command: "dida task parent <task-id> --parent <task-id> --project <project-id> --dry-run --json", HTTP: []string{"POST /batch/taskParent"}, Status: "stable", AuthRequired: true, DryRun: true},
		{ID: "comment.list", Title: "List task comments", Resource: "comment", Operation: "read", Command: "dida comment list --project <project-id> --task <task-id> --json", HTTP: []string{"GET /project/{projectId}/task/{taskId}/comments"}, Status: "stable", AuthRequired: true},
		{ID: "comment.create", Title: "Create task comment", Resource: "comment", Operation: "write", Command: "dida comment create --project <project-id> --task <task-id> --text <text> --dry-run --json", HTTP: []string{"POST /project/{projectId}/task/{taskId}/comment"}, Status: "stable", AuthRequired: true, DryRun: true},
		{ID: "comment.update", Title: "Update task comment", Resource: "comment", Operation: "write", Command: "dida comment update --project <project-id> --task <task-id> --comment <comment-id> --text <text> --dry-run --json", HTTP: []string{"PUT /project/{projectId}/task/{taskId}/comment/{commentId}"}, Status: "stable", AuthRequired: true, DryRun: true},
		{ID: "comment.delete", Title: "Delete task comment", Resource: "comment", Operation: "delete", Command: "dida comment delete --project <project-id> --task <task-id> --comment <comment-id> --yes --json", HTTP: []string{"DELETE /project/{projectId}/task/{taskId}/comment/{commentId}"}, Status: "stable", AuthRequired: true, DryRun: true, ConfirmationRequired: true},
		{ID: "completed.list", Title: "List completed tasks", Resource: "completed", Operation: "read", Command: "dida completed list --from YYYY-MM-DD --to YYYY-MM-DD --compact --json", HTTP: []string{"GET /project/all/completed"}, Status: "stable", AuthRequired: true, Compact: true},
		{ID: "closed.list", Title: "List closed-history items", Resource: "closed", Operation: "read", Command: "dida closed list --status 2 --limit 50 --json", HTTP: []string{"GET /project/{projectIds|all}/closed?from=...&to=...&status=..."}, Status: "stable", AuthRequired: true, Notes: "Uses the private closed-history endpoint. Dates are sent as full datetime strings."},
		{ID: "attachment.quota", Title: "Read attachment upload quota", Resource: "attachment", Operation: "read", Command: "dida attachment quota --json", HTTP: []string{"GET /api/v1/attachment/isUnderQuota", "GET /api/v1/attachment/dailyLimit"}, Status: "stable", AuthRequired: true, Notes: "Uses the legacy v1 Web API client observed as _s in the webapp bundle."},
		{ID: "reminder.daily", Title: "Read daily reminder preferences", Resource: "reminder", Operation: "read", Command: "dida reminder daily --json", HTTP: []string{"GET /user/preferences/dailyReminder"}, Status: "stable", AuthRequired: true},
		{ID: "share.contacts", Title: "Read share contacts", Resource: "share", Operation: "read", Command: "dida share contacts --json", HTTP: []string{"GET /share/shareContacts"}, Status: "stable", AuthRequired: true},
		{ID: "share.recentUsers", Title: "Read recent project users", Resource: "share", Operation: "read", Command: "dida share recent-users --json", HTTP: []string{"GET /project/share/recentProjectUsers"}, Status: "stable", AuthRequired: true},
		{ID: "share.projectShares", Title: "Read project sharing members", Resource: "share", Operation: "read", Command: "dida share project shares <project-id> --json", HTTP: []string{"GET /project/{projectId}/shares"}, Status: "stable", AuthRequired: true},
		{ID: "share.projectQuota", Title: "Read project share quota", Resource: "share", Operation: "read", Command: "dida share project quota <project-id> --json", HTTP: []string{"GET /project/{projectId}/share/check-quota"}, Status: "stable", AuthRequired: true},
		{ID: "share.projectInviteURL", Title: "Read project invite URL state", Resource: "share", Operation: "read", Command: "dida share project invite-url <project-id> --json", HTTP: []string{"GET /project/{projectId}/collaboration/invite-url"}, Status: "stable", AuthRequired: true, Notes: "Read-only; invite creation/deletion writes remain intentionally unavailable."},
		{ID: "calendar.subscriptions", Title: "Read calendar subscriptions", Resource: "calendar", Operation: "read", Command: "dida calendar subscriptions --json", HTTP: []string{"GET /calendar/subscription"}, Status: "stable", AuthRequired: true},
		{ID: "calendar.archived", Title: "Read archived calendar events", Resource: "calendar", Operation: "read", Command: "dida calendar archived --json", HTTP: []string{"GET /calendar/archivedEvent"}, Status: "stable", AuthRequired: true},
		{ID: "calendar.thirdAccounts", Title: "Read third-party calendar accounts", Resource: "calendar", Operation: "read", Command: "dida calendar third-accounts --json", HTTP: []string{"GET /calendar/third/accounts"}, Status: "stable", AuthRequired: true, Notes: "May expose connected account metadata; use only when needed."},
		{ID: "stats.general", Title: "Read general account statistics", Resource: "stats", Operation: "read", Command: "dida stats general --json", HTTP: []string{"GET /statistics/general"}, Status: "stable", AuthRequired: true},
		{ID: "template.projectList", Title: "List project templates", Resource: "template", Operation: "read", Command: "dida template project list --limit 50 --json", HTTP: []string{"GET /projectTemplates/all?timestamp=..."}, Status: "stable", AuthRequired: true},
		{ID: "search.all", Title: "Search indexed tasks and comments", Resource: "search", Operation: "read", Command: "dida search all --query <text> --limit 20 --json", HTTP: []string{"GET /search/all?keywords=..."}, Status: "stable", AuthRequired: true, Compact: true, Notes: "The observed working query parameter is keywords; output is compact and bounded by default; use --full for raw search objects."},
		{ID: "user.status", Title: "Read account status", Resource: "user", Operation: "read", Command: "dida user status --json", HTTP: []string{"GET /user/status"}, Status: "stable", AuthRequired: true, Compact: true, Notes: "Default output is compact; use --full only when raw account metadata is required."},
		{ID: "user.profile", Title: "Read account profile", Resource: "user", Operation: "read", Command: "dida user profile --json", HTTP: []string{"GET /user/profile"}, Status: "stable", AuthRequired: true, Compact: true, Notes: "Default output is compact; use --full only when raw profile metadata is required."},
		{ID: "user.sessions", Title: "Read login sessions", Resource: "user", Operation: "read", Command: "dida user sessions --limit 20 --json", HTTP: []string{"GET /user/sessions?lang=..."}, Status: "stable", AuthRequired: true, Compact: true, Notes: "Default output is compact; use --full only when raw session metadata is required."},
		{ID: "pomo.preferences", Title: "Read Pomodoro preferences", Resource: "pomo", Operation: "read", Command: "dida pomo preferences --json", HTTP: []string{"GET /user/preferences/pomodoro"}, Status: "stable", AuthRequired: true},
		{ID: "pomo.list", Title: "List Pomodoro records", Resource: "pomo", Operation: "read", Command: "dida pomo list --from YYYY-MM-DD --to YYYY-MM-DD --limit 50 --json", HTTP: []string{"GET /pomodoros?from=<millis>&to=<millis>"}, Status: "stable", AuthRequired: true},
		{ID: "pomo.timing", Title: "List Pomodoro timing records", Resource: "pomo", Operation: "read", Command: "dida pomo timing --from YYYY-MM-DD --to YYYY-MM-DD --limit 50 --json", HTTP: []string{"GET /pomodoros/timing?from=<millis>&to=<millis>"}, Status: "stable", AuthRequired: true},
		{ID: "pomo.stats", Title: "Read Pomodoro statistics", Resource: "pomo", Operation: "read", Command: "dida pomo stats --json", HTTP: []string{"GET /pomodoros/statistics/generalForDesktop"}, Status: "stable", AuthRequired: true},
		{ID: "pomo.timeline", Title: "List Pomodoro timeline records", Resource: "pomo", Operation: "read", Command: "dida pomo timeline --limit 50 --json", HTTP: []string{"GET /pomodoros/timeline"}, Status: "stable", AuthRequired: true},
		{ID: "pomo.task", Title: "List Pomodoro records for a task", Resource: "pomo", Operation: "read", Command: "dida pomo task --project <project-id> --task <task-id> --json", HTTP: []string{"GET /pomodoros/task?projectId=...&taskId=..."}, Status: "stable", AuthRequired: true},
		{ID: "habit.preferences", Title: "Read habit preferences", Resource: "habit", Operation: "read", Command: "dida habit preferences --json", HTTP: []string{"GET /user/preferences/habit?platform=web"}, Status: "stable", AuthRequired: true},
		{ID: "habit.list", Title: "List habits", Resource: "habit", Operation: "read", Command: "dida habit list --json", HTTP: []string{"GET /habits"}, Status: "stable", AuthRequired: true},
		{ID: "habit.sections", Title: "List habit sections", Resource: "habit", Operation: "read", Command: "dida habit sections --json", HTTP: []string{"GET /habitSections"}, Status: "stable", AuthRequired: true},
		{ID: "habit.checkins", Title: "Read habit check-ins", Resource: "habit", Operation: "read", Command: "dida habit checkins --habit <habit-id> --after-stamp <millis> --json", HTTP: []string{"POST /habitCheckins/query"}, Status: "stable", AuthRequired: true, Notes: "POST read; accepts repeated --habit and optional afterStamp cursor."},
		{ID: "quadrant.list", Title: "Group active tasks into Eisenhower quadrants", Resource: "quadrant", Operation: "read", Command: "dida quadrant list --json", HTTP: []string{"GET /batch/check/0"}, Status: "stable", AuthRequired: true},
		{ID: "raw.get", Title: "Read-only raw Web API probe", Resource: "raw", Operation: "read", Command: "dida raw get /path --api-version v1|v2 --json", HTTP: []string{"GET <path>"}, Status: "stable", AuthRequired: true, Notes: "Raw writes are intentionally unavailable."},
	}
}
