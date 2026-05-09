package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/DeliciousBuding/dida-cli/internal/auth"
	"github.com/DeliciousBuding/dida-cli/internal/config"
	"github.com/DeliciousBuding/dida-cli/internal/model"
	"github.com/DeliciousBuding/dida-cli/internal/webapi"
)

type envelope struct {
	OK      bool      `json:"ok"`
	Command string    `json:"command"`
	Meta    any       `json:"meta,omitempty"`
	Data    any       `json:"data,omitempty"`
	Error   *cliError `json:"error,omitempty"`
}

type cliError struct {
	Type    string `json:"type,omitempty"`
	Message string `json:"message"`
	Hint    string `json:"hint,omitempty"`
}

func Run(args []string, version string, stdout io.Writer, stderr io.Writer) int {
	if len(args) == 0 || args[0] == "-h" || args[0] == "--help" {
		printHelp(stdout)
		return 0
	}
	if args[0] == "--version" || args[0] == "version" {
		fmt.Fprintln(stdout, version)
		return 0
	}

	jsonOut, args := consumeJSONFlag(args)
	command := args[0]

	switch command {
	case "+today":
		return runTask(append([]string{"today"}, args[1:]...), jsonOut, stdout, stderr)
	case "doctor":
		return runDoctor(args[1:], version, jsonOut, stdout, stderr)
	case "auth":
		return runAuth(args[1:], jsonOut, stdout, stderr)
	case "sync":
		return runSync(args[1:], jsonOut, stdout, stderr)
	case "settings":
		return runSettings(args[1:], jsonOut, stdout, stderr)
	case "completed":
		return runCompleted(args[1:], jsonOut, stdout, stderr)
	case "raw":
		return runRaw(args[1:], jsonOut, stdout, stderr)
	case "project":
		return runProject(args[1:], jsonOut, stdout, stderr)
	case "task":
		return runTask(args[1:], jsonOut, stdout, stderr)
	case "report":
		return notImplemented(command, jsonOut, stdout, stderr)
	default:
		return fail(command, fmt.Sprintf("unknown command %q", command), jsonOut, stdout, stderr)
	}
}

func consumeJSONFlag(args []string) (bool, []string) {
	out := args[:0]
	jsonOut := false
	for _, arg := range args {
		if arg == "--json" || arg == "-j" {
			jsonOut = true
			continue
		}
		out = append(out, arg)
	}
	return jsonOut, out
}

func runDoctor(args []string, version string, jsonOut bool, stdout io.Writer, stderr io.Writer) int {
	if len(args) > 0 && (args[0] == "-h" || args[0] == "--help") {
		fmt.Fprintln(stdout, "Usage: dida doctor [--json]")
		return 0
	}

	cfgDir := config.DefaultDir()
	cookiePath := filepath.Join(cfgDir, "cookie.json")
	oauthPath := filepath.Join(cfgDir, "oauth.json")
	cookieExists := fileExists(cookiePath)
	oauthExists := fileExists(oauthPath)

	data := map[string]any{
		"version":       version,
		"goos":          runtime.GOOS,
		"goarch":        runtime.GOARCH,
		"config_dir":    cfgDir,
		"auth_sources":  map[string]bool{"cookie": cookieExists, "oauth": oauthExists},
		"cookie_status": auth.CookieStatus(),
		"network_check": "not_run",
	}

	if jsonOut {
		return writeJSON(stdout, envelope{OK: true, Command: "doctor", Data: data})
	}

	fmt.Fprintf(stdout, "DidaCLI %s\n", version)
	fmt.Fprintf(stdout, "Config: %s\n", cfgDir)
	fmt.Fprintf(stdout, "Cookie auth: %s\n", yesNo(cookieExists))
	fmt.Fprintf(stdout, "OAuth auth: %s\n", yesNo(oauthExists))
	fmt.Fprintln(stdout, "Network check: not run")
	return 0
}

