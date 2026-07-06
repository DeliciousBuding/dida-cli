package e2e

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/DeliciousBuding/dida-cli/internal/auth"
	"github.com/DeliciousBuding/dida-cli/internal/cli"
	"github.com/DeliciousBuding/dida-cli/internal/openapi"
)

// === Helpers ===

func saveCookieToken(t *testing.T, token string) {
	t.Helper()
	if _, err := auth.SaveCookieToken(token); err != nil {
		t.Fatalf("save cookie token: %v", err)
	}
}

func saveOAuthToken(t *testing.T, accessToken string) {
	t.Helper()
	if err := openapi.SaveToken(&openapi.TokenResponse{
		OAuthToken: openapi.OAuthToken{
			AccessToken: accessToken,
			TokenType:   "Bearer",
			CreatedAt:   time.Now().Unix(),
		},
	}); err != nil {
		t.Fatalf("save oauth token: %v", err)
	}
}

func mustJSON(t *testing.T, raw []byte) map[string]any {
	t.Helper()
	var payload map[string]any
	if err := json.Unmarshal(raw, &payload); err != nil {
		t.Fatalf("invalid json: %v\n%s", err, string(raw))
	}
	return payload
}

func assertOK(t *testing.T, payload map[string]any) {
	t.Helper()
	if ok, _ := payload["ok"].(bool); !ok {
		t.Fatalf("ok = %v, want true. payload: %#v", payload["ok"], payload)
	}
}

func assertFail(t *testing.T, payload map[string]any) {
	t.Helper()
	if ok, _ := payload["ok"].(bool); ok {
		t.Fatalf("ok = %v, want false. payload: %#v", payload["ok"], payload)
	}
}

// === Mock Dida365 Web API ===

type webAPIMock struct {
	syncPayload string
	taskResult  string
	binaryData  []byte
	binaryType  string
	// record last task POST body
	lastTaskBody []byte
}

func (m *webAPIMock) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch {
	case r.Method == http.MethodGet && r.URL.Path == "/batch/check/0":
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(m.syncPayload))
	case r.Method == http.MethodPost && r.URL.Path == "/batch/task":
		m.lastTaskBody, _ = io.ReadAll(r.Body)
		w.Header().Set("Content-Type", "application/json")
		if m.taskResult != "" {
			w.Write([]byte(m.taskResult))
		} else {
			w.Write([]byte(`{"id":"task-new-001","ok":true}`))
		}
	case r.URL.Path == "/attachment/p1/t1/a1" && r.URL.Query().Get("action") == "download":
		w.Header().Set("Content-Type", m.binaryType)
		w.Write(m.binaryData)
	default:
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(`{"error":"not found"}`))
	}
}

func makeSyncPayload() string {
	return `{
	"inboxId": "inbox-001",
	"checkPoint": 42,
	"syncTaskBean": {
		"add": [
			{
				"id": "t1",
				"projectId": "inbox-001",
				"title": "Buy groceries",
				"priority": 0,
				"status": 0,
				"deleted": 0,
				"createdTime": "2026-06-25T09:00:00+08:00",
				"modifiedTime": "2026-06-25T09:00:00+08:00",
				"dueDate": "2026-06-25T12:00:00+08:00",
				"tags": ["personal"],
				"timeZone": "Asia/Shanghai"
			},
			{
				"id": "t2",
				"projectId": "p2",
				"title": "Write quarterly report",
				"priority": 5,
				"status": 0,
				"deleted": 0,
				"createdTime": "2026-06-24T08:00:00+08:00",
				"modifiedTime": "2026-06-24T08:00:00+08:00",
				"dueDate": "2026-06-24T17:00:00+08:00",
				"tags": ["work"],
				"timeZone": "Asia/Shanghai"
			},
			{
				"id": "t3",
				"projectId": "p2",
				"title": "Review design doc",
				"priority": 3,
				"status": 0,
				"deleted": 0,
				"createdTime": "2026-06-25T10:00:00+08:00",
				"modifiedTime": "2026-06-25T10:00:00+08:00",
				"dueDate": "2099-06-28T18:00:00+08:00",
				"tags": ["work", "deep"],
				"timeZone": "Asia/Shanghai",
				"items": [{"title": "Read spec", "status": 0}, {"title": "Write feedback", "status": 0}],
				"content": "Detailed review notes go here",
				"desc": "# Review\n\nCheck all sections."
			}
		]
	},
	"projects": [
		{"id": "inbox-001", "name": "Inbox", "closed": false, "color": "#246FE0"},
		{"id": "p2", "name": "Work", "closed": false, "color": "#DB4035", "kind": "TASK", "viewMode": "list"},
		{"id": "p3", "name": "Personal", "closed": false, "color": "#808080"}
	],
	"projectGroups": [
		{"id": "g1", "name": "Career", "sortOrder": 1},
		{"id": "g2", "name": "Home", "sortOrder": 2}
	],
	"tags": [
		{"name": "deep", "color": "#111111"},
		{"name": "personal", "color": "#246FE0"},
		{"name": "work", "color": "#DB4035"},
		{"name": "urgent", "color": "#FF9933"}
	],
	"filters": [
		{"id": "f1", "name": "Today", "query": "today"},
		{"id": "f2", "name": "High Priority", "query": "priority:high"}
	],
	"checks": []
}`
}

