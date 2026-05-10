package cli

import (
	"fmt"
	"io"
	"strings"
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

type compactCommandSchema struct {
	ID                   string `json:"id"`
	Resource             string `json:"resource"`
	Operation            string `json:"operation"`
	Command              string `json:"command"`
	Status               string `json:"status"`
	AuthRequired         bool   `json:"authRequired"`
	DryRun               bool   `json:"dryRun"`
	ConfirmationRequired bool   `json:"confirmationRequired"`
	Compact              bool   `json:"compact"`
}

type schemaListOptions struct {
	Compact   bool
	Resource  string
	Operation string
	Status    string
}

func runSchema(args []string, jsonOut bool, stdout io.Writer, stderr io.Writer) int {
	if len(args) == 0 || args[0] == "-h" || args[0] == "--help" {
		printSchemaHelp(stdout)
		return 0
	}
	switch args[0] {
	case "list":
		return runSchemaList(args[1:], jsonOut, stdout, stderr)
	case "show":
		if len(args) < 2 {
			return fail("schema show", "missing schema id", jsonOut, stdout, stderr)
		}
		return runSchemaShow(args[1], jsonOut, stdout, stderr)
	default:
		return fail("schema", fmt.Sprintf("unknown schema command %q", args[0]), jsonOut, stdout, stderr)
	}
}

func runSchemaList(args []string, jsonOut bool, stdout io.Writer, stderr io.Writer) int {
	opts, err := parseSchemaListOptions(args)
	if err != nil {
		return failTyped("schema list", "validation", err.Error(), "run: dida schema list --compact --json", jsonOut, stdout, stderr)
	}
	schemas := didaCommandSchemas()
	schemas = filterCommandSchemas(schemas, opts)
	meta := map[string]any{"count": len(schemas), "compact": opts.Compact}
	if opts.Resource != "" {
		meta["resource"] = opts.Resource
	}
	if opts.Operation != "" {
		meta["operation"] = opts.Operation
	}
	if opts.Status != "" {
		meta["status"] = opts.Status
	}
	data := map[string]any{}
	if opts.Compact {
		data["schemas"] = compactCommandSchemas(schemas)
	} else {
		data["schemas"] = schemas
	}
	if jsonOut {
		return writeJSON(stdout, envelope{OK: true, Command: "schema list", Meta: meta, Data: data})
	}
	fmt.Fprintf(stdout, "%-24s  %-9s  %-10s  %s\n", "ID", "STATUS", "RESOURCE", "COMMAND")
	for _, schema := range schemas {
		fmt.Fprintf(stdout, "%-24s  %-9s  %-10s  %s\n", schema.ID, schema.Status, schema.Resource, schema.Command)
	}
	return 0
}

func parseSchemaListOptions(args []string) (schemaListOptions, error) {
	opts := schemaListOptions{}
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--compact", "--brief":
			opts.Compact = true
		case "--resource":
			if i+1 >= len(args) || strings.HasPrefix(args[i+1], "-") {
				return opts, fmt.Errorf("--resource requires a value")
			}
			opts.Resource = args[i+1]
			i++
		case "--operation":
			if i+1 >= len(args) || strings.HasPrefix(args[i+1], "-") {
				return opts, fmt.Errorf("--operation requires a value")
			}
			opts.Operation = args[i+1]
			i++
		case "--status":
			if i+1 >= len(args) || strings.HasPrefix(args[i+1], "-") {
				return opts, fmt.Errorf("--status requires a value")
			}
			opts.Status = args[i+1]
			i++
		default:
			return opts, fmt.Errorf("unknown schema list option %q", args[i])
		}
	}
	return opts, nil
}

func filterCommandSchemas(schemas []commandSchema, opts schemaListOptions) []commandSchema {
	if opts.Resource == "" && opts.Operation == "" && opts.Status == "" {
		return schemas
	}
	out := make([]commandSchema, 0, len(schemas))
	for _, schema := range schemas {
		if opts.Resource != "" && schema.Resource != opts.Resource {
			continue
		}
		if opts.Operation != "" && schema.Operation != opts.Operation {
			continue
		}
		if opts.Status != "" && schema.Status != opts.Status {
			continue
		}
		out = append(out, schema)
	}
	return out
}

