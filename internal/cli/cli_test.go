package cli

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/DeliciousBuding/dida-cli/internal/model"
	"github.com/DeliciousBuding/dida-cli/internal/openapi"
	"github.com/DeliciousBuding/dida-cli/internal/webapi"
)

func TestDoctorJSON(t *testing.T) {
	var stdout, stderr bytes.Buffer
	code := Run([]string{"--json", "doctor"}, "test-version", &stdout, &stderr)
	if code != 0 {
		t.Fatalf("exit code = %d, stderr=%s", code, stderr.String())
	}
	var payload map[string]any
	if err := json.Unmarshal(stdout.Bytes(), &payload); err != nil {
		t.Fatalf("invalid json: %v\n%s", err, stdout.String())
	}
	if payload["ok"] != true {
		t.Fatalf("ok = %v, want true", payload["ok"])
	}
	if payload["command"] != "doctor" {
		t.Fatalf("command = %v, want doctor", payload["command"])
	}
}

func TestDoctorVerifyMissingAuthIncludesDiagnosticData(t *testing.T) {
	t.Setenv("DIDA_CONFIG_DIR", t.TempDir())
	var stdout, stderr bytes.Buffer
	code := Run([]string{"doctor", "--verify", "--json"}, "test-version", &stdout, &stderr)
	if code != 1 {
		t.Fatalf("exit code = %d, want 1, stdout=%s", code, stdout.String())
	}
	if stderr.Len() != 0 {
		t.Fatalf("stderr = %q, want empty for json errors", stderr.String())
	}
	var payload map[string]any
	if err := json.Unmarshal(stdout.Bytes(), &payload); err != nil {
		t.Fatalf("invalid json: %v\n%s", err, stdout.String())
	}
	if payload["ok"] != false {
		t.Fatalf("ok = %v, want false", payload["ok"])
	}
	if payload["command"] != "doctor" {
		t.Fatalf("command = %v, want doctor", payload["command"])
	}
	data := payload["data"].(map[string]any)
	networkCheck := data["network_check"].(map[string]any)
	if networkCheck["status"] != "failed" {
		t.Fatalf("network_check.status = %v, want failed", networkCheck["status"])
	}
	if networkCheck["channel"] != "webapi" {
		t.Fatalf("network_check.channel = %v, want webapi", networkCheck["channel"])
	}
}

func TestJSONFlagWithoutCommandReturnsError(t *testing.T) {
	var stdout, stderr bytes.Buffer
	code := Run([]string{"--json"}, "test-version", &stdout, &stderr)
	if code != 1 {
		t.Fatalf("exit code = %d, want 1", code)
	}
	if stderr.Len() != 0 {
		t.Fatalf("stderr = %q, want empty for json errors", stderr.String())
	}
	var payload map[string]any
	if err := json.Unmarshal(stdout.Bytes(), &payload); err != nil {
		t.Fatalf("invalid json: %v\n%s", err, stdout.String())
	}
	if payload["ok"] != false {
		t.Fatalf("ok = %v, want false", payload["ok"])
	}
	errPayload := payload["error"].(map[string]any)
	if errPayload["type"] != "validation" {
		t.Fatalf("error.type = %v, want validation", errPayload["type"])
	}
}

func TestAuthStatusVerifyMissingAuthFailsJSON(t *testing.T) {
	t.Setenv("DIDA_CONFIG_DIR", t.TempDir())
	var stdout, stderr bytes.Buffer
	code := Run([]string{"auth", "status", "--verify", "--json"}, "test-version", &stdout, &stderr)
	if code != 1 {
		t.Fatalf("exit code = %d, want 1", code)
	}
	if stderr.Len() != 0 {
		t.Fatalf("stderr = %q, want empty for json errors", stderr.String())
	}
	var payload map[string]any
	if err := json.Unmarshal(stdout.Bytes(), &payload); err != nil {
		t.Fatalf("invalid json: %v\n%s", err, stdout.String())
	}
	if payload["command"] != "auth status" {
		t.Fatalf("command = %v, want auth status", payload["command"])
	}
	errPayload := payload["error"].(map[string]any)
	if errPayload["type"] != "auth" {
		t.Fatalf("error.type = %v, want auth", errPayload["type"])
	}
}

func TestOpenAPIDoctorIncludesRedirectAndNextActions(t *testing.T) {
	t.Setenv("DIDA_CONFIG_DIR", t.TempDir())
	t.Setenv("DIDA365_OPENAPI_CLIENT_ID", "")
	t.Setenv("DIDA365_OPENAPI_CLIENT_SECRET", "")
	var stdout, stderr bytes.Buffer
	code := Run([]string{"openapi", "doctor", "--json"}, "test-version", &stdout, &stderr)
	if code != 0 {
		t.Fatalf("exit code = %d, stderr=%s", code, stderr.String())
	}
	var payload map[string]any
	if err := json.Unmarshal(stdout.Bytes(), &payload); err != nil {
		t.Fatalf("invalid json: %v\n%s", err, stdout.String())
	}
	data := payload["data"].(map[string]any)
	if data["default_redirect_uri"] != "http://127.0.0.1:17890/callback" {
		t.Fatalf("default_redirect_uri = %v", data["default_redirect_uri"])
	}
	next := data["next"].([]any)
	if len(next) < 2 {
		t.Fatalf("next actions too short: %v", next)
	}
	if !strings.Contains(next[0].(string), "openapi client set") {
		t.Fatalf("first next action = %v", next[0])
	}
}

func TestSyncMissingAuthJSON(t *testing.T) {
	t.Setenv("DIDA_CONFIG_DIR", t.TempDir())
	var stdout, stderr bytes.Buffer
	code := Run([]string{"sync", "all", "--json"}, "test-version", &stdout, &stderr)
	if code != 1 {
		t.Fatalf("exit code = %d, want 1", code)
	}
	if !strings.Contains(stdout.String(), "missing cookie auth") {
		t.Fatalf("stdout missing auth hint: %s", stdout.String())
	}
	if stderr.Len() != 0 {
		t.Fatalf("stderr = %q, want empty for json errors", stderr.String())
	}
}

func TestAuthLoginJSON(t *testing.T) {
	var stdout, stderr bytes.Buffer
	code := Run([]string{"auth", "login", "--json"}, "test-version", &stdout, &stderr)
	if code != 0 {
		t.Fatalf("exit code = %d, stderr=%s", code, stderr.String())
	}
	var payload map[string]any
	if err := json.Unmarshal(stdout.Bytes(), &payload); err != nil {
		t.Fatalf("invalid json: %v\n%s", err, stdout.String())
	}
	data := payload["data"].(map[string]any)
	if data["cookie_name"] != "t" {
		t.Fatalf("cookie_name = %v, want t", data["cookie_name"])
	}
	if !strings.Contains(data["recommended_next"].(string), "--token-stdin") {
		t.Fatalf("recommended_next missing stdin guidance: %v", data["recommended_next"])
	}
}

func TestAuthCookieTokenArgRejectedByDefault(t *testing.T) {
	t.Setenv("DIDA_CONFIG_DIR", t.TempDir())
	var stdout, stderr bytes.Buffer
	code := Run([]string{"auth", "cookie", "set", "--token", "secret-token", "--json"}, "test-version", &stdout, &stderr)
	if code != 1 {
		t.Fatalf("exit code = %d, want 1", code)
	}
	if stderr.Len() != 0 {
		t.Fatalf("stderr = %q, want empty for json errors", stderr.String())
	}
	text := stdout.String()
	if strings.Contains(text, "secret-token") {
		t.Fatalf("json error leaked token: %s", text)
	}
	if !strings.Contains(text, "--token-stdin") {
		t.Fatalf("stdout missing stdin hint: %s", text)
	}
}

