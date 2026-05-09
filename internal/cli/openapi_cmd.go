package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"os/exec"
	"runtime"
	"time"

	"github.com/DeliciousBuding/dida-cli/internal/openapi"
)

func runOpenAPI(args []string, jsonOut bool, stdout io.Writer, stderr io.Writer) int {
	if len(args) == 0 || args[0] == "-h" || args[0] == "--help" {
		printOpenAPIHelp(stdout)
		return 0
	}
	switch args[0] {
	case "doctor":
		return runOpenAPIDoctor(jsonOut, stdout, stderr)
	case "status":
		return runOpenAPIStatus(jsonOut, stdout, stderr)
	case "logout":
		return runOpenAPILogout(jsonOut, stdout, stderr)
	case "login":
		return runOpenAPILogin(args[1:], jsonOut, stdout, stderr)
	case "auth-url":
		return runOpenAPIAuthURL(args[1:], jsonOut, stdout, stderr)
	case "exchange-code":
		return runOpenAPIExchangeCode(args[1:], jsonOut, stdout, stderr)
	case "listen-callback":
		return runOpenAPIListenCallback(args[1:], jsonOut, stdout, stderr)
	case "project":
		return runOpenAPIProject(args[1:], jsonOut, stdout, stderr)
	case "task":
		return runOpenAPITask(args[1:], jsonOut, stdout, stderr)
	case "focus":
		return runOpenAPIFocus(args[1:], jsonOut, stdout, stderr)
	case "habit":
		return runOpenAPIHabit(args[1:], jsonOut, stdout, stderr)
	default:
		return fail("openapi", fmt.Sprintf("unknown openapi command %q", args[0]), jsonOut, stdout, stderr)
	}
}

func runOpenAPIDoctor(jsonOut bool, stdout io.Writer, stderr io.Writer) int {
	clientID, err := openapi.ResolveClientID("")
	clientSecret, err2 := openapi.ResolveClientSecret("")
	tokenStatus := openapi.TokenStatus()
	data := map[string]any{
		"client_id_available":     err == nil && clientID != "",
		"client_secret_available": err2 == nil && clientSecret != "",
		"token":                   tokenStatus,
		"base_url":                openapi.DefaultAPIBaseURL,
		"auth_url":                openapi.DefaultAuthBaseURL + "/oauth/authorize",
	}
	if jsonOut {
		return writeJSON(stdout, envelope{OK: true, Command: "openapi doctor", Data: data})
	}
	fmt.Fprintf(stdout, "OpenAPI client id: %v\n", data["client_id_available"])
	fmt.Fprintf(stdout, "OpenAPI client secret: %v\n", data["client_secret_available"])
	fmt.Fprintf(stdout, "OpenAPI token: %v\n", tokenStatus["available"])
	return 0
}

func runOpenAPIStatus(jsonOut bool, stdout io.Writer, stderr io.Writer) int {
	data := map[string]any{"token": openapi.TokenStatus()}
	if jsonOut {
		return writeJSON(stdout, envelope{OK: true, Command: "openapi status", Data: data})
	}
	fmt.Fprintf(stdout, "OpenAPI token available: %v\n", data["token"].(map[string]any)["available"])
	return 0
}

func runOpenAPILogout(jsonOut bool, stdout io.Writer, stderr io.Writer) int {
	if err := openapi.ClearToken(); err != nil {
		return failTyped("openapi logout", "auth", err.Error(), "", jsonOut, stdout, stderr)
	}
	if jsonOut {
		return writeJSON(stdout, envelope{OK: true, Command: "openapi logout", Data: map[string]any{"token_cleared": true}})
	}
	fmt.Fprintln(stdout, "OpenAPI token cleared.")
	return 0
}

func runOpenAPILogin(args []string, jsonOut bool, stdout io.Writer, stderr io.Writer) int {
	redirectURI, scope, state, host, port, timeout, noOpen, err := parseOpenAPILoginFlags(args)
	if err != nil {
		return failTyped("openapi login", "validation", err.Error(), "run: dida openapi --help", jsonOut, stdout, stderr)
	}
	clientID, err := openapi.ResolveClientID("")
	if err != nil {
		return failTyped("openapi login", "auth", err.Error(), "set DIDA365_OPENAPI_CLIENT_ID", jsonOut, stdout, stderr)
	}
	clientSecret, err := openapi.ResolveClientSecret("")
	if err != nil {
		return failTyped("openapi login", "auth", err.Error(), "set DIDA365_OPENAPI_CLIENT_SECRET", jsonOut, stdout, stderr)
	}
	redirectURI = fmt.Sprintf("http://%s:%d/callback", host, port)
	authURL := openapi.AuthorizationURL(clientID, redirectURI, scope, state)
	type callbackResult struct {
		code  string
		state string
	}
	codeCh := make(chan callbackResult, 1)
	mux := http.NewServeMux()
	mux.HandleFunc("/callback", func(w http.ResponseWriter, r *http.Request) {
		select {
		case codeCh <- callbackResult{code: r.URL.Query().Get("code"), state: r.URL.Query().Get("state")}:
		default:
		}
		_, _ = w.Write([]byte("DidaCLI OpenAPI callback received. You can return to the terminal."))
	})
	server := &http.Server{Addr: fmt.Sprintf("%s:%d", host, port), Handler: mux}
	go func() { _ = server.ListenAndServe() }()
	if !noOpen {
		_ = openBrowserURL(authURL)
	}
	if jsonOut {
		_ = writeJSON(stdout, envelope{OK: true, Command: "openapi login", Data: map[string]any{
			"authorization_url": authURL,
			"redirect_uri":      redirectURI,
			"state":             state,
			"scope":             scope,
			"waiting":           true,
		}})
	} else {
		fmt.Fprintln(stdout, "Open this URL in a browser and finish authorization:")
		fmt.Fprintln(stdout, authURL)
	}
	select {
	case result := <-codeCh:
		_ = server.Close()
		if err := validateOpenAPICallback(state, result.code, result.state); err != nil {
			return failTyped("openapi login", "auth", err.Error(), "", jsonOut, stdout, stderr)
		}
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		token, err := openapi.ExchangeCode(ctx, clientID, clientSecret, result.code, redirectURI, scope)
		if err != nil {
			return failTyped("openapi login", "api", err.Error(), "", jsonOut, stdout, stderr)
		}
		if err := openapi.SaveToken(token); err != nil {
			return failTyped("openapi login", "auth", err.Error(), "", jsonOut, stdout, stderr)
		}
		data := map[string]any{"saved": true, "state": result.state, "token": openapi.TokenStatus()}
		if jsonOut {
			return writeJSON(stdout, envelope{OK: true, Command: "openapi login", Data: data})
		}
		fmt.Fprintln(stdout, "OpenAPI token saved.")
		return 0
	case <-time.After(timeout):
		_ = server.Close()
		return failTyped("openapi login", "timeout", "timed out waiting for OAuth callback", "", jsonOut, stdout, stderr)
	}
}

