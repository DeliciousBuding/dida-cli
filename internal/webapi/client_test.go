package webapi

import (
	"context"
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