func runAuth(args []string, jsonOut bool, stdout io.Writer, stderr io.Writer) int {
	if len(args) == 0 || args[0] == "-h" || args[0] == "--help" {
		printAuthHelp(stdout)
		return 0
	}
	switch args[0] {
	case "status":
		verify := hasFlag(args[1:], "--verify")
		data := map[string]any{"cookie": auth.CookieStatus(), "oauth": map[string]any{"available": false, "message": "not implemented"}}
		if verify {
			data["verify"] = verifyCookieAuth()
		}
		if jsonOut {
			return writeJSON(stdout, envelope{OK: true, Command: "auth status", Data: data})
		}
		cookie := data["cookie"].(map[string]any)
		fmt.Fprintf(stdout, "Cookie auth: %v\n", cookie["available"])
		fmt.Fprintf(stdout, "Cookie path: %v\n", cookie["path"])
		if cookie["available"] == true {
			fmt.Fprintf(stdout, "Token: %v\n", cookie["token_preview"])
			fmt.Fprintf(stdout, "Saved at: %v\n", cookie["saved_at"])
		}
		if verify {
			fmt.Fprintf(stdout, "Verify: %v\n", data["verify"])
		}
		return 0
	case "login":
		return runAuthLogin(args[1:], jsonOut, stdout, stderr)
	case "logout":
		return runAuthLogout(args[1:], jsonOut, stdout, stderr)
	case "cookie":
		return runAuthCookie(args[1:], jsonOut, stdout, stderr)
	default:
		return fail("auth", fmt.Sprintf("unknown auth command %q", args[0]), jsonOut, stdout, stderr)
	}
}

func runAuthLogin(args []string, jsonOut bool, stdout io.Writer, stderr io.Writer) int {
	if len(args) > 0 && (args[0] == "-h" || args[0] == "--help") {
		printAuthLoginHelp(stdout)
		return 0
	}
	if hasFlag(args, "--browser") {
		return runAuthLoginBrowser(args, jsonOut, stdout, stderr)
	}
	data := map[string]any{
		"mode":             "manual_cookie",
		"login_url":        "https://dida365.com/signin",
		"cookie_name":      "t",
		"recommended_next": "dida auth cookie set --token-stdin",
		"agent_hint":       "Ask the user to sign in in a browser, copy only the Dida365 cookie named 't', then paste it to stdin for `dida auth cookie set --token-stdin`. Do not ask the user to paste cookies into chat.",
		"wechat_hint":      "If the website shows WeChat or QR login, complete it in the browser first; the CLI only stores the resulting 't' cookie.",
	}
	if jsonOut {
		return writeJSON(stdout, envelope{OK: true, Command: "auth login", Data: data})
	}
	fmt.Fprintln(stdout, "Open Dida365 login in your browser:")
	fmt.Fprintln(stdout, "  https://dida365.com/signin")
	fmt.Fprintln(stdout)
	fmt.Fprintln(stdout, "After login, copy only the cookie named 't' and import it with:")
	fmt.Fprintln(stdout, "  dida auth cookie set --token-stdin")
	fmt.Fprintln(stdout)
	fmt.Fprintln(stdout, "If WeChat/QR login appears, finish it in the browser first. The CLI stores only the resulting 't' cookie.")
	return 0
}

func runAuthLoginBrowser(args []string, jsonOut bool, stdout io.Writer, stderr io.Writer) int {
	timeout := 180 * time.Second
	for i := 0; i < len(args); i++ {
		if args[i] == "--timeout" {
			if i+1 >= len(args) {
				return failTyped("auth login", "validation", "--timeout requires seconds", "example: dida auth login --browser --timeout 300", jsonOut, stdout, stderr)
			}
			var seconds int
			if _, err := fmt.Sscanf(args[i+1], "%d", &seconds); err != nil || seconds <= 0 {
				return failTyped("auth login", "validation", "--timeout must be a positive integer", "example: dida auth login --browser --timeout 300", jsonOut, stdout, stderr)
			}
			timeout = time.Duration(seconds) * time.Second
			i++
		}
	}
	if !jsonOut {
		fmt.Fprintln(stderr, "Opening Dida365 login in a local browser. Complete password/WeChat/QR login there.")
		fmt.Fprintln(stderr, "Waiting for cookie 't'...")
	}
	ctx, cancel := context.WithTimeout(context.Background(), timeout+15*time.Second)
	defer cancel()
	result, err := auth.CaptureCookieWithBrowser(ctx, timeout)
	if err != nil {
		return failTyped("auth login", "auth", err.Error(), "fallback: dida auth cookie set --token-stdin", jsonOut, stdout, stderr)
	}
	item, err := auth.SaveCookieToken(result.Token)
	if err != nil {
		return failTyped("auth login", "auth", err.Error(), "", jsonOut, stdout, stderr)
	}
	data := map[string]any{
		"mode":          "browser_cookie",
		"cookie_saved":  true,
		"path":          auth.CookiePath(),
		"domain":        result.Domain,
		"saved_at":      time.UnixMilli(item.SavedAt).Format(time.RFC3339),
		"token_length":  len(item.Token),
		"token_preview": auth.RedactToken(item.Token),
		"next":          "dida auth status --verify --json",
	}
	if jsonOut {
		return writeJSON(stdout, envelope{OK: true, Command: "auth login", Data: data})
	}
	fmt.Fprintf(stdout, "Cookie token saved: %s\n", auth.CookiePath())
	fmt.Fprintf(stdout, "Token: %s\n", auth.RedactToken(item.Token))
	fmt.Fprintln(stdout, "Next: dida auth status --verify --json")
	return 0
}