func TestAuthCookieTokenArgAllowedWithOptIn(t *testing.T) {
	t.Setenv("DIDA_CONFIG_DIR", t.TempDir())
	t.Setenv("DIDA_ALLOW_TOKEN_ARG", "1")
	var stdout, stderr bytes.Buffer
	code := Run([]string{"auth", "cookie", "set", "--token", "abc123", "--json"}, "test-version", &stdout, &stderr)
	if code != 0 {
		t.Fatalf("exit code = %d, stderr=%s stdout=%s", code, stderr.String(), stdout.String())
	}
	if strings.Contains(stdout.String(), "abc123") {
		t.Fatalf("json output leaked token: %s", stdout.String())
	}
}

func TestParseTokenInputRejectsLargeStdin(t *testing.T) {
	oldStdin := os.Stdin
	readFile, writeFile, err := os.Pipe()
	if err != nil {
		t.Fatalf("Pipe() error = %v", err)
	}
	defer func() { os.Stdin = oldStdin }()
	os.Stdin = readFile
	go func() {
		_, _ = writeFile.Write(bytes.Repeat([]byte("x"), int(maxTokenStdinBytes)+1))
		_ = writeFile.Close()
	}()
	_, err = parseTokenInput([]string{"--token-stdin"})
	if err == nil {
		t.Fatalf("parseTokenInput() error = nil, want size error")
	}
	if !strings.Contains(err.Error(), "exceeded") {
		t.Fatalf("error = %v, want exceeded", err)
	}
}

func TestShortcutTodayPreservesFlags(t *testing.T) {
	t.Setenv("DIDA_CONFIG_DIR", t.TempDir())
	var stdout, stderr bytes.Buffer
	code := Run([]string{"+today", "--limit", "2", "--json"}, "test-version", &stdout, &stderr)
	if code != 1 {
		t.Fatalf("exit code = %d, want 1", code)
	}
	if !strings.Contains(stdout.String(), `"command": "task list"`) {
		t.Fatalf("stdout missing task list envelope: %s", stdout.String())
	}
}

func TestTaskListRejectsNegativeLimit(t *testing.T) {
	var stdout, stderr bytes.Buffer
	code := Run([]string{"task", "list", "--limit", "-1", "--json"}, "test-version", &stdout, &stderr)
	if code != 1 {
		t.Fatalf("exit code = %d, want 1", code)
	}
	var payload map[string]any
	if err := json.Unmarshal(stdout.Bytes(), &payload); err != nil {
		t.Fatalf("invalid json: %v\n%s", err, stdout.String())
	}
	errPayload := payload["error"].(map[string]any)
	if errPayload["type"] != "validation" {
		t.Fatalf("error.type = %v, want validation", errPayload["type"])
	}
}

func TestProjectTasksArgsLimit(t *testing.T) {
	projectID, limit, compact, err := parseProjectTasksArgs([]string{"p1", "--limit", "2", "--compact"})
	if err != nil {
		t.Fatalf("parseProjectTasksArgs() error = %v", err)
	}
	if projectID != "p1" || limit != 2 || !compact {
		t.Fatalf("projectID=%q limit=%d compact=%v", projectID, limit, compact)
	}
	_, _, _, err = parseProjectTasksArgs([]string{"p1", "--limit", "-1"})
	if err == nil {
		t.Fatalf("parseProjectTasksArgs() error = nil, want error")
	}
}

func TestParseRangeListFlags(t *testing.T) {
	now := time.Date(2026, 5, 9, 12, 0, 0, 0, time.Local)
	opts, err := parseRangeListFlags([]string{"--from", "2026-05-01", "--to", "2026-05-09", "--limit", "3"}, now, 30)
	if err != nil {
		t.Fatalf("parseRangeListFlags() error = %v", err)
	}
	if opts.Limit != 3 {
		t.Fatalf("limit = %d, want 3", opts.Limit)
	}
	if opts.From.Format("2006-01-02 15:04:05") != "2026-05-01 00:00:00" {
		t.Fatalf("from = %s", opts.From)
	}
	if opts.To.Format("2006-01-02 15:04:05") != "2026-05-09 23:59:59" {
		t.Fatalf("to = %s", opts.To)
	}
	_, err = parseRangeListFlags([]string{"--limit", "-1"}, now, 30)
	if err == nil {
		t.Fatalf("parseRangeListFlags() error = nil, want error")
	}
}

func TestTaskOutputCompactOmitsLargeFields(t *testing.T) {
	out := taskOutput([]model.Task{{
		ID:        "t1",
		ProjectID: "p1",
		Title:     "Compact me",
		Content:   strings.Repeat("large ", 20),
		Desc:      strings.Repeat("markdown ", 20),
		Items:     []map[string]any{{"title": "step"}},
		Reminders: []any{"TRIGGER:P0DT9H0M0S"},
		Raw:       map[string]any{"huge": true},
	}}, true)
	data, err := json.Marshal(out)
	if err != nil {
		t.Fatalf("marshal compact output: %v", err)
	}
	text := string(data)
	for _, forbidden := range []string{"content", "desc", "items", "reminders", "raw", "large", "markdown"} {
		if strings.Contains(text, forbidden) {
			t.Fatalf("compact output leaked %q: %s", forbidden, text)
		}
	}
	if !strings.Contains(text, `"title":"Compact me"`) {
		t.Fatalf("compact output missing title: %s", text)
	}
}

func TestSyncPayloadValueAcceptsPointer(t *testing.T) {
	payload := syncPayloadValue(&webapi.SyncPayload{
		CheckPoint: 42,
		Tasks:      []map[string]any{{"id": "t1"}},
	})
	if payload.CheckPoint != 42 || len(payload.Tasks) != 1 {
		t.Fatalf("payload = %#v", payload)
	}
}

func TestUnknownCommandText(t *testing.T) {
	var stdout, stderr bytes.Buffer
	code := Run([]string{"nope"}, "test-version", &stdout, &stderr)
	if code != 1 {
		t.Fatalf("exit code = %d, want 1", code)
	}
	if !strings.Contains(stderr.String(), `unknown command "nope"`) {
		t.Fatalf("stderr = %q", stderr.String())
	}
}

func TestTaskCreateDryRunJSON(t *testing.T) {
	var stdout, stderr bytes.Buffer
	code := Run([]string{"task", "create", "--project", "p1", "--title", "Smoke", "--tag", "agent", "--tags", "work,deep", "--item", "Step 1", "--column", "c1", "--all-day", "--dry-run", "--json"}, "test-version", &stdout, &stderr)
	if code != 0 {
		t.Fatalf("exit code = %d, stderr=%s", code, stderr.String())
	}
	var payload map[string]any
	if err := json.Unmarshal(stdout.Bytes(), &payload); err != nil {
		t.Fatalf("invalid json: %v\n%s", err, stdout.String())
	}
	if payload["command"] != "task create" {
		t.Fatalf("command = %v, want task create", payload["command"])
	}
	data := payload["data"].(map[string]any)
	if data["dryRun"] != true {
		t.Fatalf("dryRun = %v, want true", data["dryRun"])
	}
	if !strings.Contains(data["hint"].(string), "remove --dry-run") {
		t.Fatalf("hint = %v", data["hint"])
	}
	requestPayload := data["payload"].(map[string]any)
	add := requestPayload["add"].([]any)
	task := add[0].(map[string]any)
	if task["columnId"] != "c1" {
		t.Fatalf("columnId = %v, want c1", task["columnId"])
	}
	if tags := task["tags"].([]any); len(tags) != 3 {
		t.Fatalf("tags len = %d, want 3", len(tags))
	}
	if items := task["items"].([]any); len(items) != 1 {
		t.Fatalf("items len = %d, want 1", len(items))
	}
	if task["allDay"] != true {
		t.Fatalf("allDay = %v, want true", task["allDay"])
	}
}

