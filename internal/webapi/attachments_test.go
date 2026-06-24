package webapi

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestDownloadTaskAttachmentUsesV1DownloadEndpoint(t *testing.T) {
	var gotMethod, gotPath, gotQuery, gotCookie string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		gotPath = r.URL.EscapedPath()
		gotQuery = r.URL.RawQuery
		gotCookie = r.Header.Get("Cookie")
		w.Header().Set("Content-Type", "application/msword")
		_, _ = w.Write([]byte("doc-bytes"))
	}))
	defer server.Close()

	client := NewClient("test-token")
	client.BaseURLV1 = server.URL
	var out bytes.Buffer
	n, contentType, err := client.DownloadTaskAttachment(context.Background(), "p 1", "t/1", "a?1", &out)
	if err != nil {
		t.Fatalf("DownloadTaskAttachment() error = %v", err)
	}
	if gotMethod != http.MethodGet {
		t.Fatalf("method = %q, want GET", gotMethod)
	}
	if gotPath != "/attachment/p%201/t%2F1/a%3F1" {
		t.Fatalf("path = %q", gotPath)
	}
	if gotQuery != "action=download" {
		t.Fatalf("query = %q, want action=download", gotQuery)
	}
	if gotCookie != "t=test-token" {
		t.Fatalf("cookie = %q, want t=test-token", gotCookie)
	}
	if n != int64(len("doc-bytes")) || out.String() != "doc-bytes" {
		t.Fatalf("downloaded n=%d body=%q", n, out.String())
	}
	if contentType != "application/msword" {
		t.Fatalf("contentType = %q", contentType)
	}
}

func TestDownloadTaskAttachmentHandlesHTTPErrors(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"errorCode":"attachment_not_exist"}`))
	}))
	defer server.Close()

	client := NewClient("test-token")
	client.BaseURLV1 = server.URL
	var out bytes.Buffer
	_, _, err := client.DownloadTaskAttachment(context.Background(), "p1", "t1", "a1", &out)
	if err == nil {
		t.Fatalf("DownloadTaskAttachment() error = nil, want error")
	}
	apiErr, ok := err.(*APIError)
	if !ok {
		t.Fatalf("error type = %T, want *APIError", err)
	}
	if apiErr.StatusCode != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", apiErr.StatusCode)
	}
}
