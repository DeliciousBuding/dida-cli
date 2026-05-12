package officialmcp

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"
)

func TestToolsHandshakeFlow(t *testing.T) {
	var initSeen, notifySeen, toolsSeen bool
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var payload map[string]any
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			t.Fatalf("decode request: %v", err)
		}
		switch payload["method"] {
		case "initialize":
			initSeen = true
			w.Header().Set("Mcp-Session-Id", "sess-1")
			_ = json.NewEncoder(w).Encode(map[string]any{"jsonrpc": "2.0", "id": 1, "result": map[string]any{"protocolVersion": InitProtocolVersion}})
		case "notifications/initialized":
			notifySeen = true
			w.WriteHeader(http.StatusNoContent)
		case "tools/list":
			toolsSeen = true
			if got := r.Header.Get("Mcp-Session-Id"); got != "sess-1" {
				t.Fatalf("session header = %q, want sess-1", got)
			}
			if got := r.Header.Get("MCP-Protocol-Version"); got != ProtocolVersion {
				t.Fatalf("protocol header = %q, want %q", got, ProtocolVersion)
			}
			_ = json.NewEncoder(w).Encode(map[string]any{"jsonrpc": "2.0", "id": 2, "result": map[string]any{"tools": []map[string]any{{"name": "list_projects", "description": "List"}}}})
		default:
			t.Fatalf("unexpected method %#v", payload["method"])
		}
	}))
	defer server.Close()

	client := NewClient("token")
	client.URL = server.URL
	if err := client.Initialize(context.Background(), "test", "0.1.0"); err != nil {
		t.Fatalf("Initialize() error = %v", err)
	}
	tools, err := client.Tools(context.Background())
	if err != nil {
		t.Fatalf("Tools() error = %v", err)
	}
	if !initSeen || !notifySeen || !toolsSeen {
		t.Fatalf("handshake flags = init:%v notify:%v tools:%v", initSeen, notifySeen, toolsSeen)
	}
	if len(tools) != 1 || tools[0].Name != "list_projects" {
		t.Fatalf("tools = %#v", tools)
	}
}

func TestResolveTokenFromEnv(t *testing.T) {
	t.Setenv("DIDA365_TOKEN", "dp_test")
	token, err := ResolveToken("")
	if err != nil {
		t.Fatalf("ResolveToken() error = %v", err)
	}
	if token != "dp_test" {
		t.Fatalf("token = %q, want dp_test", token)
	}
}

func TestCallToolUnwrapsStructuredContent(t *testing.T) {
	step := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var payload map[string]any
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			t.Fatalf("decode request: %v", err)
		}
		method := payload["method"]
		switch method {
		case "initialize":
			w.Header().Set("Mcp-Session-Id", "sess-1")
			_ = json.NewEncoder(w).Encode(map[string]any{"jsonrpc": "2.0", "id": 1, "result": map[string]any{"protocolVersion": InitProtocolVersion}})
		case "notifications/initialized":
			w.WriteHeader(http.StatusNoContent)
		case "tools/call":
			step++
			params := payload["params"].(map[string]any)
			if params["name"] != "list_projects" {
				t.Fatalf("tool name = %#v", params["name"])
			}
			_ = json.NewEncoder(w).Encode(map[string]any{
				"jsonrpc": "2.0",
				"id":      3,
				"result": map[string]any{
					"structuredContent": map[string]any{"result": []map[string]any{{"id": "p1"}}},
				},
			})
		default:
			t.Fatalf("unexpected method %#v", method)
		}
	}))
	defer server.Close()

	client := NewClient("token")
	client.URL = server.URL
	if err := client.Initialize(context.Background(), "test", "0.1.0"); err != nil {
		t.Fatalf("Initialize() error = %v", err)
	}
	result, err := client.CallTool(context.Background(), "list_projects", map[string]any{})
	if err != nil {
		t.Fatalf("CallTool() error = %v", err)
	}
	if step != 1 {
		t.Fatalf("tools/call count = %d, want 1", step)
	}
	items := result.([]any)
	if len(items) != 1 {
		t.Fatalf("result = %#v", result)
	}
}