func TestTaskDeleteRequiresYesJSON(t *testing.T) {
	var stdout, stderr bytes.Buffer
	code := Run([]string{"task", "delete", "t1", "--project", "p1", "--json"}, "test-version", &stdout, &stderr)
	if code != 1 {
		t.Fatalf("exit code = %d, want 1", code)
	}
	if stderr.Len() != 0 {
		t.Fatalf("stderr = %q, want empty for json errors", stderr.String())
	}
	var payload map[string]any
	if err := json.Unmarshal(stdout.Bytes(), &payload); err != nil {
		t.Fatalf("invalid json: %v\n%s", err, stdout.String())
	}
	errPayload := payload["error"].(map[string]any)
	if errPayload["type"] != "confirmation_required" {
		t.Fatalf("error.type = %v, want confirmation_required", errPayload["type"])
	}
}

func TestCommentCreateDryRunJSON(t *testing.T) {
	var stdout, stderr bytes.Buffer
	code := Run([]string{"comment", "create", "--project", "p1", "--task", "t1", "--text", "Looks good", "--dry-run", "--json"}, "test-version", &stdout, &stderr)
	if code != 0 {
		t.Fatalf("exit code = %d, stderr=%s", code, stderr.String())
	}
	var payload map[string]any
	if err := json.Unmarshal(stdout.Bytes(), &payload); err != nil {
		t.Fatalf("invalid json: %v\n%s", err, stdout.String())
	}
	if payload["command"] != "comment create" {
		t.Fatalf("command = %v, want comment create", payload["command"])
	}
	data := payload["data"].(map[string]any)
	if data["dryRun"] != true {
		t.Fatalf("dryRun = %v, want true", data["dryRun"])
	}
	requestPayload := data["payload"].(map[string]any)
	comment := requestPayload["comment"].(map[string]any)
	if comment["title"] != "Looks good" {
		t.Fatalf("comment title = %v, want Looks good", comment["title"])
	}
}

func TestCommentCreateWithFileDryRunJSON(t *testing.T) {
	var stdout, stderr bytes.Buffer
	code := Run([]string{"comment", "create", "--project", "p1", "--task", "t1", "--text", "Looks good", "--file", "probe.png", "--dry-run", "--json"}, "test-version", &stdout, &stderr)
	if code != 0 {
		t.Fatalf("exit code = %d, stderr=%s", code, stderr.String())
	}
	var payload map[string]any
	if err := json.Unmarshal(stdout.Bytes(), &payload); err != nil {
		t.Fatalf("invalid json: %v\n%s", err, stdout.String())
	}
	if payload["command"] != "comment create" {
		t.Fatalf("command = %v, want comment create", payload["command"])
	}
	data := payload["data"].(map[string]any)
	requestPayload := data["payload"].(map[string]any)
	files := requestPayload["files"].([]any)
	if len(files) != 1 || files[0].(map[string]any)["name"] != "probe.png" {
		t.Fatalf("files = %#v", files)
	}
	upload := requestPayload["upload"].(map[string]any)
	if upload["field"] != "file" {
		t.Fatalf("upload = %#v", upload)
	}
}

func TestCommentDeleteRequiresYesJSON(t *testing.T) {
	var stdout, stderr bytes.Buffer
	code := Run([]string{"comment", "delete", "--project", "p1", "--task", "t1", "--comment", "c1", "--json"}, "test-version", &stdout, &stderr)
	if code != 1 {
		t.Fatalf("exit code = %d, want 1", code)
	}
	if stderr.Len() != 0 {
		t.Fatalf("stderr = %q, want empty for json errors", stderr.String())
	}
	var payload map[string]any
	if err := json.Unmarshal(stdout.Bytes(), &payload); err != nil {
		t.Fatalf("invalid json: %v\n%s", err, stdout.String())
	}
	errPayload := payload["error"].(map[string]any)
	if errPayload["type"] != "confirmation_required" {
		t.Fatalf("error.type = %v, want confirmation_required", errPayload["type"])
	}
}

func TestFilterListMissingAuthJSON(t *testing.T) {
	t.Setenv("DIDA_CONFIG_DIR", t.TempDir())
	var stdout, stderr bytes.Buffer
	code := Run([]string{"filter", "list", "--json"}, "test-version", &stdout, &stderr)
	if code != 1 {
		t.Fatalf("exit code = %d, want 1", code)
	}
	if stderr.Len() != 0 {
		t.Fatalf("stderr = %q, want empty for json errors", stderr.String())
	}
	var payload map[string]any
	if err := json.Unmarshal(stdout.Bytes(), &payload); err != nil {
		t.Fatalf("invalid json: %v\n%s", err, stdout.String())
	}
	if payload["command"] != "filter list" {
		t.Fatalf("command = %v, want filter list", payload["command"])
	}
	errPayload := payload["error"].(map[string]any)
	if errPayload["type"] != "auth" {
		t.Fatalf("error.type = %v, want auth", errPayload["type"])
	}
}

func TestSchemaListJSONDoesNotRequireAuth(t *testing.T) {
	t.Setenv("DIDA_CONFIG_DIR", t.TempDir())
	var stdout, stderr bytes.Buffer
	code := Run([]string{"schema", "list", "--json"}, "test-version", &stdout, &stderr)
	if code != 0 {
		t.Fatalf("exit code = %d, stderr=%s", code, stderr.String())
	}
	var payload map[string]any
	if err := json.Unmarshal(stdout.Bytes(), &payload); err != nil {
		t.Fatalf("invalid json: %v\n%s", err, stdout.String())
	}
	if payload["command"] != "schema list" {
		t.Fatalf("command = %v, want schema list", payload["command"])
	}
	data := payload["data"].(map[string]any)
	schemas := data["schemas"].([]any)
	if len(schemas) < 20 {
		t.Fatalf("schema count = %d, want broad command coverage", len(schemas))
	}
}

func TestSchemaShowTaskCreateJSON(t *testing.T) {
	var stdout, stderr bytes.Buffer
	code := Run([]string{"schema", "show", "task.create", "--json"}, "test-version", &stdout, &stderr)
	if code != 0 {
		t.Fatalf("exit code = %d, stderr=%s", code, stderr.String())
	}
	var payload map[string]any
	if err := json.Unmarshal(stdout.Bytes(), &payload); err != nil {
		t.Fatalf("invalid json: %v\n%s", err, stdout.String())
	}
	data := payload["data"].(map[string]any)
	schema := data["schema"].(map[string]any)
	if schema["command"] != "dida task create --project <project-id> --title <title> --dry-run --json" {
		t.Fatalf("command = %v", schema["command"])
	}
	if schema["dryRun"] != true {
		t.Fatalf("dryRun = %v, want true", schema["dryRun"])
	}
}

func TestSchemaAuthMetadataForOfficialAndOpenAPI(t *testing.T) {
	cases := map[string]bool{
		"official.doctor":        true,
		"official.project.list":  true,
		"official.habit.list":    true,
		"official.task.get":      true,
		"official.task.batchAdd": true,
		"openapi.doctor":         false,
		"openapi.clientSet":      false,
		"openapi.authUrl":        true,
		"openapi.projectList":    true,
		"openapi.taskCreate":     true,
	}
	for id, want := range cases {
		t.Run(id, func(t *testing.T) {
			var stdout, stderr bytes.Buffer
			code := Run([]string{"schema", "show", id, "--json"}, "test-version", &stdout, &stderr)
			if code != 0 {
				t.Fatalf("Run() code = %d stderr=%s stdout=%s", code, stderr.String(), stdout.String())
			}
			var payload map[string]any
			if err := json.Unmarshal(stdout.Bytes(), &payload); err != nil {
				t.Fatalf("invalid json: %v\n%s", err, stdout.String())
			}
			data := payload["data"].(map[string]any)
			schema := data["schema"].(map[string]any)
			if schema["authRequired"] != want {
				t.Fatalf("authRequired = %v, want %v", schema["authRequired"], want)
			}
		})
	}
}