// === Test 1: Task Today with real project data ===

func TestE2E_TaskTodayWithRealProjectData(t *testing.T) {
	t.Setenv("DIDA_CONFIG_DIR", t.TempDir())
	saveCookieToken(t, "test-cookie-token-1234567890")

	mock := &webAPIMock{syncPayload: makeSyncPayload()}
	srv := httptest.NewServer(mock)
	defer srv.Close()
	t.Setenv("DIDA_WEBAPI_BASE_URL", srv.URL)

	var stdout, stderr bytes.Buffer
	code := cli.Run([]string{"task", "today", "--limit", "2", "--json"}, "e2e-test", &stdout, &stderr)
	if code != 0 {
		t.Fatalf("exit code = %d, stderr=%s", code, stderr.String())
	}
	if stderr.Len() != 0 {
		t.Fatalf("stderr = %q, want empty", stderr.String())
	}

	payload := mustJSON(t, stdout.Bytes())
	assertOK(t, payload)

	if cmd, _ := payload["command"].(string); cmd != "task list" {
		t.Fatalf("command = %q, want task list", cmd)
	}
	meta := payload["meta"].(map[string]any)
	if count, _ := meta["count"].(float64); count != 2 {
		t.Fatalf("meta.count = %v, want 2", count)
	}
	if total, _ := meta["total"].(float64); total != 2 {
		t.Fatalf("meta.total = %v, want 2 (all 3 tasks minus 1 with start date in future makes today=2)", total)
	}
	data := payload["data"].(map[string]any)
	tasks := data["tasks"].([]any)
	if len(tasks) != 2 {
		t.Fatalf("tasks len = %d, want 2", len(tasks))
	}
	task0 := tasks[0].(map[string]any)
	if task0["title"] != "Write quarterly report" {
		t.Fatalf("task[0].title = %v, want 'Write quarterly report'", task0["title"])
	}
	if task0["id"] != "t2" {
		t.Fatalf("task[0].id = %v, want t2", task0["id"])
	}
	task1 := tasks[1].(map[string]any)
	if task1["title"] != "Buy groceries" {
		t.Fatalf("task[1].title = %v, want 'Buy groceries'", task1["title"])
	}
	if task1["id"] != "t1" {
		t.Fatalf("task[1].id = %v, want t1", task1["id"])
	}
}

// === Test 2: Task Create and Read ===

