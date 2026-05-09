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

func TestClientConfigResolveAndStatus(t *testing.T) {
	t.Setenv("DIDA_CONFIG_DIR", t.TempDir())
	t.Setenv("DIDA365_OPENAPI_CLIENT_ID", "")
	t.Setenv("DIDA365_OPENAPI_CLIENT_SECRET", "")
	if _, err := SaveClientConfig("client-id", "client-secret"); err != nil {
		t.Fatalf("SaveClientConfig() error = %v", err)
	}
	clientID, err := ResolveClientID("")
	if err != nil {
		t.Fatalf("ResolveClientID() error = %v", err)
	}
	if clientID != "client-id" {
		t.Fatalf("clientID = %q", clientID)
	}
	clientSecret, err := ResolveClientSecret("")
	if err != nil {
		t.Fatalf("ResolveClientSecret() error = %v", err)
	}
	if clientSecret != "client-secret" {
		t.Fatalf("clientSecret = %q", clientSecret)
	}
	status := ClientConfigStatus()
	if status["available"] != true {
		t.Fatalf("status = %#v", status)
	}
	if _, ok := status["client_secret"]; ok {
		t.Fatalf("status exposed client secret: %#v", status)
	}
}