func runAuthLogout(args []string, jsonOut bool, stdout io.Writer, stderr io.Writer) int {
	if len(args) > 0 && (args[0] == "-h" || args[0] == "--help") {
		fmt.Fprintln(stdout, "Usage: dida auth logout [--json]")
		return 0
	}
	if err := auth.ClearCookieToken(); err != nil {
		return failTyped("auth logout", "auth", err.Error(), "", jsonOut, stdout, stderr)
	}
	data := map[string]any{"cookie_cleared": true, "path": auth.CookiePath()}
	if jsonOut {
		return writeJSON(stdout, envelope{OK: true, Command: "auth logout", Data: data})
	}
	fmt.Fprintf(stdout, "Cookie auth cleared: %s\n", auth.CookiePath())
	return 0
}

func runAuthCookie(args []string, jsonOut bool, stdout io.Writer, stderr io.Writer) int {
	if len(args) == 0 || args[0] == "-h" || args[0] == "--help" {
		printAuthCookieHelp(stdout)
		return 0
	}
	if args[0] != "set" {
		return fail("auth cookie", fmt.Sprintf("unknown auth cookie command %q", args[0]), jsonOut, stdout, stderr)
	}
	token, err := parseTokenInput(args[1:])
	if err != nil {
		return fail("auth cookie set", err.Error(), jsonOut, stdout, stderr)
	}
	item, err := auth.SaveCookieToken(token)
	if err != nil {
		return fail("auth cookie set", err.Error(), jsonOut, stdout, stderr)
	}
	data := map[string]any{
		"path":          auth.CookiePath(),
		"saved_at":      time.UnixMilli(item.SavedAt).Format(time.RFC3339),
		"token_length":  len(item.Token),
		"token_preview": auth.RedactToken(item.Token),
	}
	if jsonOut {
		return writeJSON(stdout, envelope{OK: true, Command: "auth cookie set", Data: data})
	}
	fmt.Fprintf(stdout, "Cookie token saved: %s\n", auth.CookiePath())
	fmt.Fprintf(stdout, "Token: %s\n", auth.RedactToken(item.Token))
	return 0
}

func parseTokenInput(args []string) (string, error) {
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--token":
			if i+1 >= len(args) {
				return "", fmt.Errorf("--token requires a value")
			}
			return args[i+1], nil
		case "--token-stdin":
			data, err := io.ReadAll(os.Stdin)
			if err != nil {
				return "", fmt.Errorf("read token from stdin: %w", err)
			}
			return string(data), nil
		}
	}
	return "", fmt.Errorf("missing token; use --token-stdin to avoid shell history")
}

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
	if jsonOut {
		return writeJSON(stdout, envelope{OK: true, Command: "sync checkpoint", Meta: meta, Data: data})
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

func runCompleted(args []string, jsonOut bool, stdout io.Writer, stderr io.Writer) int {
	if len(args) == 0 || args[0] == "-h" || args[0] == "--help" {
		printCompletedHelp(stdout)
		return 0
	}
	now := time.Now()
	limit := 100
	var from, to time.Time
	command := "completed " + args[0]
	switch args[0] {
	case "today":
		from, to = dayRange(now)
	case "yesterday":
		from, to = dayRange(now.AddDate(0, 0, -1))
	case "week":
		start := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location()).AddDate(0, 0, -int(now.Weekday()))
		from = start
		to = start.AddDate(0, 0, 7).Add(-time.Second)
	case "list":
		parsedFrom, parsedTo, parsedLimit, err := parseCompletedListFlags(args[1:], now)
		if err != nil {
			return failTyped("completed list", "validation", err.Error(), "run: dida completed --help", jsonOut, stdout, stderr)
		}
		from, to, limit = parsedFrom, parsedTo, parsedLimit
	default:
		return fail("completed", fmt.Sprintf("unknown completed command %q", args[0]), jsonOut, stdout, stderr)
	}
	tasks, err := loadCompletedTasks(from, to, limit)
	if err != nil {
		return failTyped(command, "auth", err.Error(), "run: dida auth login", jsonOut, stdout, stderr)
	}
	view, _ := loadSyncView()
	projectNames := map[string]string{}
	for _, project := range view.Projects {
		projectNames[project.ID] = project.Name
	}
	normalized := model.NormalizeTasks(tasks, projectNames, now)
	data := map[string]any{
		"from":  formatDidaQueryTime(from),
		"to":    formatDidaQueryTime(to),
		"tasks": stripTaskRaw(normalized),
	}
	meta := map[string]any{"count": len(normalized), "limit": limit}
	if jsonOut {
		return writeJSON(stdout, envelope{OK: true, Command: command, Meta: meta, Data: data})
	}
	printTasks(stdout, normalized, len(normalized))
	return 0
}

