package cli

import (
	"context"
	"fmt"
	"io"

	"github.com/DeliciousBuding/dida-cli/internal/webapi"
)

func runCalendar(args []string, jsonOut bool, stdout io.Writer, stderr io.Writer) int {
	if len(args) == 0 || args[0] == "-h" || args[0] == "--help" {
		printCalendarHelp(stdout)
		return 0
	}
	switch args[0] {
	case "subscriptions":
		return runCalendarSubscriptions(jsonOut, stdout, stderr)
	case "archived":
		return runCalendarArchivedEvents(jsonOut, stdout, stderr)
	case "third-accounts":
		return runCalendarThirdAccounts(jsonOut, stdout, stderr)
	default:
		return fail("calendar", fmt.Sprintf("unknown calendar command %q", args[0]), jsonOut, stdout, stderr)
	}
}

func runCalendarSubscriptions(jsonOut bool, stdout io.Writer, stderr io.Writer) int {
	result, err := executeRead(func(ctx context.Context, client *webapi.Client) (any, error) {
		return client.CalendarSubscriptions(ctx)
	})
	if err != nil {
		return failTyped("calendar subscriptions", "api", err.Error(), "", jsonOut, stdout, stderr)
	}
	items := result.([]map[string]any)
	meta := map[string]any{"count": len(items)}
	data := map[string]any{"subscriptions": items}
	if jsonOut {
		return writeJSON(stdout, envelope{OK: true, Command: "calendar subscriptions", Meta: meta, Data: data})
	}
	printMapList(stdout, items, "calendar subscriptions")
	return 0
}

func runCalendarArchivedEvents(jsonOut bool, stdout io.Writer, stderr io.Writer) int {
	result, err := executeRead(func(ctx context.Context, client *webapi.Client) (any, error) {
		return client.CalendarArchivedEvents(ctx)
	})
	if err != nil {
		return failTyped("calendar archived", "api", err.Error(), "", jsonOut, stdout, stderr)
	}
	items := result.([]map[string]any)
	meta := map[string]any{"count": len(items)}
	data := map[string]any{"events": items}
	if jsonOut {
		return writeJSON(stdout, envelope{OK: true, Command: "calendar archived", Meta: meta, Data: data})
	}
	printMapList(stdout, items, "archived calendar events")
	return 0
}

func runCalendarThirdAccounts(jsonOut bool, stdout io.Writer, stderr io.Writer) int {
	result, err := executeRead(func(ctx context.Context, client *webapi.Client) (any, error) {
		return client.CalendarThirdAccounts(ctx)
	})
	if err != nil {
		return failTyped("calendar third-accounts", "api", err.Error(), "", jsonOut, stdout, stderr)
	}
	data := map[string]any{"accounts": result}
	if jsonOut {
		return writeJSON(stdout, envelope{OK: true, Command: "calendar third-accounts", Data: data})
	}
	fmt.Fprintln(stdout, "Calendar third-party accounts read.")
	return 0
}
