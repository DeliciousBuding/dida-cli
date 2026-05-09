package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"time"
)

type BrowserLoginResult struct {
	Token   string `json:"token"`
	Domain  string `json:"domain"`
	Expires any    `json:"expires,omitempty"`
}

func CaptureCookieWithBrowser(ctx context.Context, timeout time.Duration) (*BrowserLoginResult, error) {
	python, err := findPython()
	if err != nil {
		return nil, err
	}
	profileDir := filepath.Join(DefaultBrowserProfileDir(), "dida-web-login")
	if err := os.MkdirAll(profileDir, 0o700); err != nil {
		return nil, fmt.Errorf("create browser profile dir: %w", err)
	}
	cmd := exec.CommandContext(ctx, python, "-c", browserCaptureScript, profileDir, strconv.Itoa(int(timeout.Seconds())))
	output, err := cmd.Output()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			return nil, fmt.Errorf("browser login failed: %s", string(exitErr.Stderr))
		}
		return nil, fmt.Errorf("start browser login helper: %w", err)
	}
	var result BrowserLoginResult
	if err := json.Unmarshal(output, &result); err != nil {
		return nil, fmt.Errorf("decode browser login result: %w", err)
	}
	if result.Token == "" {
		return nil, fmt.Errorf("browser login did not return cookie 't'")
	}
	return &result, nil
}

func DefaultBrowserProfileDir() string {
	base, err := os.UserConfigDir()
	if err != nil || base == "" {
		base = os.TempDir()
	}
	return filepath.Join(base, "dida-cli", "browser")
}

func findPython() (string, error) {
	candidates := []string{"python3", "python"}
	if runtime.GOOS == "windows" {
		candidates = []string{"python", "py", "python3"}
	}
	for _, candidate := range candidates {
		path, err := exec.LookPath(candidate)
		if err == nil {
			return path, nil
		}
	}
	return "", fmt.Errorf("python not found; install Python with Playwright, or use: dida auth cookie set --token-stdin")
}

const browserCaptureScript = `
import asyncio
import json
import sys

async def main():
    profile_dir = sys.argv[1]
    timeout = int(sys.argv[2])
    try:
        from playwright.async_api import async_playwright
    except Exception as exc:
        print("Python Playwright is not installed. Install it or use manual cookie import: dida auth cookie set --token-stdin", file=sys.stderr)
        raise SystemExit(2) from exc

    async with async_playwright() as p:
        browser_name = "chromium"
        browser = p.chromium
        context = await browser.launch_persistent_context(
            profile_dir,
            headless=False,
            args=["--disable-blink-features=AutomationControlled"],
        )
        page = context.pages[0] if context.pages else await context.new_page()
        await page.goto("https://dida365.com/signin", wait_until="domcontentloaded")
        deadline = asyncio.get_event_loop().time() + timeout
        token_cookie = None
        while asyncio.get_event_loop().time() < deadline:
            cookies = await context.cookies()
            for cookie in cookies:
                domain = cookie.get("domain", "")
                if cookie.get("name") == "t" and "dida365.com" in domain and cookie.get("value"):
                    token_cookie = cookie
                    break
            if token_cookie:
                break
            await asyncio.sleep(1)
        await context.close()
        if not token_cookie:
            print("Timed out waiting for Dida365 cookie 't'. Complete login in the opened browser or retry with a larger timeout.", file=sys.stderr)
            raise SystemExit(3)
        print(json.dumps({
            "token": token_cookie.get("value", ""),
            "domain": token_cookie.get("domain", ""),
            "expires": token_cookie.get("expires"),
        }, ensure_ascii=False))

asyncio.run(main())
`