func TestE2E_TaskCreateAndRead(t *testing.T) {
	t.Setenv("DIDA_CONFIG_DIR", t.TempDir())
	saveCookieToken(t, "test-cookie-token-1234567890")

	// Part A: create dry-run — no server needed, verify schema
	var stdout, stderr bytes.Buffer
	code := cli.Run([]string{
		"task", "create",
		"--project", "inbox-001",
		"--title", "E2E smoke test task",
		"--tag", "deep",
		"--tags", "work,urgent",
		"--item", "Step 1: verify",
		"--column", "c1",
		"--dry-run", "--json",
	}, "e2e-test", &stdout, &stderr)
	if code != 0 {
		t.Fatalf("exit code = %d, stderr=%s", code, stderr.String())
	}
	payload := mustJSON(t, stdout.Bytes())
	assertOK(t, payload)
	if cmd, _ := payload["command"].(string); cmd != "task create" {
		t.Fatalf("command = %q, want task create", cmd)
	}
	createData := payload["data"].(map[string]any)
	if dryRun, _ := createData["dryRun"].(bool); !dryRun {
		t.Fatal("dryRun = false, want true")
	}
	createPayload := createData["payload"].(map[string]any)
	add := createPayload["add"].([]any)
	task := add[0].(map[string]any)
	if task["projectId"] != "inbox-001" {
		t.Fatalf("projectId = %v, want inbox-001", task["projectId"])
	}
	if task["title"] != "E2E smoke test task" {
		t.Fatalf("title = %v", task["title"])
	}
	tags := task["tags"].([]any)
	if len(tags) != 3 {
		t.Fatalf("tags len = %d, want 3", len(tags))
	}
	items := task["items"].([]any)
	if len(items) != 1 {
		t.Fatalf("items len = %d, want 1", len(items))
	}

	// Part B: create real — fake server records POST /batch/task payload
	mock := &webAPIMock{
		syncPayload: makeSyncPayload(),
		taskResult:  `{"id":"task-new-001","created":true}`,
	}
	srv := httptest.NewServer(mock)
	defer srv.Close()
	t.Setenv("DIDA_WEBAPI_BASE_URL", srv.URL)

	stdout.Reset()
	stderr.Reset()
	code = cli.Run([]string{
		"task", "create",
		"--project", "inbox-001",
		"--title", "Real create test",
		"--priority", "5",
		"--json",
	}, "e2e-test", &stdout, &stderr)
	if code != 0 {
		t.Fatalf("exit code = %d, stderr=%s", code, stderr.String())
	}
	payload = mustJSON(t, stdout.Bytes())
	assertOK(t, payload)
	if cmd, _ := payload["command"].(string); cmd != "task create" {
		t.Fatalf("command = %q, want task create", cmd)
	}
	createResult := payload["data"].(map[string]any)
	if _, ok := createResult["result"]; !ok {
		t.Fatalf("data missing result: %#v", createResult)
	}
	if mock.lastTaskBody == nil {
		t.Fatal("server did not receive task POST body")
	}
	// Verify the server received title in POST body
	if !strings.Contains(string(mock.lastTaskBody), "Real create test") {
		t.Fatalf("server body missing title: %s", string(mock.lastTaskBody))
	}

	// Part C: read from sync data — fake server returns sync with the created task
	stdout.Reset()
	stderr.Reset()
	code = cli.Run([]string{"task", "get", "t1", "--json"}, "e2e-test", &stdout, &stderr)
	if code != 0 {
		t.Fatalf("exit code = %d, stderr=%s", code, stderr.String())
	}
	payload = mustJSON(t, stdout.Bytes())
	assertOK(t, payload)
	if cmd, _ := payload["command"].(string); cmd != "task get" {
		t.Fatalf("command = %q, want task get", cmd)
	}
	getData := payload["data"].(map[string]any)
	getTask := getData["task"].(map[string]any)
	if getTask["title"] != "Buy groceries" {
		t.Fatalf("task.title = %v, want 'Buy groceries'", getTask["title"])
	}
}

// === Test 3: Sync All and Agent Context ===

func TestE2E_SyncAllAndAgentContext(t *testing.T) {
	t.Setenv("DIDA_CONFIG_DIR", t.TempDir())
	saveCookieToken(t, "test-cookie-token-1234567890")

	mock := &webAPIMock{syncPayload: makeSyncPayload()}
	srv := httptest.NewServer(mock)
	defer srv.Close()
	t.Setenv("DIDA_WEBAPI_BASE_URL", srv.URL)

	// sync all
	var stdout, stderr bytes.Buffer
	code := cli.Run([]string{"sync", "all", "--json"}, "e2e-test", &stdout, &stderr)
	if code != 0 {
		t.Fatalf("exit code = %d, stderr=%s", code, stderr.String())
	}
	syncPayload := mustJSON(t, stdout.Bytes())
	assertOK(t, syncPayload)
	if cmd, _ := syncPayload["command"].(string); cmd != "sync all" {
		t.Fatalf("command = %q, want sync all", cmd)
	}
	syncMeta := syncPayload["meta"].(map[string]any)
	if cp, _ := syncMeta["checkpoint"].(float64); cp != 42 {
		t.Fatalf("checkpoint = %v, want 42", cp)
	}
	syncData := syncPayload["data"].(map[string]any)
	counts := syncData["counts"].(map[string]any)
	if tc, _ := counts["tasks"].(float64); tc != 3 {
		t.Fatalf("data.counts.tasks = %v, want 3", tc)
	}
	if pc, _ := counts["projects"].(float64); pc != 3 {
		t.Fatalf("data.counts.projects = %v, want 3", pc)
	}
	if gc, _ := counts["projectGroups"].(float64); gc != 2 {
		t.Fatalf("data.counts.projectGroups = %v, want 2", gc)
	}

	// agent context
	stdout.Reset()
	stderr.Reset()
	code = cli.Run([]string{"agent", "context", "--json"}, "e2e-test", &stdout, &stderr)
	if code != 0 {
		t.Fatalf("exit code = %d, stderr=%s", code, stderr.String())
	}
	agentPayload := mustJSON(t, stdout.Bytes())
	assertOK(t, agentPayload)
	if cmd, _ := agentPayload["command"].(string); cmd != "agent context" {
		t.Fatalf("command = %q, want agent context", cmd)
	}
	agentMeta := agentPayload["meta"].(map[string]any)
	if p, _ := agentMeta["projects"].(float64); p != 3 {
		t.Fatalf("agent meta.projects = %v, want 3", p)
	}
	if tg, _ := agentMeta["tags"].(float64); tg != 4 {
		t.Fatalf("agent meta.tags = %v, want 4", tg)
	}
	agentData := agentPayload["data"].(map[string]any)
	projects := agentData["projects"].([]any)
	if len(projects) != 3 {
		t.Fatalf("agent data.projects len = %d, want 3", len(projects))
	}
	// Verify today bucket exists
	if _, ok := agentData["today"]; !ok {
		t.Fatal("agent data missing 'today'")
	}
	if _, ok := agentData["upcoming"]; !ok {
		t.Fatal("agent data missing 'upcoming'")
	}
	if _, ok := agentData["tags"]; !ok {
		t.Fatal("agent data missing 'tags'")
	}
	if _, ok := agentData["projectGroups"]; !ok {
		t.Fatal("agent data missing 'projectGroups'")
	}
}