func validateOpenAPICallback(expectedState string, code string, gotState string) error {
	if code == "" {
		return fmt.Errorf("oauth callback did not include code")
	}
	if expectedState != "" && gotState != expectedState {
		return fmt.Errorf("oauth callback state mismatch")
	}
	return nil
}

func runOpenAPIAuthURL(args []string, jsonOut bool, stdout io.Writer, stderr io.Writer) int {
	redirectURI, scope, state, err := parseOpenAPIAuthURLFlags(args)
	if err != nil {
		return failTyped("openapi auth-url", "validation", err.Error(), "run: dida openapi --help", jsonOut, stdout, stderr)
	}
	clientID, err := openapi.ResolveClientID("")
	if err != nil {
		return failTyped("openapi auth-url", "auth", err.Error(), "set DIDA365_OPENAPI_CLIENT_ID", jsonOut, stdout, stderr)
	}
	url := openapi.AuthorizationURL(clientID, redirectURI, scope, state)
	data := map[string]any{"authorization_url": url, "redirect_uri": redirectURI, "scope": scope, "state": state}
	if jsonOut {
		return writeJSON(stdout, envelope{OK: true, Command: "openapi auth-url", Data: data})
	}
	fmt.Fprintln(stdout, url)
	return 0
}

func runOpenAPIExchangeCode(args []string, jsonOut bool, stdout io.Writer, stderr io.Writer) int {
	code, redirectURI, scope, err := parseOpenAPIExchangeFlags(args)
	if err != nil {
		return failTyped("openapi exchange-code", "validation", err.Error(), "run: dida openapi --help", jsonOut, stdout, stderr)
	}
	clientID, err := openapi.ResolveClientID("")
	if err != nil {
		return failTyped("openapi exchange-code", "auth", err.Error(), "set DIDA365_OPENAPI_CLIENT_ID", jsonOut, stdout, stderr)
	}
	clientSecret, err := openapi.ResolveClientSecret("")
	if err != nil {
		return failTyped("openapi exchange-code", "auth", err.Error(), "set DIDA365_OPENAPI_CLIENT_SECRET", jsonOut, stdout, stderr)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	token, err := openapi.ExchangeCode(ctx, clientID, clientSecret, code, redirectURI, scope)
	if err != nil {
		return failTyped("openapi exchange-code", "api", err.Error(), "", jsonOut, stdout, stderr)
	}
	if err := openapi.SaveToken(token); err != nil {
		return failTyped("openapi exchange-code", "auth", err.Error(), "", jsonOut, stdout, stderr)
	}
	data := map[string]any{"saved": true, "token": openapi.TokenStatus()}
	if jsonOut {
		return writeJSON(stdout, envelope{OK: true, Command: "openapi exchange-code", Data: data})
	}
	fmt.Fprintln(stdout, "OpenAPI token saved.")
	return 0
}

func runOpenAPIListenCallback(args []string, jsonOut bool, stdout io.Writer, stderr io.Writer) int {
	host, port, err := parseOpenAPIListenFlags(args)
	if err != nil {
		return failTyped("openapi listen-callback", "validation", err.Error(), "run: dida openapi --help", jsonOut, stdout, stderr)
	}
	redirectURI := fmt.Sprintf("http://%s:%d/callback", host, port)
	codeCh := make(chan map[string]string, 1)
	mux := http.NewServeMux()
	mux.HandleFunc("/callback", func(w http.ResponseWriter, r *http.Request) {
		values := map[string]string{
			"code":  r.URL.Query().Get("code"),
			"state": r.URL.Query().Get("state"),
		}
		select {
		case codeCh <- values:
		default:
		}
		_, _ = w.Write([]byte("OpenAPI callback received. You can return to the CLI."))
	})
	server := &http.Server{Addr: fmt.Sprintf("%s:%d", host, port), Handler: mux}
	go func() { _ = server.ListenAndServe() }()
	select {
	case values := <-codeCh:
		_ = server.Close()
		data := map[string]any{"redirect_uri": redirectURI, "code": values["code"], "state": values["state"]}
		if jsonOut {
			_ = writeJSON(stdout, envelope{OK: true, Command: "openapi listen-callback", Data: data})
			return 0
		}
		fmt.Fprintf(stdout, "Code: %s\nState: %s\n", values["code"], values["state"])
		return 0
	case <-time.After(10 * time.Minute):
		_ = server.Close()
		return failTyped("openapi listen-callback", "timeout", "timed out waiting for OAuth callback", "", jsonOut, stdout, stderr)
	}
}

func runOpenAPIProject(args []string, jsonOut bool, stdout io.Writer, stderr io.Writer) int {
	if len(args) == 0 {
		return failTyped("openapi project", "validation", "usage: dida openapi project list|get|data", "run: dida openapi --help", jsonOut, stdout, stderr)
	}
	token, err := openapi.LoadToken()
	if err != nil {
		return failTyped("openapi project "+args[0], "auth", err.Error(), "run: dida openapi login first", jsonOut, stdout, stderr)
	}
	client := openapi.NewClient(token.AccessToken)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	switch args[0] {
	case "list":
		projects, err := client.Projects(ctx)
		if err != nil {
			return failTyped("openapi project list", "api", err.Error(), "", jsonOut, stdout, stderr)
		}
		meta := map[string]any{"count": len(projects)}
		data := map[string]any{"projects": projects}
		if jsonOut {
			return writeJSON(stdout, envelope{OK: true, Command: "openapi project list", Meta: meta, Data: data})
		}
		printMapList(stdout, projects, "openapi projects")
		return 0
	case "get":
		if len(args) != 2 {
			return failTyped("openapi project get", "validation", "usage: dida openapi project get <project-id>", "run: dida openapi --help", jsonOut, stdout, stderr)
		}
		project, err := client.Project(ctx, args[1])
		if err != nil {
			return failTyped("openapi project get", "api", err.Error(), "", jsonOut, stdout, stderr)
		}
		if jsonOut {
			return writeJSON(stdout, envelope{OK: true, Command: "openapi project get", Data: map[string]any{"project": project}})
		}
		return writeJSON(stdout, project)
	case "data":
		if len(args) != 2 {
			return failTyped("openapi project data", "validation", "usage: dida openapi project data <project-id>", "run: dida openapi --help", jsonOut, stdout, stderr)
		}
		data, err := client.ProjectData(ctx, args[1])
		if err != nil {
			return failTyped("openapi project data", "api", err.Error(), "", jsonOut, stdout, stderr)
		}
		if jsonOut {
			return writeJSON(stdout, envelope{OK: true, Command: "openapi project data", Data: data})
		}
		return writeJSON(stdout, data)
	default:
		return failTyped("openapi project", "validation", fmt.Sprintf("unknown openapi project command %q", args[0]), "run: dida openapi --help", jsonOut, stdout, stderr)
	}
}

func runOpenAPITask(args []string, jsonOut bool, stdout io.Writer, stderr io.Writer) int {
	if len(args) == 0 {
		return failTyped("openapi task", "validation", "usage: dida openapi task get|create|update|complete|delete|move|completed|filter", "run: dida openapi --help", jsonOut, stdout, stderr)
	}
	if handled, code := runOpenAPITaskDryRun(args, jsonOut, stdout, stderr); handled {
		return code
	}
	token, err := openapi.LoadToken()
	if err != nil {
		return failTyped("openapi task "+args[0], "auth", err.Error(), "run: dida openapi login first", jsonOut, stdout, stderr)
	}
	client := openapi.NewClient(token.AccessToken)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	switch args[0] {
	case "get":
		projectID, taskID, err := parseOpenAPITaskTargetFlags(args[1:])
		if err != nil {
			return failTyped("openapi task get", "validation", err.Error(), "run: dida openapi --help", jsonOut, stdout, stderr)
		}
		task, err := client.Task(ctx, projectID, taskID)
		if err != nil {
			return failTyped("openapi task get", "api", err.Error(), "", jsonOut, stdout, stderr)
		}
		if jsonOut {
			return writeJSON(stdout, envelope{OK: true, Command: "openapi task get", Data: map[string]any{"task": task}})
		}
		return writeJSON(stdout, task)
	case "create":
		payload, dryRun, err := parseOpenAPIJSONWriteFlags(args[1:])
		if err != nil {
			return failTyped("openapi task create", "validation", err.Error(), "run: dida openapi --help", jsonOut, stdout, stderr)
		}
		if dryRun {
			return writeJSON(stdout, envelope{OK: true, Command: "openapi task create", Data: map[string]any{"dry_run": true, "payload": payload}})
		}
		task, err := client.CreateTask(ctx, payload)
		if err != nil {
			return failTyped("openapi task create", "api", err.Error(), "", jsonOut, stdout, stderr)
		}
		if jsonOut {
			return writeJSON(stdout, envelope{OK: true, Command: "openapi task create", Data: map[string]any{"task": task}})
		}
		return writeJSON(stdout, task)
	case "update":
		taskID, payload, dryRun, err := parseOpenAPITaskUpdateFlags(args[1:])
		if err != nil {
			return failTyped("openapi task update", "validation", err.Error(), "run: dida openapi --help", jsonOut, stdout, stderr)
		}
		if _, ok := payload["id"]; !ok {
			payload["id"] = taskID
		}
		if dryRun {
			return writeJSON(stdout, envelope{OK: true, Command: "openapi task update", Data: map[string]any{"dry_run": true, "task_id": taskID, "payload": payload}})
		}
		task, err := client.UpdateTask(ctx, taskID, payload)
		if err != nil {
			return failTyped("openapi task update", "api", err.Error(), "", jsonOut, stdout, stderr)
		}
		if jsonOut {
			return writeJSON(stdout, envelope{OK: true, Command: "openapi task update", Data: map[string]any{"task": task}})
		}
		return writeJSON(stdout, task)
	case "complete":
		projectID, taskID, dryRun, err := parseOpenAPITaskTargetWriteFlags(args[1:], false)
		if err != nil {
			return failTyped("openapi task complete", "validation", err.Error(), "run: dida openapi --help", jsonOut, stdout, stderr)
		}
		if dryRun {
			return writeJSON(stdout, envelope{OK: true, Command: "openapi task complete", Data: map[string]any{"dry_run": true, "project_id": projectID, "task_id": taskID}})
		}
		if err := client.CompleteTask(ctx, projectID, taskID); err != nil {
			return failTyped("openapi task complete", "api", err.Error(), "", jsonOut, stdout, stderr)
		}
		return writeJSON(stdout, envelope{OK: true, Command: "openapi task complete", Data: map[string]any{"project_id": projectID, "task_id": taskID}})
	case "delete":
		projectID, taskID, dryRun, err := parseOpenAPITaskTargetWriteFlags(args[1:], true)
		if err != nil {
			return failTyped("openapi task delete", "validation", err.Error(), "run: dida openapi --help", jsonOut, stdout, stderr)
		}
		if dryRun {
			return writeJSON(stdout, envelope{OK: true, Command: "openapi task delete", Data: map[string]any{"dry_run": true, "project_id": projectID, "task_id": taskID}})
		}
		if err := client.DeleteTask(ctx, projectID, taskID); err != nil {
			return failTyped("openapi task delete", "api", err.Error(), "", jsonOut, stdout, stderr)
		}
		return writeJSON(stdout, envelope{OK: true, Command: "openapi task delete", Data: map[string]any{"project_id": projectID, "task_id": taskID}})
	case "move":
		payload, dryRun, err := parseOpenAPIAnyJSONWriteFlags(args[1:])
		if err != nil {
			return failTyped("openapi task move", "validation", err.Error(), "run: dida openapi --help", jsonOut, stdout, stderr)
		}
		if dryRun {
			return writeJSON(stdout, envelope{OK: true, Command: "openapi task move", Data: map[string]any{"dry_run": true, "payload": payload}})
		}
		result, err := client.MoveTasks(ctx, payload)
		if err != nil {
			return failTyped("openapi task move", "api", err.Error(), "", jsonOut, stdout, stderr)
		}
		return writeJSON(stdout, envelope{OK: true, Command: "openapi task move", Data: map[string]any{"result": result}})
	case "completed":
		payload, err := parseOpenAPIJSONReadFlags(args[1:])
		if err != nil {
			return failTyped("openapi task completed", "validation", err.Error(), "run: dida openapi --help", jsonOut, stdout, stderr)
		}
		tasks, err := client.CompletedTasks(ctx, payload)
		if err != nil {
			return failTyped("openapi task completed", "api", err.Error(), "", jsonOut, stdout, stderr)
		}
		return writeJSON(stdout, envelope{OK: true, Command: "openapi task completed", Meta: map[string]any{"count": len(tasks)}, Data: map[string]any{"tasks": tasks}})
	case "filter":
		payload, err := parseOpenAPIJSONReadFlags(args[1:])
		if err != nil {
			return failTyped("openapi task filter", "validation", err.Error(), "run: dida openapi --help", jsonOut, stdout, stderr)
		}
		tasks, err := client.FilterTasks(ctx, payload)
		if err != nil {
			return failTyped("openapi task filter", "api", err.Error(), "", jsonOut, stdout, stderr)
		}
		return writeJSON(stdout, envelope{OK: true, Command: "openapi task filter", Meta: map[string]any{"count": len(tasks)}, Data: map[string]any{"tasks": tasks}})
	default:
		return failTyped("openapi task", "validation", fmt.Sprintf("unknown openapi task command %q", args[0]), "run: dida openapi --help", jsonOut, stdout, stderr)
	}
}

func runOpenAPIFocus(args []string, jsonOut bool, stdout io.Writer, stderr io.Writer) int {
	if len(args) == 0 {
		return failTyped("openapi focus", "validation", "usage: dida openapi focus get|list|delete", "run: dida openapi --help", jsonOut, stdout, stderr)
	}
	if args[0] == "delete" && openAPIHasFlag(args[1:], "--dry-run") {
		focusID, focusType, _, _, err := parseOpenAPIFocusDeleteFlags(args[1:])
		if err != nil {
			return failTyped("openapi focus delete", "validation", err.Error(), "run: dida openapi --help", jsonOut, stdout, stderr)
		}
		return writeJSON(stdout, envelope{OK: true, Command: "openapi focus delete", Data: map[string]any{"dry_run": true, "focus_id": focusID, "type": focusType}})
	}
	token, err := openapi.LoadToken()
	if err != nil {
		return failTyped("openapi focus "+args[0], "auth", err.Error(), "run: dida openapi login first", jsonOut, stdout, stderr)
	}
	client := openapi.NewClient(token.AccessToken)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	switch args[0] {
	case "get":
		focusID, focusType, err := parseOpenAPIFocusGetFlags(args[1:])
		if err != nil {
			return failTyped("openapi focus get", "validation", err.Error(), "run: dida openapi --help", jsonOut, stdout, stderr)
		}
		focus, err := client.Focus(ctx, focusID, focusType)
		if err != nil {
			return failTyped("openapi focus get", "api", err.Error(), "", jsonOut, stdout, stderr)
		}
		return writeJSON(stdout, envelope{OK: true, Command: "openapi focus get", Data: map[string]any{"focus": focus}})
	case "list":
		from, to, focusType, err := parseOpenAPIFocusListFlags(args[1:])
		if err != nil {
			return failTyped("openapi focus list", "validation", err.Error(), "run: dida openapi --help", jsonOut, stdout, stderr)
		}
		focuses, err := client.Focuses(ctx, from, to, focusType)
		if err != nil {
			return failTyped("openapi focus list", "api", err.Error(), "", jsonOut, stdout, stderr)
		}
		return writeJSON(stdout, envelope{OK: true, Command: "openapi focus list", Meta: map[string]any{"count": len(focuses)}, Data: map[string]any{"focuses": focuses}})
	case "delete":
		focusID, focusType, dryRun, yes, err := parseOpenAPIFocusDeleteFlags(args[1:])
		if err != nil {
			return failTyped("openapi focus delete", "validation", err.Error(), "run: dida openapi --help", jsonOut, stdout, stderr)
		}
		if dryRun {
			return writeJSON(stdout, envelope{OK: true, Command: "openapi focus delete", Data: map[string]any{"dry_run": true, "focus_id": focusID, "type": focusType}})
		}
		if !yes {
			return failTyped("openapi focus delete", "confirmation_required", "openapi focus delete requires --yes", "preview first with --dry-run", jsonOut, stdout, stderr)
		}
		result, err := client.DeleteFocus(ctx, focusID, focusType)
		if err != nil {
			return failTyped("openapi focus delete", "api", err.Error(), "", jsonOut, stdout, stderr)
		}
		return writeJSON(stdout, envelope{OK: true, Command: "openapi focus delete", Data: map[string]any{"result": result}})
	default:
		return failTyped("openapi focus", "validation", fmt.Sprintf("unknown openapi focus command %q", args[0]), "run: dida openapi --help", jsonOut, stdout, stderr)
	}
}

func runOpenAPIHabit(args []string, jsonOut bool, stdout io.Writer, stderr io.Writer) int {
	if len(args) == 0 {
		return failTyped("openapi habit", "validation", "usage: dida openapi habit list|get|create|update|checkin|checkins", "run: dida openapi --help", jsonOut, stdout, stderr)
	}
	if handled, code := runOpenAPIHabitDryRun(args, jsonOut, stdout, stderr); handled {
		return code
	}
	token, err := openapi.LoadToken()
	if err != nil {
		return failTyped("openapi habit "+args[0], "auth", err.Error(), "run: dida openapi login first", jsonOut, stdout, stderr)
	}
	client := openapi.NewClient(token.AccessToken)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	switch args[0] {
	case "list":
		habits, err := client.Habits(ctx)
		if err != nil {
			return failTyped("openapi habit list", "api", err.Error(), "", jsonOut, stdout, stderr)
		}
		return writeJSON(stdout, envelope{OK: true, Command: "openapi habit list", Meta: map[string]any{"count": len(habits)}, Data: map[string]any{"habits": habits}})
	case "get":
		habitID, err := parseOpenAPISingleID(args[1:], "habit")
		if err != nil {
			return failTyped("openapi habit get", "validation", err.Error(), "run: dida openapi --help", jsonOut, stdout, stderr)
		}
		habit, err := client.Habit(ctx, habitID)
		if err != nil {
			return failTyped("openapi habit get", "api", err.Error(), "", jsonOut, stdout, stderr)
		}
		return writeJSON(stdout, envelope{OK: true, Command: "openapi habit get", Data: map[string]any{"habit": habit}})
	case "create":
		payload, dryRun, err := parseOpenAPIJSONWriteFlags(args[1:])
		if err != nil {
			return failTyped("openapi habit create", "validation", err.Error(), "run: dida openapi --help", jsonOut, stdout, stderr)
		}
		if dryRun {
			return writeJSON(stdout, envelope{OK: true, Command: "openapi habit create", Data: map[string]any{"dry_run": true, "payload": payload}})
		}
		habit, err := client.CreateHabit(ctx, payload)
		if err != nil {
			return failTyped("openapi habit create", "api", err.Error(), "", jsonOut, stdout, stderr)
		}
		return writeJSON(stdout, envelope{OK: true, Command: "openapi habit create", Data: map[string]any{"habit": habit}})
	case "update":
		habitID, payload, dryRun, err := parseOpenAPIIDJSONWriteFlags(args[1:], "habit")
		if err != nil {
			return failTyped("openapi habit update", "validation", err.Error(), "run: dida openapi --help", jsonOut, stdout, stderr)
		}
		if dryRun {
			return writeJSON(stdout, envelope{OK: true, Command: "openapi habit update", Data: map[string]any{"dry_run": true, "habit_id": habitID, "payload": payload}})
		}
		habit, err := client.UpdateHabit(ctx, habitID, payload)
		if err != nil {
			return failTyped("openapi habit update", "api", err.Error(), "", jsonOut, stdout, stderr)
		}
		return writeJSON(stdout, envelope{OK: true, Command: "openapi habit update", Data: map[string]any{"habit": habit}})
	case "checkin":
		habitID, payload, dryRun, err := parseOpenAPIIDJSONWriteFlags(args[1:], "habit")
		if err != nil {
			return failTyped("openapi habit checkin", "validation", err.Error(), "run: dida openapi --help", jsonOut, stdout, stderr)
		}
		if dryRun {
			return writeJSON(stdout, envelope{OK: true, Command: "openapi habit checkin", Data: map[string]any{"dry_run": true, "habit_id": habitID, "payload": payload}})
		}
		checkin, err := client.UpsertHabitCheckin(ctx, habitID, payload)
		if err != nil {
			return failTyped("openapi habit checkin", "api", err.Error(), "", jsonOut, stdout, stderr)
		}
		return writeJSON(stdout, envelope{OK: true, Command: "openapi habit checkin", Data: map[string]any{"checkin": checkin}})
	case "checkins":
		habitIDs, from, to, err := parseOpenAPIHabitCheckinsFlags(args[1:])
		if err != nil {
			return failTyped("openapi habit checkins", "validation", err.Error(), "run: dida openapi --help", jsonOut, stdout, stderr)
		}
		checkins, err := client.HabitCheckins(ctx, habitIDs, from, to)
		if err != nil {
			return failTyped("openapi habit checkins", "api", err.Error(), "", jsonOut, stdout, stderr)
		}
		return writeJSON(stdout, envelope{OK: true, Command: "openapi habit checkins", Meta: map[string]any{"count": len(checkins)}, Data: map[string]any{"checkins": checkins}})
	default:
		return failTyped("openapi habit", "validation", fmt.Sprintf("unknown openapi habit command %q", args[0]), "run: dida openapi --help", jsonOut, stdout, stderr)
	}
}

func runOpenAPITaskDryRun(args []string, jsonOut bool, stdout io.Writer, stderr io.Writer) (bool, int) {
	if !openAPIHasFlag(args[1:], "--dry-run") {
		return false, 0
	}
	switch args[0] {
	case "create":
		payload, _, err := parseOpenAPIJSONWriteFlags(args[1:])
		if err != nil {
			return true, failTyped("openapi task create", "validation", err.Error(), "run: dida openapi --help", jsonOut, stdout, stderr)
		}
		return true, writeJSON(stdout, envelope{OK: true, Command: "openapi task create", Data: map[string]any{"dry_run": true, "payload": payload}})
	case "update":
		taskID, payload, _, err := parseOpenAPITaskUpdateFlags(args[1:])
		if err != nil {
			return true, failTyped("openapi task update", "validation", err.Error(), "run: dida openapi --help", jsonOut, stdout, stderr)
		}
		if _, ok := payload["id"]; !ok {
			payload["id"] = taskID
		}
		return true, writeJSON(stdout, envelope{OK: true, Command: "openapi task update", Data: map[string]any{"dry_run": true, "task_id": taskID, "payload": payload}})
	case "complete":
		projectID, taskID, _, err := parseOpenAPITaskTargetWriteFlags(args[1:], false)
		if err != nil {
			return true, failTyped("openapi task complete", "validation", err.Error(), "run: dida openapi --help", jsonOut, stdout, stderr)
		}
		return true, writeJSON(stdout, envelope{OK: true, Command: "openapi task complete", Data: map[string]any{"dry_run": true, "project_id": projectID, "task_id": taskID}})
	case "delete":
		projectID, taskID, _, err := parseOpenAPITaskTargetWriteFlags(args[1:], true)
		if err != nil {
			return true, failTyped("openapi task delete", "validation", err.Error(), "run: dida openapi --help", jsonOut, stdout, stderr)
		}
		return true, writeJSON(stdout, envelope{OK: true, Command: "openapi task delete", Data: map[string]any{"dry_run": true, "project_id": projectID, "task_id": taskID}})
	case "move":
		payload, _, err := parseOpenAPIAnyJSONWriteFlags(args[1:])
		if err != nil {
			return true, failTyped("openapi task move", "validation", err.Error(), "run: dida openapi --help", jsonOut, stdout, stderr)
		}
		return true, writeJSON(stdout, envelope{OK: true, Command: "openapi task move", Data: map[string]any{"dry_run": true, "payload": payload}})
	default:
		return false, 0
	}
}

func runOpenAPIHabitDryRun(args []string, jsonOut bool, stdout io.Writer, stderr io.Writer) (bool, int) {
	if !openAPIHasFlag(args[1:], "--dry-run") {
		return false, 0
	}
	switch args[0] {
	case "create":
		payload, _, err := parseOpenAPIJSONWriteFlags(args[1:])
		if err != nil {
			return true, failTyped("openapi habit create", "validation", err.Error(), "run: dida openapi --help", jsonOut, stdout, stderr)
		}
		return true, writeJSON(stdout, envelope{OK: true, Command: "openapi habit create", Data: map[string]any{"dry_run": true, "payload": payload}})
	case "update":
		habitID, payload, _, err := parseOpenAPIIDJSONWriteFlags(args[1:], "habit")
		if err != nil {
			return true, failTyped("openapi habit update", "validation", err.Error(), "run: dida openapi --help", jsonOut, stdout, stderr)
		}
		return true, writeJSON(stdout, envelope{OK: true, Command: "openapi habit update", Data: map[string]any{"dry_run": true, "habit_id": habitID, "payload": payload}})
	case "checkin":
		habitID, payload, _, err := parseOpenAPIIDJSONWriteFlags(args[1:], "habit")
		if err != nil {
			return true, failTyped("openapi habit checkin", "validation", err.Error(), "run: dida openapi --help", jsonOut, stdout, stderr)
		}
		return true, writeJSON(stdout, envelope{OK: true, Command: "openapi habit checkin", Data: map[string]any{"dry_run": true, "habit_id": habitID, "payload": payload}})
	default:
		return false, 0
	}
}

func parseOpenAPIAuthURLFlags(args []string) (string, string, string, error) {
	redirectURI := "http://127.0.0.1:17890/callback"
	scope := openapi.DefaultScopes
	state := fmt.Sprintf("dida-%d-%d", time.Now().Unix(), rand.Intn(100000))
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--redirect-uri":
			if i+1 >= len(args) {
				return "", "", "", fmt.Errorf("--redirect-uri requires a value")
			}
			redirectURI = args[i+1]
			i++
		case "--scope":
			if i+1 >= len(args) {
				return "", "", "", fmt.Errorf("--scope requires a value")
			}
			scope = args[i+1]
			i++
		case "--state":
			if i+1 >= len(args) {
				return "", "", "", fmt.Errorf("--state requires a value")
			}
			state = args[i+1]
			i++
		default:
			return "", "", "", fmt.Errorf("unknown flag %q", args[i])
		}
	}
	return redirectURI, scope, state, nil
}

