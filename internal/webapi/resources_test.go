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
		"POST /batch/taskProject",
		"POST /batch/taskParent",
	}
	if strings.Join(seen, "\n") != strings.Join(want, "\n") {
		t.Fatalf("seen endpoints:\n%s\nwant:\n%s", strings.Join(seen, "\n"), strings.Join(want, "\n"))
	}
}