func compactCommandSchemas(schemas []commandSchema) []compactCommandSchema {
	out := make([]compactCommandSchema, len(schemas))
	for i, schema := range schemas {
		out[i] = compactCommandSchema{
			ID:                   schema.ID,
			Resource:             schema.Resource,
			Operation:            schema.Operation,
			Command:              schema.Command,
			Status:               schema.Status,
			AuthRequired:         schema.AuthRequired,
			DryRun:               schema.DryRun,
			ConfirmationRequired: schema.ConfirmationRequired,
			Compact:              schema.Compact,
		}
	}
	return out
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
	schemas := []commandSchema{
		{ID: "auth.login", Title: "Browser or manual cookie login", Resource: "auth", Operation: "auth", Command: "dida auth login --browser --json", Status: "stable", AuthRequired: false, Notes: "Stores only the Dida365 t cookie under the local config directory."},
		{ID: "auth.status", Title: "Check local auth", Resource: "auth", Operation: "read", Command: "dida auth status --verify --json", HTTP: []string{"GET /batch/check/0"}, Status: "stable", AuthRequired: false},
		{ID: "doctor", Title: "Check CLI, config, auth, and optional endpoint health", Resource: "system", Operation: "read", Command: "dida doctor --verify --json", HTTP: []string{"GET /batch/check/0 when --verify is set"}, Status: "stable", AuthRequired: false, Notes: "Without --verify this command is local-only and reports network_check: not_run."},
		{ID: "schema.list", Title: "List local command contracts", Resource: "schema", Operation: "read", Command: "dida schema list --compact --json", Status: "stable", AuthRequired: false, Compact: true, Notes: "Local-only command index. Use schema show for full details, HTTP surfaces, and notes."},
		{ID: "schema.show", Title: "Show one local command contract", Resource: "schema", Operation: "read", Command: "dida schema show <schema-id> --json", Status: "stable", AuthRequired: false, Notes: "Local-only detailed schema for one command."},
		{ID: "channel.list", Title: "Explain API channel selection and auth boundaries", Resource: "channel", Operation: "read", Command: "dida channel list --json", Status: "stable", AuthRequired: false, Notes: "Local-only guide for choosing Web API, Official MCP, or Official OpenAPI without mixing auth models."},
		{ID: "official.doctor", Title: "Check official dida365 MCP channel", Resource: "official", Operation: "read", Command: "dida official doctor --json", HTTP: []string{"POST https://mcp.dida365.com initialize", "POST https://mcp.dida365.com tools/list"}, Status: "stable", AuthRequired: false, Notes: "Requires DIDA365_TOKEN or saved local official token config."},
		{ID: "official.tokenStatus", Title: "Check saved official MCP token config", Resource: "official", Operation: "auth", Command: "dida official token status --json", Status: "stable", AuthRequired: false, Notes: "Reports whether DIDA365_TOKEN or saved local token config is available without printing the full token."},
		{ID: "official.tokenSet", Title: "Save official MCP token config", Resource: "official", Operation: "auth", Command: "dida official token set --token-stdin --json", Status: "stable", AuthRequired: false, Notes: "Stores the official MCP token locally. DIDA365_TOKEN still takes precedence."},
		{ID: "official.tokenClear", Title: "Clear saved official MCP token config", Resource: "official", Operation: "auth", Command: "dida official token clear --json", Status: "stable", AuthRequired: false},
		{ID: "official.tools", Title: "List official dida365 MCP tools", Resource: "official", Operation: "read", Command: "dida official tools --limit 100 --json", HTTP: []string{"POST https://mcp.dida365.com tools/list"}, Status: "stable", AuthRequired: false, Compact: true, Notes: "Requires DIDA365_TOKEN or saved local official token config. Default output is compact; use --full for raw schemas."},
		{ID: "official.show", Title: "Show official dida365 MCP tool schema", Resource: "official", Operation: "read", Command: "dida official show <tool-name> --json", HTTP: []string{"POST https://mcp.dida365.com tools/list"}, Status: "stable", AuthRequired: false, Notes: "Requires DIDA365_TOKEN or saved local official token config."},
		{ID: "official.call", Title: "Call an official dida365 MCP tool", Resource: "official", Operation: "write", Command: "dida official call <tool-name> --args-json '{...}' --json", HTTP: []string{"POST https://mcp.dida365.com tools/call"}, Status: "stable", AuthRequired: false, Notes: "Requires DIDA365_TOKEN or saved local official token config. Prefer read-only exploration; write-capable tools have no dry-run here, so use first-class wrappers or explicit operator approval."},
		{ID: "official.project.list", Title: "List projects through official MCP", Resource: "official-project", Operation: "read", Command: "dida official project list --json", HTTP: []string{"POST https://mcp.dida365.com tools/call list_projects"}, Status: "stable", AuthRequired: false, Notes: "Requires DIDA365_TOKEN or saved local official token config."},
		{ID: "official.project.get", Title: "Read one project through official MCP", Resource: "official-project", Operation: "read", Command: "dida official project get <project-id> --json", HTTP: []string{"POST https://mcp.dida365.com tools/call get_project_by_id"}, Status: "stable", AuthRequired: false, Notes: "Requires DIDA365_TOKEN or saved local official token config."},
		{ID: "official.project.data", Title: "Read project and undone tasks through official MCP", Resource: "official-project", Operation: "read", Command: "dida official project data <project-id> --json", HTTP: []string{"POST https://mcp.dida365.com tools/call get_project_with_undone_tasks"}, Status: "stable", AuthRequired: false, Compact: true, Notes: "Requires DIDA365_TOKEN or saved local official token config."},
		{ID: "official.task.get", Title: "Read one task through official MCP", Resource: "official-task", Operation: "read", Command: "dida official task get <task-id> --project <project-id> --json", HTTP: []string{"POST https://mcp.dida365.com tools/call get_task_by_id", "POST https://mcp.dida365.com tools/call get_task_in_project when --project is set"}, Status: "stable", AuthRequired: false, Notes: "Requires DIDA365_TOKEN or saved local official token config."},
		{ID: "official.task.search", Title: "Search tasks through official MCP", Resource: "official-task", Operation: "read", Command: "dida official task search --query <text> --json", HTTP: []string{"POST https://mcp.dida365.com tools/call search_task"}, Status: "stable", AuthRequired: false, Notes: "Requires DIDA365_TOKEN or saved local official token config."},
		{ID: "official.task.query", Title: "List undone tasks through official MCP time query", Resource: "official-task", Operation: "read", Command: "dida official task query --query today --json", HTTP: []string{"POST https://mcp.dida365.com tools/call list_undone_tasks_by_time_query"}, Status: "stable", AuthRequired: false, Compact: true, Notes: "Requires DIDA365_TOKEN or saved local official token config. Sends the query as query_command to the upstream MCP tool."},
		{ID: "official.task.undone", Title: "List undone tasks by date through official MCP", Resource: "official-task", Operation: "read", Command: "dida official task undone --start RFC3339 --end RFC3339 --json", HTTP: []string{"POST https://mcp.dida365.com tools/call list_undone_tasks_by_date"}, Status: "stable", AuthRequired: false, Compact: true, Notes: "Requires DIDA365_TOKEN or saved local official token config."},
		{ID: "official.task.filter", Title: "Filter tasks through official MCP", Resource: "official-task", Operation: "read", Command: "dida official task filter --project <project-id> --status 0 --json", HTTP: []string{"POST https://mcp.dida365.com tools/call filter_tasks"}, Status: "stable", AuthRequired: false, Compact: true, Notes: "Requires DIDA365_TOKEN or saved local official token config."},
		{ID: "official.task.batchAdd", Title: "Batch add tasks through official MCP", Resource: "official-task", Operation: "write", Command: "dida official task batch-add --args-json '{...}' --dry-run --json", HTTP: []string{"POST https://mcp.dida365.com tools/call batch_add_tasks"}, Status: "stable", AuthRequired: false, DryRun: true, Notes: "Requires DIDA365_TOKEN or saved local official token config unless --dry-run is used. Payload follows the upstream MCP tool schema."},
		{ID: "official.task.batchUpdate", Title: "Batch update tasks through official MCP", Resource: "official-task", Operation: "write", Command: "dida official task batch-update --args-json '{...}' --dry-run --json", HTTP: []string{"POST https://mcp.dida365.com tools/call batch_update_tasks"}, Status: "stable", AuthRequired: false, DryRun: true, Notes: "Requires DIDA365_TOKEN or saved local official token config unless --dry-run is used. Payload follows the upstream MCP tool schema."},
		{ID: "official.task.completeProject", Title: "Complete tasks in one project through official MCP", Resource: "official-task", Operation: "write", Command: "dida official task complete-project --project <project-id> --task <task-id> --dry-run --json", HTTP: []string{"POST https://mcp.dida365.com tools/call complete_tasks_in_project"}, Status: "stable", AuthRequired: false, DryRun: true, Notes: "Requires DIDA365_TOKEN or saved local official token config unless --dry-run is used. Supports repeated --task or comma-separated --tasks."},
		{ID: "official.habit.list", Title: "List habits through official MCP", Resource: "official-habit", Operation: "read", Command: "dida official habit list --json", HTTP: []string{"POST https://mcp.dida365.com tools/call list_habits"}, Status: "stable", AuthRequired: false, Notes: "Requires DIDA365_TOKEN or saved local official token config."},
		{ID: "official.habit.sections", Title: "List habit sections through official MCP", Resource: "official-habit", Operation: "read", Command: "dida official habit sections --json", HTTP: []string{"POST https://mcp.dida365.com tools/call list_habit_sections"}, Status: "stable", AuthRequired: false, Notes: "Requires DIDA365_TOKEN or saved local official token config."},
		{ID: "official.habit.get", Title: "Read one habit through official MCP", Resource: "official-habit", Operation: "read", Command: "dida official habit get <habit-id> --json", HTTP: []string{"POST https://mcp.dida365.com tools/call get_habit"}, Status: "stable", AuthRequired: false, Notes: "Requires DIDA365_TOKEN or saved local official token config."},
		{ID: "official.habit.create", Title: "Create habit through official MCP", Resource: "official-habit", Operation: "write", Command: "dida official habit create --args-json '{...}' --dry-run --json", HTTP: []string{"POST https://mcp.dida365.com tools/call create_habit"}, Status: "stable", AuthRequired: false, DryRun: true, Notes: "Requires DIDA365_TOKEN or saved local official token config unless --dry-run is used. Payload follows the official MCP tool schema."},
		{ID: "official.habit.update", Title: "Update habit through official MCP", Resource: "official-habit", Operation: "write", Command: "dida official habit update <habit-id> --args-json '{...}' --dry-run --json", HTTP: []string{"POST https://mcp.dida365.com tools/call update_habit"}, Status: "stable", AuthRequired: false, DryRun: true, Notes: "Requires DIDA365_TOKEN or saved local official token config unless --dry-run is used. Payload follows the official MCP tool schema."},
		{ID: "official.habit.checkin", Title: "Upsert a habit check-in through official MCP", Resource: "official-habit", Operation: "write", Command: "dida official habit checkin <habit-id> --date YYYY-MM-DD --value 1 --dry-run --json", HTTP: []string{"POST https://mcp.dida365.com tools/call upsert_habit_checkins"}, Status: "stable", AuthRequired: false, DryRun: true, Notes: "Requires DIDA365_TOKEN or saved local official token config unless --dry-run is used."},
		{ID: "official.habit.checkins", Title: "Read habit check-ins through official MCP", Resource: "official-habit", Operation: "read", Command: "dida official habit checkins --habit-ids <ids> --from YYYYMMDD --to YYYYMMDD --json", HTTP: []string{"POST https://mcp.dida365.com tools/call get_habit_checkins"}, Status: "stable", AuthRequired: false, Notes: "Requires DIDA365_TOKEN or saved local official token config."},
		{ID: "official.focus.get", Title: "Read one focus record through official MCP", Resource: "official-focus", Operation: "read", Command: "dida official focus get <focus-id> --type 0|1 --json", HTTP: []string{"POST https://mcp.dida365.com tools/call get_focus"}, Status: "stable", AuthRequired: false, Notes: "Requires DIDA365_TOKEN or saved local official token config. Type 0 is Pomodoro; type 1 is timing."},
		{ID: "official.focus.list", Title: "List focus records through official MCP", Resource: "official-focus", Operation: "read", Command: "dida official focus list --from-time RFC3339 --to-time RFC3339 --type 0|1 --json", HTTP: []string{"POST https://mcp.dida365.com tools/call get_focuses_by_time"}, Status: "stable", AuthRequired: false, Notes: "Requires DIDA365_TOKEN or saved local official token config. Type 0 is Pomodoro; type 1 is timing."},
		{ID: "official.focus.delete", Title: "Delete focus record through official MCP", Resource: "official-focus", Operation: "delete", Command: "dida official focus delete <focus-id> --type 0|1 --yes --dry-run --json", HTTP: []string{"POST https://mcp.dida365.com tools/call delete_focus"}, Status: "stable", AuthRequired: false, DryRun: true, ConfirmationRequired: true, Notes: "Requires DIDA365_TOKEN or saved local official token config unless --dry-run is used. Use only when the target focus record is known."},
		{ID: "openapi.doctor", Title: "Check official OAuth OpenAPI setup", Resource: "openapi", Operation: "read", Command: "dida openapi doctor --json", HTTP: []string{"GET/POST oauth metadata only"}, Status: "stable", AuthRequired: false, Notes: "Uses DIDA365_OPENAPI_CLIENT_ID / DIDA365_OPENAPI_CLIENT_SECRET or saved local client config."},
		{ID: "openapi.clientStatus", Title: "Check saved official OpenAPI client config", Resource: "openapi", Operation: "read", Command: "dida openapi client status --json", Status: "stable", AuthRequired: false, Notes: "Reports whether local OpenAPI client config is available without printing the secret."},
		{ID: "openapi.clientSet", Title: "Save official OpenAPI client config", Resource: "openapi", Operation: "auth", Command: "dida openapi client set --id <client-id> --secret-stdin --json", Status: "stable", AuthRequired: false, Notes: "Stores client id and client secret locally for OAuth commands. Environment variables still take precedence."},
		{ID: "openapi.clientClear", Title: "Clear saved official OpenAPI client config", Resource: "openapi", Operation: "auth", Command: "dida openapi client clear --json", Status: "stable", AuthRequired: false},
		{ID: "openapi.status", Title: "Check saved official OpenAPI OAuth token", Resource: "openapi", Operation: "auth", Command: "dida openapi status --json", Status: "stable", AuthRequired: false, Notes: "Reports whether a local OAuth access token is available without printing it."},
		{ID: "openapi.authUrl", Title: "Generate official OpenAPI authorization URL", Resource: "openapi", Operation: "auth", Command: "dida openapi auth-url --json", HTTP: []string{"GET https://dida365.com/oauth/authorize"}, Status: "stable", AuthRequired: false},
		{ID: "openapi.listenCallback", Title: "Listen for one official OpenAPI OAuth callback", Resource: "openapi", Operation: "auth", Command: "dida openapi listen-callback --json", HTTP: []string{"GET local callback /callback?code=...&state=..."}, Status: "stable", AuthRequired: false, Notes: "Local-only helper for manual no-browser OAuth flows."},
		{ID: "openapi.exchangeCode", Title: "Exchange OAuth authorization code for access token", Resource: "openapi", Operation: "auth", Command: "dida openapi exchange-code --code <code> --json", HTTP: []string{"POST https://dida365.com/oauth/token"}, Status: "stable", AuthRequired: false},
		{ID: "openapi.login", Title: "Complete browser-based official OpenAPI OAuth login", Resource: "openapi", Operation: "auth", Command: "dida openapi login --browser --json", HTTP: []string{"GET https://dida365.com/oauth/authorize", "GET local callback /callback", "POST https://dida365.com/oauth/token"}, Status: "stable", AuthRequired: false, Notes: "Requires client config and a developer app redirect URL matching the reported local callback."},
		{ID: "openapi.logout", Title: "Clear saved official OpenAPI OAuth token", Resource: "openapi", Operation: "auth", Command: "dida openapi logout --json", Status: "stable", AuthRequired: false, Notes: "Removes the saved local OpenAPI access token."},
		{ID: "openapi.projectList", Title: "List projects through official OpenAPI", Resource: "openapi", Operation: "read", Command: "dida openapi project list --json", HTTP: []string{"GET /open/v1/project"}, Status: "stable", AuthRequired: false, Notes: "Requires an already saved OAuth access token."},
		{ID: "openapi.projectGet", Title: "Get project through official OpenAPI", Resource: "openapi", Operation: "read", Command: "dida openapi project get <project-id> --json", HTTP: []string{"GET /open/v1/project/{projectId}"}, Status: "stable", AuthRequired: false, Notes: "Requires an already saved OAuth access token."},
		{ID: "openapi.projectData", Title: "Get project data through official OpenAPI", Resource: "openapi", Operation: "read", Command: "dida openapi project data <project-id> --json", HTTP: []string{"GET /open/v1/project/{projectId}/data"}, Status: "stable", AuthRequired: false, Notes: "Requires an already saved OAuth access token. Project id may be inbox."},
		{ID: "openapi.projectCreate", Title: "Create project through official OpenAPI", Resource: "openapi-project", Operation: "write", Command: "dida openapi project create --args-json '{...}' --dry-run --json", HTTP: []string{"POST /open/v1/project"}, Status: "stable", AuthRequired: false, DryRun: true, Notes: "Requires an already saved OAuth access token unless --dry-run is used."},
		{ID: "openapi.projectUpdate", Title: "Update project through official OpenAPI", Resource: "openapi-project", Operation: "write", Command: "dida openapi project update <project-id> --args-json '{...}' --dry-run --json", HTTP: []string{"POST /open/v1/project/{projectId}"}, Status: "stable", AuthRequired: false, DryRun: true, Notes: "Requires an already saved OAuth access token unless --dry-run is used."},
		{ID: "openapi.projectDelete", Title: "Delete project through official OpenAPI", Resource: "openapi-project", Operation: "delete", Command: "dida openapi project delete <project-id> --dry-run --yes --json", HTTP: []string{"DELETE /open/v1/project/{projectId}"}, Status: "stable", AuthRequired: false, DryRun: true, ConfirmationRequired: true, Notes: "Requires an already saved OAuth access token. Preview with --dry-run first."},
		{ID: "openapi.taskGet", Title: "Get task through official OpenAPI", Resource: "openapi-task", Operation: "read", Command: "dida openapi task get --project <project-id> --task <task-id> --json", HTTP: []string{"GET /open/v1/project/{projectId}/task/{taskId}"}, Status: "stable", AuthRequired: false, Notes: "Requires an already saved OAuth access token."},
		{ID: "openapi.taskCreate", Title: "Create task through official OpenAPI", Resource: "openapi-task", Operation: "write", Command: "dida openapi task create --args-json '{...}' --dry-run --json", HTTP: []string{"POST /open/v1/task"}, Status: "stable", AuthRequired: false, DryRun: true, Notes: "Requires an already saved OAuth access token."},
		{ID: "openapi.taskUpdate", Title: "Update task through official OpenAPI", Resource: "openapi-task", Operation: "write", Command: "dida openapi task update <task-id> --args-json '{...}' --dry-run --json", HTTP: []string{"POST /open/v1/task/{taskId}"}, Status: "stable", AuthRequired: false, DryRun: true, Notes: "Requires an already saved OAuth access token."},
		{ID: "openapi.taskComplete", Title: "Complete task through official OpenAPI", Resource: "openapi-task", Operation: "write", Command: "dida openapi task complete --project <project-id> --task <task-id> --dry-run --json", HTTP: []string{"POST /open/v1/project/{projectId}/task/{taskId}/complete"}, Status: "stable", AuthRequired: false, DryRun: true, Notes: "Requires an already saved OAuth access token."},
		{ID: "openapi.taskDelete", Title: "Delete task through official OpenAPI", Resource: "openapi-task", Operation: "delete", Command: "dida openapi task delete --project <project-id> --task <task-id> --dry-run --yes --json", HTTP: []string{"DELETE /open/v1/project/{projectId}/task/{taskId}"}, Status: "stable", AuthRequired: false, DryRun: true, ConfirmationRequired: true, Notes: "Requires an already saved OAuth access token."},
		{ID: "openapi.taskMove", Title: "Move tasks through official OpenAPI", Resource: "openapi-task", Operation: "write", Command: "dida openapi task move --args-json '[...]' --dry-run --json", HTTP: []string{"POST /open/v1/task/move"}, Status: "stable", AuthRequired: false, DryRun: true, Notes: "Requires an already saved OAuth access token."},
		{ID: "openapi.taskCompleted", Title: "List completed tasks through official OpenAPI", Resource: "openapi-task", Operation: "read", Command: "dida openapi task completed --args-json '{...}' --json", HTTP: []string{"POST /open/v1/task/completed"}, Status: "stable", AuthRequired: false, Notes: "Requires an already saved OAuth access token."},
		{ID: "openapi.taskFilter", Title: "Filter tasks through official OpenAPI", Resource: "openapi-task", Operation: "read", Command: "dida openapi task filter --args-json '{...}' --json", HTTP: []string{"POST /open/v1/task/filter"}, Status: "stable", AuthRequired: false, Notes: "Requires an already saved OAuth access token."},
		{ID: "openapi.focusGet", Title: "Get focus record through official OpenAPI", Resource: "openapi-focus", Operation: "read", Command: "dida openapi focus get <focus-id> --type 0 --json", HTTP: []string{"GET /open/v1/focus/{focusId}?type=0"}, Status: "stable", AuthRequired: false, Notes: "Requires an already saved OAuth access token."},
		{ID: "openapi.focusList", Title: "List focus records through official OpenAPI", Resource: "openapi-focus", Operation: "read", Command: "dida openapi focus list --from TIME --to TIME --type 1 --json", HTTP: []string{"GET /open/v1/focus?from=...&to=...&type=..."}, Status: "stable", AuthRequired: false, Notes: "Requires an already saved OAuth access token. Time range above 30 days may be adjusted by the server."},
		{ID: "openapi.focusDelete", Title: "Delete focus record through official OpenAPI", Resource: "openapi-focus", Operation: "delete", Command: "dida openapi focus delete <focus-id> --type 0 --dry-run --yes --json", HTTP: []string{"DELETE /open/v1/focus/{focusId}?type=0"}, Status: "stable", AuthRequired: false, DryRun: true, ConfirmationRequired: true, Notes: "Requires an already saved OAuth access token."},
		{ID: "openapi.habitList", Title: "List habits through official OpenAPI", Resource: "openapi-habit", Operation: "read", Command: "dida openapi habit list --json", HTTP: []string{"GET /open/v1/habit"}, Status: "stable", AuthRequired: false, Notes: "Requires an already saved OAuth access token."},
		{ID: "openapi.habitGet", Title: "Get habit through official OpenAPI", Resource: "openapi-habit", Operation: "read", Command: "dida openapi habit get <habit-id> --json", HTTP: []string{"GET /open/v1/habit/{habitId}"}, Status: "stable", AuthRequired: false, Notes: "Requires an already saved OAuth access token."},
		{ID: "openapi.habitCreate", Title: "Create habit through official OpenAPI", Resource: "openapi-habit", Operation: "write", Command: "dida openapi habit create --args-json '{...}' --dry-run --json", HTTP: []string{"POST /open/v1/habit"}, Status: "stable", AuthRequired: false, DryRun: true, Notes: "Requires an already saved OAuth access token."},
		{ID: "openapi.habitUpdate", Title: "Update habit through official OpenAPI", Resource: "openapi-habit", Operation: "write", Command: "dida openapi habit update <habit-id> --args-json '{...}' --dry-run --json", HTTP: []string{"POST /open/v1/habit/{habitId}"}, Status: "stable", AuthRequired: false, DryRun: true, Notes: "Requires an already saved OAuth access token."},
		{ID: "openapi.habitCheckin", Title: "Create or update habit check-in through official OpenAPI", Resource: "openapi-habit", Operation: "write", Command: "dida openapi habit checkin <habit-id> --args-json '{...}' --dry-run --json", HTTP: []string{"POST /open/v1/habit/{habitId}/checkin"}, Status: "stable", AuthRequired: false, DryRun: true, Notes: "Requires an already saved OAuth access token."},
		{ID: "openapi.habitCheckins", Title: "List habit check-ins through official OpenAPI", Resource: "openapi-habit", Operation: "read", Command: "dida openapi habit checkins --habit-ids h1,h2 --from YYYYMMDD --to YYYYMMDD --json", HTTP: []string{"GET /open/v1/habit/checkins?habitIds=...&from=...&to=..."}, Status: "stable", AuthRequired: false, Notes: "Requires an already saved OAuth access token."},
		{ID: "agent.context", Title: "One-call compact context pack", Resource: "agent", Operation: "read", Command: "dida agent context --json", HTTP: []string{"GET /batch/check/0"}, Status: "stable", AuthRequired: true, Compact: true, Notes: "Returns projects, folders, tags, filters, today, upcoming, and quadrants from one sync. Use --outline for task id references plus a deduplicated taskIndex."},
		{ID: "sync.all", Title: "Full account sync", Resource: "sync", Operation: "read", Command: "dida sync all --json", HTTP: []string{"GET /batch/check/0"}, Status: "stable", AuthRequired: true},
		{ID: "sync.checkpoint", Title: "Incremental account sync", Resource: "sync", Operation: "read", Command: "dida sync checkpoint <checkpoint> --json", HTTP: []string{"GET /batch/check/{checkpoint}"}, Status: "stable", AuthRequired: true, Notes: "Preserves add/update/delete/order/reminder deltas."},
		{ID: "settings.get", Title: "Read user settings", Resource: "settings", Operation: "read", Command: "dida settings get --json", HTTP: []string{"GET /user/preferences/settings", "GET /user/preferences/settings?includeWeb=true"}, Status: "stable", AuthRequired: true, Notes: "Use --include-web to include Web-side preference fields such as smartProjects and shortcut config."},
		{ID: "project.list", Title: "List projects", Resource: "project", Operation: "read", Command: "dida project list --json", HTTP: []string{"GET /batch/check/0"}, Status: "stable", AuthRequired: true},
		{ID: "project.tasks", Title: "List tasks in a project", Resource: "project", Operation: "read", Command: "dida project tasks <project-id> --limit 50 --compact --json", HTTP: []string{"GET /project/{projectId}/tasks"}, Status: "stable", AuthRequired: true, Compact: true},
		{ID: "project.create", Title: "Create project", Resource: "project", Operation: "write", Command: "dida project create --name <name> --dry-run --json", HTTP: []string{"POST /batch/project"}, Status: "stable", AuthRequired: true, DryRun: true},
		{ID: "project.update", Title: "Update project", Resource: "project", Operation: "write", Command: "dida project update <project-id> --name <name> --dry-run --json", HTTP: []string{"POST /batch/project"}, Status: "stable", AuthRequired: true, DryRun: true},
		{ID: "project.delete", Title: "Delete project", Resource: "project", Operation: "delete", Command: "dida project delete <project-id> --dry-run --yes --json", HTTP: []string{"POST /batch/project"}, Status: "stable", AuthRequired: true, DryRun: true, ConfirmationRequired: true},
		{ID: "folder.list", Title: "List project folders", Resource: "folder", Operation: "read", Command: "dida folder list --json", HTTP: []string{"GET /batch/check/0"}, Status: "stable", AuthRequired: true},
		{ID: "folder.create", Title: "Create project folder", Resource: "folder", Operation: "write", Command: "dida folder create --name <name> --dry-run --json", HTTP: []string{"POST /batch/projectGroup"}, Status: "stable", AuthRequired: true, DryRun: true},
		{ID: "folder.update", Title: "Update project folder", Resource: "folder", Operation: "write", Command: "dida folder update <folder-id> --name <name> --dry-run --json", HTTP: []string{"POST /batch/projectGroup"}, Status: "stable", AuthRequired: true, DryRun: true},
		{ID: "folder.delete", Title: "Delete project folder", Resource: "folder", Operation: "delete", Command: "dida folder delete <folder-id> --dry-run --yes --json", HTTP: []string{"POST /batch/projectGroup"}, Status: "stable", AuthRequired: true, DryRun: true, ConfirmationRequired: true},
		{ID: "tag.list", Title: "List tags", Resource: "tag", Operation: "read", Command: "dida tag list --json", HTTP: []string{"GET /batch/check/0"}, Status: "stable", AuthRequired: true},
		{ID: "tag.create", Title: "Create tag", Resource: "tag", Operation: "write", Command: "dida tag create <name> --dry-run --json", HTTP: []string{"POST /batch/tag"}, Status: "stable", AuthRequired: true, DryRun: true},
		{ID: "tag.update", Title: "Update tag metadata", Resource: "tag", Operation: "write", Command: "dida tag update <name> --dry-run --json", HTTP: []string{"POST /batch/tag"}, Status: "stable", AuthRequired: true, DryRun: true},
		{ID: "tag.rename", Title: "Rename tag", Resource: "tag", Operation: "write", Command: "dida tag rename <old-name> <new-name> --dry-run --json", HTTP: []string{"PUT /tag/rename"}, Status: "stable", AuthRequired: true, DryRun: true},
		{ID: "tag.merge", Title: "Merge tag associations", Resource: "tag", Operation: "write", Command: "dida tag merge <from-name> <to-name> --dry-run --yes --json", HTTP: []string{"PUT /tag/merge"}, Status: "stable", AuthRequired: true, DryRun: true, ConfirmationRequired: true, Notes: "Observed endpoint may leave the source tag object present; list tags after merge."},
		{ID: "tag.delete", Title: "Delete tag", Resource: "tag", Operation: "delete", Command: "dida tag delete <name> --dry-run --yes --json", HTTP: []string{"DELETE /tag?name=..."}, Status: "stable", AuthRequired: true, DryRun: true, ConfirmationRequired: true},
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
		{ID: "task.delete", Title: "Delete task", Resource: "task", Operation: "delete", Command: "dida task delete <task-id> --project <project-id> --dry-run --yes --json", HTTP: []string{"POST /batch/task"}, Status: "stable", AuthRequired: true, DryRun: true, ConfirmationRequired: true},
		{ID: "task.move", Title: "Move task between projects", Resource: "task", Operation: "write", Command: "dida task move <task-id> --from <project-id> --to <project-id> --dry-run --json", HTTP: []string{"POST /batch/taskProject"}, Status: "stable", AuthRequired: true, DryRun: true},
		{ID: "task.parent", Title: "Set task parent", Resource: "task", Operation: "write", Command: "dida task parent <task-id> --parent <task-id> --project <project-id> --dry-run --json", HTTP: []string{"POST /batch/taskParent"}, Status: "stable", AuthRequired: true, DryRun: true},
		{ID: "comment.list", Title: "List task comments", Resource: "comment", Operation: "read", Command: "dida comment list --project <project-id> --task <task-id> --json", HTTP: []string{"GET /project/{projectId}/task/{taskId}/comments"}, Status: "stable", AuthRequired: true},
		{ID: "comment.create", Title: "Create task comment", Resource: "comment", Operation: "write", Command: "dida comment create --project <project-id> --task <task-id> --text <text> --dry-run --json", HTTP: []string{"POST /project/{projectId}/task/{taskId}/comment"}, Status: "stable", AuthRequired: true, DryRun: true},
		{ID: "comment.createWithFile", Title: "Create task comment with uploaded attachment", Resource: "comment", Operation: "write", Command: "dida comment create --project <project-id> --task <task-id> --text <text> --file <path> --dry-run --json", HTTP: []string{"POST /api/v1/attachment/upload/comment/{projectId}/{taskId}", "POST /project/{projectId}/task/{taskId}/comment"}, Status: "stable", AuthRequired: true, DryRun: true, Notes: "Uses multipart field file, then sends comment attachments [{id}]. Use the real project id, not logical inbox."},
		{ID: "comment.update", Title: "Update task comment", Resource: "comment", Operation: "write", Command: "dida comment update --project <project-id> --task <task-id> --comment <comment-id> --text <text> --dry-run --json", HTTP: []string{"PUT /project/{projectId}/task/{taskId}/comment/{commentId}"}, Status: "stable", AuthRequired: true, DryRun: true},
		{ID: "comment.delete", Title: "Delete task comment", Resource: "comment", Operation: "delete", Command: "dida comment delete --project <project-id> --task <task-id> --comment <comment-id> --dry-run --yes --json", HTTP: []string{"DELETE /project/{projectId}/task/{taskId}/comment/{commentId}"}, Status: "stable", AuthRequired: true, DryRun: true, ConfirmationRequired: true},
		{ID: "completed.list", Title: "List completed tasks", Resource: "completed", Operation: "read", Command: "dida completed list --from YYYY-MM-DD --to YYYY-MM-DD --compact --json", HTTP: []string{"GET /project/all/completed"}, Status: "stable", AuthRequired: true, Compact: true},
		{ID: "closed.list", Title: "List closed-history items", Resource: "closed", Operation: "read", Command: "dida closed list --status 2 --limit 50 --json", HTTP: []string{"GET /project/{projectIds|all}/closed?from=...&to=...&status=..."}, Status: "stable", AuthRequired: true, Notes: "Uses the private closed-history endpoint. Dates are sent as full datetime strings."},
		{ID: "trash.list", Title: "List deleted tasks from trash", Resource: "trash", Operation: "read", Command: "dida trash list --cursor 20 --compact --json", HTTP: []string{"GET /project/all/trash/page", "GET /project/all/trash/page?from=<cursor>"}, Status: "stable", AuthRequired: true, Compact: true, Notes: "The server returns next as the follow-up cursor. Avoid type=task; it returned HTTP 500 in live probing."},
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
	applyChannelAuthMetadata(schemas)
	return schemas
}

func applyChannelAuthMetadata(schemas []commandSchema) {
	officialNoAuth := map[string]bool{
		"official.tokenStatus": true,
		"official.tokenSet":    true,
		"official.tokenClear":  true,
	}
	openAPINoAuth := map[string]bool{
		"openapi.doctor":         true,
		"openapi.clientStatus":   true,
		"openapi.clientSet":      true,
		"openapi.clientClear":    true,
		"openapi.status":         true,
		"openapi.listenCallback": true,
		"openapi.logout":         true,
	}
	for i := range schemas {
		id := schemas[i].ID
		if strings.HasPrefix(id, "official.") && !officialNoAuth[id] {
			schemas[i].AuthRequired = true
		}
		if strings.HasPrefix(id, "openapi.") && !openAPINoAuth[id] {
			schemas[i].AuthRequired = true
		}
	}
}