func runProject(args []string, jsonOut bool, stdout io.Writer, stderr io.Writer) int {
	if len(args) == 0 || args[0] == "-h" || args[0] == "--help" {
		printProjectHelp(stdout)
		return 0
	}
	switch args[0] {
	case "list":
		return runProjectList(jsonOut, stdout, stderr)
	case "tasks":
		if len(args) != 2 {
			return failTyped("project tasks", "validation", "usage: dida project tasks <project-id>", "run: dida project --help", jsonOut, stdout, stderr)
		}
		return runProjectTasks(args[1], jsonOut, stdout, stderr)
	case "columns":
		if len(args) != 2 {
			return failTyped("project columns", "validation", "usage: dida project columns <project-id>", "run: dida project --help", jsonOut, stdout, stderr)
		}
		return runProjectColumns(args[1], jsonOut, stdout, stderr)
	default:
		return fail("project", fmt.Sprintf("unknown project command %q", args[0]), jsonOut, stdout, stderr)
	}
}

func runProjectList(jsonOut bool, stdout io.Writer, stderr io.Writer) int {
	view, err := loadSyncView()
	if err != nil {
		return failTyped("project list", "auth", err.Error(), "run: dida auth login", jsonOut, stdout, stderr)
	}
	projects := stripProjectRaw(view.Projects)
	data := map[string]any{"projects": projects}
	meta := map[string]any{"count": len(projects)}
	if jsonOut {
		return writeJSON(stdout, envelope{OK: true, Command: "project list", Meta: meta, Data: data})
	}
	printProjects(stdout, view.Projects)
	return 0
}

func runProjectTasks(projectID string, jsonOut bool, stdout io.Writer, stderr io.Writer) int {
	view, err := loadSyncView()
	if err != nil {
		return failTyped("project tasks", "auth", err.Error(), "run: dida auth login", jsonOut, stdout, stderr)
	}
	tasks := model.ProjectTasks(view.Tasks, projectID)
	tasks = model.ActiveTasks(tasks)
	data := map[string]any{"projectId": projectID, "tasks": stripTaskRaw(tasks)}
	meta := map[string]any{"count": len(tasks)}
	if jsonOut {
		return writeJSON(stdout, envelope{OK: true, Command: "project tasks", Meta: meta, Data: data})
	}
	printTasks(stdout, tasks, len(tasks))
	return 0
}

func runProjectColumns(projectID string, jsonOut bool, stdout io.Writer, stderr io.Writer) int {
	view, err := loadSyncView()
	if err != nil {
		return failTyped("project columns", "auth", err.Error(), "run: dida auth login", jsonOut, stdout, stderr)
	}
	columns := model.InferColumns(projectID, view.Tasks)
	data := map[string]any{"projectId": projectID, "columns": columns}
	meta := map[string]any{"count": len(columns), "source": "inferred_from_tasks"}
	if jsonOut {
		return writeJSON(stdout, envelope{OK: true, Command: "project columns", Meta: meta, Data: data})
	}
	if len(columns) == 0 {
		fmt.Fprintln(stdout, "No columns found.")
		return 0
	}
	fmt.Fprintf(stdout, "%-28s  %-8s\n", "ID", "TASKS")
	for _, column := range columns {
		fmt.Fprintf(stdout, "%-28s  %-8d\n", column.ID, column.TaskCount)
	}
	return 0
}