func TestSchemaShowUnknownJSON(t *testing.T) {
	var stdout, stderr bytes.Buffer
	code := Run([]string{"schema", "show", "missing.id", "--json"}, "test-version", &stdout, &stderr)
	if code != 1 {
		t.Fatalf("exit code = %d, want 1", code)
	}
	if stderr.Len() != 0 {
		t.Fatalf("stderr = %q, want empty for json errors", stderr.String())
	}
	var payload map[string]any
	if err := json.Unmarshal(stdout.Bytes(), &payload); err != nil {
		t.Fatalf("invalid json: %v\n%s", err, stdout.String())
	}
	errPayload := payload["error"].(map[string]any)
	if errPayload["type"] != "not_found" {
		t.Fatalf("error.type = %v, want not_found", errPayload["type"])
	}
}

func TestRawAPIErrorEnvelopeIncludesProbeDetails(t *testing.T) {
	err := &webapi.APIError{
		Method:      "GET",
		Path:        "/task/activity/t1",
		StatusCode:  http.StatusInternalServerError,
		BodySnippet: `{"error":"need_pro"}`,
	}
	var stdout bytes.Buffer
	code := failRawAPIError(err, "v1", &stdout)
	if code != 1 {
		t.Fatalf("code = %d, want 1", code)
	}
	var payload map[string]any
	if err := json.Unmarshal(stdout.Bytes(), &payload); err != nil {
		t.Fatalf("invalid json: %v\n%s", err, stdout.String())
	}
	errorPayload := payload["error"].(map[string]any)
	details := errorPayload["details"].(map[string]any)
	if details["statusCode"] != float64(http.StatusInternalServerError) {
		t.Fatalf("details = %#v", details)
	}
	if details["bodySnippet"] != `{"error":"need_pro"}` {
		t.Fatalf("bodySnippet = %#v", details["bodySnippet"])
	}
}

func TestSettingsGetIncludeWebJSON(t *testing.T) {
	t.Setenv("DIDA_CONFIG_DIR", t.TempDir())
	var stdout, stderr bytes.Buffer
	code := Run([]string{"settings", "get", "--include-web", "--json"}, "test-version", &stdout, &stderr)
	if code != 1 {
		t.Fatalf("exit code = %d, want 1 because auth is missing", code)
	}
	var payload map[string]any
	if err := json.Unmarshal(stdout.Bytes(), &payload); err != nil {
		t.Fatalf("invalid json: %v\n%s", err, stdout.String())
	}
	if payload["command"] != "settings get" {
		t.Fatalf("command = %v, want settings get", payload["command"])
	}
}

func TestParseClosedListFlags(t *testing.T) {
	now := time.Date(2026, 5, 9, 12, 0, 0, 0, time.Local)
	opts, err := parseClosedListFlags([]string{"--project", "p1", "--project", "p2", "--status", "2", "--from", "2026-05-01", "--to", "2026-05-09", "--completed-user", "u1", "--limit", "3"}, now)
	if err != nil {
		t.Fatalf("parseClosedListFlags() error = %v", err)
	}
	if len(opts.ProjectIDs) != 2 || opts.ProjectIDs[0] != "p1" || opts.ProjectIDs[1] != "p2" {
		t.Fatalf("projectIds = %#v", opts.ProjectIDs)
	}
	if len(opts.Statuses) != 1 || opts.Statuses[0] != 2 {
		t.Fatalf("statuses = %#v", opts.Statuses)
	}
	if opts.From != "2026-05-01 00:00:00" || opts.To != "2026-05-09 23:59:59" {
		t.Fatalf("from/to = %q / %q", opts.From, opts.To)
	}
	if opts.CompletedUserID != "u1" || opts.Limit != 3 {
		t.Fatalf("completedUserID/limit = %q / %d", opts.CompletedUserID, opts.Limit)
	}
}

func TestParseTrashListFlags(t *testing.T) {
	opts, err := parseTrashListFlags([]string{"--cursor", "20", "--limit", "5", "--full"})
	if err != nil {
		t.Fatalf("parseTrashListFlags() error = %v", err)
	}
	if opts.Cursor != 20 || opts.Limit != 5 || opts.Compact {
		t.Fatalf("opts = %#v", opts)
	}
	_, err = parseTrashListFlags([]string{"--cursor", "-1"})
	if err == nil {
		t.Fatalf("parseTrashListFlags() error = nil, want cursor error")
	}
}

func TestTrashTaskOutputCompactOmitsLargeFields(t *testing.T) {
	out := trashTaskOutput([]map[string]any{{
		"id":          "t1",
		"projectId":   "p1",
		"title":       "Deleted",
		"kind":        "TEXT",
		"status":      float64(0),
		"priority":    float64(1),
		"deletedTime": float64(1778321394641),
		"content":     strings.Repeat("large", 20),
		"desc":        strings.Repeat("markdown", 20),
		"items":       []any{map[string]any{"title": "step"}},
	}}, true)
	encoded, err := json.Marshal(out)
	if err != nil {
		t.Fatalf("marshal compact trash: %v", err)
	}
	text := string(encoded)
	for _, forbidden := range []string{"content", "desc", "items", "large", "markdown"} {
		if strings.Contains(text, forbidden) {
			t.Fatalf("compact trash leaked %q: %s", forbidden, text)
		}
	}
	if !strings.Contains(text, "Deleted") {
		t.Fatalf("compact trash missing title: %s", text)
	}
}

func TestBuildAgentContextCompact(t *testing.T) {
	now := time.Unix(1893456000, 0) // 2030-01-01T00:00:00Z
	view := model.SyncView{
		InboxID: "inbox",
		Projects: []model.Project{{
			ID:   "p1",
			Name: "Inbox",
			Raw:  map[string]any{"large": true},
		}},
		ProjectGroups: []map[string]any{{"id": "g1", "name": "Work", "ignored": strings.Repeat("x", 20)}},
		Tags:          []map[string]any{{"name": "deep", "color": "#111111", "ignored": true}},
		Filters:       []map[string]any{{"id": "f1", "name": "Today", "ignored": true}},
		Tasks: []model.Task{{
			ID:        "t1",
			ProjectID: "p1",
			Title:     "Compact context",
			Content:   strings.Repeat("large ", 20),
			Desc:      strings.Repeat("markdown ", 20),
			DueUnix:   now.Add(time.Hour).Unix(),
			Priority:  5,
			Raw:       map[string]any{"large": true},
		}},
	}
	data, meta := buildAgentContext(view, agentContextOptions{Days: 14, Limit: 50, Compact: true}, now)
	if meta["today"] != 1 || meta["upcoming"] != 1 {
		t.Fatalf("meta = %#v", meta)
	}
	encoded, err := json.Marshal(data)
	if err != nil {
		t.Fatalf("marshal context: %v", err)
	}
	text := string(encoded)
	for _, forbidden := range []string{"content", "raw", "markdown", "ignored"} {
		if strings.Contains(text, forbidden) {
			t.Fatalf("agent context leaked %q: %s", forbidden, text)
		}
	}
	if !strings.Contains(text, "Compact context") {
		t.Fatalf("agent context missing task title: %s", text)
	}
}