func TestTokenConfigSaveLoadClear(t *testing.T) {
	t.Setenv("DIDA_CONFIG_DIR", t.TempDir())
	t.Setenv("DIDA365_TOKEN", "")

	cfg, err := SaveTokenConfig("dp_test_token_12345")
	if err != nil {
		t.Fatalf("SaveTokenConfig() error = %v", err)
	}
	if cfg.Token != "dp_test_token_12345" {
		t.Fatalf("cfg.Token = %q", cfg.Token)
	}
	if cfg.SavedAt == 0 {
		t.Fatalf("cfg.SavedAt = 0, want non-zero")
	}

	loaded, err := LoadTokenConfig()
	if err != nil {
		t.Fatalf("LoadTokenConfig() error = %v", err)
	}
	if loaded.Token != "dp_test_token_12345" {
		t.Fatalf("loaded.Token = %q", loaded.Token)
	}

	if err := ClearTokenConfig(); err != nil {
		t.Fatalf("ClearTokenConfig() error = %v", err)
	}

	_, err = LoadTokenConfig()
	if err == nil {
		t.Fatalf("LoadTokenConfig() after clear: error = nil")
	}
}

func TestSaveTokenConfigRejectsEmpty(t *testing.T) {
	t.Setenv("DIDA_CONFIG_DIR", t.TempDir())
	_, err := SaveTokenConfig("")
	if err == nil {
		t.Fatalf("SaveTokenConfig(\"\") error = nil, want error")
	}
	_, err = SaveTokenConfig("   ")
	if err == nil {
		t.Fatalf("SaveTokenConfig(\"   \") error = nil, want error")
	}
}

func TestLoadTokenConfigRejectsEmptyToken(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("DIDA_CONFIG_DIR", dir)
	// Write a config with empty token
	data := []byte(`{"token":"","saved_at":12345}`)
	path := TokenConfigPath()
	if err := os.MkdirAll(dir, 0o700); err != nil {
		t.Fatalf("MkdirAll: %v", err)
	}
	if err := os.WriteFile(path, data, 0o600); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}
	_, err := LoadTokenConfig()
	if err == nil {
		t.Fatalf("LoadTokenConfig() error = nil, want empty token error")
	}
	if !strings.Contains(err.Error(), "no token") {
		t.Fatalf("error = %v, want 'no token'", err)
	}
}

func TestTokenConfigStatusFromEnv(t *testing.T) {
	t.Setenv("DIDA_CONFIG_DIR", t.TempDir())
	t.Setenv("DIDA365_TOKEN", "dp_env_token")
	status := TokenConfigStatus()
	if status["source"] != "env" {
		t.Fatalf("source = %v, want env", status["source"])
	}
	if status["available"] != true {
		t.Fatalf("available = %v, want true", status["available"])
	}
}

func TestTokenConfigStatusMissing(t *testing.T) {
	t.Setenv("DIDA_CONFIG_DIR", t.TempDir())
	t.Setenv("DIDA365_TOKEN", "")
	status := TokenConfigStatus()
	if status["source"] != "missing" {
		t.Fatalf("source = %v, want missing", status["source"])
	}
	if status["available"] != false {
		t.Fatalf("available = %v, want false", status["available"])
	}
}

func TestTokenConfigStatusFromFile(t *testing.T) {
	t.Setenv("DIDA_CONFIG_DIR", t.TempDir())
	t.Setenv("DIDA365_TOKEN", "")
	_, err := SaveTokenConfig("dp_file_token_abcdefgh")
	if err != nil {
		t.Fatalf("SaveTokenConfig() error = %v", err)
	}
	status := TokenConfigStatus()
	if status["source"] != "config" {
		t.Fatalf("source = %v, want config", status["source"])
	}
	if status["available"] != true {
		t.Fatalf("available = %v, want true", status["available"])
	}
	preview, _ := status["token_preview"].(string)
	if !strings.Contains(preview, "...") {
		t.Fatalf("token_preview = %q, want redacted", preview)
	}
}

