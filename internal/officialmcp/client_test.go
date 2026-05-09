package officialmcp

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
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
