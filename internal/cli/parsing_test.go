package cli

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/DeliciousBuding/dida-cli/internal/model"
)

func TestParseRawGetArgs(t *testing.T) {
	cases := []struct {
		name    string
		args    []string
		path    string
		version string
		err     string
	}{
		{name: "simple v2", args: []string{"/task/123"}, path: "/task/123", version: "v2"},
		{name: "explicit v1", args: []string{"/task/123", "--api-version", "v1"}, path: "/task/123", version: "v1"},
		{name: "shorthand v1", args: []string{"/task/123", "--v1"}, path: "/task/123", version: "v1"},
		{name: "shorthand v2", args: []string{"/task/123", "--v2"}, path: "/task/123", version: "v2"},
		{name: "path first then flags", args: []string{"/project/list", "--v1"}, path: "/project/list", version: "v1"},
		{name: "missing path", args: []string{}, err: "missing path"},
		{name: "missing path flags only", args: []string{"--v1"}, err: "missing path"},
		{name: "missing version value", args: []string{"/task", "--api-version"}, err: "--api-version requires v1 or v2"},
		{name: "invalid version", args: []string{"/task", "--api-version", "v3"}, err: "--api-version must be v1 or v2"},
		{name: "unknown flag", args: []string{"/task", "--surprise"}, err: `unknown flag "--surprise"`},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			path, version, err := parseRawGetArgs(tc.args)
			if tc.err != "" {
				if err == nil || !strings.Contains(err.Error(), tc.err) {
					t.Fatalf("error = %v, want %q", err, tc.err)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if path != tc.path {
				t.Fatalf("path = %q, want %q", path, tc.path)
			}
			if version != tc.version {
				t.Fatalf("version = %q, want %q", version, tc.version)
			}
		})
	}
}

func TestParseSearchAllFlags(t *testing.T) {
	cases := []struct {
		name  string
		args  []string
		query string
		limit int
		full  bool
		err   string
	}{
		{name: "positional query", args: []string{"test"}, query: "test", limit: 20},
		{name: "flagged query", args: []string{"--query", "hello"}, query: "hello", limit: 20},
		{name: "short flag", args: []string{"-q", "hello"}, query: "hello", limit: 20},
		{name: "full flag", args: []string{"test", "--full"}, query: "test", limit: 20, full: true},
		{name: "custom limit", args: []string{"test", "--limit", "5"}, query: "test", limit: 5},
		{name: "all flags", args: []string{"--query", "work", "--limit", "10", "--full"}, query: "work", limit: 10, full: true},
		{name: "missing query", args: []string{}, err: "missing query"},
		{name: "missing query value", args: []string{"--query"}, err: "--query requires text"},
		{name: "negative limit", args: []string{"test", "--limit", "-1"}, err: "--limit must be a non-negative integer"},
		{name: "invalid limit", args: []string{"test", "--limit", "abc"}, err: "--limit must be a non-negative integer"},
		{name: "unknown flag", args: []string{"test", "--surprise"}, err: `unknown flag "--surprise"`},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			query, limit, full, err := parseSearchAllFlags(tc.args)
			if tc.err != "" {
				if err == nil || !strings.Contains(err.Error(), tc.err) {
					t.Fatalf("error = %v, want %q", err, tc.err)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if query != tc.query {
				t.Fatalf("query = %q, want %q", query, tc.query)
			}
			if limit != tc.limit {
				t.Fatalf("limit = %d, want %d", limit, tc.limit)
			}
			if full != tc.full {
				t.Fatalf("full = %v, want %v", full, tc.full)
			}
		})
	}
}

func TestParseFullFlag(t *testing.T) {
	full, err := parseFullFlag([]string{"--full"})
	if err != nil {
		t.Fatalf("error = %v", err)
	}
	if !full {
		t.Fatalf("full = false, want true")
	}

	full, err = parseFullFlag(nil)
	if err != nil {
		t.Fatalf("error = %v", err)
	}
	if full {
		t.Fatalf("full = true, want false")
	}

	_, err = parseFullFlag([]string{"--surprise"})
	if err == nil {
		t.Fatalf("error = nil, want unknown flag error")
	}
}

