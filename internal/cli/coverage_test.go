package cli

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/DeliciousBuding/dida-cli/internal/auth"
	"github.com/DeliciousBuding/dida-cli/internal/model"
)

func runCLIJSON(t *testing.T, args ...string) map[string]any {
	t.Helper()
	var stdout, stderr bytes.Buffer
	code := Run(append(args, "--json"), "test-version", &stdout, &stderr)
	if code != 0 {
		t.Fatalf("Run(%v) code = %d stderr=%s stdout=%s", args, code, stderr.String(), stdout.String())
	}
	if stderr.Len() != 0 {
		t.Fatalf("stderr = %q, want empty", stderr.String())
	}
	var payload map[string]any
	if err := json.Unmarshal(stdout.Bytes(), &payload); err != nil {
		t.Fatalf("invalid json: %v\n%s", err, stdout.String())
	}
	return payload
}

func TestRootHelpAndVersionText(t *testing.T) {
	var stdout, stderr bytes.Buffer
	code := Run(nil, "test-version", &stdout, &stderr)
	if code != 0 {
		t.Fatalf("Run(nil) code = %d stderr=%s", code, stderr.String())
	}
	if !strings.Contains(stdout.String(), "DidaCLI - Dida365 / TickTick command line client") {
		t.Fatalf("root help missing title: %s", stdout.String())
	}

	stdout.Reset()
	stderr.Reset()
	code = Run([]string{"--version"}, "test-version", &stdout, &stderr)
	if code != 0 {
		t.Fatalf("version code = %d stderr=%s", code, stderr.String())
	}
	if strings.TrimSpace(stdout.String()) != "test-version" {
		t.Fatalf("version stdout = %q", stdout.String())
	}
}

func TestLocalHelpCommandsDoNotRequireAuth(t *testing.T) {
	cases := []struct {
		name string
		args []string
		want string
	}{
		{name: "auth", args: []string{"auth", "--help"}, want: "dida auth cookie set --token-stdin"},
		{name: "completion", args: []string{"completion", "--help"}, want: "dida completion powershell"},
		{name: "schema", args: []string{"schema", "--help"}, want: "dida schema list"},
		{name: "channel", args: []string{"channel", "--help"}, want: "dida channel list"},
		{name: "raw", args: []string{"raw", "--help"}, want: "Only GET is supported"},
		{name: "filter", args: []string{"filter", "--help"}, want: "dida filter list"},
		{name: "sync", args: []string{"sync", "--help"}, want: "dida sync all"},
		{name: "settings", args: []string{"settings", "--help"}, want: "dida settings get"},
		{name: "reminder", args: []string{"reminder", "--help"}, want: "dida reminder daily"},
		{name: "calendar", args: []string{"calendar", "--help"}, want: "dida calendar subscriptions"},
		{name: "stats", args: []string{"stats", "--help"}, want: "dida stats general"},
		{name: "template", args: []string{"template", "--help"}, want: "dida template project list"},
		{name: "search", args: []string{"search", "--help"}, want: "dida search all"},
		{name: "user", args: []string{"user", "--help"}, want: "dida user sessions"},
		{name: "pomo", args: []string{"pomo", "--help"}, want: "dida pomo timeline"},
		{name: "habit", args: []string{"habit", "--help"}, want: "dida habit checkins"},
		{name: "trash", args: []string{"trash", "--help"}, want: "dida trash list"},
		{name: "attachment", args: []string{"attachment", "--help"}, want: "dida attachment download"},
		{name: "completed", args: []string{"completed", "--help"}, want: "dida completed list"},
		{name: "closed", args: []string{"closed", "--help"}, want: "dida closed list"},
		{name: "share", args: []string{"share", "--help"}, want: "dida share project invite-url"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var stdout, stderr bytes.Buffer
			code := Run(tc.args, "test-version", &stdout, &stderr)
			if code != 0 {
				t.Fatalf("Run(%v) code = %d stderr=%s", tc.args, code, stderr.String())
			}
			if stderr.Len() != 0 {
				t.Fatalf("stderr = %q, want empty", stderr.String())
			}
			if !strings.Contains(stdout.String(), tc.want) {
				t.Fatalf("stdout missing %q: %s", tc.want, stdout.String())
			}
		})
	}
}