func TestResourceCreateDryRunJSON(t *testing.T) {
	cases := [][]string{
		{"project", "create", "--name", "Smoke", "--dry-run", "--json"},
		{"folder", "create", "--name", "Smoke", "--dry-run", "--json"},
		{"tag", "create", "smoke", "--dry-run", "--json"},
		{"column", "create", "--project", "p1", "--name", "Doing", "--dry-run", "--json"},
	}
	for _, args := range cases {
		var stdout, stderr bytes.Buffer
		code := Run(args, "test-version", &stdout, &stderr)
		if code != 0 {
			t.Fatalf("%v exit code = %d, stderr=%s", args, code, stderr.String())
		}
		var payload map[string]any
		if err := json.Unmarshal(stdout.Bytes(), &payload); err != nil {
			t.Fatalf("%v invalid json: %v\n%s", args, err, stdout.String())
		}
		data := payload["data"].(map[string]any)
		if data["dryRun"] != true {
			t.Fatalf("%v dryRun = %v, want true", args, data["dryRun"])
		}
	}
}

func TestResourceDeleteRequiresYesJSON(t *testing.T) {
	cases := [][]string{
		{"project", "delete", "p1", "--json"},
		{"folder", "delete", "g1", "--json"},
		{"tag", "delete", "smoke", "--json"},
	}
	for _, args := range cases {
		var stdout, stderr bytes.Buffer
		code := Run(args, "test-version", &stdout, &stderr)
		if code != 1 {
			t.Fatalf("%v exit code = %d, want 1", args, code)
		}
		if stderr.Len() != 0 {
			t.Fatalf("%v stderr = %q, want empty for json errors", args, stderr.String())
		}
		var payload map[string]any
		if err := json.Unmarshal(stdout.Bytes(), &payload); err != nil {
			t.Fatalf("%v invalid json: %v\n%s", args, err, stdout.String())
		}
		errPayload := payload["error"].(map[string]any)
		if errPayload["type"] != "confirmation_required" {
			t.Fatalf("%v error.type = %v, want confirmation_required", args, errPayload["type"])
		}
	}
}

func TestFolderRejectsUnsupportedGroupFlag(t *testing.T) {
	var stdout, stderr bytes.Buffer
	code := Run([]string{"folder", "create", "--name", "Folder", "--group", "g1", "--dry-run", "--json"}, "test-version", &stdout, &stderr)
	if code != 1 {
		t.Fatalf("exit code = %d, want 1", code)
	}
	var payload map[string]any
	if err := json.Unmarshal(stdout.Bytes(), &payload); err != nil {
		t.Fatalf("invalid json: %v\n%s", err, stdout.String())
	}
	errPayload := payload["error"].(map[string]any)
	if errPayload["type"] != "validation" {
		t.Fatalf("error.type = %v, want validation", errPayload["type"])
	}
}

func TestCompactSearchPayloadDropsLargeFields(t *testing.T) {
	payload := map[string]any{
		"hits": []any{map[string]any{
			"id":    "h1",
			"index": "task",
			"source": map[string]any{
				"title":        "Needle",
				"projectId":    "p1",
				"modifiedTime": "now",
				"content":      strings.Repeat("large", 20),
				"desc":         strings.Repeat("markdown", 20),
			},
		}},
		"tasks": []any{map[string]any{
			"id":      "t1",
			"title":   "Needle",
			"content": strings.Repeat("large", 20),
			"desc":    strings.Repeat("markdown", 20),
			"items":   []any{map[string]any{"title": "step"}},
		}},
	}
	compactSearchPayload(payload)
	encoded, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("marshal compact search: %v", err)
	}
	text := string(encoded)
	for _, forbidden := range []string{"content", "desc", "items", "large", "markdown"} {
		if strings.Contains(text, forbidden) {
			t.Fatalf("compact search leaked %q: %s", forbidden, text)
		}
	}
	if !strings.Contains(text, "Needle") {
		t.Fatalf("compact search missing title: %s", text)
	}
}

func TestCompactUserOutputsStaySmall(t *testing.T) {
	status := compactUserStatus(map[string]any{
		"userId":   "u1",
		"phone":    "13500000000",
		"username": "user@example.com",
	})
	profile := compactUserProfile(map[string]any{
		"name":    "Tester",
		"phone":   "13500000000",
		"email":   "user@example.com",
		"picture": "https://example.com/a.png",
	})
	session := compactUserSession(map[string]any{
		"id":   "s1",
		"ip":   "1.2.3.4",
		"city": "Changsha",
		"deviceInfo": map[string]any{
			"platform": "windows",
			"os":       "Windows 11",
			"device":   "Chrome",
		},
	})
	encoded, err := json.Marshal(map[string]any{
		"status":  status,
		"profile": profile,
		"session": session,
	})
	if err != nil {
		t.Fatalf("marshal user compact output: %v", err)
	}
	text := string(encoded)
	for _, forbidden := range []string{"13500000000", "user@example.com", "1.2.3.4", "Changsha"} {
		if strings.Contains(text, forbidden) {
			t.Fatalf("compact user output leaked %q: %s", forbidden, text)
		}
	}
	for _, expected := range []string{"userId", "name", "deviceInfo"} {
		if !strings.Contains(text, expected) {
			t.Fatalf("compact user output missing %q: %s", expected, text)
		}
	}
}

func TestOfficialToolsFlagValidation(t *testing.T) {
	_, _, err := parseOfficialToolsFlags([]string{"--limit", "-1"})
	if err == nil {
		t.Fatalf("parseOfficialToolsFlags() error = nil, want validation error")
	}
}

func TestOfficialCallFlagParsing(t *testing.T) {
	tool, payload, err := parseOfficialCallFlags([]string{"list_projects", "--args-json", "{\"query\":\"today\"}"})
	if err != nil {
		t.Fatalf("parseOfficialCallFlags() error = %v", err)
	}
	if tool != "list_projects" {
		t.Fatalf("tool = %q, want list_projects", tool)
	}
	if payload["query"] != "today" {
		t.Fatalf("payload = %#v", payload)
	}
}

func TestOfficialPayloadFlagParsing(t *testing.T) {
	payload, err := parseOfficialPayloadFlags([]string{"--args-json", "{\"name\":\"Read\",\"goal\":1}"})
	if err != nil {
		t.Fatalf("parseOfficialPayloadFlags() error = %v", err)
	}
	if payload["name"] != "Read" {
		t.Fatalf("payload = %#v", payload)
	}
	if payload["goal"] != float64(1) {
		t.Fatalf("payload goal = %#v, want 1", payload["goal"])
	}
}

func TestOfficialTokenSetStatusJSON(t *testing.T) {
	t.Setenv("DIDA_CONFIG_DIR", t.TempDir())
	t.Setenv("DIDA365_TOKEN", "")
	oldStdin := os.Stdin
	readFile, writeFile, err := os.Pipe()
	if err != nil {
		t.Fatalf("Pipe() error = %v", err)
	}
	defer func() { os.Stdin = oldStdin }()
	os.Stdin = readFile
	go func() {
		_, _ = writeFile.Write([]byte("dp_test_secret_token"))
		_ = writeFile.Close()
	}()

	var stdout, stderr bytes.Buffer
	code := Run([]string{"official", "token", "set", "--token-stdin", "--json"}, "test-version", &stdout, &stderr)
	if code != 0 {
		t.Fatalf("Run(set) code = %d stderr=%s stdout=%s", code, stderr.String(), stdout.String())
	}
	if strings.Contains(stdout.String(), "dp_test_secret_token") {
		t.Fatalf("token set leaked token: %s", stdout.String())
	}

	stdout.Reset()
	stderr.Reset()
	code = Run([]string{"official", "token", "status", "--json"}, "test-version", &stdout, &stderr)
	if code != 0 {
		t.Fatalf("Run(status) code = %d stderr=%s stdout=%s", code, stderr.String(), stdout.String())
	}
	if strings.Contains(stdout.String(), "dp_test_secret_token") {
		t.Fatalf("token status leaked token: %s", stdout.String())
	}
	if !strings.Contains(stdout.String(), `"source": "config"`) {
		t.Fatalf("status missing config source: %s", stdout.String())
	}
}

