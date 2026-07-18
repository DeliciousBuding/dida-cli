package cli

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/DeliciousBuding/dida-cli/internal/auth"
	"github.com/DeliciousBuding/dida-cli/internal/identity"
	"github.com/DeliciousBuding/dida-cli/internal/model"
	"github.com/DeliciousBuding/dida-cli/internal/openapi"
	"github.com/DeliciousBuding/dida-cli/internal/webapi"
)

// applyTaskFieldNormalization normalizes --start/--due/--reminder after flag parse.
func applyTaskFieldNormalization(start, due, tz string, reminders []string) (string, string, string, []string, error) {
	if strings.TrimSpace(tz) == "" {
		tz = "Asia/Shanghai"
	}
	var err error
	if start, err = normalizeTaskTime(start, tz); err != nil {
		return "", "", "", nil, fmt.Errorf("--start: %w", err)
	}
	if due, err = normalizeTaskTime(due, tz); err != nil {
		return "", "", "", nil, fmt.Errorf("--due: %w", err)
	}
	reminders, err = normalizeReminders(reminders)
	if err != nil {
		return "", "", "", nil, fmt.Errorf("--reminder: %w", err)
	}
	return start, due, tz, reminders, nil
}

func writeTaskWithOptionalReminders(
	command string,
	task webapi.TaskMutation,
	isCreate bool,
	dryRun bool,
	yes bool,
	jsonOut bool,
	stdout io.Writer,
	stderr io.Writer,
) int {
	reminders := append([]string(nil), task.Reminders...)
	task.Reminders = nil // never send reminders on Web batch/task (known HTTP 500)

	if len(reminders) == 0 {
		payload := map[string]any{}
		if isCreate {
			payload["add"] = []webapi.TaskMutation{task}
		} else {
			payload["update"] = []webapi.TaskMutation{task}
		}
		if dryRun {
			return writeMutationPreview(command, payload, yes, jsonOut, stdout, stderr)
		}
		var result map[string]any
		var err error
		if isCreate {
			result, err = executeMutation(func(ctx context.Context, client *webapi.Client) (map[string]any, error) {
				return client.CreateTask(ctx, task)
			})
		} else {
			result, err = executeMutation(func(ctx context.Context, client *webapi.Client) (map[string]any, error) {
				return client.UpdateTask(ctx, task)
			})
		}
		if err != nil {
			return failTyped(command, "api", err.Error(), "", jsonOut, stdout, stderr)
		}
		data := map[string]any{"taskId": task.ID, "projectId": task.ProjectID, "result": result, "writePlan": "webapi"}
		if jsonOut {
			return writeJSON(stdout, envelope{OK: true, Command: command, Meta: map[string]any{"writePlan": "webapi"}, Data: data})
		}
		fmt.Fprintf(stdout, "%s: %s\n", commandLabel(command), task.ID)
		return 0
	}

	webPayload := map[string]any{}
	if isCreate {
		webPayload["add"] = []webapi.TaskMutation{task}
	} else {
		webPayload["update"] = []webapi.TaskMutation{task}
	}
	openPayload := map[string]any{
		"id":        task.ID,
		"projectId": task.ProjectID,
		"reminders": reminders,
	}
	action := "POST /batch/task update"
	if isCreate {
		action = "POST /batch/task add"
	}
	preview := map[string]any{
		"writePlan": "web+openapi-reminders",
		"steps": []map[string]any{
			{"channel": "webapi", "action": action, "payload": webPayload},
			{"channel": "openapi", "action": "POST /open/v1/task/{taskId}", "payload": openPayload, "note": "reminders only; requires matching account identity"},
		},
		"hint": "Web API rejects task reminders (HTTP 500); OpenAPI applies reminders after identity match",
	}
	// Dry-run previews the two-step plan without network or identity gate.
	if dryRun {
		if jsonOut {
			return writeJSON(stdout, envelope{
				OK:      true,
				Command: command,
				Meta:    map[string]any{"dryRun": true, "writePlan": "web+openapi-reminders"},
				Data: map[string]any{
					"dryRun":  true,
					"hint":    "remove --dry-run to execute this write",
					"payload": preview,
				},
			})
		}
		fmt.Fprintf(stdout, "Dry-run %s (web+openapi-reminders)\n", command)
		fmt.Fprintf(stdout, "  1) webapi task body without reminders\n")
		fmt.Fprintf(stdout, "  2) openapi reminders: %s\n", strings.Join(reminders, ", "))
		return 0
	}

	// Execute path: identity guard + OpenAPI token required.
	if err := identity.GuardMultiChannelWrite([]string{identity.ChannelWebAPI, identity.ChannelOpenAPI}); err != nil {
		return failTyped(command, "identity", err.Error(), "run: dida account verify --json", jsonOut, stdout, stderr)
	}

	oaToken, err := openapi.LoadToken()
	if err != nil || oaToken == nil || strings.TrimSpace(oaToken.AccessToken) == "" {
		return failTyped(command, "auth",
			"reminders require OpenAPI OAuth (Web API reminder writes return HTTP 500)",
			"run: dida openapi login --browser --json && dida account verify --json",
			jsonOut, stdout, stderr)
	}

	var webResult map[string]any
	if isCreate {
		webResult, err = executeMutation(func(ctx context.Context, client *webapi.Client) (map[string]any, error) {
			return client.CreateTask(ctx, task)
		})
	} else {
		webResult, err = executeMutation(func(ctx context.Context, client *webapi.Client) (map[string]any, error) {
			return client.UpdateTask(ctx, task)
		})
	}
	if err != nil {
		return failTyped(command, "api", err.Error(), "", jsonOut, stdout, stderr)
	}

	oaClient := openapi.NewClient(oaToken.AccessToken)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	oaResult, oaErr := oaClient.UpdateTask(ctx, task.ID, openPayload)

	userID := ""
	var match *bool
	if store, loadErr := identity.Load(); loadErr == nil {
		if web, ok := store.Get(identity.ChannelWebAPI); ok {
			userID = web.UserID
		}
		res := identity.EvaluateMatch(store)
		match = res.Match
	}

	data := map[string]any{
		"taskId":           task.ID,
		"projectId":        task.ProjectID,
		"result":           webResult,
		"writePlan":        "web+openapi-reminders",
		"remindersApplied": reminders,
		"openapiResult":    oaResult,
	}
	if userID != "" {
		data["accountUserId"] = userID
	}
	meta := map[string]any{
		"writePlan":     "web+openapi-reminders",
		"accountUserId": userID,
		"identityMatch": match,
		"reminders":     reminders,
	}
	if oaErr != nil {
		data["warnings"] = []string{"openapi reminders failed: " + oaErr.Error()}
		if jsonOut {
			_ = writeJSON(stdout, envelope{
				OK:      false,
				Command: command,
				Meta:    meta,
				Data:    data,
				Error: &cliError{
					Type:    "api",
					Message: "web task write succeeded but openapi reminders failed: " + oaErr.Error(),
					Hint:    "task exists; re-run account verify and openapi task update for reminders",
				},
			})
			return 1
		}
		fmt.Fprintf(stderr, "dida: web task write succeeded but openapi reminders failed: %v\n", oaErr)
		fmt.Fprintf(stderr, "task id: %s\n", task.ID)
		return 1
	}
	if jsonOut {
		return writeJSON(stdout, envelope{OK: true, Command: command, Meta: meta, Data: data})
	}
	fmt.Fprintf(stdout, "%s: %s (reminders via openapi)\n", commandLabel(command), task.ID)
	return 0
}