func TestParseUserSessionsFlags(t *testing.T) {
	cases := []struct {
		name string
		args []string
		lang string
		lim  int
		full bool
		err  string
	}{
		{name: "defaults", args: nil, lang: "zh_CN", lim: 20},
		{name: "custom lang", args: []string{"--lang", "en"}, lang: "en", lim: 20},
		{name: "custom limit", args: []string{"--limit", "5"}, lang: "zh_CN", lim: 5},
		{name: "full", args: []string{"--full"}, lang: "zh_CN", lim: 20, full: true},
		{name: "all flags", args: []string{"--lang", "en", "--limit", "3", "--full"}, lang: "en", lim: 3, full: true},
		{name: "missing lang value", args: []string{"--lang"}, err: "--lang requires a locale"},
		{name: "negative limit", args: []string{"--limit", "-1"}, err: "--limit must be a non-negative integer"},
		{name: "unknown flag", args: []string{"--surprise"}, err: `unknown flag "--surprise"`},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			lang, lim, full, err := parseUserSessionsFlags(tc.args)
			if tc.err != "" {
				if err == nil || !strings.Contains(err.Error(), tc.err) {
					t.Fatalf("error = %v, want %q", err, tc.err)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if lang != tc.lang {
				t.Fatalf("lang = %q, want %q", lang, tc.lang)
			}
			if lim != tc.lim {
				t.Fatalf("limit = %d, want %d", lim, tc.lim)
			}
			if full != tc.full {
				t.Fatalf("full = %v, want %v", full, tc.full)
			}
		})
	}
}

func TestParseTemplateListFlags(t *testing.T) {
	ts, limit, err := parseTemplateListFlags([]string{"--timestamp", "1700000000", "--limit", "10"})
	if err != nil {
		t.Fatalf("error = %v", err)
	}
	if ts != 1700000000 {
		t.Fatalf("timestamp = %d, want 1700000000", ts)
	}
	if limit != 10 {
		t.Fatalf("limit = %d, want 10", limit)
	}

	// defaults
	ts, limit, err = parseTemplateListFlags(nil)
	if err != nil {
		t.Fatalf("error = %v", err)
	}
	if ts != 0 || limit != 50 {
		t.Fatalf("defaults: ts=%d limit=%d", ts, limit)
	}

	_, _, err = parseTemplateListFlags([]string{"--timestamp", "-1"})
	if err == nil {
		t.Fatalf("error = nil, want negative timestamp error")
	}

	_, _, err = parseTemplateListFlags([]string{"--surprise"})
	if err == nil {
		t.Fatalf("error = nil, want unknown flag error")
	}
}

func TestLimitSearchPayload(t *testing.T) {
	payload := map[string]any{
		"hits": []any{
			map[string]any{"id": "h1"},
			map[string]any{"id": "h2"},
			map[string]any{"id": "h3"},
		},
		"tasks": []any{
			map[string]any{"id": "t1"},
			map[string]any{"id": "t2"},
		},
	}
	counts := limitSearchPayload(payload, 1)
	if counts["hitsTotal"] != 3 || counts["hitsCount"] != 1 {
		t.Fatalf("hits counts = %#v", counts)
	}
	if counts["tasksTotal"] != 2 || counts["tasksCount"] != 1 {
		t.Fatalf("tasks counts = %#v", counts)
	}

	// limit=0 means no limit
	payload2 := map[string]any{
		"hits": []any{map[string]any{"id": "h1"}, map[string]any{"id": "h2"}},
	}
	counts2 := limitSearchPayload(payload2, 0)
	if counts2["hitsCount"] != 2 {
		t.Fatalf("limit=0: hitsCount = %d, want 2", counts2["hitsCount"])
	}

	// missing key
	counts3 := limitSearchPayload(map[string]any{}, 5)
	if counts3["hitsTotal"] != 0 {
		t.Fatalf("missing key: hitsTotal = %d, want 0", counts3["hitsTotal"])
	}
}