func runTask(args []string, jsonOut bool, stdout io.Writer, stderr io.Writer) int {
	if len(args) == 0 || args[0] == "-h" || args[0] == "--help" {
		printTaskHelp(stdout)
		return 0
	}
	switch args[0] {
	case "today":
		return runTaskList(append([]string{"list", "--filter", "today"}, args[1:]...), jsonOut, stdout, stderr)
	case "list":
		return runTaskList(args, jsonOut, stdout, stderr)
	case "search":
		return runTaskSearch(args[1:], jsonOut, stdout, stderr)
	case "upcoming":
		return runTaskUpcoming(args[1:], jsonOut, stdout, stderr)
	case "get":
		if len(args) != 2 {
			return failTyped("task get", "validation", "usage: dida task get <task-id>", "run: dida task --help", jsonOut, stdout, stderr)
		}
		return runTaskGet(args[1], jsonOut, stdout, stderr)
	default:
		return fail("task", fmt.Sprintf("unknown task command %q", args[0]), jsonOut, stdout, stderr)
	}
}

func runTaskList(args []string, jsonOut bool, stdout io.Writer, stderr io.Writer) int {
	filter, limit, err := parseTaskListFlags(args[1:])
	if err != nil {
		return failTyped("task list", "validation", err.Error(), "run: dida task list --help", jsonOut, stdout, stderr)
	}
	view, err := loadSyncView()
	if err != nil {
		return failTyped("task list", "auth", err.Error(), "run: dida auth login", jsonOut, stdout, stderr)
	}
	now := time.Now()
	var tasks []model.Task
	switch filter {
	case "all":
		tasks = model.ActiveTasks(view.Tasks)
	case "today":
		tasks = model.TodayTasks(view.Tasks, now)
	default:
		return failTyped("task list", "validation", "unknown filter; supported filters: today, all", "run: dida task list --help", jsonOut, stdout, stderr)
	}
	total := len(tasks)
	if limit > 0 && len(tasks) > limit {
		tasks = tasks[:limit]
	}
	data := map[string]any{
		"filter": filter,
		"tasks":  stripTaskRaw(tasks),
	}
	meta := map[string]any{"count": len(tasks), "total": total}
	command := "task list"
	if filter == "today" && len(args) > 0 && args[0] != "list" {
		command = "task today"
	}
	if jsonOut {
		return writeJSON(stdout, envelope{OK: true, Command: command, Meta: meta, Data: data})
	}
	printTasks(stdout, tasks, total)
	return 0
}

func runTaskSearch(args []string, jsonOut bool, stdout io.Writer, stderr io.Writer) int {
	query, limit, err := parseSearchFlags(args)
	if err != nil {
		return failTyped("task search", "validation", err.Error(), "run: dida task --help", jsonOut, stdout, stderr)
	}
	view, err := loadSyncView()
	if err != nil {
		return failTyped("task search", "auth", err.Error(), "run: dida auth login", jsonOut, stdout, stderr)
	}
	tasks := model.SearchTasks(model.ActiveTasks(view.Tasks), query)
	total := len(tasks)
	if limit > 0 && len(tasks) > limit {
		tasks = tasks[:limit]
	}
	data := map[string]any{"query": query, "tasks": stripTaskRaw(tasks)}
	meta := map[string]any{"count": len(tasks), "total": total}
	if jsonOut {
		return writeJSON(stdout, envelope{OK: true, Command: "task search", Meta: meta, Data: data})
	}
	printTasks(stdout, tasks, total)
	return 0
}

func runTaskUpcoming(args []string, jsonOut bool, stdout io.Writer, stderr io.Writer) int {
	days, limit, err := parseUpcomingFlags(args)
	if err != nil {
		return failTyped("task upcoming", "validation", err.Error(), "run: dida task --help", jsonOut, stdout, stderr)
	}
	view, err := loadSyncView()
	if err != nil {
		return failTyped("task upcoming", "auth", err.Error(), "run: dida auth login", jsonOut, stdout, stderr)
	}
	tasks := model.UpcomingTasks(view.Tasks, time.Now(), days)
	total := len(tasks)
	if limit > 0 && len(tasks) > limit {
		tasks = tasks[:limit]
	}
	data := map[string]any{"days": days, "tasks": stripTaskRaw(tasks)}
	meta := map[string]any{"count": len(tasks), "total": total}
	if jsonOut {
		return writeJSON(stdout, envelope{OK: true, Command: "task upcoming", Meta: meta, Data: data})
	}
	printTasks(stdout, tasks, total)
	return 0
}

