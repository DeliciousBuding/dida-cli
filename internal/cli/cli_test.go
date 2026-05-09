package cli

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
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
	code := Run([]string{"task", "create", "--project", "p1", "--title", "Smoke", "--dry-run", "--json"}, "test-version", &stdout, &stderr)
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