func parseOpenAPITaskTargetFlags(args []string) (string, string, error) {
	projectID := ""
	taskID := ""
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--project":
			if i+1 >= len(args) {
				return "", "", fmt.Errorf("--project requires a value")
			}
			projectID = args[i+1]
			i++
		case "--task":
			if i+1 >= len(args) {
				return "", "", fmt.Errorf("--task requires a value")
			}
			taskID = args[i+1]
			i++
		default:
			return "", "", fmt.Errorf("unknown flag %q", args[i])
		}
	}
	if projectID == "" {
		return "", "", fmt.Errorf("missing required --project")
	}
	if taskID == "" {
		return "", "", fmt.Errorf("missing required --task")
	}
	return projectID, taskID, nil
}

func parseOpenAPITaskTargetWriteFlags(args []string, requireYes bool) (string, string, bool, error) {
	projectID := ""
	taskID := ""
	dryRun := false
	yes := false
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--project":
			if i+1 >= len(args) {
				return "", "", false, fmt.Errorf("--project requires a value")
			}
			projectID = args[i+1]
			i++
		case "--task":
			if i+1 >= len(args) {
				return "", "", false, fmt.Errorf("--task requires a value")
			}
			taskID = args[i+1]
			i++
		case "--dry-run":
			dryRun = true
		case "--yes":
			yes = true
		default:
			return "", "", false, fmt.Errorf("unknown flag %q", args[i])
		}
	}
	if projectID == "" {
		return "", "", false, fmt.Errorf("missing required --project")
	}
	if taskID == "" {
		return "", "", false, fmt.Errorf("missing required --task")
	}
	if requireYes && !dryRun && !yes {
		return "", "", false, fmt.Errorf("openapi task delete requires --yes")
	}
	return projectID, taskID, dryRun, nil
}