// === Test 4: Attachment Download ===

func TestE2E_AttachmentDownload(t *testing.T) {
	t.Setenv("DIDA_CONFIG_DIR", t.TempDir())
	saveCookieToken(t, "test-cookie-token-1234567890")

	expectedData := []byte("fake-binary-attachment-content-0123456789")
	mock := &webAPIMock{
		binaryData: expectedData,
		binaryType: "application/pdf",
	}
	srv := httptest.NewServer(mock)
	defer srv.Close()
	t.Setenv("DIDA_WEBAPI_BASE_URL_V1", srv.URL)

	outputDir := t.TempDir()
	output := filepath.Join(outputDir, "downloaded.pdf")

	var stdout, stderr bytes.Buffer
	code := cli.Run([]string{
		"attachment", "download",
		"--project", "p1",
		"--task", "t1",
		"--attachment", "a1",
		"--output", output,
		"--force",
		"--json",
	}, "e2e-test", &stdout, &stderr)
	if code != 0 {
		t.Fatalf("exit code = %d, stderr=%s", code, stderr.String())
	}

	payload := mustJSON(t, stdout.Bytes())
	assertOK(t, payload)
	if cmd, _ := payload["command"].(string); cmd != "attachment download" {
		t.Fatalf("command = %q, want attachment download", cmd)
	}
	downloadData := payload["data"].(map[string]any)
	if pid, _ := downloadData["projectId"].(string); pid != "p1" {
		t.Fatalf("data.projectId = %v, want p1", pid)
	}
	if size, _ := downloadData["bytes"].(float64); int64(size) != int64(len(expectedData)) {
		t.Fatalf("data.bytes = %v, want %d", size, len(expectedData))
	}
	if ct, _ := downloadData["contentType"].(string); ct != "application/pdf" {
		t.Fatalf("data.contentType = %v, want application/pdf", ct)
	}

	// Verify file contents
	actual, err := os.ReadFile(output)
	if err != nil {
		t.Fatalf("read downloaded file: %v", err)
	}
	if !bytes.Equal(actual, expectedData) {
		t.Fatalf("downloaded file mismatch: got %d bytes, want %d bytes", len(actual), len(expectedData))
	}
}

// === Mock MCP Server ===

type mcpMock struct {
	projectsJSON string
}

