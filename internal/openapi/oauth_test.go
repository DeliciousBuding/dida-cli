package openapi

import (
	"strings"
	"testing"
)

func TestAuthorizationURL(t *testing.T) {
	url := AuthorizationURL("cid", "http://127.0.0.1:17890/callback", "tasks:read tasks:write", "abc123")
	for _, want := range []string{"client_id=cid", "response_type=code", "state=abc123", "tasks%3Aread+tasks%3Awrite"} {
		if !strings.Contains(url, want) {
			t.Fatalf("authorization url missing %q: %s", want, url)
		}
	}
}

func TestRedactToken(t *testing.T) {
	if got := redactToken("abcd1234wxyz"); got != "abcd...wxyz" {
		t.Fatalf("redactToken() = %q", got)
	}
}

func TestTokenStatusDoesNotExposeLocalPath(t *testing.T) {
	t.Setenv("DIDA_CONFIG_DIR", t.TempDir())
	status := TokenStatus()
	if _, ok := status["path"]; ok {
		t.Fatalf("TokenStatus exposed local path: %#v", status)
	}
	if status["available"] != false {
		t.Fatalf("available = %#v, want false", status["available"])
	}
}
