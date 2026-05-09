package cli

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/DeliciousBuding/dida-cli/internal/auth"
	"github.com/DeliciousBuding/dida-cli/internal/model"
	"github.com/DeliciousBuding/dida-cli/internal/webapi"
)

func runSync(args []string, jsonOut bool, stdout io.Writer, stderr io.Writer) int {
	if len(args) == 0 || args[0] == "-h" || args[0] == "--help" {
		printSyncHelp(stdout)
		return 0
	}
	if args[0] == "checkpoint" {
		if len(args) != 2 {
			return failTyped("sync checkpoint", "validation", "usage: dida sync checkpoint <checkpoint>", "run: dida sync --help", jsonOut, stdout, stderr)
		}
		var checkpoint int64
		if _, err := fmt.Sscanf(args[1], "%d", &checkpoint); err != nil || checkpoint < 0 {
			return failTyped("sync checkpoint", "validation", "checkpoint must be a non-negative integer", "run: dida sync all --json to get latest checkpoint", jsonOut, stdout, stderr)
		}
		return runSyncCheckpoint(checkpoint, jsonOut, stdout, stderr)
	}
	if args[0] != "all" {
		return fail("sync", fmt.Sprintf("unknown sync command %q", args[0]), jsonOut, stdout, stderr)
	}
	token, err := auth.LoadCookieToken()
	if err != nil {
		return missingAuth("sync all", jsonOut, stdout, stderr)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	payload, err := webapi.NewClient(token.Token).FullSync(ctx)
	if err != nil {
		return fail("sync all", err.Error(), jsonOut, stdout, stderr)
	}
	data := model.BuildSyncView(payload.InboxID, payload.Projects, payload.Tasks, payload.ProjectGroups, payload.Tags, time.Now())
	meta := map[string]any{"checkpoint": payload.CheckPoint}
	if jsonOut {
		return writeJSON(stdout, envelope{OK: true, Command: "sync all", Meta: meta, Data: data})
	}
	fmt.Fprintln(stdout, "Sync complete")
	fmt.Fprintf(stdout, "Tasks: %d\n", data.Counts["tasks"])
	fmt.Fprintf(stdout, "Projects: %d\n", data.Counts["projects"])
	fmt.Fprintf(stdout, "Project groups: %d\n", data.Counts["projectGroups"])
	fmt.Fprintf(stdout, "Tags: %d\n", data.Counts["tags"])
	return 0
}

func runSyncCheckpoint(checkpoint int64, jsonOut bool, stdout io.Writer, stderr io.Writer) int {
	token, err := auth.LoadCookieToken()
	if err != nil {
		return missingAuth("sync checkpoint", jsonOut, stdout, stderr)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	payload, err := webapi.NewClient(token.Token).SyncSince(ctx, checkpoint)
	if err != nil {
		return fail("sync checkpoint", err.Error(), jsonOut, stdout, stderr)
	}
	data := model.BuildSyncView(payload.InboxID, payload.Projects, payload.Tasks, payload.ProjectGroups, payload.Tags, time.Now())
	meta := map[string]any{
		"requestedCheckpoint": checkpoint,
		"checkpoint":          payload.CheckPoint,
		"checks":              len(payload.Checks),
		"filters":             len(payload.Filters),
	}
	deltas := map[string]any{
		"taskAdds":      payload.TaskAdds,
		"taskUpdates":   payload.TaskUpdates,
		"taskDeletes":   payload.TaskDeletes,
		"syncOrder":     payload.SyncOrder,
		"syncTaskOrder": payload.SyncTaskOrder,
		"reminders":     payload.Reminders,
	}
	if jsonOut {
		return writeJSON(stdout, envelope{OK: true, Command: "sync checkpoint", Meta: meta, Data: map[string]any{"view": data, "deltas": deltas}})
	}
	fmt.Fprintf(stdout, "Checkpoint: %d\n", payload.CheckPoint)
	fmt.Fprintf(stdout, "Tasks: %d\n", data.Counts["tasks"])
	fmt.Fprintf(stdout, "Projects: %d\n", data.Counts["projects"])
	return 0
}

func runSettings(args []string, jsonOut bool, stdout io.Writer, stderr io.Writer) int {
	if len(args) == 0 || args[0] == "-h" || args[0] == "--help" {
		printSettingsHelp(stdout)
		return 0
	}
	if args[0] != "get" {
		return fail("settings", fmt.Sprintf("unknown settings command %q", args[0]), jsonOut, stdout, stderr)
	}
	token, err := auth.LoadCookieToken()
	if err != nil {
		return missingAuth("settings get", jsonOut, stdout, stderr)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	settings, err := webapi.NewClient(token.Token).Settings(ctx)
	if err != nil {
		return fail("settings get", err.Error(), jsonOut, stdout, stderr)
	}
	data := map[string]any{
		"settings": settings,
	}
	meta := map[string]any{
		"count":    len(settings),
		"timeZone": settings["timeZone"],
		"locale":   settings["locale"],
	}
	if jsonOut {
		return writeJSON(stdout, envelope{OK: true, Command: "settings get", Meta: meta, Data: data})
	}
	fmt.Fprintf(stdout, "Settings: %d keys\n", len(settings))
	fmt.Fprintf(stdout, "Timezone: %v\n", settings["timeZone"])
	fmt.Fprintf(stdout, "Locale: %v\n", settings["locale"])
	return 0
}

func loadSyncView() (model.SyncView, error) {
	token, err := auth.LoadCookieToken()
	if err != nil {
		return model.SyncView{}, fmt.Errorf("missing cookie auth; run: dida auth cookie set --token-stdin")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	payload, err := webapi.NewClient(token.Token).FullSync(ctx)
	if err != nil {
		return model.SyncView{}, err
	}
	return model.BuildSyncView(payload.InboxID, payload.Projects, payload.Tasks, payload.ProjectGroups, payload.Tags, time.Now()), nil
}
