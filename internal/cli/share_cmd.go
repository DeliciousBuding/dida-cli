package cli

import (
	"context"
	"fmt"
	"io"

	"github.com/DeliciousBuding/dida-cli/internal/webapi"
)

func runShare(args []string, jsonOut bool, stdout io.Writer, stderr io.Writer) int {
	if len(args) == 0 || args[0] == "-h" || args[0] == "--help" {
		printShareHelp(stdout)
		return 0
	}
	switch args[0] {
	case "contacts":
		return runShareContacts(jsonOut, stdout, stderr)
	case "recent-users":
		return runRecentProjectUsers(jsonOut, stdout, stderr)
	case "project":
		return runShareProject(args[1:], jsonOut, stdout, stderr)
	default:
		return fail("share", fmt.Sprintf("unknown share command %q", args[0]), jsonOut, stdout, stderr)
	}
}

func runShareContacts(jsonOut bool, stdout io.Writer, stderr io.Writer) int {
	result, err := executeRead(func(ctx context.Context, client *webapi.Client) (any, error) {
		return client.ShareContacts(ctx)
	})
	if err != nil {
		return failTyped("share contacts", "api", err.Error(), "", jsonOut, stdout, stderr)
	}
	data := map[string]any{"contacts": result}
	if jsonOut {
		return writeJSON(stdout, envelope{OK: true, Command: "share contacts", Data: data})
	}
	fmt.Fprintln(stdout, "Share contacts read.")
	return 0
}

func runRecentProjectUsers(jsonOut bool, stdout io.Writer, stderr io.Writer) int {
	result, err := executeRead(func(ctx context.Context, client *webapi.Client) (any, error) {
		return client.RecentProjectUsers(ctx)
	})
	if err != nil {
		return failTyped("share recent-users", "api", err.Error(), "", jsonOut, stdout, stderr)
	}
	items := result.([]map[string]any)
	meta := map[string]any{"count": len(items)}
	data := map[string]any{"recentUsers": items}
	if jsonOut {
		return writeJSON(stdout, envelope{OK: true, Command: "share recent-users", Meta: meta, Data: data})
	}
	printMapList(stdout, items, "recent project users")
	return 0
}

func runShareProject(args []string, jsonOut bool, stdout io.Writer, stderr io.Writer) int {
	if len(args) < 2 {
		return failTyped("share project", "validation", "usage: dida share project <shares|quota|invite-url> <project-id>", "run: dida share --help", jsonOut, stdout, stderr)
	}
	action, projectID := args[0], args[1]
	switch action {
	case "shares":
		return runProjectShares(projectID, jsonOut, stdout, stderr)
	case "quota":
		return runProjectShareQuota(projectID, jsonOut, stdout, stderr)
	case "invite-url":
		return runProjectInviteURL(projectID, jsonOut, stdout, stderr)
	default:
		return failTyped("share project", "validation", fmt.Sprintf("unknown project share action %q", action), "run: dida share --help", jsonOut, stdout, stderr)
	}
}

func runProjectShares(projectID string, jsonOut bool, stdout io.Writer, stderr io.Writer) int {
	result, err := executeRead(func(ctx context.Context, client *webapi.Client) (any, error) {
		return client.ProjectShares(ctx, projectID)
	})
	if err != nil {
		return failTyped("share project shares", "api", err.Error(), "", jsonOut, stdout, stderr)
	}
	items := result.([]map[string]any)
	meta := map[string]any{"count": len(items)}
	data := map[string]any{"projectId": projectID, "shares": items}
	if jsonOut {
		return writeJSON(stdout, envelope{OK: true, Command: "share project shares", Meta: meta, Data: data})
	}
	printMapList(stdout, items, "project shares")
	return 0
}

func runProjectShareQuota(projectID string, jsonOut bool, stdout io.Writer, stderr io.Writer) int {
	result, err := executeRead(func(ctx context.Context, client *webapi.Client) (any, error) {
		return client.ProjectShareQuota(ctx, projectID)
	})
	if err != nil {
		return failTyped("share project quota", "api", err.Error(), "", jsonOut, stdout, stderr)
	}
	data := map[string]any{"projectId": projectID, "quota": result}
	if jsonOut {
		return writeJSON(stdout, envelope{OK: true, Command: "share project quota", Data: data})
	}
	fmt.Fprintf(stdout, "Project share quota: %v\n", result)
	return 0
}

func runProjectInviteURL(projectID string, jsonOut bool, stdout io.Writer, stderr io.Writer) int {
	result, err := executeRead(func(ctx context.Context, client *webapi.Client) (any, error) {
		return client.ProjectInviteURL(ctx, projectID)
	})
	if err != nil {
		return failTyped("share project invite-url", "api", err.Error(), "", jsonOut, stdout, stderr)
	}
	data := map[string]any{"projectId": projectID, "invite": result}
	if jsonOut {
		return writeJSON(stdout, envelope{OK: true, Command: "share project invite-url", Data: data})
	}
	fmt.Fprintln(stdout, "Project invite URL read.")
	return 0
}