func (m *mcpMock) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var body map[string]any
	data, _ := io.ReadAll(r.Body)
	_ = json.Unmarshal(data, &body)

	method, _ := body["method"].(string)
	id, _ := body["id"].(float64)

	w.Header().Set("Content-Type", "application/json")

	switch method {
	case "initialize":
		w.Header().Set("Mcp-Session-Id", "mcp-session-001")
		resp := fmt.Sprintf(`{"jsonrpc":"2.0","id":%d,"result":{"protocolVersion":"2025-03-26","serverInfo":{"name":"Dida365 MCP","version":"1.0"},"capabilities":{"tools":{}}}}`, int(id))
		w.Write([]byte(resp))
	case "notifications/initialized":
		w.WriteHeader(http.StatusNoContent)
	case "tools/list":
		resp := fmt.Sprintf(`{"jsonrpc":"2.0","id":%d,"result":{"tools":[{"name":"list_projects","description":"List all projects","inputSchema":{},"outputSchema":{}}]}}`, int(id))
		w.Write([]byte(resp))
	case "tools/call":
		params, _ := body["params"].(map[string]any)
		toolName, _ := params["name"].(string)
		if toolName == "list_projects" {
			resp := fmt.Sprintf(`{"jsonrpc":"2.0","id":%d,"result":{"content":[{"type":"text","text":%s}]}}`, int(id), m.projectsJSON)
			w.Write([]byte(resp))
		} else {
			w.Write([]byte(fmt.Sprintf(`{"jsonrpc":"2.0","id":%d,"error":{"code":-32601,"message":"tool not found: %s"}}`, int(id), toolName)))
		}
	default:
		w.WriteHeader(http.StatusBadRequest)
	}
}

func makeMCPProjectsJSON() string {
	return `"[{\"id\":\"p1\",\"name\":\"Inbox\",\"color\":\"#246FE0\",\"closed\":false},{\"id\":\"p2\",\"name\":\"Work\",\"color\":\"#DB4035\",\"closed\":false,\"kind\":\"TASK\",\"viewMode\":\"list\"}]"`
}

// === Test 5: Official Project List ===

func TestE2E_OfficialProjectList(t *testing.T) {
	t.Setenv("DIDA_CONFIG_DIR", t.TempDir())
	t.Setenv("DIDA365_TOKEN", "test_mcp_token_12345678901234")

	mock := &mcpMock{projectsJSON: makeMCPProjectsJSON()}
	srv := httptest.NewServer(mock)
	defer srv.Close()
	t.Setenv("DIDA_MCP_BASE_URL", srv.URL)

	var stdout, stderr bytes.Buffer
	code := cli.Run([]string{"official", "project", "list", "--json"}, "e2e-test", &stdout, &stderr)
	if code != 0 {
		t.Fatalf("exit code = %d, stderr=%s", code, stderr.String())
	}

	payload := mustJSON(t, stdout.Bytes())
	assertOK(t, payload)
	if cmd, _ := payload["command"].(string); cmd != "official project list" {
		t.Fatalf("command = %q, want official project list", cmd)
	}

	// Data should contain the unwrapped project list
	data := payload["data"]
	// Try as array
	if arr, ok := data.([]any); ok {
		if len(arr) != 2 {
			t.Fatalf("project array len = %d, want 2", len(arr))
		}
		proj0 := arr[0].(map[string]any)
		if proj0["name"] != "Inbox" {
			t.Fatalf("project[0].name = %v, want Inbox", proj0["name"])
		}
	} else if m, ok := data.(map[string]any); ok {
		// Might be an envelope unwrap
		if results, ok := m["results"].([]any); ok {
			if len(results) != 2 {
				t.Fatalf("results len = %d, want 2", len(results))
			}
		} else {
			t.Fatalf("unexpected data shape: %#v", m)
		}
	}
}

// === Test 6: Official Task Batch-Add Dry Run (no server call) ===

func TestE2E_OfficialCallDryRun(t *testing.T) {
	t.Setenv("DIDA_CONFIG_DIR", t.TempDir())
	t.Setenv("DIDA365_TOKEN", "")

	var stdout, stderr bytes.Buffer
	code := cli.Run([]string{
		"official", "task", "batch-add",
		"--args-json", `{"tasks":[{"title":"E2E Smoke","projectId":"p1","priority":5}]}`,
		"--dry-run", "--json",
	}, "e2e-test", &stdout, &stderr)
	if code != 0 {
		t.Fatalf("exit code = %d, stderr=%s", code, stderr.String())
	}

	payload := mustJSON(t, stdout.Bytes())
	assertOK(t, payload)
	if cmd, _ := payload["command"].(string); cmd != "official task batch-add" {
		t.Fatalf("command = %q, want official task batch-add", cmd)
	}
	data := payload["data"].(map[string]any)
	if dryRun, _ := data["dry_run"].(bool); !dryRun {
		t.Fatal("dry_run = false, want true")
	}
	if tool, _ := data["tool"].(string); tool != "batch_add_tasks" {
		t.Fatalf("tool = %q, want batch_add_tasks", tool)
	}
	// Verify arguments shape
	args, _ := data["arguments"].(map[string]any)
	tasks, _ := args["tasks"].([]any)
	if len(tasks) != 1 {
		t.Fatalf("tasks len = %d, want 1", len(tasks))
	}
	task := tasks[0].(map[string]any)
	if task["title"] != "E2E Smoke" {
		t.Fatalf("task.title = %v, want E2E Smoke", task["title"])
	}
}

