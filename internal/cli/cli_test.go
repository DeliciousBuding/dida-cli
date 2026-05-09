package cli

import (
	"bytes"
	"encoding/json"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/DeliciousBuding/dida-cli/internal/model"
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