func commandLabel(command string) string {
	switch command {
	case "task create":
		return "Task created"
	case "task update":
		return "Task updated"
	default:
		return command
	}
}

func projectIDsFromSyncView(view model.SyncView) []string {
	ids := make([]string, 0, len(view.Projects))
	for _, p := range view.Projects {
		if strings.TrimSpace(p.ID) != "" {
			ids = append(ids, p.ID)
		}
	}
	return ids
}

// verifyAccountIdentities hits live APIs and rewrites identity.json.
func verifyAccountIdentities(ctx context.Context) (map[string]any, error) {
	store, err := identity.Load()
	if err != nil {
		store = &identity.Store{Channels: map[string]identity.ChannelIdentity{}}
	}

	// Web via cookie + full sync project list
	if _, err := auth.LoadCookieToken(); err == nil {
		statusResult, err := executeRead(func(ctx context.Context, client *webapi.Client) (any, error) {
			return client.UserStatus(ctx)
		})
		if err != nil {
			return nil, fmt.Errorf("webapi user status: %w", err)
		}
		status, _ := statusResult.(map[string]any)
		profileResult, _ := executeRead(func(ctx context.Context, client *webapi.Client) (any, error) {
			return client.UserProfile(ctx)
		})
		profile, _ := profileResult.(map[string]any)
		projectIDs := []string{}
		if view, err := loadSyncView(); err == nil {
			projectIDs = projectIDsFromSyncView(view)
		}
		store.Put(identity.ChannelIdentity{
			Channel:    identity.ChannelWebAPI,
			UserID:     identity.StringField(status, "userId"),
			Name:       identity.StringField(profile, "displayName", "name"),
			ProjectFP:  identity.ProjectFingerprint(projectIDs),
			VerifiedAt: time.Now().UnixMilli(),
			Source:     "account-verify",
		})
	}

	// OpenAPI via project list fingerprint
	if token, err := openapi.LoadToken(); err == nil && token != nil && strings.TrimSpace(token.AccessToken) != "" {
		client := openapi.NewClient(token.AccessToken)
		projects, err := client.Projects(ctx)
		if err != nil {
			return nil, fmt.Errorf("openapi project list: %w", err)
		}
		fp := identity.ProjectFingerprint(identity.ExtractProjectIDs(projects))
		userID, name := "", ""
		if web, ok := store.Get(identity.ChannelWebAPI); ok && web.ProjectFP != "" && web.ProjectFP == fp {
			userID, name = web.UserID, web.Name
		}
		store.Put(identity.ChannelIdentity{
			Channel:    identity.ChannelOpenAPI,
			UserID:     userID,
			Name:       name,
			ProjectFP:  fp,
			VerifiedAt: time.Now().UnixMilli(),
			Source:     "account-verify",
		})
	}

	if err := store.Save(); err != nil {
		return nil, err
	}
	match := identity.EvaluateMatch(store)
	return map[string]any{
		"identities":          store.Channels,
		"identity_match":      match.Match,
		"match_reason":        match.Reason,
		"identity_path":       identity.Path(),
		"allow_cross_account": strings.TrimSpace(os.Getenv("DIDA_ALLOW_CROSS_ACCOUNT")) == "1",
	}, nil
}

func accountWhoami() map[string]any {
	store, err := identity.Load()
	if err != nil {
		store = &identity.Store{Channels: map[string]identity.ChannelIdentity{}}
	}
	match := identity.EvaluateMatch(store)
	return map[string]any{
		"identities":          store.Channels,
		"identity_match":      match.Match,
		"match_reason":        match.Reason,
		"identity_path":       identity.Path(),
		"allow_cross_account": strings.TrimSpace(os.Getenv("DIDA_ALLOW_CROSS_ACCOUNT")) == "1",
		"note":                "cached identity only; run dida account verify --json to refresh from live APIs",
	}
}