func TestHabitCreateFlagParsingDoesNotRequireToolName(t *testing.T) {
	payload, err := parseHabitCreateArgs([]string{"--args-json", "{\"name\":\"Read\"}"})
	if err != nil {
		t.Fatalf("parseHabitCreateArgs() error = %v", err)
	}
	if payload["name"] != "Read" {
		t.Fatalf("payload = %#v", payload)
	}
}

func TestOfficialYesFlagParsing(t *testing.T) {
	confirmed, err := parseOfficialYesFlag([]string{"--yes"})
	if err != nil {
		t.Fatalf("parseOfficialYesFlag() error = %v", err)
	}
	if !confirmed {
		t.Fatalf("confirmed = false, want true")
	}

	confirmed, err = parseOfficialYesFlag(nil)
	if err != nil {
		t.Fatalf("parseOfficialYesFlag(nil) error = %v", err)
	}
	if confirmed {
		t.Fatalf("confirmed = true, want false")
	}
}

func TestOfficialTaskFilterFlagParsing(t *testing.T) {
	filter, err := parseOfficialTaskFilterFlags([]string{
		"--project", "p1",
		"--project", "p2",
		"--start", "2026-05-01T00:00:00+08:00",
		"--end", "2026-05-09T23:59:59+08:00",
		"--priority", "0,5",
		"--tag", "work,urgent",
		"--kind", "TEXT,CHECKLIST",
		"--status", "0,2",
	})
	if err != nil {
		t.Fatalf("parseOfficialTaskFilterFlags() error = %v", err)
	}
	if got := filter["startDate"]; got != "2026-05-01T00:00:00+08:00" {
		t.Fatalf("startDate = %#v", got)
	}
	if got := filter["projectIds"].([]string); len(got) != 2 || got[0] != "p1" || got[1] != "p2" {
		t.Fatalf("projectIds = %#v", got)
	}
	if got := filter["priority"].([]int); len(got) != 2 || got[0] != 0 || got[1] != 5 {
		t.Fatalf("priority = %#v", got)
	}
	if got := filter["tag"].([]string); len(got) != 2 || got[0] != "work" || got[1] != "urgent" {
		t.Fatalf("tag = %#v", got)
	}
}

func TestOfficialTaskSearchFlagParsing(t *testing.T) {
	query, err := parseOfficialTaskSearchFlags([]string{"--query", "today"})
	if err != nil {
		t.Fatalf("parseOfficialTaskSearchFlags() error = %v", err)
	}
	if query != "today" {
		t.Fatalf("query = %q, want today", query)
	}
}

func TestOfficialTaskQueryFlagParsing(t *testing.T) {
	query, err := parseOfficialTaskQueryFlags([]string{"--query", "next 7 days"})
	if err != nil {
		t.Fatalf("parseOfficialTaskQueryFlags() error = %v", err)
	}
	if query != "next 7 days" {
		t.Fatalf("query = %q, want next 7 days", query)
	}
}

func TestOfficialTaskGetFlagParsing(t *testing.T) {
	taskID, projectID, err := parseOfficialTaskGetFlags([]string{"t1", "--project", "p1"})
	if err != nil {
		t.Fatalf("parseOfficialTaskGetFlags() error = %v", err)
	}
	if taskID != "t1" || projectID != "p1" {
		t.Fatalf("taskID=%q projectID=%q", taskID, projectID)
	}
}

func TestOfficialTaskBatchDryRunDoesNotRequireToken(t *testing.T) {
	t.Setenv("DIDA365_TOKEN", "")
	var stdout, stderr bytes.Buffer
	code := Run([]string{"official", "task", "batch-add", "--args-json", "{\"tasks\":[{\"title\":\"Smoke\"}]}", "--dry-run", "--json"}, "test-version", &stdout, &stderr)
	if code != 0 {
		t.Fatalf("Run() code = %d stderr = %s stdout = %s", code, stderr.String(), stdout.String())
	}
	var payload map[string]any
	if err := json.Unmarshal(stdout.Bytes(), &payload); err != nil {
		t.Fatalf("decode stdout: %v\n%s", err, stdout.String())
	}
	data := payload["data"].(map[string]any)
	if data["dry_run"] != true || data["tool"] != "batch_add_tasks" {
		t.Fatalf("data = %#v", data)
	}
}

func TestOfficialTaskCompleteProjectFlagParsing(t *testing.T) {
	payload, dryRun, err := parseOfficialTaskCompleteProjectFlags([]string{"--project", "p1", "--task", "t1", "--tasks", "t2,t3", "--dry-run"})
	if err != nil {
		t.Fatalf("parseOfficialTaskCompleteProjectFlags() error = %v", err)
	}
	if !dryRun || payload["project_id"] != "p1" {
		t.Fatalf("payload = %#v dryRun = %v", payload, dryRun)
	}
	taskIDs := payload["task_ids"].([]string)
	if len(taskIDs) != 3 || taskIDs[0] != "t1" || taskIDs[2] != "t3" {
		t.Fatalf("task_ids = %#v", taskIDs)
	}
}

func TestOfficialFocusFlagParsing(t *testing.T) {
	getPayload, err := parseOfficialFocusIDTypeArgs([]string{"f1", "--type", "1"})
	if err != nil {
		t.Fatalf("parseOfficialFocusIDTypeArgs() error = %v", err)
	}
	if getPayload["focus_id"] != "f1" || getPayload["type"] != 1 {
		t.Fatalf("get payload = %#v", getPayload)
	}

	listPayload, err := parseFocusListArgs([]string{"--from-time", "2026-05-01T00:00:00+08:00", "--to-time", "2026-05-09T23:59:59+08:00", "--type", "0"})
	if err != nil {
		t.Fatalf("parseFocusListArgs() error = %v", err)
	}
	if listPayload["from_time"] != "2026-05-01T00:00:00+08:00" || listPayload["to_time"] != "2026-05-09T23:59:59+08:00" || listPayload["type"] != 0 {
		t.Fatalf("list payload = %#v", listPayload)
	}

	deletePayload, confirmed, dryRun, err := parseOfficialFocusDeleteArgs([]string{"f1", "--type", "1", "--yes", "--dry-run"})
	if err != nil {
		t.Fatalf("parseOfficialFocusDeleteArgs() error = %v", err)
	}
	if !confirmed || !dryRun || deletePayload["focus_id"] != "f1" || deletePayload["type"] != 1 {
		t.Fatalf("delete payload = %#v confirmed = %v dryRun = %v", deletePayload, confirmed, dryRun)
	}
}

func TestOfficialFocusListRequiresType(t *testing.T) {
	_, err := parseFocusListArgs([]string{"--from-time", "2026-05-01T00:00:00+08:00", "--to-time", "2026-05-09T23:59:59+08:00"})
	if err == nil {
		t.Fatalf("parseFocusListArgs() error = nil, want missing type")
	}
}

func TestOfficialHabitCheckinsFlagParsing(t *testing.T) {
	payload, err := parseOfficialHabitCheckinsArgs([]string{"--habit-ids", "h1,h2", "--from", "20260501", "--to", "20260510"})
	if err != nil {
		t.Fatalf("parseOfficialHabitCheckinsArgs() error = %v", err)
	}
	ids := payload["habit_ids"].([]string)
	if len(ids) != 2 || ids[0] != "h1" || ids[1] != "h2" {
		t.Fatalf("habit_ids = %#v", ids)
	}
	if payload["from_stamp"] != 20260501 || payload["to_stamp"] != 20260510 {
		t.Fatalf("payload = %#v", payload)
	}
}