func TestRedactForStatus(t *testing.T) {
	cases := []struct {
		input string
		want  string
	}{
		{"dp_abcdefgh", "dp_a...efgh"},
		{"short", "***"},
		{"12345678", "***"},
		{"", "***"},
		{"abc", "***"},
		{"abcdefghi", "abcd...fghi"},
	}
	for _, tc := range cases {
		got := RedactForStatus(tc.input)
		if got != tc.want {
			t.Errorf("RedactForStatus(%q) = %q, want %q", tc.input, got, tc.want)
		}
	}
}

func TestUnwrapEnvelope(t *testing.T) {
	// single result
	got := unwrapEnvelope(map[string]any{"result": "hello"})
	if got != "hello" {
		t.Fatalf("unwrapEnvelope result = %#v", got)
	}

	// nested result
	got = unwrapEnvelope(map[string]any{"result": map[string]any{"result": "deep"}})
	if got != "deep" {
		t.Fatalf("unwrapEnvelope nested = %#v", got)
	}

	// results key
	got = unwrapEnvelope(map[string]any{"results": []any{"a", "b"}})
	items, ok := got.([]any)
	if !ok || len(items) != 2 {
		t.Fatalf("unwrapEnvelope results = %#v", got)
	}

	// multiple keys - no unwrap
	got = unwrapEnvelope(map[string]any{"result": "a", "other": "b"})
	if _, ok := got.(map[string]any); !ok {
		t.Fatalf("unwrapEnvelope multi-key should return original")
	}

	// non-map - no unwrap
	got = unwrapEnvelope("plain string")
	if got != "plain string" {
		t.Fatalf("unwrapEnvelope string = %#v", got)
	}
}

func TestUnwrapToolResult(t *testing.T) {
	// structuredContent path
	got := unwrapToolResult(map[string]any{
		"structuredContent": map[string]any{"result": "data"},
	})
	if got != "data" {
		t.Fatalf("structuredContent: got %#v", got)
	}

	// content array with text JSON
	got = unwrapToolResult(map[string]any{
		"content": []any{
			map[string]any{"type": "text", "text": `{"result":{"id":"p1"}}`},
		},
	})
	item, ok := got.(map[string]any)
	if !ok || item["id"] != "p1" {
		t.Fatalf("content text JSON: got %#v", got)
	}

	// content array with plain text
	got = unwrapToolResult(map[string]any{
		"content": []any{
			map[string]any{"type": "text", "text": "hello world"},
		},
	})
	if got != "hello world" {
		t.Fatalf("content plain text: got %#v", got)
	}

	// content with non-text item - fallback
	got = unwrapToolResult(map[string]any{
		"content": []any{
			map[string]any{"type": "image", "data": "base64"},
		},
	})
	if _, ok := got.(map[string]any); !ok {
		t.Fatalf("non-text content should return original map")
	}

	// empty content - fallback
	got = unwrapToolResult(map[string]any{
		"content": []any{},
	})
	if _, ok := got.(map[string]any); !ok {
		t.Fatalf("empty content should return original map")
	}
}