func TestProjectUpdateDryRunJSON(t *testing.T) {
	payload := runCLIJSON(t, "project", "update", "p1", "--name", "Renamed", "--group", "g1", "--dry-run")
	if payload["command"] != "project update" {
		t.Fatalf("command = %v, want project update", payload["command"])
	}
	data := payload["data"].(map[string]any)
	requestPayload := data["payload"].(map[string]any)
	update := requestPayload["update"].([]any)
	project := update[0].(map[string]any)
	if project["id"] != "p1" || project["name"] != "Renamed" || project["groupId"] != "g1" {
		t.Fatalf("project payload = %#v", project)
	}
}

func TestTaskMoveAndParentDryRunJSON(t *testing.T) {
	movePayload := runCLIJSON(t, "task", "move", "t1", "--from", "p1", "--to", "p2", "--dry-run")
	if movePayload["command"] != "task move" {
		t.Fatalf("command = %v, want task move", movePayload["command"])
	}
	moveData := movePayload["data"].(map[string]any)
	moveRequest := moveData["payload"].([]any)[0].(map[string]any)
	if moveRequest["taskId"] != "t1" || moveRequest["fromProjectId"] != "p1" || moveRequest["toProjectId"] != "p2" {
		t.Fatalf("move payload = %#v", moveRequest)
	}

	parentPayload := runCLIJSON(t, "task", "parent", "t1", "--parent", "parent1", "--project", "p1", "--dry-run")
	if parentPayload["command"] != "task parent" {
		t.Fatalf("command = %v, want task parent", parentPayload["command"])
	}
	parentData := parentPayload["data"].(map[string]any)
	parentRequest := parentData["payload"].([]any)[0].(map[string]any)
	if parentRequest["taskId"] != "t1" || parentRequest["parentId"] != "parent1" || parentRequest["projectId"] != "p1" {
		t.Fatalf("parent payload = %#v", parentRequest)
	}
}

func TestTaskCreateFullFlagDryRunJSON(t *testing.T) {
	payload := runCLIJSON(t,
		"task", "create",
		"--id", "task-1",
		"--project", "p1",
		"--title", "Write report",
		"--content", "plain body",
		"--desc", "markdown body",
		"--start", "2026-07-07T09:00:00+08:00",
		"--due", "2026-07-07T18:00:00+08:00",
		"--timezone", "Asia/Hong_Kong",
		"--reminder", "TRIGGER:P0DT9H0M0S",
		"--repeat", "RRULE:FREQ=DAILY",
		"--repeat-from", "dueDate",
		"--repeat-flag", "0",
		"--priority", "5",
		"--column", "c1",
		"--tag", "agent",
		"--tags", "work,deep",
		"--item", "First step",
		"--not-all-day",
		"--floating",
		"--dry-run",
	)
	data := payload["data"].(map[string]any)
	requestPayload := data["payload"].(map[string]any)
	add := requestPayload["add"].([]any)
	task := add[0].(map[string]any)
	if task["id"] != "task-1" || task["projectId"] != "p1" || task["title"] != "Write report" {
		t.Fatalf("task identity payload = %#v", task)
	}
	if int(task["priority"].(float64)) != 5 || task["timeZone"] != "Asia/Hong_Kong" {
		t.Fatalf("task scheduling payload = %#v", task)
	}
	if tags := task["tags"].([]any); len(tags) != 3 {
		t.Fatalf("tags = %#v, want 3 entries", tags)
	}
	if items := task["items"].([]any); len(items) != 1 {
		t.Fatalf("items = %#v, want one checklist item", items)
	}
}