func TestOfficialTaskCompleteProjectDryRunDoesNotRequireToken(t *testing.T) {
	t.Setenv("DIDA365_TOKEN", "")
	var stdout, stderr bytes.Buffer
	code := Run([]string{"official", "task", "complete-project", "--project", "p1", "--task", "t1", "--dry-run", "--json"}, "test-version", &stdout, &stderr)
	if code != 0 {
		t.Fatalf("Run() code = %d stderr = %s stdout = %s", code, stderr.String(), stdout.String())
	}
	if !strings.Contains(stdout.String(), "complete_tasks_in_project") {
		t.Fatalf("stdout missing tool name: %s", stdout.String())
	}
}

func TestOfficialHabitDryRunDoesNotRequireToken(t *testing.T) {
	t.Setenv("DIDA365_TOKEN", "")
	cases := [][]string{
		{"official", "habit", "create", "--args-json", "{\"name\":\"Read\",\"type\":\"Boolean\"}", "--dry-run", "--json"},
		{"official", "habit", "update", "h1", "--args-json", "{\"name\":\"Read more\"}", "--dry-run", "--json"},
		{"official", "habit", "checkin", "h1", "--date", "2026-05-10", "--value", "1", "--dry-run", "--json"},
		{"official", "focus", "delete", "f1", "--type", "0", "--dry-run", "--json"},
	}
	for _, args := range cases {
		var stdout, stderr bytes.Buffer
		code := Run(args, "test-version", &stdout, &stderr)
		if code != 0 {
			t.Fatalf("Run(%v) code = %d stderr = %s stdout = %s", args, code, stderr.String(), stdout.String())
		}
		if !strings.Contains(stdout.String(), `"dry_run": true`) {
			t.Fatalf("stdout missing dry_run for %v: %s", args, stdout.String())
		}
	}
}

func TestOpenAPIAuthURLFlagParsing(t *testing.T) {
	redirectURI, scope, state, err := parseOpenAPIAuthURLFlags([]string{"--redirect-uri", "http://127.0.0.1:17890/callback", "--scope", "tasks:read", "--state", "abc"})
	if err != nil {
		t.Fatalf("parseOpenAPIAuthURLFlags() error = %v", err)
	}
	if redirectURI == "" || scope != "tasks:read" || state != "abc" {
		t.Fatalf("values = %q %q %q", redirectURI, scope, state)
	}
}

func TestOpenAPIExchangeFlagParsing(t *testing.T) {
	code, redirectURI, scope, err := parseOpenAPIExchangeFlags([]string{"--code", "xyz", "--redirect-uri", "http://127.0.0.1:17890/callback", "--scope", "tasks:read"})
	if err != nil {
		t.Fatalf("parseOpenAPIExchangeFlags() error = %v", err)
	}
	if code != "xyz" || redirectURI == "" || scope != "tasks:read" {
		t.Fatalf("values = %q %q %q", code, redirectURI, scope)
	}
}

func TestOpenAPILoginFlagParsingAndCallbackNormalization(t *testing.T) {
	redirectURI, scope, state, host, port, timeout, noOpen, err := parseOpenAPILoginFlags([]string{"--browser", "--redirect-uri", "http://127.0.0.1:17999/callback", "--scope", "tasks:read", "--state", "abc", "--timeout", "7"})
	if err != nil {
		t.Fatalf("parseOpenAPILoginFlags() error = %v", err)
	}
	if scope != "tasks:read" || state != "abc" || timeout != 7*time.Second || noOpen {
		t.Fatalf("unexpected parsed values: scope=%q state=%q timeout=%v noOpen=%v", scope, state, timeout, noOpen)
	}
	redirectURI, host, port, err = normalizeOpenAPICallback(redirectURI, host, port)
	if err != nil {
		t.Fatalf("normalizeOpenAPICallback() error = %v", err)
	}
	if redirectURI != "http://127.0.0.1:17999/callback" || host != "127.0.0.1" || port != 17999 {
		t.Fatalf("callback = %q %q %d", redirectURI, host, port)
	}
}

func TestNormalizeOpenAPICallbackRejectsNonLocalShape(t *testing.T) {
	if _, _, _, err := normalizeOpenAPICallback("https://example.com/callback", "127.0.0.1", 17890); err == nil {
		t.Fatalf("normalizeOpenAPICallback() error = nil, want scheme error")
	}
	if _, _, _, err := normalizeOpenAPICallback("http://127.0.0.1:17890/not-callback", "127.0.0.1", 17890); err == nil {
		t.Fatalf("normalizeOpenAPICallback() error = nil, want path error")
	}
	if _, _, _, err := normalizeOpenAPICallback("http://example.com:17890/callback", "127.0.0.1", 17890); err == nil {
		t.Fatalf("normalizeOpenAPICallback() error = nil, want loopback host error")
	}
}

func TestOpenAPILoginJSONNoOpenFailsSingleEnvelope(t *testing.T) {
	var stdout, stderr bytes.Buffer
	code := Run([]string{"openapi", "login", "--no-open", "--json"}, "test-version", &stdout, &stderr)
	if code != 1 {
		t.Fatalf("Run() code = %d, want 1", code)
	}
	if stderr.Len() != 0 {
		t.Fatalf("stderr = %q, want empty for json errors", stderr.String())
	}
	decoder := json.NewDecoder(&stdout)
	var payload map[string]any
	if err := decoder.Decode(&payload); err != nil {
		t.Fatalf("decode first JSON envelope: %v\n%s", err, stdout.String())
	}
	var extra any
	if err := decoder.Decode(&extra); err != io.EOF {
		t.Fatalf("stdout contains multiple JSON values or trailing data: extra=%#v err=%v stdout=%s", extra, err, stdout.String())
	}
	if payload["command"] != "openapi login" || payload["ok"] != false {
		t.Fatalf("payload = %#v", payload)
	}
	errPayload := payload["error"].(map[string]any)
	if errPayload["type"] != "validation" {
		t.Fatalf("error.type = %v, want validation", errPayload["type"])
	}
}

func TestOpenAPIClientSetStatusJSON(t *testing.T) {
	t.Setenv("DIDA_CONFIG_DIR", t.TempDir())
	oldStdin := os.Stdin
	readFile, writeFile, err := os.Pipe()
	if err != nil {
		t.Fatalf("Pipe() error = %v", err)
	}
	defer func() { os.Stdin = oldStdin }()
	os.Stdin = readFile
	go func() {
		_, _ = writeFile.Write([]byte("client-secret"))
		_ = writeFile.Close()
	}()

	var stdout, stderr bytes.Buffer
	code := Run([]string{"openapi", "client", "set", "--id", "client-id", "--secret-stdin", "--json"}, "test-version", &stdout, &stderr)
	if code != 0 {
		t.Fatalf("Run(set) code = %d stderr=%s stdout=%s", code, stderr.String(), stdout.String())
	}
	if strings.Contains(stdout.String(), "client-secret") {
		t.Fatalf("client set leaked secret: %s", stdout.String())
	}

	stdout.Reset()
	stderr.Reset()
	code = Run([]string{"openapi", "client", "status", "--json"}, "test-version", &stdout, &stderr)
	if code != 0 {
		t.Fatalf("Run(status) code = %d stderr=%s stdout=%s", code, stderr.String(), stdout.String())
	}
	if strings.Contains(stdout.String(), "client-secret") {
		t.Fatalf("client status leaked secret: %s", stdout.String())
	}
	if !strings.Contains(stdout.String(), `"available": true`) {
		t.Fatalf("status missing available true: %s", stdout.String())
	}
}

