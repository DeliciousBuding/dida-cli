package webapi

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestResourceMutationsUseExpectedEndpoints(t *testing.T) {
	var seen []string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		seen = append(seen, r.Method+" "+r.URL.RequestURI())
		if r.Body != nil && r.ContentLength != 0 {
			var payload any
			if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
				t.Fatalf("decode request for %s: %v", r.URL.Path, err)
			}
		}
		if r.Method == http.MethodGet && strings.HasPrefix(r.URL.Path, "/column/project/") {
			_ = json.NewEncoder(w).Encode([]map[string]any{{"id": "c1", "name": "Doing"}})
			return
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"ok": true})
	}))
	defer server.Close()

	client := NewClient("test-token")
	client.BaseURL = server.URL
	ctx := context.Background()

	calls := []func() error{
		func() error { _, err := client.CreateProject(ctx, ProjectMutation{ID: "p1", Name: "P"}); return err },
		func() error { _, err := client.UpdateProject(ctx, ProjectMutation{ID: "p1", Name: "P2"}); return err },
		func() error { _, err := client.DeleteProject(ctx, "p1"); return err },
		func() error {
			_, err := client.CreateProjectGroup(ctx, ProjectGroupMutation{ID: "g1", Name: "G"})
			return err
		},
		func() error {
			_, err := client.UpdateProjectGroup(ctx, ProjectGroupMutation{ID: "g1", Name: "G2"})
			return err
		},
		func() error { _, err := client.DeleteProjectGroup(ctx, "g1"); return err },
		func() error { _, err := client.CreateTag(ctx, TagMutation{Name: "tag-a"}); return err },
		func() error { _, err := client.UpdateTag(ctx, TagMutation{Name: "tag-a", Label: "Tag A"}); return err },
		func() error { _, err := client.RenameTag(ctx, "tag-a", "tag-b"); return err },
		func() error { _, err := client.MergeTags(ctx, "tag-a", "tag-b"); return err },
		func() error { _, err := client.DeleteTag(ctx, "tag-a"); return err },
		func() error { _, err := client.CreateColumn(ctx, "p1", "Doing"); return err },
		func() error { _, err := client.ProjectColumns(ctx, "p1"); return err },
		func() error { _, err := client.MoveTask(ctx, "t1", "p1", "p2"); return err },
		func() error { _, err := client.SetTaskParent(ctx, "child", "parent", "p1"); return err },
	}
	for _, call := range calls {
		if err := call(); err != nil {
			t.Fatalf("call failed: %v", err)
		}
	}

	want := []string{
		"POST /batch/project",
		"POST /batch/project",
		"POST /batch/project",
		"POST /batch/projectGroup",
		"POST /batch/projectGroup",
		"POST /batch/projectGroup",
		"POST /batch/tag",
		"POST /batch/tag",
		"PUT /tag/rename",
		"PUT /tag/merge",
		"DELETE /tag?name=tag-a",
		"POST /column",
		"GET /column/project/p1",
		"POST /batch/taskProject",
		"POST /batch/taskParent",
	}
	if strings.Join(seen, "\n") != strings.Join(want, "\n") {
		t.Fatalf("seen endpoints:\n%s\nwant:\n%s", strings.Join(seen, "\n"), strings.Join(want, "\n"))
	}
}

func TestResourceMutationBodies(t *testing.T) {
	var bodies []any
	var paths []string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		paths = append(paths, r.Method+" "+r.URL.RequestURI())
		var payload any
		if r.Body != nil && r.ContentLength != 0 {
			if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
				t.Fatalf("decode request for %s: %v", r.URL.Path, err)
			}
		}
		bodies = append(bodies, payload)
		_ = json.NewEncoder(w).Encode(map[string]any{"ok": true})
	}))
	defer server.Close()

	client := NewClient("test-token")
	client.BaseURL = server.URL
	ctx := context.Background()

	if _, err := client.CreateProject(ctx, ProjectMutation{ID: "p1", Name: "P", ViewMode: "list", Kind: "TASK"}); err != nil {
		t.Fatalf("CreateProject() error = %v", err)
	}
	if _, err := client.RenameTag(ctx, "old tag", "new tag"); err != nil {
		t.Fatalf("RenameTag() error = %v", err)
	}
	if _, err := client.MergeTags(ctx, "from", "to"); err != nil {
		t.Fatalf("MergeTags() error = %v", err)
	}
	if _, err := client.CreateColumn(ctx, "p1", "Doing"); err != nil {
		t.Fatalf("CreateColumn() error = %v", err)
	}
	if _, err := client.MoveTask(ctx, "t1", "p1", "p2"); err != nil {
		t.Fatalf("MoveTask() error = %v", err)
	}
	if _, err := client.SetTaskParent(ctx, "child", "parent", "p1"); err != nil {
		t.Fatalf("SetTaskParent() error = %v", err)
	}
	if _, err := client.DeleteTag(ctx, "old tag"); err != nil {
		t.Fatalf("DeleteTag() error = %v", err)
	}

	if paths[6] != "DELETE /tag?name=old+tag" {
		t.Fatalf("delete path = %s, want escaped tag query", paths[6])
	}
	projectBody := bodies[0].(map[string]any)
	projectAdd := projectBody["add"].([]any)[0].(map[string]any)
	if projectAdd["viewMode"] != "list" || projectAdd["kind"] != "TASK" {
		t.Fatalf("project add body = %#v", projectAdd)
	}
	renameBody := bodies[1].(map[string]any)
	if renameBody["name"] != "old tag" || renameBody["newName"] != "new tag" {
		t.Fatalf("rename body = %#v", bodies[1])
	}
	mergeBody := bodies[2].(map[string]any)
	if mergeBody["from"] != "from" || mergeBody["to"] != "to" {
		t.Fatalf("merge body = %#v", bodies[2])
	}
	columnBody := bodies[3].(map[string]any)
	if columnBody["projectId"] != "p1" || columnBody["name"] != "Doing" {
		t.Fatalf("column body = %#v", bodies[3])
	}
	moveBody := bodies[4].([]any)[0].(map[string]any)
	if moveBody["taskId"] != "t1" || moveBody["fromProjectId"] != "p1" || moveBody["toProjectId"] != "p2" {
		t.Fatalf("move body = %#v", bodies[4])
	}
	parentBody := bodies[5].([]any)[0].(map[string]any)
	if parentBody["taskId"] != "child" || parentBody["parentId"] != "parent" || parentBody["projectId"] != "p1" {
		t.Fatalf("parent body = %#v", bodies[5])
	}
}