func parseOpenAPITaskUpdateFlags(args []string) (string, map[string]any, bool, error) {
	if len(args) == 0 {
		return "", nil, false, fmt.Errorf("usage: dida openapi task update <task-id> --args-json <json> [--dry-run]")
	}
	taskID := args[0]
	payload, dryRun, err := parseOpenAPIJSONWriteFlags(args[1:])
	return taskID, payload, dryRun, err
}

func parseOpenAPIFocusGetFlags(args []string) (string, string, error) {
	if len(args) == 0 {
		return "", "", fmt.Errorf("usage: dida openapi focus get <focus-id> --type 0|1")
	}
	focusID := args[0]
	focusType := "0"
	for i := 1; i < len(args); i++ {
		switch args[i] {
		case "--type":
			if i+1 >= len(args) {
				return "", "", fmt.Errorf("--type requires 0 or 1")
			}
			focusType = args[i+1]
			i++
		default:
			return "", "", fmt.Errorf("unknown flag %q", args[i])
		}
	}
	if focusType != "0" && focusType != "1" {
		return "", "", fmt.Errorf("--type must be 0 or 1")
	}
	return focusID, focusType, nil
}

func parseOpenAPIFocusListFlags(args []string) (string, string, string, error) {
	from := ""
	to := ""
	focusType := "0"
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--from":
			if i+1 >= len(args) {
				return "", "", "", fmt.Errorf("--from requires a value")
			}
			from = args[i+1]
			i++
		case "--to":
			if i+1 >= len(args) {
				return "", "", "", fmt.Errorf("--to requires a value")
			}
			to = args[i+1]
			i++
		case "--type":
			if i+1 >= len(args) {
				return "", "", "", fmt.Errorf("--type requires 0 or 1")
			}
			focusType = args[i+1]
			i++
		default:
			return "", "", "", fmt.Errorf("unknown flag %q", args[i])
		}
	}
	if from == "" {
		return "", "", "", fmt.Errorf("missing required --from")
	}
	if to == "" {
		return "", "", "", fmt.Errorf("missing required --to")
	}
	if focusType != "0" && focusType != "1" {
		return "", "", "", fmt.Errorf("--type must be 0 or 1")
	}
	return from, to, focusType, nil
}