func TestParseTaskMoveFlags(t *testing.T) {
	opts, err := parseTaskMoveFlags([]string{"t1", "--from", "p1", "--to", "p2"})
	if err != nil {
		t.Fatalf("error = %v", err)
	}
	if opts.TaskID != "t1" || opts.FromProjectID != "p1" || opts.ToProjectID != "p2" {
		t.Fatalf("opts = %+v", opts)
	}

	// dry-run
	opts, err = parseTaskMoveFlags([]string{"t1", "--from", "p1", "--to", "p2", "--dry-run"})
	if err != nil {
		t.Fatalf("error = %v", err)
	}
	if !opts.DryRun {
		t.Fatalf("DryRun = false, want true")
	}

	// missing args
	_, err = parseTaskMoveFlags(nil)
	if err == nil {
		t.Fatalf("error = nil, want usage error")
	}

	// missing --from
	_, err = parseTaskMoveFlags([]string{"t1", "--to", "p2"})
	if err == nil {
		t.Fatalf("error = nil, want missing --from error")
	}

	// missing --to
	_, err = parseTaskMoveFlags([]string{"t1", "--from", "p1"})
	if err == nil {
		t.Fatalf("error = nil, want missing --to error")
	}

	// missing --from value
	_, err = parseTaskMoveFlags([]string{"t1", "--from"})
	if err == nil {
		t.Fatalf("error = nil, want --from requires value")
	}

	// missing --to value
	_, err = parseTaskMoveFlags([]string{"t1", "--from", "p1", "--to"})
	if err == nil {
		t.Fatalf("error = nil, want --to requires value")
	}

	// unknown flag
	_, err = parseTaskMoveFlags([]string{"t1", "--from", "p1", "--to", "p2", "--surprise"})
	if err == nil {
		t.Fatalf("error = nil, want unknown flag")
	}
}

func TestParseTaskParentFlags(t *testing.T) {
	opts, err := parseTaskParentFlags([]string{"t1", "--parent", "p1", "--project", "proj1"})
	if err != nil {
		t.Fatalf("error = %v", err)
	}
	if opts.TaskID != "t1" || opts.ParentID != "p1" || opts.ProjectID != "proj1" {
		t.Fatalf("opts = %+v", opts)
	}

	// short form -p
	opts, err = parseTaskParentFlags([]string{"t1", "--parent", "p1", "-p", "proj1"})
	if err != nil {
		t.Fatalf("error = %v", err)
	}
	if opts.ProjectID != "proj1" {
		t.Fatalf("ProjectID = %q", opts.ProjectID)
	}

	// dry-run
	opts, err = parseTaskParentFlags([]string{"t1", "--parent", "p1", "--project", "proj1", "--dry-run"})
	if err != nil {
		t.Fatalf("error = %v", err)
	}
	if !opts.DryRun {
		t.Fatalf("DryRun = false, want true")
	}

	// missing args
	_, err = parseTaskParentFlags(nil)
	if err == nil {
		t.Fatalf("error = nil, want usage error")
	}

	// missing --parent
	_, err = parseTaskParentFlags([]string{"t1", "--project", "proj1"})
	if err == nil {
		t.Fatalf("error = nil, want missing --parent error")
	}

	// missing --project
	_, err = parseTaskParentFlags([]string{"t1", "--parent", "p1"})
	if err == nil {
		t.Fatalf("error = nil, want missing --project error")
	}
}

func TestTaskOutputNonCompactReturnsRawStripped(t *testing.T) {
	tasks := []model.Task{
		{ID: "t1", ProjectID: "p1", Title: "Task 1", Raw: map[string]any{"secret": true}},
	}
	out := taskOutput(tasks, false)
	encoded, _ := json.Marshal(out)
	if strings.Contains(string(encoded), "secret") {
		t.Fatalf("non-compact output should strip raw: %s", string(encoded))
	}
}