func runTaskGet(taskID string, jsonOut bool, stdout io.Writer, stderr io.Writer) int {
	view, err := loadSyncView()
	if err != nil {
		return failTyped("task get", "auth", err.Error(), "run: dida auth login", jsonOut, stdout, stderr)
	}
	task, ok := model.FindTask(view.Tasks, taskID)
	if !ok {
		return failTyped("task get", "not_found", "task not found", "run: dida task list --filter all --json", jsonOut, stdout, stderr)
	}
	if jsonOut {
		return writeJSON(stdout, envelope{OK: true, Command: "task get", Data: map[string]any{"task": stripSingleTaskRaw(task)}})
	}
	printTasks(stdout, []model.Task{task}, 1)
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

func parseTaskListFlags(args []string) (string, int, error) {
	filter := "all"
	limit := 50
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--filter":
			if i+1 >= len(args) {
				return "", 0, fmt.Errorf("--filter requires a value")
			}
			filter = args[i+1]
			i++
		case "--limit":
			if i+1 >= len(args) {
				return "", 0, fmt.Errorf("--limit requires a value")
			}
			var parsed int
			if _, err := fmt.Sscanf(args[i+1], "%d", &parsed); err != nil {
				return "", 0, fmt.Errorf("--limit must be an integer")
			}
			limit = parsed
			i++
		default:
			return "", 0, fmt.Errorf("unknown flag %q", args[i])
		}
	}
	return filter, limit, nil
}

func parseSearchFlags(args []string) (string, int, error) {
	query := ""
	limit := 50
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--query", "-q":
			if i+1 >= len(args) {
				return "", 0, fmt.Errorf("%s requires a value", args[i])
			}
			query = args[i+1]
			i++
		case "--limit":
			if i+1 >= len(args) {
				return "", 0, fmt.Errorf("--limit requires a value")
			}
			if _, err := fmt.Sscanf(args[i+1], "%d", &limit); err != nil || limit < 0 {
				return "", 0, fmt.Errorf("--limit must be a non-negative integer")
			}
			i++
		default:
			if query == "" {
				query = args[i]
				continue
			}
			return "", 0, fmt.Errorf("unknown flag %q", args[i])
		}
	}
	if strings.TrimSpace(query) == "" {
		return "", 0, fmt.Errorf("missing query; use: dida task search --query <text>")
	}
	return query, limit, nil
}

func parseUpcomingFlags(args []string) (int, int, error) {
	days := 7
	limit := 50
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--days":
			if i+1 >= len(args) {
				return 0, 0, fmt.Errorf("--days requires a value")
			}
			if _, err := fmt.Sscanf(args[i+1], "%d", &days); err != nil || days <= 0 {
				return 0, 0, fmt.Errorf("--days must be a positive integer")
			}
			i++
		case "--limit":
			if i+1 >= len(args) {
				return 0, 0, fmt.Errorf("--limit requires a value")
			}
			if _, err := fmt.Sscanf(args[i+1], "%d", &limit); err != nil || limit < 0 {
				return 0, 0, fmt.Errorf("--limit must be a non-negative integer")
			}
			i++
		default:
			return 0, 0, fmt.Errorf("unknown flag %q", args[i])
		}
	}
	return days, limit, nil
}

func parseCompletedListFlags(args []string, now time.Time) (time.Time, time.Time, int, error) {
	from, to := dayRange(now)
	limit := 100
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--from":
			if i+1 >= len(args) {
				return time.Time{}, time.Time{}, 0, fmt.Errorf("--from requires YYYY-MM-DD")
			}
			parsed, err := time.ParseInLocation("2006-01-02", args[i+1], now.Location())
			if err != nil {
				return time.Time{}, time.Time{}, 0, fmt.Errorf("--from must be YYYY-MM-DD")
			}
			from = parsed
			i++
		case "--to":
			if i+1 >= len(args) {
				return time.Time{}, time.Time{}, 0, fmt.Errorf("--to requires YYYY-MM-DD")
			}
			parsed, err := time.ParseInLocation("2006-01-02", args[i+1], now.Location())
			if err != nil {
				return time.Time{}, time.Time{}, 0, fmt.Errorf("--to must be YYYY-MM-DD")
			}
			to = parsed.AddDate(0, 0, 1).Add(-time.Second)
			i++
		case "--limit":
			if i+1 >= len(args) {
				return time.Time{}, time.Time{}, 0, fmt.Errorf("--limit requires a value")
			}
			if _, err := fmt.Sscanf(args[i+1], "%d", &limit); err != nil || limit <= 0 {
				return time.Time{}, time.Time{}, 0, fmt.Errorf("--limit must be a positive integer")
			}
			i++
		default:
			return time.Time{}, time.Time{}, 0, fmt.Errorf("unknown flag %q", args[i])
		}
	}
	if from.After(to) {
		return time.Time{}, time.Time{}, 0, fmt.Errorf("--from must be before or equal to --to")
	}
	return from, to, limit, nil
}

