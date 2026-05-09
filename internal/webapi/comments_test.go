package webapi

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestCommentEndpointsAndBodies(t *testing.T) {
	var seen []string
	var bodies []any
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		seen = append(seen, r.Method+" "+r.URL.RequestURI())
		var payload any
		if r.Body != nil && r.ContentLength != 0 {
			if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
				t.Fatalf("decode request for %s: %v", r.URL.Path, err)
			}
		}
		bodies = append(bodies, payload)
		if r.Method == http.MethodGet {
			_ = json.NewEncoder(w).Encode([]map[string]any{{"id": "c1", "title": "hello"}})
			return
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"id": "c1", "ok": true})
	}))
	defer server.Close()

	client := NewClient("test-token")
	client.BaseURL = server.URL
	ctx := context.Background()

	if _, err := client.TaskComments(ctx, "p 1", "t/1"); err != nil {
		t.Fatalf("TaskComments() error = %v", err)
	}
	if _, err := client.CreateComment(ctx, "p 1", "t/1", CommentMutation{Title: "hello"}); err != nil {
		t.Fatalf("CreateComment() error = %v", err)
	}
	if _, err := client.UpdateComment(ctx, "p 1", "t/1", "c 1", CommentMutation{Title: "updated"}); err != nil {
		t.Fatalf("UpdateComment() error = %v", err)
	}
	if _, err := client.DeleteComment(ctx, "p 1", "t/1", "c 1"); err != nil {
		t.Fatalf("DeleteComment() error = %v", err)
	}

	want := []string{
		"GET /project/p%201/task/t%2F1/comments",
		"POST /project/p%201/task/t%2F1/comment",
		"PUT /project/p%201/task/t%2F1/comment/c%201",
		"DELETE /project/p%201/task/t%2F1/comment/c%201",
	}
	if strings.Join(seen, "\n") != strings.Join(want, "\n") {
		t.Fatalf("seen endpoints:\n%s\nwant:\n%s", strings.Join(seen, "\n"), strings.Join(want, "\n"))
	}
	createBody := bodies[1].(map[string]any)
	if createBody["title"] != "hello" {
		t.Fatalf("create body = %#v", createBody)
	}
	if createBody["id"] == "" || createBody["createdTime"] == "" || createBody["taskId"] != "t/1" || createBody["projectId"] != "p 1" {
		t.Fatalf("create body missing webapp-generated fields: %#v", createBody)
	}
	userProfile := createBody["userProfile"].(map[string]any)
	if userProfile["isMyself"] != true {
		t.Fatalf("userProfile = %#v", userProfile)
	}
	updateBody := bodies[2].(map[string]any)
	if updateBody["title"] != "updated" {
		t.Fatalf("update body = %#v", updateBody)
	}
}

func TestUploadCommentAttachmentMultipart(t *testing.T) {
	var seenMethod string
	var seenPath string
	var seenCookie string
	var seenContentType string
	var seenFileName string
	var seenFileBody string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		seenMethod = r.Method
		seenPath = r.URL.RequestURI()
		seenCookie = r.Header.Get("Cookie")
		seenContentType = r.Header.Get("Content-Type")
		file, header, err := r.FormFile("file")
		if err != nil {
			t.Fatalf("FormFile(file) error = %v", err)
		}
		defer file.Close()
		body, err := io.ReadAll(file)
		if err != nil {
			t.Fatalf("ReadAll(file) error = %v", err)
		}
		seenFileName = header.Filename
		seenFileBody = string(body)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"id":          "att1",
			"refId":       "ref1",
			"path":        "comment/path",
			"size":        7,
			"fileName":    header.Filename,
			"fileType":    "image/png",
			"createdTime": "2026-05-10T00:00:00.000+0000",
		})
	}))
	defer server.Close()

	client := NewClient("test-token")
	client.BaseURLV1 = server.URL
	out, err := client.UploadCommentAttachment(context.Background(), "p 1", "t/1", "probe.png", "image/png", strings.NewReader("pngdata"))
	if err != nil {
		t.Fatalf("UploadCommentAttachment() error = %v", err)
	}
	if out["id"] != "att1" {
		t.Fatalf("upload response = %#v", out)
	}
	if seenMethod != http.MethodPost || seenPath != "/attachment/upload/comment/p%201/t%2F1" {
		t.Fatalf("request = %s %s", seenMethod, seenPath)
	}
	if seenCookie != "t=test-token" {
		t.Fatalf("Cookie = %q", seenCookie)
	}
	if !strings.HasPrefix(seenContentType, "multipart/form-data; boundary=") {
		t.Fatalf("Content-Type = %q", seenContentType)
	}
	if seenFileName != "probe.png" || seenFileBody != "pngdata" {
		t.Fatalf("file = %q %q", seenFileName, seenFileBody)
	}
}
