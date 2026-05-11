package webapi

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestClientRedactsSensitiveErrorContent(t *testing.T) {
	t.Setenv("DIDA_DEBUG_API_ERRORS", "1")
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, `{"token":"secret-token","cookie":"t=secret-token","message":"secret-token Authorization: Bearer abc.def"}`, http.StatusUnauthorized)
	}))
	defer server.Close()

	client := NewClient("secret-token")
	client.BaseURL = server.URL

	err := client.Do(context.Background(), http.MethodGet, "/fail", nil, nil)
	if err == nil {
		t.Fatalf("Do() error = nil, want error")
	}
	text := err.Error()
	for _, leaked := range []string{"secret-token", "abc.def"} {
		if strings.Contains(text, leaked) {
			t.Fatalf("error leaked %q: %s", leaked, text)
		}
	}
	if !strings.Contains(text, "[REDACTED]") {
		t.Fatalf("error missing redaction marker: %s", text)
	}
}

func TestClientHidesErrorBodyByDefault(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, `{"message":"private task title"}`, http.StatusBadRequest)
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL

	err := client.Do(context.Background(), http.MethodGet, "/fail", nil, nil)
	if err == nil {
		t.Fatalf("Do() error = nil, want error")
	}
	text := err.Error()
	if strings.Contains(text, "private task title") {
		t.Fatalf("error leaked response body: %s", text)
	}
	if !strings.Contains(text, "returned 400") {
		t.Fatalf("error missing status: %s", text)
	}
	var apiErr *APIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("error type = %T, want *APIError", err)
	}
	if apiErr.StatusCode != http.StatusBadRequest || apiErr.Method != http.MethodGet || apiErr.Path != "/fail" {
		t.Fatalf("api error = %#v", apiErr)
	}
	if !strings.Contains(apiErr.BodySnippet, "private task title") {
		t.Fatalf("api error body snippet = %q", apiErr.BodySnippet)
	}
}

func TestClientRejectsOversizedResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"data":"0123456789"}`))
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	client.MaxResponseBytes = 8

	err := client.Do(context.Background(), http.MethodGet, "/large", nil, nil)
	if err == nil {
		t.Fatalf("Do() error = nil, want response size error")
	}
	if !strings.Contains(err.Error(), "response exceeded 8 bytes") {
		t.Fatalf("unexpected error: %s", err)
	}
}

func TestClientRejectsMissingToken(t *testing.T) {
	client := NewClient("")
	client.BaseURL = "http://localhost"
	err := client.Do(context.Background(), "GET", "/test", nil, nil)
	if err == nil {
		t.Fatalf("Do() error = nil, want missing token error")
	}
	if !strings.Contains(err.Error(), "missing Dida web cookie token") {
		t.Fatalf("error = %v", err)
	}
}

func TestClientHandlesNoContentResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	err := client.Do(context.Background(), "POST", "/test", nil, nil)
	if err != nil {
		t.Fatalf("Do() 204: error = %v", err)
	}
}

func TestClientHandlesNilOutputWithBody(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"ok":true}`))
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURL = server.URL
	// nil output should succeed silently
	err := client.Do(context.Background(), "GET", "/test", nil, nil)
	if err != nil {
		t.Fatalf("Do() nil output: error = %v", err)
	}
}

func TestClientSendsCookieAndHeaders(t *testing.T) {
	var gotCookie, gotUserAgent, gotContentType string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotCookie = r.Header.Get("Cookie")
		gotUserAgent = r.Header.Get("User-Agent")
		gotContentType = r.Header.Get("Content-Type")
		_ = json.NewEncoder(w).Encode(map[string]any{"ok": true})
	}))
	defer server.Close()

	client := NewClient("mytoken123")
	client.BaseURL = server.URL
	client.UserAgent = "TestAgent/1.0"
	var out map[string]any
	_ = client.Do(context.Background(), "POST", "/test", map[string]any{"key": "val"}, &out)

	if gotCookie != "t=mytoken123" {
		t.Fatalf("cookie = %q, want t=mytoken123", gotCookie)
	}
	if gotUserAgent != "TestAgent/1.0" {
		t.Fatalf("user-agent = %q", gotUserAgent)
	}
	if gotContentType != "application/json" {
		t.Fatalf("content-type = %q", gotContentType)
	}
}

func TestDoV1UsesV1BaseURL(t *testing.T) {
	var gotPath string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		_ = json.NewEncoder(w).Encode(map[string]any{"ok": true})
	}))
	defer server.Close()

	client := NewClient("token")
	client.BaseURLV1 = server.URL
	var out map[string]any
	_ = client.DoV1(context.Background(), "GET", "/task/activity/t1", nil, &out)
	if gotPath != "/task/activity/t1" {
		t.Fatalf("path = %q, want /task/activity/t1", gotPath)
	}
}