// === Mock OpenAPI Server ===

type openAPIMock struct {
	projectsJSON string
}

func (m *openAPIMock) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	switch {
	case r.Method == http.MethodGet && r.URL.Path == "/project":
		w.Write([]byte(m.projectsJSON))
	default:
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(`{"error":"not found"}`))
	}
}

func makeOpenAPIProjectsJSON() string {
	return `[{"id":"p1","name":"Inbox","color":"#246FE0","closed":false,"kind":"TASK","viewMode":"list"},{"id":"p2","name":"Work","color":"#DB4035","closed":false,"kind":"TASK","viewMode":"kanban"}]`
}

// === Test 7: OpenAPI Project List ===

func TestE2E_OpenAPIProjectList(t *testing.T) {
	t.Setenv("DIDA_CONFIG_DIR", t.TempDir())
	saveOAuthToken(t, "test-openapi-access-token-9876543210")

	mock := &openAPIMock{projectsJSON: makeOpenAPIProjectsJSON()}
	srv := httptest.NewServer(mock)
	defer srv.Close()
	t.Setenv("DIDA_OPENAPI_BASE_URL", srv.URL)

	var stdout, stderr bytes.Buffer
	code := cli.Run([]string{"openapi", "project", "list", "--json"}, "e2e-test", &stdout, &stderr)
	if code != 0 {
		t.Fatalf("exit code = %d, stderr=%s", code, stderr.String())
	}

	payload := mustJSON(t, stdout.Bytes())
	assertOK(t, payload)
	if cmd, _ := payload["command"].(string); cmd != "openapi project list" {
		t.Fatalf("command = %q, want openapi project list", cmd)
	}
	meta := payload["meta"].(map[string]any)
	if count, _ := meta["count"].(float64); count != 2 {
		t.Fatalf("meta.count = %v, want 2", count)
	}
	data := payload["data"].(map[string]any)
	projects := data["projects"].([]any)
	if len(projects) != 2 {
		t.Fatalf("projects len = %d, want 2", len(projects))
	}
	proj0 := projects[0].(map[string]any)
	if proj0["name"] != "Inbox" {
		t.Fatalf("project[0].name = %v, want Inbox", proj0["name"])
	}
	if proj0["closed"] != false {
		t.Fatalf("project[0].closed = %v, want false", proj0["closed"])
	}
	proj1 := projects[1].(map[string]any)
	if proj1["name"] != "Work" {
		t.Fatalf("project[1].name = %v, want Work", proj1["name"])
	}
}

// === Test 8: OpenAPI Task Create Dry Run (no token, no server) ===

func TestE2E_OpenAPITaskCreateDryRun(t *testing.T) {
	t.Setenv("DIDA_CONFIG_DIR", t.TempDir())

	var stdout, stderr bytes.Buffer
	code := cli.Run([]string{
		"openapi", "task", "create",
		"--args-json", `{"title":"E2E smoke","projectId":"p1","priority":5}`,
		"--dry-run", "--json",
	}, "e2e-test", &stdout, &stderr)
	if code != 0 {
		t.Fatalf("exit code = %d, stderr=%s", code, stderr.String())
	}

	payload := mustJSON(t, stdout.Bytes())
	assertOK(t, payload)
	if cmd, _ := payload["command"].(string); cmd != "openapi task create" {
		t.Fatalf("command = %q, want openapi task create", cmd)
	}
	data := payload["data"].(map[string]any)
	if dryRun, _ := data["dry_run"].(bool); !dryRun {
		t.Fatal("dry_run = false, want true")
	}
	createPayload := data["payload"].(map[string]any)
	if createPayload["title"] != "E2E smoke" {
		t.Fatalf("payload.title = %v, want E2E smoke", createPayload["title"])
	}
	if createPayload["projectId"] != "p1" {
		t.Fatalf("payload.projectId = %v, want p1", createPayload["projectId"])
	}
	if priority, _ := createPayload["priority"].(float64); priority != 5 {
		t.Fatalf("payload.priority = %v, want 5", priority)
	}
}