func TestPickKeys(t *testing.T) {
	item := map[string]any{
		"id":    "t1",
		"title": "Test",
		"extra": "data",
	}
	result := pickKeys(item, "id", "title")
	if len(result) != 2 {
		t.Fatalf("result len = %d, want 2", len(result))
	}
	if result["id"] != "t1" || result["title"] != "Test" {
		t.Fatalf("result = %#v", result)
	}
	if _, ok := result["extra"]; ok {
		t.Fatalf("result should not contain 'extra'")
	}
}

func TestCompactListSkipsNonObjects(t *testing.T) {
	payload := map[string]any{
		"items": []any{"string", 42, map[string]any{"id": "obj1", "title": "Test"}},
	}
	compactList(payload, "items", func(item map[string]any) map[string]any {
		return pickKeys(item, "id")
	})
	items := payload["items"].([]any)
	if len(items) != 3 {
		t.Fatalf("items len = %d, want 3", len(items))
	}
	if items[0] != "string" {
		t.Fatalf("non-object should pass through: %v", items[0])
	}
	if items[2].(map[string]any)["id"] != "obj1" {
		t.Fatalf("object should be compacted: %v", items[2])
	}
}

func TestCompactListMissingKeyNoOp(t *testing.T) {
	payload := map[string]any{}
	compactList(payload, "missing", func(item map[string]any) map[string]any {
		return item
	})
	// should not panic or modify payload
	if _, ok := payload["missing"]; ok {
		t.Fatalf("missing key should not be added")
	}
}

func TestParseTaskListFlags(t *testing.T) {
	cases := []struct {
		name   string
		args   []string
		filter string
		limit  int
		compact bool
		err    string
	}{
		{name: "defaults", args: nil, filter: "all", limit: 50},
		{name: "filter today", args: []string{"--filter", "today"}, filter: "today", limit: 50},
		{name: "custom limit", args: []string{"--limit", "10"}, filter: "all", limit: 10},
		{name: "compact", args: []string{"--compact"}, filter: "all", limit: 50, compact: true},
		{name: "brief alias", args: []string{"--brief"}, filter: "all", limit: 50, compact: true},
		{name: "all flags", args: []string{"--filter", "week", "--limit", "5", "--compact"}, filter: "week", limit: 5, compact: true},
		{name: "missing filter value", args: []string{"--filter"}, err: "--filter requires a value"},
		{name: "negative limit", args: []string{"--limit", "-1"}, err: "--limit must be a non-negative integer"},
		{name: "unknown flag", args: []string{"--surprise"}, err: `unknown flag "--surprise"`},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			filter, limit, compact, err := parseTaskListFlags(tc.args)
			if tc.err != "" {
				if err == nil || !strings.Contains(err.Error(), tc.err) {
					t.Fatalf("error = %v, want %q", err, tc.err)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if filter != tc.filter {
				t.Fatalf("filter = %q, want %q", filter, tc.filter)
			}
			if limit != tc.limit {
				t.Fatalf("limit = %d, want %d", limit, tc.limit)
			}
			if compact != tc.compact {
				t.Fatalf("compact = %v, want %v", compact, tc.compact)
			}
		})
	}
}