func TestProjectAndTaskValidationErrorsStayLocal(t *testing.T) {
	cases := []struct {
		name string
		args []string
		want string
	}{
		{name: "project update missing changes", args: []string{"project", "update", "p1", "--json"}, want: "no updates provided"},
		{name: "task update missing changes", args: []string{"task", "update", "t1", "--project", "p1", "--json"}, want: "no updates provided"},
		{name: "task move missing target", args: []string{"task", "move", "t1", "--from", "p1", "--json"}, want: "missing --from or --to"},
		{name: "task parent missing parent", args: []string{"task", "parent", "t1", "--project", "p1", "--json"}, want: "missing --parent or --project"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Setenv("DIDA_CONFIG_DIR", t.TempDir())
			var stdout, stderr bytes.Buffer
			code := Run(tc.args, "test-version", &stdout, &stderr)
			if code != 1 {
				t.Fatalf("Run(%v) code = %d, want 1", tc.args, code)
			}
			if stderr.Len() != 0 {
				t.Fatalf("stderr = %q, want empty for json errors", stderr.String())
			}
			if !strings.Contains(stdout.String(), tc.want) {
				t.Fatalf("stdout missing %q: %s", tc.want, stdout.String())
			}
		})
	}
}

func TestOpenAPITaskDryRunsDoNotRequireToken(t *testing.T) {
	t.Setenv("DIDA_CONFIG_DIR", t.TempDir())
	cases := []struct {
		name    string
		args    []string
		command string
	}{
		{
			name:    "create",
			args:    []string{"openapi", "task", "create", "--args-json", `{"title":"Draft","projectId":"p1"}`, "--dry-run"},
			command: "openapi task create",
		},
		{
			name:    "update",
			args:    []string{"openapi", "task", "update", "t1", "--args-json", `{"title":"Renamed"}`, "--dry-run"},
			command: "openapi task update",
		},
		{
			name:    "complete",
			args:    []string{"openapi", "task", "complete", "--project", "p1", "--task", "t1", "--dry-run"},
			command: "openapi task complete",
		},
		{
			name:    "delete",
			args:    []string{"openapi", "task", "delete", "--project", "p1", "--task", "t1", "--yes", "--dry-run"},
			command: "openapi task delete",
		},
		{
			name:    "move",
			args:    []string{"openapi", "task", "move", "--args-json", `[{"taskId":"t1","fromProjectId":"p1","toProjectId":"p2"}]`, "--dry-run"},
			command: "openapi task move",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			payload := runCLIJSON(t, tc.args...)
			if payload["command"] != tc.command {
				t.Fatalf("command = %v, want %s", payload["command"], tc.command)
			}
			data := payload["data"].(map[string]any)
			if data["dry_run"] != true {
				t.Fatalf("dry_run = %v, want true", data["dry_run"])
			}
		})
	}
}

