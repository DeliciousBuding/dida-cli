package cli

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/DeliciousBuding/dida-cli/internal/auth"
	"github.com/DeliciousBuding/dida-cli/internal/model"
	"github.com/DeliciousBuding/dida-cli/internal/openapi"
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

func TestOpenAPIReadCommandsUseLocalTestServer(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer test-openapi-token" {
			t.Fatalf("Authorization = %q", r.Header.Get("Authorization"))
		}
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/project":
			_, _ = io.WriteString(w, `[{"id":"p1","name":"Work"}]`)
		case "/project/p1":
			_, _ = io.WriteString(w, `{"id":"p1","name":"Work"}`)
		case "/project/p1/data":
			_, _ = io.WriteString(w, `{"project":{"id":"p1"},"tasks":[{"id":"task-alpha","title":"Alpha"}]}`)
		case "/project/p1/task/task-alpha":
			_, _ = io.WriteString(w, `{"id":"task-alpha","projectId":"p1","title":"Alpha"}`)
		case "/task/completed":
			if r.Method != http.MethodPost {
				t.Fatalf("completed method = %s, want POST", r.Method)
			}
			_, _ = io.WriteString(w, `[{"id":"done-1","projectId":"p1","title":"Done"}]`)
		case "/task/filter":
			if r.Method != http.MethodPost {
				t.Fatalf("filter method = %s, want POST", r.Method)
			}
			_, _ = io.WriteString(w, `[{"id":"task-alpha","projectId":"p1","title":"Alpha"}]`)
		case "/focus/focus-1":
			if r.URL.Query().Get("type") != "0" {
				t.Fatalf("focus get query = %q, want type=0", r.URL.RawQuery)
			}
			_, _ = io.WriteString(w, `{"id":"focus-1","type":0}`)
		case "/focus":
			if r.URL.Query().Get("from") == "" || r.URL.Query().Get("to") == "" || r.URL.Query().Get("type") != "1" {
				t.Fatalf("focus list query = %q", r.URL.RawQuery)
			}
			_, _ = io.WriteString(w, `[{"id":"focus-2","type":1}]`)
		case "/habit":
			_, _ = io.WriteString(w, `[{"id":"habit-1","name":"Read"}]`)
		case "/habit/habit-1":
			_, _ = io.WriteString(w, `{"id":"habit-1","name":"Read"}`)
		case "/habit/checkins":
			if r.URL.Query().Get("habitIds") != "habit-1" {
				t.Fatalf("habit checkins query = %q", r.URL.RawQuery)
			}
			_, _ = io.WriteString(w, `[{"habitId":"habit-1","stamp":20260707}]`)
		default:
			t.Fatalf("unexpected openapi request: %s %s", r.Method, r.URL.String())
		}
	}))
	defer server.Close()

	t.Setenv("DIDA_CONFIG_DIR", t.TempDir())
	t.Setenv("DIDA_OPENAPI_BASE_URL", server.URL)
	if err := openapi.SaveToken(&openapi.TokenResponse{OAuthToken: openapi.OAuthToken{AccessToken: "test-openapi-token", TokenType: "Bearer", Scope: "tasks:read", CreatedAt: 123}}); err != nil {
		t.Fatalf("SaveToken() error = %v", err)
	}

	cases := []struct {
		name    string
		args    []string
		command string
	}{
		{name: "project list", args: []string{"openapi", "project", "list"}, command: "openapi project list"},
		{name: "project get", args: []string{"openapi", "project", "get", "p1"}, command: "openapi project get"},
		{name: "project data", args: []string{"openapi", "project", "data", "p1"}, command: "openapi project data"},
		{name: "task get", args: []string{"openapi", "task", "get", "--project", "p1", "--task", "task-alpha"}, command: "openapi task get"},
		{name: "task completed", args: []string{"openapi", "task", "completed", "--args-json", `{"from":"2026-07-01","to":"2026-07-07"}`}, command: "openapi task completed"},
		{name: "task filter", args: []string{"openapi", "task", "filter", "--args-json", `{"projectId":"p1"}`}, command: "openapi task filter"},
		{name: "focus get", args: []string{"openapi", "focus", "get", "focus-1", "--type", "0"}, command: "openapi focus get"},
		{name: "focus list", args: []string{"openapi", "focus", "list", "--from", "2026-07-01T00:00:00+08:00", "--to", "2026-07-07T00:00:00+08:00", "--type", "1"}, command: "openapi focus list"},
		{name: "habit list", args: []string{"openapi", "habit", "list"}, command: "openapi habit list"},
		{name: "habit get", args: []string{"openapi", "habit", "get", "habit-1"}, command: "openapi habit get"},
		{name: "habit checkins", args: []string{"openapi", "habit", "checkins", "--habit-ids", "habit-1", "--from", "20260701", "--to", "20260707"}, command: "openapi habit checkins"},
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

func TestSyncBackedReadCommandsUseLocalTestServer(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/batch/check/0":
			_, _ = io.WriteString(w, `{
				"inboxId":"inbox-real",
				"checkPoint":42,
				"syncTaskBean":{"add":[
					{"id":"task-alpha","projectId":"p1","title":"Alpha report","content":"quarterly notes","dueDate":"2099-07-08T09:00:00+08:00","createdTime":"2026-07-07T08:00:00+08:00","priority":5,"status":0},
					{"id":"task-beta","projectId":"p2","title":"Beta followup","createdTime":"2026-07-06T08:00:00+08:00","priority":1,"status":0}
				]},
				"projects":[{"id":"p1","name":"Work"},{"id":"p2","name":"Side"}],
				"projectGroups":[{"id":"g1","name":"Folders"}],
				"tags":[{"name":"deep","label":"Deep Work","color":"blue"}],
				"filters":[{"id":"f1","name":"Today"}]
			}`)
		case "/batch/check/41":
			_, _ = io.WriteString(w, `{
				"inboxId":"inbox-real",
				"checkPoint":42,
				"syncTaskBean":{"add":[{"id":"task-delta","projectId":"p1","title":"Delta","status":0}]},
				"projects":[{"id":"p1","name":"Work"}],
				"checks":[{"id":"check-1"}],
				"filters":[{"id":"f1","name":"Today"}]
			}`)
		case "/project/p1/tasks":
			_, _ = io.WriteString(w, `[{"id":"task-alpha","projectId":"p1","title":"Alpha report","status":0}]`)
		case "/column/project/p1":
			_, _ = io.WriteString(w, `[{"id":"col-1","name":"Doing","sortOrder":1}]`)
		case "/project/p1/task/task-alpha/comments":
			_, _ = io.WriteString(w, `[{"id":"comment-1","title":"Looks good","taskId":"task-alpha","projectId":"p1"}]`)
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
		case "/user/preferences/dailyReminder":
			_, _ = io.WriteString(w, `{"enabled":true,"time":"09:00"}`)
		case "/user/preferences/pomodoro":
			_, _ = io.WriteString(w, `{"duration":25}`)
		case "/user/preferences/habit":
			if r.URL.Query().Get("platform") != "web" {
				t.Fatalf("habit preferences query = %q, want platform=web", r.URL.RawQuery)
			}
			_, _ = io.WriteString(w, `{"showCompleted":true}`)
		case "/project/all/trash/page":
			if r.URL.Query().Get("from") != "2" {
				t.Fatalf("trash cursor query = %q, want from=2", r.URL.RawQuery)
			}
			_, _ = io.WriteString(w, `{"next":3,"tasks":[{"id":"deleted-1","projectId":"p1","title":"Deleted","kind":"TEXT","status":2,"priority":1,"deletedTime":123456}]}`)
		case "/project/all/completed":
			if r.URL.Query().Get("limit") != "2" {
				t.Fatalf("completed query = %q, want limit=2", r.URL.RawQuery)
			}
			_, _ = io.WriteString(w, `[{"id":"done-1","projectId":"p1","title":"Done","status":2}]`)
		case "/project/p1/closed":
			if r.URL.Query().Get("status") != "2" {
				t.Fatalf("closed query = %q, want status=2", r.URL.RawQuery)
			}
			_, _ = io.WriteString(w, `[{"id":"closed-1","projectId":"p1","title":"Closed","status":2}]`)
		case "/attachment/isUnderQuota":
			_, _ = io.WriteString(w, `true`)
		case "/attachment/dailyLimit":
			_, _ = io.WriteString(w, `20`)
		case "/attachment/p1/task-alpha/attach-1":
			if r.URL.Query().Get("action") != "download" {
				t.Fatalf("attachment query = %q, want action=download", r.URL.RawQuery)
			}
			w.Header().Set("Content-Type", "text/plain")
			_, _ = io.WriteString(w, `download-body`)
		case "/share/shareContacts":
			_, _ = io.WriteString(w, `{"contacts":[{"id":"u1","name":"Friend"}]}`)
		case "/project/share/recentProjectUsers":
			_, _ = io.WriteString(w, `[{"id":"u1","name":"Recent"}]`)
		case "/project/p1/shares":
			_, _ = io.WriteString(w, `[{"id":"share-1","name":"Member"}]`)
		case "/project/p1/share/check-quota":
			_, _ = io.WriteString(w, `5`)
		case "/project/p1/collaboration/invite-url":
			_, _ = io.WriteString(w, `{"enabled":true,"url":"https://example.invalid/invite"}`)
		case "/calendar/subscription":
			_, _ = io.WriteString(w, `[{"id":"cal-1","name":"Main"}]`)
		case "/calendar/archivedEvent":
			_, _ = io.WriteString(w, `[{"id":"event-1","name":"Archived"}]`)
		case "/calendar/third/accounts":
			_, _ = io.WriteString(w, `{"accounts":[{"id":"third-1","provider":"google"}]}`)
		case "/statistics/general":
			_, _ = io.WriteString(w, `{"completedCount":7}`)
		case "/projectTemplates/all":
			if r.URL.Query().Get("timestamp") != "123" {
				t.Fatalf("template query = %q, want timestamp=123", r.URL.RawQuery)
			}
			_, _ = io.WriteString(w, `{"projectTemplates":[{"id":"tpl-1"},{"id":"tpl-2"}]}`)
		case "/search/all":
			if r.URL.Query().Get("keywords") != "Alpha" {
				t.Fatalf("search query = %q, want Alpha", r.URL.RawQuery)
			}
			_, _ = io.WriteString(w, `{
				"hits":[{"index":"task","id":"task-alpha","source":{"title":"Alpha report","projectId":"p1","modifiedTime":"2026-07-07T08:00:00+08:00"}}],
				"tasks":[{"id":"task-alpha","projectId":"p1","title":"Alpha report","status":0,"priority":1}],
				"comments":[{"id":"comment-1","projectId":"p1","taskId":"task-alpha","title":"Looks good"}]
			}`)
		case "/user/status":
			_, _ = io.WriteString(w, `{"userId":"u1","inboxId":"inbox-real","pro":true,"private":"ignored"}`)
		case "/user/profile":
			_, _ = io.WriteString(w, `{"name":"Ada","displayName":"Ada Lovelace","locale":"en_US","email":"hidden@example.invalid"}`)
		case "/user/sessions":
			if r.URL.Query().Get("lang") != "en_US" {
				t.Fatalf("sessions query = %q, want lang=en_US", r.URL.RawQuery)
			}
			_, _ = io.WriteString(w, `[{"id":"session-1","deviceInfo":{"platform":"web","name":"Firefox"}},{"id":"session-2","deviceInfo":{"platform":"ios"}}]`)
		case "/pomodoros":
			if r.URL.Query().Get("from") == "" || r.URL.Query().Get("to") == "" {
				t.Fatalf("pomo range query = %q, want from and to", r.URL.RawQuery)
			}
			_, _ = io.WriteString(w, `[{"id":"pomo-1","name":"Focus"}]`)
		case "/pomodoros/timing":
			if r.URL.Query().Get("from") == "" || r.URL.Query().Get("to") == "" {
				t.Fatalf("pomo timing query = %q, want from and to", r.URL.RawQuery)
			}
			_, _ = io.WriteString(w, `[{"id":"timing-1","name":"Timing"}]`)
		case "/pomodoros/task":
			if r.URL.Query().Get("projectId") != "p1" || r.URL.Query().Get("taskId") != "task-alpha" {
				t.Fatalf("pomo task query = %q", r.URL.RawQuery)
			}
			_, _ = io.WriteString(w, `[{"id":"pomo-task-1","name":"Task Focus"}]`)
		case "/pomodoros/statistics/generalForDesktop":
			_, _ = io.WriteString(w, `{"total":3}`)
		case "/pomodoros/timeline":
			if r.URL.Query().Get("to") != "cursor-1" {
				t.Fatalf("pomo timeline query = %q, want to=cursor-1", r.URL.RawQuery)
			}
			_, _ = io.WriteString(w, `[{"id":"timeline-1","name":"Timeline"}]`)
		case "/habits":
			_, _ = io.WriteString(w, `[{"id":"habit-1","name":"Read"}]`)
		case "/habitSections":
			_, _ = io.WriteString(w, `[{"id":"section-1","name":"Health"}]`)
		case "/habitCheckins/query":
			if r.Method != http.MethodPost {
				t.Fatalf("habit checkins method = %s, want POST", r.Method)
			}
			_, _ = io.WriteString(w, `{"checkins":[{"habitId":"habit-1","stamp":20260707}]}`)
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
		{name: "task upcoming", args: []string{"task", "upcoming", "--days", "36500", "--limit", "1", "--compact"}, command: "task upcoming"},
		{name: "task get", args: []string{"task", "get", "task-alpha"}, command: "task get"},
		{name: "project list", args: []string{"project", "list"}, command: "project list"},
		{name: "project tasks", args: []string{"project", "tasks", "p1", "--limit", "1", "--compact"}, command: "project tasks"},
		{name: "project columns", args: []string{"project", "columns", "p1"}, command: "project columns"},
		{name: "column list", args: []string{"column", "list", "p1"}, command: "project columns"},
		{name: "comment list", args: []string{"comment", "list", "--project", "p1", "--task", "task-alpha"}, command: "comment list"},
		{name: "filter list", args: []string{"filter", "list"}, command: "filter list"},
		{name: "folder list", args: []string{"folder", "list"}, command: "folder list"},
		{name: "tag list", args: []string{"tag", "list"}, command: "tag list"},
		{name: "quadrant list", args: []string{"quadrant", "list"}, command: "quadrant list"},
		{name: "quadrant view", args: []string{"quadrant", "view", "q1"}, command: "quadrant view"},
		{name: "agent context", args: []string{"agent", "context", "--outline", "--limit", "1"}, command: "agent context"},
		{name: "sync all", args: []string{"sync", "all"}, command: "sync all"},
		{name: "sync checkpoint", args: []string{"sync", "checkpoint", "41"}, command: "sync checkpoint"},
		{name: "completed list", args: []string{"completed", "list", "--from", "2026-07-01", "--to", "2026-07-07", "--limit", "2", "--compact"}, command: "completed list"},
		{name: "closed list", args: []string{"closed", "list", "--project", "p1", "--status", "2", "--from", "2026-07-01", "--to", "2026-07-07", "--limit", "1"}, command: "closed list"},
		{name: "task due-counts", args: []string{"task", "due-counts"}, command: "task due-counts"},
		{name: "settings get", args: []string{"settings", "get", "--include-web"}, command: "settings get"},
		{name: "trash list", args: []string{"trash", "list", "--cursor", "2", "--limit", "1", "--compact"}, command: "trash list"},
		{name: "attachment quota", args: []string{"attachment", "quota"}, command: "attachment quota"},
		{name: "reminder daily", args: []string{"reminder", "daily"}, command: "reminder daily"},
		{name: "calendar subscriptions", args: []string{"calendar", "subscriptions"}, command: "calendar subscriptions"},
		{name: "calendar archived", args: []string{"calendar", "archived"}, command: "calendar archived"},
		{name: "calendar third accounts", args: []string{"calendar", "third-accounts"}, command: "calendar third-accounts"},
		{name: "share contacts", args: []string{"share", "contacts"}, command: "share contacts"},
		{name: "share recent", args: []string{"share", "recent-users"}, command: "share recent-users"},
		{name: "share project shares", args: []string{"share", "project", "shares", "p1"}, command: "share project shares"},
		{name: "share project quota", args: []string{"share", "project", "quota", "p1"}, command: "share project quota"},
		{name: "share project invite", args: []string{"share", "project", "invite-url", "p1"}, command: "share project invite-url"},
		{name: "stats general", args: []string{"stats", "general"}, command: "stats general"},
		{name: "template project list", args: []string{"template", "project", "list", "--timestamp", "123", "--limit", "1"}, command: "template project list"},
		{name: "search all", args: []string{"search", "all", "--query", "Alpha", "--limit", "1"}, command: "search all"},
		{name: "user status", args: []string{"user", "status"}, command: "user status"},
		{name: "user profile", args: []string{"user", "profile"}, command: "user profile"},
		{name: "user sessions", args: []string{"user", "sessions", "--lang", "en_US", "--limit", "1"}, command: "user sessions"},
		{name: "pomo preferences", args: []string{"pomo", "preferences"}, command: "pomo preferences"},
		{name: "pomo list", args: []string{"pomo", "list", "--from", "2026-07-01", "--to", "2026-07-07", "--limit", "1"}, command: "pomo list"},
		{name: "pomo timing", args: []string{"pomo", "timing", "--from", "2026-07-01", "--to", "2026-07-07", "--limit", "1"}, command: "pomo timing"},
		{name: "pomo task", args: []string{"pomo", "task", "--project", "p1", "--task", "task-alpha"}, command: "pomo task"},
		{name: "pomo stats", args: []string{"pomo", "stats"}, command: "pomo stats"},
		{name: "pomo timeline", args: []string{"pomo", "timeline", "--to", "cursor-1", "--limit", "1"}, command: "pomo timeline"},
		{name: "habit preferences", args: []string{"habit", "preferences"}, command: "habit preferences"},
		{name: "habit list", args: []string{"habit", "list"}, command: "habit list"},
		{name: "habit sections", args: []string{"habit", "sections"}, command: "habit sections"},
		{name: "habit checkins", args: []string{"habit", "checkins", "--habit", "habit-1", "--after-stamp", "123"}, command: "habit checkins"},
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

	output := filepath.Join(t.TempDir(), "attachment.txt")
	payload := runCLIJSON(t, "attachment", "download", "--project", "p1", "--task", "task-alpha", "--attachment", "attach-1", "--output", output)
	if payload["command"] != "attachment download" {
		t.Fatalf("attachment command = %v, want attachment download", payload["command"])
	}
	data, err := os.ReadFile(output)
	if err != nil {
		t.Fatalf("read downloaded attachment: %v", err)
	}
	if string(data) != "download-body" {
		t.Fatalf("downloaded attachment = %q", string(data))
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