func parseOpenAPIFocusDeleteFlags(args []string) (string, string, bool, bool, error) {
	focusID, focusType, err := parseOpenAPIFocusGetFlags(argsWithoutBooleans(args, "--dry-run", "--yes"))
	if err != nil {
		return "", "", false, false, err
	}
	return focusID, focusType, openAPIHasFlag(args, "--dry-run"), openAPIHasFlag(args, "--yes"), nil
}

func parseOpenAPISingleID(args []string, name string) (string, error) {
	if len(args) != 1 {
		return "", fmt.Errorf("usage: dida openapi %s get <%s-id>", name, name)
	}
	return args[0], nil
}

func parseOpenAPIIDJSONWriteFlags(args []string, name string) (string, map[string]any, bool, error) {
	if len(args) == 0 {
		return "", nil, false, fmt.Errorf("usage: dida openapi %s <command> <%s-id> --args-json <json> [--dry-run]", name, name)
	}
	id := args[0]
	payload, dryRun, err := parseOpenAPIJSONWriteFlags(args[1:])
	return id, payload, dryRun, err
}

func parseOpenAPIHabitCheckinsFlags(args []string) (string, string, string, error) {
	habitIDs := ""
	from := ""
	to := ""
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--habit-ids", "--habits":
			if i+1 >= len(args) {
				return "", "", "", fmt.Errorf("%s requires a comma-separated value", args[i])
			}
			habitIDs = args[i+1]
			i++
		case "--from":
			if i+1 >= len(args) {
				return "", "", "", fmt.Errorf("--from requires YYYYMMDD")
			}
			from = args[i+1]
			i++
		case "--to":
			if i+1 >= len(args) {
				return "", "", "", fmt.Errorf("--to requires YYYYMMDD")
			}
			to = args[i+1]
			i++
		default:
			return "", "", "", fmt.Errorf("unknown flag %q", args[i])
		}
	}
	if habitIDs == "" {
		return "", "", "", fmt.Errorf("missing required --habit-ids")
	}
	if from == "" {
		return "", "", "", fmt.Errorf("missing required --from")
	}
	if to == "" {
		return "", "", "", fmt.Errorf("missing required --to")
	}
	return habitIDs, from, to, nil
}

