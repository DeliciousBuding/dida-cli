package cli

import (
	"context"
	"fmt"
	"io"

	"github.com/DeliciousBuding/dida-cli/internal/webapi"
)

func runUser(args []string, jsonOut bool, stdout io.Writer, stderr io.Writer) int {
	if len(args) == 0 || args[0] == "-h" || args[0] == "--help" {
		printUserHelp(stdout)
		return 0
	}
	switch args[0] {
	case "status":
		full, err := parseFullFlag(args[1:])
		if err != nil {
			return failTyped("user status", "validation", err.Error(), "run: dida user --help", jsonOut, stdout, stderr)
		}
		return runUserStatus(full, jsonOut, stdout, stderr)
	case "profile":
		full, err := parseFullFlag(args[1:])
		if err != nil {
			return failTyped("user profile", "validation", err.Error(), "run: dida user --help", jsonOut, stdout, stderr)
		}
		return runUserProfile(full, jsonOut, stdout, stderr)
	case "sessions":
		lang, limit, full, err := parseUserSessionsFlags(args[1:])
		if err != nil {
			return failTyped("user sessions", "validation", err.Error(), "run: dida user --help", jsonOut, stdout, stderr)
		}
		return runUserSessions(lang, limit, full, jsonOut, stdout, stderr)
	default:
		return fail("user", fmt.Sprintf("unknown user command %q", args[0]), jsonOut, stdout, stderr)
	}
}

func runUserStatus(full bool, jsonOut bool, stdout io.Writer, stderr io.Writer) int {
	result, err := executeRead(func(ctx context.Context, client *webapi.Client) (any, error) {
		return client.UserStatus(ctx)
	})
	if err != nil {
		return failTyped("user status", "api", err.Error(), "", jsonOut, stdout, stderr)
	}
	payload := result.(map[string]any)
	if !full {
		payload = compactUserStatus(payload)
	}
	meta := map[string]any{"compact": !full}
	data := map[string]any{"status": payload}
	if jsonOut {
		return writeJSON(stdout, envelope{OK: true, Command: "user status", Meta: meta, Data: data})
	}
	fmt.Fprintf(stdout, "User status: %d keys\n", len(payload))
	return 0
}

func runUserProfile(full bool, jsonOut bool, stdout io.Writer, stderr io.Writer) int {
	result, err := executeRead(func(ctx context.Context, client *webapi.Client) (any, error) {
		return client.UserProfile(ctx)
	})
	if err != nil {
		return failTyped("user profile", "api", err.Error(), "", jsonOut, stdout, stderr)
	}
	payload := result.(map[string]any)
	if !full {
		payload = compactUserProfile(payload)
	}
	meta := map[string]any{"compact": !full}
	data := map[string]any{"profile": payload}
	if jsonOut {
		return writeJSON(stdout, envelope{OK: true, Command: "user profile", Meta: meta, Data: data})
	}
	fmt.Fprintf(stdout, "User profile: %d keys\n", len(payload))
	return 0
}

func runUserSessions(lang string, limit int, full bool, jsonOut bool, stdout io.Writer, stderr io.Writer) int {
	result, err := executeRead(func(ctx context.Context, client *webapi.Client) (any, error) {
		return client.UserSessions(ctx, lang)
	})
	if err != nil {
		return failTyped("user sessions", "api", err.Error(), "", jsonOut, stdout, stderr)
	}
	items := result.([]map[string]any)
	total := len(items)
	if !full {
		out := make([]map[string]any, 0, len(items))
		for _, item := range items {
			out = append(out, compactUserSession(item))
		}
		items = out
	}
	if limit > 0 && len(items) > limit {
		items = items[:limit]
	}
	meta := map[string]any{"compact": !full, "count": len(items), "total": total, "limit": limit, "lang": lang}
	data := map[string]any{"sessions": items}
	if jsonOut {
		return writeJSON(stdout, envelope{OK: true, Command: "user sessions", Meta: meta, Data: data})
	}
	printMapList(stdout, items, "user sessions")
	return 0
}

func parseFullFlag(args []string) (bool, error) {
	full := false
	for _, arg := range args {
		if arg == "--full" {
			full = true
			continue
		}
		return false, fmt.Errorf("unknown flag %q", arg)
	}
	return full, nil
}

func parseUserSessionsFlags(args []string) (string, int, bool, error) {
	lang := "zh_CN"
	limit := 20
	full := false
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--lang":
			if i+1 >= len(args) {
				return "", 0, false, fmt.Errorf("--lang requires a locale")
			}
			lang = args[i+1]
			i++
		case "--limit":
			if i+1 >= len(args) {
				return "", 0, false, fmt.Errorf("--limit requires a value")
			}
			parsed, err := parseIntStrict(args[i+1])
			if err != nil || parsed < 0 {
				return "", 0, false, fmt.Errorf("--limit must be a non-negative integer")
			}
			limit = parsed
			i++
		case "--full":
			full = true
		default:
			return "", 0, false, fmt.Errorf("unknown flag %q", args[i])
		}
	}
	return lang, limit, full, nil
}

func compactUserStatus(item map[string]any) map[string]any {
	return pickKeys(item, "userId", "inboxId", "teamPro", "teamUser", "activeTeamUser", "freeTrial", "pro", "ds", "needSubscribe", "registerDate", "proEndDate")
}

func compactUserProfile(item map[string]any) map[string]any {
	return pickKeys(item, "name", "displayName", "siteDomain", "accountDomain", "locale", "verifiedEmail", "fakedEmail", "filledPassword", "picture")
}

func compactUserSession(item map[string]any) map[string]any {
	out := pickKeys(item, "id", "createdTime", "modifiedTime")
	if deviceInfo, ok := item["deviceInfo"].(map[string]any); ok {
		out["deviceInfo"] = pickKeys(deviceInfo, "platform", "os", "device", "version", "channel", "name")
	}
	return out
}