func TestParseSearchFlags(t *testing.T) {
	cases := []struct {
		name    string
		args    []string
		query   string
		limit   int
		compact bool
		err     string
	}{
		{name: "positional", args: []string{"test"}, query: "test", limit: 50},
		{name: "flagged", args: []string{"--query", "hello"}, query: "hello", limit: 50},
		{name: "short flag", args: []string{"-q", "hi"}, query: "hi", limit: 50},
		{name: "compact", args: []string{"test", "--compact"}, query: "test", limit: 50, compact: true},
		{name: "custom limit", args: []string{"test", "--limit", "3"}, query: "test", limit: 3},
		{name: "missing query", args: []string{}, err: "missing query"},
		{name: "empty query", args: []string{"--query", "  "}, err: "missing query"},
		{name: "missing query value", args: []string{"--query"}, err: "--query requires a value"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			query, limit, compact, err := parseSearchFlags(tc.args)
			if tc.err != "" {
				if err == nil || !strings.Contains(err.Error(), tc.err) {
					t.Fatalf("error = %v, want %q", err, tc.err)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if query != tc.query {
				t.Fatalf("query = %q, want %q", query, tc.query)
			}
			if limit != tc.limit {
				t.Fatalf("limit = %d, want %d", limit, tc.limit)
			}
			if compact != tc.compact {
				t.Fatalf("compact = %v", compact)
			}
		})
	}
}

func TestParseUpcomingFlags(t *testing.T) {
	days, limit, compact, err := parseUpcomingFlags([]string{"--days", "14", "--limit", "20"})
	if err != nil {
		t.Fatalf("error = %v", err)
	}
	if days != 14 || limit != 20 || compact {
		t.Fatalf("days=%d limit=%d compact=%v", days, limit, compact)
	}

	days, limit, compact, err = parseUpcomingFlags([]string{"--compact"})
	if err != nil {
		t.Fatalf("error = %v", err)
	}
	if days != 7 || limit != 50 || !compact {
		t.Fatalf("defaults: days=%d limit=%d compact=%v", days, limit, compact)
	}

	_, _, _, err = parseUpcomingFlags([]string{"--days", "0"})
	if err == nil {
		t.Fatalf("error = nil, want positive days error")
	}

	_, _, _, err = parseUpcomingFlags([]string{"--days", "-1"})
	if err == nil {
		t.Fatalf("error = nil, want positive days error")
	}

	_, _, _, err = parseUpcomingFlags([]string{"--surprise"})
	if err == nil {
		t.Fatalf("error = nil, want unknown flag")
	}
}

func TestParseTaskIDProjectFlags(t *testing.T) {
	opts, err := parseTaskIDProjectFlags([]string{"t1", "--project", "p1"}, "complete")
	if err != nil {
		t.Fatalf("error = %v", err)
	}
	if opts.TaskID != "t1" || opts.ProjectID != "p1" {
		t.Fatalf("opts = %+v", opts)
	}

	// positional project
	opts, err = parseTaskIDProjectFlags([]string{"t1", "p1"}, "complete")
	if err != nil {
		t.Fatalf("error = %v", err)
	}
	if opts.ProjectID != "p1" {
		t.Fatalf("ProjectID = %q", opts.ProjectID)
	}

	// dry-run + yes
	opts, err = parseTaskIDProjectFlags([]string{"t1", "--project", "p1", "--dry-run", "--yes"}, "delete")
	if err != nil {
		t.Fatalf("error = %v", err)
	}
	if !opts.DryRun || !opts.Yes {
		t.Fatalf("DryRun=%v Yes=%v", opts.DryRun, opts.Yes)
	}

	// short form -p
	opts, err = parseTaskIDProjectFlags([]string{"t1", "-p", "p1"}, "complete")
	if err != nil {
		t.Fatalf("error = %v", err)
	}
	if opts.ProjectID != "p1" {
		t.Fatalf("ProjectID = %q", opts.ProjectID)
	}

	// missing args
	_, err = parseTaskIDProjectFlags(nil, "complete")
	if err == nil {
		t.Fatalf("error = nil, want usage error")
	}

	// missing project
	_, err = parseTaskIDProjectFlags([]string{"t1"}, "complete")
	if err == nil {
		t.Fatalf("error = nil, want missing project")
	}

	// missing project value
	_, err = parseTaskIDProjectFlags([]string{"t1", "--project"}, "complete")
	if err == nil {
		t.Fatalf("error = nil, want project value error")
	}
}

func TestParsePriority(t *testing.T) {
	cases := []struct {
		input string
		want  int
		err   string
	}{
		{"0", 0, ""},
		{"1", 1, ""},
		{"3", 3, ""},
		{"5", 5, ""},
		{"2", 0, "must be one of"},
		{"abc", 0, "must be one of"},
	}
	for _, tc := range cases {
		got, err := parsePriority(tc.input)
		if tc.err != "" {
			if err == nil || !strings.Contains(err.Error(), tc.err) {
				t.Errorf("parsePriority(%q) error = %v, want %q", tc.input, err, tc.err)
			}
			continue
		}
		if err != nil {
			t.Errorf("parsePriority(%q) error = %v", tc.input, err)
			continue
		}
		if got != tc.want {
			t.Errorf("parsePriority(%q) = %d, want %d", tc.input, got, tc.want)
		}
	}
}

func TestParseCommentWriteFlags(t *testing.T) {
	opts, err := parseCommentWriteFlags([]string{"--project", "p1", "--task", "t1", "--text", "Hello"}, "create", false)
	if err != nil {
		t.Fatalf("error = %v", err)
	}
	if opts.ProjectID != "p1" || opts.TaskID != "t1" || opts.Title != "Hello" {
		t.Fatalf("opts = %+v", opts)
	}

	// missing project
	_, err = parseCommentWriteFlags([]string{"--task", "t1", "--text", "Hi"}, "create", false)
	if err == nil {
		t.Fatalf("error = nil, want missing project")
	}

	// missing text for create
	_, err = parseCommentWriteFlags([]string{"--project", "p1", "--task", "t1"}, "create", true)
	if err == nil {
		t.Fatalf("error = nil, want missing text")
	}
}

func TestParseCompactOnlyFlags(t *testing.T) {
	compact, err := parseCompactOnlyFlags([]string{"--compact"})
	if err != nil {
		t.Fatalf("error = %v", err)
	}
	if !compact {
		t.Fatalf("compact = false, want true")
	}

	compact, err = parseCompactOnlyFlags(nil)
	if err != nil {
		t.Fatalf("error = %v", err)
	}
	if compact {
		t.Fatalf("compact = true, want false")
	}

	_, err = parseCompactOnlyFlags([]string{"--surprise"})
	if err == nil {
		t.Fatalf("error = nil, want unknown flag")
	}
}

func TestParseSettingsGetFlags(t *testing.T) {
	includeWeb, err := parseSettingsGetFlags([]string{"--include-web"})
	if err != nil {
		t.Fatalf("error = %v", err)
	}
	if !includeWeb {
		t.Fatalf("includeWeb = false, want true")
	}

	includeWeb, err = parseSettingsGetFlags(nil)
	if err != nil {
		t.Fatalf("error = %v", err)
	}
	if includeWeb {
		t.Fatalf("includeWeb = true, want false")
	}

	_, err = parseSettingsGetFlags([]string{"--surprise"})
	if err == nil {
		t.Fatalf("error = nil, want unknown flag")
	}
}

func TestParseHabitCheckinFlags(t *testing.T) {
	ids, stamp, err := parseHabitCheckinFlags([]string{"--habit", "h1", "--habit", "h2", "--after-stamp", "1715328000000"})
	if err != nil {
		t.Fatalf("error = %v", err)
	}
	if len(ids) != 2 || ids[0] != "h1" || ids[1] != "h2" {
		t.Fatalf("ids = %v", ids)
	}
	if stamp != 1715328000000 {
		t.Fatalf("stamp = %d", stamp)
	}

	// empty
	ids, stamp, err = parseHabitCheckinFlags(nil)
	if err != nil {
		t.Fatalf("error = %v", err)
	}
	if len(ids) != 0 || stamp != 0 {
		t.Fatalf("defaults: ids=%v stamp=%d", ids, stamp)
	}
}

func TestParseTimelineFlags(t *testing.T) {
	to, limit, err := parseTimelineFlags([]string{"--to", "cursor123", "--limit", "14"})
	if err != nil {
		t.Fatalf("error = %v", err)
	}
	if to != "cursor123" || limit != 14 {
		t.Fatalf("to=%q limit=%d", to, limit)
	}

	// defaults
	to, limit, err = parseTimelineFlags(nil)
	if err != nil {
		t.Fatalf("error = %v", err)
	}
	if to != "" || limit != 50 {
		t.Fatalf("defaults: to=%q limit=%d", to, limit)
	}

	_, _, err = parseTimelineFlags([]string{"--to"})
	if err == nil {
		t.Fatalf("error = nil, want missing value")
	}
}