func argsWithoutBooleans(args []string, names ...string) []string {
	out := make([]string, 0, len(args))
	for _, arg := range args {
		if containsString(names, arg) {
			continue
		}
		out = append(out, arg)
	}
	return out
}

func containsString(values []string, value string) bool {
	for _, item := range values {
		if item == value {
			return true
		}
	}
	return false
}

func openAPIHasFlag(args []string, flag string) bool {
	for _, arg := range args {
		if arg == flag {
			return true
		}
	}
	return false
}

func parseOpenAPIJSONReadFlags(args []string) (map[string]any, error) {
	payload, _, err := parseOpenAPIJSONWriteFlags(args)
	return payload, err
}

func parseOpenAPIJSONWriteFlags(args []string) (map[string]any, bool, error) {
	raw, dryRun, err := parseOpenAPIAnyJSONWriteFlags(args)
	if err != nil {
		return nil, false, err
	}
	payload, ok := raw.(map[string]any)
	if !ok {
		return nil, false, fmt.Errorf("--args-json must decode to a JSON object")
	}
	return payload, dryRun, nil
}

func parseOpenAPIAnyJSONWriteFlags(args []string) (any, bool, error) {
	var payload any = map[string]any{}
	dryRun := false
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--args-json":
			if i+1 >= len(args) {
				return nil, false, fmt.Errorf("--args-json requires a value")
			}
			if err := decodeJSONArg(args[i+1], &payload); err != nil {
				return nil, false, err
			}
			i++
		case "--dry-run":
			dryRun = true
		default:
			return nil, false, fmt.Errorf("unknown flag %q", args[i])
		}
	}
	return payload, dryRun, nil
}