func TestSyncBackedReadCommandsUseLocalTestServer(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/batch/check/0":
			_, _ = io.WriteString(w, `{
				"inboxId":"inbox-real",
				"checkPoint":42,
				"syncTaskBean":{"add":[
					{"id":"task-alpha","projectId":"p1","title":"Alpha report","content":"quarterly notes","dueDate":"2099-07-08T09:00:00+08:00","createdTime":"2026-07-07T08:00:00+08:00","status":0},
					{"id":"task-beta","projectId":"p2","title":"Beta followup","createdTime":"2026-07-06T08:00:00+08:00","status":0}
				]},
				"projects":[{"id":"p1","name":"Work"},{"id":"p2","name":"Side"}],
				"filters":[{"id":"f1","name":"Today"}]
			}`)
		case "/project/p1/tasks":
			_, _ = io.WriteString(w, `[{"id":"task-alpha","projectId":"p1","title":"Alpha report","status":0}]`)
		case "/task/activity/count/all":
			if r.Method != http.MethodPost {
				t.Fatalf("due-counts method = %s, want POST", r.Method)
			}
			_, _ = io.WriteString(w, `{"overdue":1}`)
		case "/user/preferences/settings":
			if r.URL.Query().Get("includeWeb") != "true" {
				t.Fatalf("includeWeb query = %q, want true", r.URL.RawQuery)
			}
			_, _ = io.WriteString(w, `{"timeZone":"Asia/Hong_Kong","locale":"en_US"}`)
		case "/project/all/trash/page":
			if r.URL.Query().Get("from") != "2" {
				t.Fatalf("trash cursor query = %q, want from=2", r.URL.RawQuery)
			}
			_, _ = io.WriteString(w, `{"next":3,"tasks":[{"id":"deleted-1","projectId":"p1","title":"Deleted","kind":"TEXT","status":2,"priority":1,"deletedTime":123456}]}`)
		case "/probe":
			_, _ = io.WriteString(w, `{"ok":true}`)
		default:
			t.Fatalf("unexpected request: %s %s", r.Method, r.URL.String())
		}
	}))
	defer server.Close()

	t.Setenv("DIDA_CONFIG_DIR", t.TempDir())
	t.Setenv("DIDA_WEBAPI_BASE_URL", server.URL)
	t.Setenv("DIDA_WEBAPI_BASE_URL_V1", server.URL)
	if _, err := auth.SaveCookieToken("test_cookie_value_12345"); err != nil {
		t.Fatalf("SaveCookieToken() error = %v", err)
	}

	cases := []struct {
		name    string
		args    []string
		command string
	}{
		{name: "task list", args: []string{"task", "list", "--filter", "all", "--limit", "1", "--compact"}, command: "task list"},
		{name: "task search", args: []string{"task", "search", "--query", "Alpha", "--compact"}, command: "task search"},
		{name: "task get", args: []string{"task", "get", "task-alpha"}, command: "task get"},
		{name: "project list", args: []string{"project", "list"}, command: "project list"},
		{name: "project tasks", args: []string{"project", "tasks", "p1", "--limit", "1", "--compact"}, command: "project tasks"},
		{name: "filter list", args: []string{"filter", "list"}, command: "filter list"},
		{name: "sync all", args: []string{"sync", "all"}, command: "sync all"},
		{name: "task due-counts", args: []string{"task", "due-counts"}, command: "task due-counts"},
		{name: "settings get", args: []string{"settings", "get", "--include-web"}, command: "settings get"},
		{name: "trash list", args: []string{"trash", "list", "--cursor", "2", "--limit", "1", "--compact"}, command: "trash list"},
		{name: "raw get", args: []string{"raw", "get", "/probe"}, command: "raw get"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			payload := runCLIJSON(t, tc.args...)
			if payload["command"] != tc.command {
				t.Fatalf("command = %v, want %s", payload["command"], tc.command)
			}
			if payload["ok"] != true {
				t.Fatalf("ok = %v, want true", payload["ok"])
			}
		})
	}
}

func TestTextPrintersAndRawStrippers(t *testing.T) {
	projects := []model.Project{
		{ID: "p1", Name: "Inbox", Raw: map[string]any{"secret": "raw"}},
	}
	strippedProjects := stripProjectRaw(projects)
	if strippedProjects[0].Raw != nil {
		t.Fatalf("stripProjectRaw kept raw payload: %#v", strippedProjects[0].Raw)
	}
	if projects[0].Raw == nil {
		t.Fatalf("stripProjectRaw mutated input")
	}

	task := model.Task{ID: "t1", ProjectID: "p1", ProjectName: "LongProjectName", Title: "Task", Priority: 3, DueDate: "2026-07-07", Raw: map[string]any{"secret": "raw"}}
	strippedTask := stripSingleTaskRaw(task)
	if strippedTask.Raw != nil || task.Raw == nil {
		t.Fatalf("stripSingleTaskRaw result=%#v original=%#v", strippedTask.Raw, task.Raw)
	}

	var stdout bytes.Buffer
	printProjects(&stdout, projects)
	if !strings.Contains(stdout.String(), "Inbox") {
		t.Fatalf("printProjects output = %s", stdout.String())
	}
	stdout.Reset()
	printTasks(&stdout, []model.Task{task}, 1)
	if !strings.Contains(stdout.String(), "Showing 1 of 1 task(s)") || !strings.Contains(stdout.String(), "Task") {
		t.Fatalf("printTasks output = %s", stdout.String())
	}
	stdout.Reset()
	printMapList(&stdout, []map[string]any{{"id": "tag1", "label": "Tag One"}}, "tags")
	if !strings.Contains(stdout.String(), "Tag One") {
		t.Fatalf("printMapList output = %s", stdout.String())
	}
}