func TestParseSSEResponse(t *testing.T) {
	// valid SSE with data lines
	body := bytes.NewBufferString("data: {\"jsonrpc\":\"2.0\",\"id\":1,\"result\":{\"ok\":true}}\n\n")
	resp, err := parseSSEResponse(body)
	if err != nil {
		t.Fatalf("parseSSEResponse() error = %v", err)
	}
	if resp.Result["ok"] != true {
		t.Fatalf("result = %#v", resp.Result)
	}

	// two separate SSE events, first is non-JSON, second is valid
	body = bytes.NewBufferString("data: not-json-at-all\n\ndata: {\"jsonrpc\":\"2.0\",\"id\":1,\"result\":{\"x\":1}}\n\n")
	resp, err = parseSSEResponse(body)
	if err != nil {
		t.Fatalf("parseSSEResponse() two-event error = %v", err)
	}
	if resp.Result == nil {
		t.Fatalf("result = nil")
	}

	// empty response
	body = bytes.NewBufferString("")
	_, err = parseSSEResponse(body)
	if err == nil {
		t.Fatalf("parseSSEResponse() empty: error = nil")
	}
	if !strings.Contains(err.Error(), "empty") {
		t.Fatalf("error = %v, want 'empty'", err)
	}

	// SSE with non-JSON data lines
	body = bytes.NewBufferString("data: not-json\n\n")
	_, err = parseSSEResponse(body)
	if err == nil {
		t.Fatalf("parseSSEResponse() non-json: error = nil")
	}
}

func TestToolSchemaFindsTool(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var payload map[string]any
		_ = json.NewDecoder(r.Body).Decode(&payload)
		switch payload["method"] {
		case "initialize":
			w.Header().Set("Mcp-Session-Id", "s1")
			_ = json.NewEncoder(w).Encode(map[string]any{"jsonrpc": "2.0", "id": 1, "result": map[string]any{"protocolVersion": InitProtocolVersion}})
		case "notifications/initialized":
			w.WriteHeader(http.StatusNoContent)
		case "tools/list":
			_ = json.NewEncoder(w).Encode(map[string]any{"jsonrpc": "2.0", "id": 2, "result": map[string]any{
				"tools": []map[string]any{
					{"name": "list_projects", "description": "List all"},
					{"name": "get_task", "description": "Get task"},
				},
			}})
		}
	}))
	defer server.Close()

	client := NewClient("token")
	client.URL = server.URL
	_ = client.Initialize(context.Background(), "test", "0.1.0")

	tool, err := client.ToolSchema(context.Background(), "get_task")
	if err != nil {
		t.Fatalf("ToolSchema() error = %v", err)
	}
	if tool.Name != "get_task" || tool.Description != "Get task" {
		t.Fatalf("tool = %#v", tool)
	}

	_, err = client.ToolSchema(context.Background(), "missing")
	if err == nil {
		t.Fatalf("ToolSchema(missing) error = nil")
	}
}

func TestResolveTokenExplicit(t *testing.T) {
	t.Setenv("DIDA365_TOKEN", "")
	token, err := ResolveToken("dp_explicit")
	if err != nil {
		t.Fatalf("error = %v", err)
	}
	if token != "dp_explicit" {
		t.Fatalf("token = %q", token)
	}
}

func TestResolveTokenMissing(t *testing.T) {
	t.Setenv("DIDA365_TOKEN", "")
	t.Setenv("DIDA_CONFIG_DIR", t.TempDir())
	_, err := ResolveToken("")
	if err == nil {
		t.Fatalf("ResolveToken(\"\") error = nil, want missing token error")
	}
	if !strings.Contains(err.Error(), "missing official mcp token") {
		t.Fatalf("error = %v", err)
	}
}

func TestClearTokenConfigNoFile(t *testing.T) {
	t.Setenv("DIDA_CONFIG_DIR", t.TempDir())
	// Clear when no file exists should succeed
	if err := ClearTokenConfig(); err != nil {
		t.Fatalf("ClearTokenConfig() on missing file: error = %v", err)
	}
}

func TestNewClient(t *testing.T) {
	c := NewClient("mytoken")
	if c.Token != "mytoken" || c.URL != DefaultURL {
		t.Fatalf("client = %+v", c)
	}
	if c.HTTPClient == nil {
		t.Fatal("HTTPClient is nil")
	}
	if c.HTTPClient.Timeout != 60*time.Second {
		t.Fatalf("HTTPClient.Timeout = %v, want 60s", c.HTTPClient.Timeout)
	}
}