func decodeJSONArg(value string, out any) error {
	if err := json.Unmarshal([]byte(value), out); err != nil {
		return fmt.Errorf("decode --args-json: %w", err)
	}
	return nil
}

func parseOpenAPIExchangeFlags(args []string) (string, string, string, error) {
	code := ""
	redirectURI := "http://127.0.0.1:17890/callback"
	scope := openapi.DefaultScopes
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--code":
			if i+1 >= len(args) {
				return "", "", "", fmt.Errorf("--code requires a value")
			}
			code = args[i+1]
			i++
		case "--redirect-uri":
			if i+1 >= len(args) {
				return "", "", "", fmt.Errorf("--redirect-uri requires a value")
			}
			redirectURI = args[i+1]
			i++
		case "--scope":
			if i+1 >= len(args) {
				return "", "", "", fmt.Errorf("--scope requires a value")
			}
			scope = args[i+1]
			i++
		default:
			return "", "", "", fmt.Errorf("unknown flag %q", args[i])
		}
	}
	if code == "" {
		return "", "", "", fmt.Errorf("missing --code")
	}
	return code, redirectURI, scope, nil
}

func parseOpenAPIListenFlags(args []string) (string, int, error) {
	host := "127.0.0.1"
	port := 17890
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--host":
			if i+1 >= len(args) {
				return "", 0, fmt.Errorf("--host requires a value")
			}
			host = args[i+1]
			i++
		case "--port":
			if i+1 >= len(args) {
				return "", 0, fmt.Errorf("--port requires a value")
			}
			if _, err := fmt.Sscanf(args[i+1], "%d", &port); err != nil || port <= 0 {
				return "", 0, fmt.Errorf("--port must be a positive integer")
			}
			i++
		default:
			return "", 0, fmt.Errorf("unknown flag %q", args[i])
		}
	}
	return host, port, nil
}

