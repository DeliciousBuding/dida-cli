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