func TestValidateOpenAPICallback(t *testing.T) {
	if err := validateOpenAPICallback("state-1", "code-1", "state-1"); err != nil {
		t.Fatalf("validateOpenAPICallback() error = %v", err)
	}
	if err := validateOpenAPICallback("state-1", "", "state-1"); err == nil {
		t.Fatalf("validateOpenAPICallback() error = nil, want missing code error")
	}
	if err := validateOpenAPICallback("state-1", "code-1", "state-2"); err == nil {
		t.Fatalf("validateOpenAPICallback() error = nil, want state mismatch")
	}
}

func TestOpenAPITaskTargetWriteFlagsRequireConfirmation(t *testing.T) {
	_, _, _, err := parseOpenAPITaskTargetWriteFlags([]string{"--project", "p1", "--task", "t1"}, true)
	if err == nil {
		t.Fatalf("parseOpenAPITaskTargetWriteFlags() error = nil, want confirmation error")
	}
	projectID, taskID, dryRun, err := parseOpenAPITaskTargetWriteFlags([]string{"--project", "p1", "--task", "t1", "--dry-run"}, true)
	if err != nil {
		t.Fatalf("parseOpenAPITaskTargetWriteFlags() error = %v", err)
	}
	if projectID != "p1" || taskID != "t1" || !dryRun {
		t.Fatalf("values = %q %q %v", projectID, taskID, dryRun)
	}
}

func TestOpenAPIJSONWriteFlags(t *testing.T) {
	payload, dryRun, err := parseOpenAPIJSONWriteFlags([]string{"--args-json", "{\"title\":\"Task\"}", "--dry-run"})
	if err != nil {
		t.Fatalf("parseOpenAPIJSONWriteFlags() error = %v", err)
	}
	if payload["title"] != "Task" || !dryRun {
		t.Fatalf("payload = %#v dryRun = %v", payload, dryRun)
	}
}

func TestOpenAPIProjectDryRunDoesNotRequireToken(t *testing.T) {
	t.Setenv("DIDA_CONFIG_DIR", t.TempDir())
	var stdout, stderr bytes.Buffer
	code := Run([]string{"openapi", "project", "create", "--args-json", "{\"name\":\"Smoke\",\"viewMode\":\"list\",\"kind\":\"TASK\"}", "--dry-run", "--json"}, "test-version", &stdout, &stderr)
	if code != 0 {
		t.Fatalf("Run(create dry-run) code = %d stderr = %s stdout = %s", code, stderr.String(), stdout.String())
	}
	if !strings.Contains(stdout.String(), `"dry_run": true`) {
		t.Fatalf("stdout missing dry_run: %s", stdout.String())
	}

	stdout.Reset()
	stderr.Reset()
	code = Run([]string{"openapi", "project", "update", "p1", "--args-json", "{\"name\":\"Renamed\"}", "--dry-run", "--json"}, "test-version", &stdout, &stderr)
	if code != 0 {
		t.Fatalf("Run(update dry-run) code = %d stderr = %s stdout = %s", code, stderr.String(), stdout.String())
	}
	if !strings.Contains(stdout.String(), `"project_id": "p1"`) {
		t.Fatalf("stdout missing project_id: %s", stdout.String())
	}
}

func TestOpenAPIProjectDeleteRequiresYesJSON(t *testing.T) {
	t.Setenv("DIDA_CONFIG_DIR", t.TempDir())
	if err := openapi.SaveToken(&openapi.TokenResponse{OAuthToken: openapi.OAuthToken{AccessToken: "test-token", CreatedAt: time.Now().Unix()}}); err != nil {
		t.Fatalf("SaveToken() error = %v", err)
	}
	var stdout, stderr bytes.Buffer
	code := Run([]string{"openapi", "project", "delete", "p1", "--json"}, "test-version", &stdout, &stderr)
	if code != 1 {
		t.Fatalf("Run() code = %d, want 1", code)
	}
	if stderr.Len() != 0 {
		t.Fatalf("stderr = %q, want empty for json errors", stderr.String())
	}
	var payload map[string]any
	if err := json.Unmarshal(stdout.Bytes(), &payload); err != nil {
		t.Fatalf("invalid json: %v\n%s", err, stdout.String())
	}
	errPayload := payload["error"].(map[string]any)
	if errPayload["type"] != "confirmation_required" {
		t.Fatalf("error.type = %v, want confirmation_required", errPayload["type"])
	}

	stdout.Reset()
	stderr.Reset()
	code = Run([]string{"openapi", "project", "delete", "p1", "--dry-run", "--json"}, "test-version", &stdout, &stderr)
	if code != 0 {
		t.Fatalf("dry-run code = %d stderr=%s stdout=%s", code, stderr.String(), stdout.String())
	}
	if !strings.Contains(stdout.String(), `"dry_run": true`) {
		t.Fatalf("stdout missing dry_run: %s", stdout.String())
	}
}

func TestOpenAPIFocusFlagParsing(t *testing.T) {
	from, to, focusType, err := parseOpenAPIFocusListFlags([]string{"--from", "2026-04-01T00:00:00+0800", "--to", "2026-04-02T00:00:00+0800", "--type", "1"})
	if err != nil {
		t.Fatalf("parseOpenAPIFocusListFlags() error = %v", err)
	}
	if from == "" || to == "" || focusType != "1" {
		t.Fatalf("values = %q %q %q", from, to, focusType)
	}

	focusID, focusType, dryRun, yes, err := parseOpenAPIFocusDeleteFlags([]string{"f1", "--type", "0", "--dry-run", "--yes"})
	if err != nil {
		t.Fatalf("parseOpenAPIFocusDeleteFlags() error = %v", err)
	}
	if focusID != "f1" || focusType != "0" || !dryRun || !yes {
		t.Fatalf("values = %q %q %v %v", focusID, focusType, dryRun, yes)
	}
}

func TestOpenAPIHabitFlagParsing(t *testing.T) {
	habitID, payload, dryRun, err := parseOpenAPIIDJSONWriteFlags([]string{"h1", "--args-json", "{\"stamp\":20260407}", "--dry-run"}, "habit")
	if err != nil {
		t.Fatalf("parseOpenAPIIDJSONWriteFlags() error = %v", err)
	}
	if habitID != "h1" || payload["stamp"] != float64(20260407) || !dryRun {
		t.Fatalf("values = %q %#v %v", habitID, payload, dryRun)
	}

	habitIDs, from, to, err := parseOpenAPIHabitCheckinsFlags([]string{"--habit-ids", "h1,h2", "--from", "20260401", "--to", "20260407"})
	if err != nil {
		t.Fatalf("parseOpenAPIHabitCheckinsFlags() error = %v", err)
	}
	if habitIDs != "h1,h2" || from != "20260401" || to != "20260407" {
		t.Fatalf("values = %q %q %q", habitIDs, from, to)
	}
}

func TestOpenAPIDryRunDoesNotRequireToken(t *testing.T) {
	t.Setenv("DIDA_CONFIG_DIR", t.TempDir())
	var stdout, stderr bytes.Buffer
	code := Run([]string{"openapi", "habit", "checkin", "h1", "--args-json", "{\"stamp\":20260407}", "--dry-run", "--json"}, "test-version", &stdout, &stderr)
	if code != 0 {
		t.Fatalf("Run() code = %d stderr = %s stdout = %s", code, stderr.String(), stdout.String())
	}
	var payload map[string]any
	if err := json.Unmarshal(stdout.Bytes(), &payload); err != nil {
		t.Fatalf("decode stdout: %v", err)
	}
	if payload["ok"] != true {
		t.Fatalf("payload = %#v", payload)
	}
}