func parseOpenAPILoginFlags(args []string) (string, string, string, string, int, time.Duration, bool, error) {
	redirectURI := ""
	scope := openapi.DefaultScopes
	state := fmt.Sprintf("dida-%d-%d", time.Now().Unix(), rand.Intn(100000))
	host := "127.0.0.1"
	port := 17890
	timeout := 10 * time.Minute
	noOpen := false
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--redirect-uri":
			if i+1 >= len(args) {
				return "", "", "", "", 0, 0, false, fmt.Errorf("--redirect-uri requires a value")
			}
			redirectURI = args[i+1]
			i++
		case "--scope":
			if i+1 >= len(args) {
				return "", "", "", "", 0, 0, false, fmt.Errorf("--scope requires a value")
			}
			scope = args[i+1]
			i++
		case "--state":
			if i+1 >= len(args) {
				return "", "", "", "", 0, 0, false, fmt.Errorf("--state requires a value")
			}
			state = args[i+1]
			i++
		case "--host":
			if i+1 >= len(args) {
				return "", "", "", "", 0, 0, false, fmt.Errorf("--host requires a value")
			}
			host = args[i+1]
			i++
		case "--port":
			if i+1 >= len(args) {
				return "", "", "", "", 0, 0, false, fmt.Errorf("--port requires a value")
			}
			if _, err := fmt.Sscanf(args[i+1], "%d", &port); err != nil || port <= 0 {
				return "", "", "", "", 0, 0, false, fmt.Errorf("--port must be a positive integer")
			}
			i++
		case "--timeout":
			if i+1 >= len(args) {
				return "", "", "", "", 0, 0, false, fmt.Errorf("--timeout requires seconds")
			}
			var seconds int
			if _, err := fmt.Sscanf(args[i+1], "%d", &seconds); err != nil || seconds <= 0 {
				return "", "", "", "", 0, 0, false, fmt.Errorf("--timeout must be a positive integer")
			}
			timeout = time.Duration(seconds) * time.Second
			i++
		case "--no-open":
			noOpen = true
		default:
			return "", "", "", "", 0, 0, false, fmt.Errorf("unknown flag %q", args[i])
		}
	}
	return redirectURI, scope, state, host, port, timeout, noOpen, nil
}

func openBrowserURL(target string) error {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "windows":
		cmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", target)
	case "darwin":
		cmd = exec.Command("open", target)
	default:
		cmd = exec.Command("xdg-open", target)
	}
	return cmd.Start()
}
