package cli

import (
	"fmt"
	"io"
	"path/filepath"
	"runtime"

	"github.com/DeliciousBuding/dida-cli/internal/auth"
	"github.com/DeliciousBuding/dida-cli/internal/config"
)

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
