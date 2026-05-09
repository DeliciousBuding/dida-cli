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