func loadCompletedTasks(from time.Time, to time.Time, limit int) ([]map[string]any, error) {
	token, err := auth.LoadCookieToken()
	if err != nil {
		return nil, fmt.Errorf("missing cookie auth; run: dida auth cookie set --token-stdin")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	return webapi.NewClient(token.Token).CompletedTasks(ctx, formatDidaQueryTime(from), formatDidaQueryTime(to), limit)
}

func dayRange(t time.Time) (time.Time, time.Time) {
	start := time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
	return start, start.AddDate(0, 0, 1).Add(-time.Second)
}

func formatDidaQueryTime(t time.Time) string {
	return t.Format("2006-01-02 15:04:05")
}

func runRaw(args []string, jsonOut bool, stdout io.Writer, stderr io.Writer) int {
	if len(args) == 0 || args[0] == "-h" || args[0] == "--help" {
		printRawHelp(stdout)
		return 0
	}
	if args[0] != "get" || len(args) != 2 {
		return fail("raw", "usage: dida raw get <path>", jsonOut, stdout, stderr)
	}
	token, err := auth.LoadCookieToken()
	if err != nil {
		return missingAuth("raw get", jsonOut, stdout, stderr)
	}
	var data any
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	if err := webapi.NewClient(token.Token).Do(ctx, "GET", args[1], nil, &data); err != nil {
		return fail("raw get", err.Error(), jsonOut, stdout, stderr)
	}
	if jsonOut {
		return writeJSON(stdout, envelope{OK: true, Command: "raw get", Data: data})
	}
	return writeJSON(stdout, data)
}

func notImplemented(command string, jsonOut bool, stdout io.Writer, stderr io.Writer) int {
	return fail(command, "command scaffolded but not implemented yet", jsonOut, stdout, stderr)
}

func fail(command string, message string, jsonOut bool, stdout io.Writer, stderr io.Writer) int {
	return failTyped(command, "", message, "", jsonOut, stdout, stderr)
}

func missingAuth(command string, jsonOut bool, stdout io.Writer, stderr io.Writer) int {
	return failTyped(command, "auth", "missing cookie auth", "run: dida auth login", jsonOut, stdout, stderr)
}

func failTyped(command string, errType string, message string, hint string, jsonOut bool, stdout io.Writer, stderr io.Writer) int {
	if jsonOut {
		_ = writeJSON(stdout, envelope{OK: false, Command: command, Error: &cliError{Type: errType, Message: message, Hint: hint}})
		return 1
	}
	fmt.Fprintf(stderr, "dida: %s\n", message)
	if hint != "" {
		fmt.Fprintf(stderr, "hint: %s\n", hint)
	}
	return 1
}

func verifyCookieAuth() map[string]any {
	token, err := auth.LoadCookieToken()
	if err != nil {
		return map[string]any{"ok": false, "message": "missing cookie auth", "hint": "run: dida auth login"}
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	payload, err := webapi.NewClient(token.Token).FullSync(ctx)
	if err != nil {
		return map[string]any{"ok": false, "message": err.Error(), "hint": "refresh the Dida365 't' cookie with: dida auth login"}
	}
	return map[string]any{
		"ok":       true,
		"projects": len(payload.Projects),
		"tasks":    len(payload.Tasks),
	}
}

func hasFlag(args []string, flag string) bool {
	for _, arg := range args {
		if arg == flag {
			return true
		}
	}
	return false
}

func writeJSON(w io.Writer, value any) int {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	if err := enc.Encode(value); err != nil {
		return 1
	}
	return 0
}

func fileExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && !info.IsDir()
}

func yesNo(v bool) string {
	if v {
		return "yes"
	}
	return "no"
}

func printHelp(w io.Writer) {
	fmt.Fprintln(w, strings.TrimSpace(`
DidaCLI - Dida365 / TickTick command line client

Usage:
  dida <command> [options]

Commands:
  doctor       Check local config and auth status
  auth         Manage OAuth and cookie auth
  sync         Sync tasks/projects/tags
  settings     Read user preferences
  completed    Read completed task history
  project      Project discovery
  task         Task reads and writes
  report       Generate reports
  raw          Raw read-only API escape hatch
  version      Print version
  +today       Shortcut for task today

Global options:
  -j, --json   Emit machine-readable JSON
  -h, --help   Show help
`))
}

func printAuthHelp(w io.Writer) {
	fmt.Fprintln(w, strings.TrimSpace(`
Usage:
  dida auth login [--json]
  dida auth status [--json]
  dida auth status --verify [--json]
  dida auth logout [--json]
  dida auth cookie set --token-stdin
  dida auth cookie set --token <token>
`))
}

func printAuthLoginHelp(w io.Writer) {
	fmt.Fprintln(w, strings.TrimSpace(`
Usage:
  dida auth login [--json]
  dida auth login --browser [--timeout 180] [--json]

This prints a browser login guide. Complete Dida365/WeChat/QR login in the browser,
then import only the resulting cookie named 't' with:
  dida auth cookie set --token-stdin

With --browser, the CLI opens a visible browser, waits for cookie 't', and saves it automatically.
`))
}

func printAuthCookieHelp(w io.Writer) {
	fmt.Fprintln(w, strings.TrimSpace(`
Usage:
  dida auth cookie set --token-stdin
  dida auth cookie set --token <token>

Prefer --token-stdin to avoid shell history.
`))
}

func printSyncHelp(w io.Writer) {
	fmt.Fprintln(w, strings.TrimSpace(`
Usage:
  dida sync all [--json]
  dida sync checkpoint <checkpoint> [--json]
`))
}

func printSettingsHelp(w io.Writer) {
	fmt.Fprintln(w, strings.TrimSpace(`
Usage:
  dida settings get [--json]
`))
}

func printCompletedHelp(w io.Writer) {
	fmt.Fprintln(w, strings.TrimSpace(`
Usage:
  dida completed today [--json]
  dida completed yesterday [--json]
  dida completed week [--json]
  dida completed list [--from YYYY-MM-DD] [--to YYYY-MM-DD] [--limit N] [--json]
`))
}

func printProjectHelp(w io.Writer) {
	fmt.Fprintln(w, strings.TrimSpace(`
Usage:
  dida project list [--json]
  dida project tasks <project-id> [--json]
  dida project columns <project-id> [--json]
`))
}

func printTaskHelp(w io.Writer) {
	fmt.Fprintln(w, strings.TrimSpace(`
Usage:
  dida task today [--json] [--limit N]
  dida task list [--json] [--filter today|all] [--limit N]
  dida task search --query <text> [--limit N] [--json]
  dida task upcoming [--days N] [--limit N] [--json]
  dida task get <task-id> [--json]
  dida +today [--json] [--limit N]
`))
}

func printRawHelp(w io.Writer) {
	fmt.Fprintln(w, strings.TrimSpace(`
Usage:
  dida raw get <path> [--json]

Only GET is supported for raw calls.
`))
}

func printProjects(w io.Writer, projects []model.Project) {
	if len(projects) == 0 {
		fmt.Fprintln(w, "No projects found.")
		return
	}
	fmt.Fprintf(w, "%-28s  %s\n", "ID", "NAME")
	for _, project := range projects {
		fmt.Fprintf(w, "%-28s  %s\n", project.ID, project.Name)
	}
}

func printTasks(w io.Writer, tasks []model.Task, total int) {
	if len(tasks) == 0 {
		fmt.Fprintln(w, "No tasks found.")
		return
	}
	fmt.Fprintf(w, "Showing %d of %d task(s)\n", len(tasks), total)
	fmt.Fprintf(w, "%-28s  %-10s  %-8s  %-16s  %s\n", "ID", "PROJECT", "PRIORITY", "DUE", "TITLE")
	for _, task := range tasks {
		due := "-"
		if task.DueDate != "" {
			due = task.DueDate
		}
		project := task.ProjectName
		if project == "" {
			project = task.ProjectID
		}
		if len(project) > 10 {
			project = project[:10]
		}
		fmt.Fprintf(w, "%-28s  %-10s  %-8d  %-16s  %s\n", task.ID, project, task.Priority, due, task.Title)
	}
}

func stripProjectRaw(projects []model.Project) []model.Project {
	out := make([]model.Project, len(projects))
	copy(out, projects)
	for i := range out {
		out[i].Raw = nil
	}
	return out
}

func stripTaskRaw(tasks []model.Task) []model.Task {
	out := make([]model.Task, len(tasks))
	copy(out, tasks)
	for i := range out {
		out[i].Raw = nil
	}
	return out
}

func stripSingleTaskRaw(task model.Task) model.Task {
	task.Raw = nil
	return task
}
