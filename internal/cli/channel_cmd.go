package cli

import (
	"fmt"
	"io"
)

type apiChannelInfo struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	Auth        string   `json:"auth"`
	Role        string   `json:"role"`
	BestFor     []string `json:"bestFor"`
	AvoidFor    []string `json:"avoidFor"`
	FirstChecks []string `json:"firstChecks"`
}

type channelJobGuide struct {
	Job      string `json:"job"`
	Prefer   string `json:"prefer"`
	Fallback string `json:"fallback,omitempty"`
	Notes    string `json:"notes"`
}

type channelBlockerGuide struct {
	Blocker        string `json:"blocker"`
	EvidenceNeeded string `json:"evidenceNeeded"`
}

func runChannel(args []string, jsonOut bool, stdout io.Writer, stderr io.Writer) int {
	if len(args) == 0 || args[0] == "-h" || args[0] == "--help" {
		printChannelHelp(stdout)
		return 0
	}
	switch args[0] {
	case "list":
		return runChannelList(jsonOut, stdout)
	default:
		return fail("channel", fmt.Sprintf("unknown channel command %q", args[0]), jsonOut, stdout, stderr)
	}
}

func runChannelList(jsonOut bool, stdout io.Writer) int {
	data := channelGuideData()
	if jsonOut {
		return writeJSON(stdout, envelope{OK: true, Command: "channel list", Data: data})
	}
	fmt.Fprintln(stdout, "Channels:")
	for _, channel := range data["channels"].([]apiChannelInfo) {
		fmt.Fprintf(stdout, "- %s (%s): %s\n", channel.ID, channel.Name, channel.Role)
	}
	fmt.Fprintln(stdout)
	fmt.Fprintln(stdout, "Run with --json for auth boundaries, job selection, and blocker exit criteria.")
	return 0
}

func channelGuideData() map[string]any {
	return map[string]any{
		"channels": []apiChannelInfo{
			{
				ID:   "webapi",
				Name: "Web API",
				Auth: "browser cookie t saved by dida auth login --browser --json",
				Role: "Primary broad-coverage web-app channel",
				BestFor: []string{
					"agent context packs",
					"normal task/project/folder/tag/comment work",
					"settings, sharing, calendar, templates, stats, trash, closed history, and search",
				},
				AvoidFor: []string{
					"public OAuth REST validation",
					"new private write flows without captured request and rollback evidence",
				},
				FirstChecks: []string{
					"dida auth status --verify --json",
					"dida agent context --outline --json",
				},
			},
			{
				ID:   "official-mcp",
				Name: "Official MCP",
				Auth: "DIDA365_TOKEN or saved local official token config",
				Role: "Official token-based tool channel",
				BestFor: []string{
					"official project and task validation",
					"official habit/focus reads when disposable ids exist",
					"schema-backed official tool exploration",
				},
				AvoidFor: []string{
					"Web API-only metadata",
					"write-capable official call payloads without dry-run wrapper or explicit approval",
				},
				FirstChecks: []string{
					"dida official token status --json",
					"dida official doctor --json",
					"dida official tools --limit 20 --json",
				},
			},
			{
				ID:   "official-openapi",
				Name: "Official OpenAPI",
				Auth: "OAuth access token saved by dida openapi login --browser --json",
				Role: "Official OAuth REST channel",
				BestFor: []string{
					"public REST contract validation",
					"OpenAPI project/task/focus/habit wrappers",
					"OAuth integration testing",
				},
				AvoidFor: []string{
					"MCP dp tokens",
					"browser cookie auth",
					"live writes before OAuth token and disposable resource are verified",
				},
				FirstChecks: []string{
					"dida openapi doctor --json",
					"dida openapi status --json",
				},
			},
		},
		"jobs": []channelJobGuide{
			{Job: "first-account-read", Prefer: "webapi", Fallback: "dida sync all --json", Notes: "Use dida agent context --outline --json for compact task references and a deduplicated taskIndex."},
			{Job: "normal-task-work", Prefer: "webapi", Fallback: "official-mcp when token auth is required", Notes: "Web API task commands have compact reads, dry-run writes, and --yes deletes."},
			{Job: "habit-focus-work", Prefer: "official-mcp or official-openapi", Fallback: "webapi habit/pomo reads for web-app-only views", Notes: "Use official surfaces first; live writes need disposable targets."},
			{Job: "public-rest-validation", Prefer: "official-openapi", Notes: "Requires OpenAPI OAuth client config and saved access token."},
			{Job: "web-app-only-metadata", Prefer: "webapi", Notes: "Use Web API for settings, comments, sharing, calendar, templates, stats, trash, closed history, and search."},
			{Job: "unknown-private-write", Prefer: "none", Notes: "Document endpoint, payload, response, rollback, and live evidence before adding a command."},
		},
		"authBoundaries": []string{
			"Do not send Web API cookie t to Official MCP or OpenAPI commands.",
			"Do not send DIDA365_TOKEN or dp tokens to Web API or OpenAPI commands.",
			"Do not treat an OpenAPI OAuth access token as a browser cookie or MCP token.",
		},
		"blockers": []channelBlockerGuide{
			{Blocker: "openapi-live-resource-calls", EvidenceNeeded: "dida openapi login --browser --json saves an OAuth token, then dida openapi project list --json succeeds."},
			{Blocker: "official-mcp-known-id-habit-focus", EvidenceNeeded: "A disposable habit or focus record exists and get-by-id succeeds against it."},
			{Blocker: "webapi-task-activity", EvidenceNeeded: "A Pro-entitled account or browser trace returns successful GET /task/activity/{taskId} fields and pagination semantics."},
			{Blocker: "task-level-attachments", EvidenceNeeded: "A reversible trace proves upload, task association, read-back/download or preview, quota behavior, and orphan cleanup."},
			{Blocker: "private-write-flows", EvidenceNeeded: "Real traffic captures request bodies, response shapes, permissions, ordering semantics, and rollback paths."},
		},
	}
}
