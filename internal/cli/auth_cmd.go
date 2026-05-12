package cli

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/DeliciousBuding/dida-cli/internal/auth"
	"github.com/DeliciousBuding/dida-cli/internal/webapi"
)

const maxTokenStdinBytes int64 = 64 << 10

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
			verifyResult := verifyCookieAuth()
			data["verify"] = verifyResult
			if verifyResult["ok"] != true {
				return failTyped("auth status", "auth", fmt.Sprint(verifyResult["message"]), fmt.Sprint(verifyResult["hint"]), jsonOut, stdout, stderr)
			}
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
			if os.Getenv("DIDA_ALLOW_TOKEN_ARG") != "1" {
				return "", fmt.Errorf("--token is disabled by default because it can leak cookies into shell history; use --token-stdin or set DIDA_ALLOW_TOKEN_ARG=1 for a one-off local test")
			}
			if i+1 >= len(args) {
				return "", fmt.Errorf("--token requires a value")
			}
			if strings.TrimSpace(args[i+1]) == "" {
				return "", fmt.Errorf("--token requires a value")
			}
			return args[i+1], nil
		case "--token-stdin":
			if fileInfo, _ := os.Stdin.Stat(); fileInfo != nil && (fileInfo.Mode()&os.ModeCharDevice) != 0 {
				fmt.Fprintln(os.Stderr, "Paste cookie value, then press Ctrl+D (Unix) or Ctrl+Z+Enter (Windows):")
			}
			data, err := io.ReadAll(io.LimitReader(os.Stdin, maxTokenStdinBytes+1))
			if err != nil {
				return "", fmt.Errorf("read token from stdin: %w", err)
			}
			if int64(len(data)) > maxTokenStdinBytes {
				return "", fmt.Errorf("token stdin exceeded %d bytes", maxTokenStdinBytes)
			}
			return string(data), nil
		}
	}
	return "", fmt.Errorf("missing token; use --token-stdin to avoid shell history")
}

func verifyCookieAuth() map[string]any {
	result, err := executeRead(func(ctx context.Context, client *webapi.Client) (any, error) {
		return client.FullSync(ctx)
	})
	if err != nil {
		return map[string]any{"ok": false, "message": err.Error(), "hint": "refresh the Dida365 't' cookie with: dida auth cookie set --token-stdin --json"}
	}
	payload := syncPayloadValue(result)
	return map[string]any{
		"ok":       true,
		"projects": len(payload.Projects),
		"tasks":    len(payload.Tasks),
	}
}
