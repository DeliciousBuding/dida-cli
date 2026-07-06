package cli

import (
	"fmt"
	"io"
	"net/http"
	"path/filepath"
	"runtime"

	"github.com/DeliciousBuding/dida-cli/internal/auth"
	"github.com/DeliciousBuding/dida-cli/internal/config"
)

func runDoctor(args []string, version string, jsonOut bool, stdout io.Writer, stderr io.Writer) int {
	if len(args) > 0 && (args[0] == "-h" || args[0] == "--help") {
		fmt.Fprintln(stdout, "Usage: dida doctor [--verify] [--check-upgrade] [--json]")
		return 0
	}
	verify := hasFlag(args, "--verify")
	checkUpgrade := hasFlag(args, "--check-upgrade")

	cfgDir := config.DefaultDir()
	cookiePath := filepath.Join(cfgDir, "cookie.json")
	oauthPath := filepath.Join(cfgDir, "oauth.json")
	openapiOAuthPath := filepath.Join(cfgDir, "openapi-oauth.json")
	cookieExists := fileExists(cookiePath)
	oauthExists := fileExists(oauthPath)
	openapiOAuthExists := fileExists(openapiOAuthPath)

	data := map[string]any{
		"version":       version,
		"goos":          runtime.GOOS,
		"goarch":        runtime.GOARCH,
		"config_dir":    cfgDir,
		"auth_sources":  map[string]bool{"cookie": cookieExists, "oauth": oauthExists, "openapi_oauth": openapiOAuthExists},
		"cookie_status": auth.CookieStatus(),
		"network_check": "not_run",
		"upgrade_check": "not_run",
	}
	if checkUpgrade {
		data["upgrade_check"] = doctorUpgradeCheck(version)
	}
	if verify {
		verifyResult := verifyCookieAuth()
		data["network_check"] = map[string]any{
			"channel": "webapi",
			"status":  doctorNetworkStatus(verifyResult),
			"result":  verifyResult,
		}
		if verifyResult["ok"] != true {
			if jsonOut {
				_ = writeJSON(stdout, envelope{
					OK:      false,
					Command: "doctor",
					Data:    data,
					Error: &cliError{
						Type:    "auth",
						Message: fmt.Sprint(verifyResult["message"]),
						Hint:    fmt.Sprint(verifyResult["hint"]),
					},
				})
				return 1
			}
			fmt.Fprintf(stderr, "dida: %s\n", verifyResult["message"])
			fmt.Fprintf(stderr, "hint: %s\n", verifyResult["hint"])
			return 1
		}
	}

	if jsonOut {
		return writeJSON(stdout, envelope{OK: true, Command: "doctor", Data: data})
	}

	fmt.Fprintf(stdout, "DidaCLI %s\n", version)
	fmt.Fprintf(stdout, "Config: %s\n", cfgDir)
	fmt.Fprintf(stdout, "Cookie auth: %s\n", yesNo(cookieExists))
	fmt.Fprintf(stdout, "OAuth auth: %s\n", yesNo(oauthExists))
	fmt.Fprintf(stdout, "OpenAPI OAuth: %s\n", yesNo(openapiOAuthExists))
	if verify {
		fmt.Fprintf(stdout, "Network check: %v\n", data["network_check"])
	} else {
		fmt.Fprintln(stdout, "Network check: not run")
	}
	if checkUpgrade {
		printDoctorUpgradeCheck(stdout, data["upgrade_check"])
	} else {
		fmt.Fprintln(stdout, "Upgrade check: not run")
	}
	return 0
}

func doctorNetworkStatus(verifyResult map[string]any) string {
	if verifyResult["ok"] == true {
		return "ok"
	}
	return "failed"
}

func doctorUpgradeCheck(version string) map[string]any {
	client := &http.Client{Timeout: metadataDownloadTimeout}
	_, info, err := latestUpgradeMetadata(version, client)
	if err != nil {
		return map[string]any{
			"status": "failed",
			"error":  fmt.Sprintf("check for updates failed: %v", err),
			"hint":   "check your internet connection or run: dida upgrade --check --json",
		}
	}
	status := "current"
	if info.NeedsUpdate {
		status = "available"
	}
	return map[string]any{
		"status":          status,
		"current_version": info.CurrentVersion,
		"latest_version":  info.LatestVersion,
		"needs_update":    info.NeedsUpdate,
		"release_url":     info.ReleaseURL,
	}
}

func printDoctorUpgradeCheck(stdout io.Writer, value any) {
	check, ok := value.(map[string]any)
	if !ok {
		fmt.Fprintf(stdout, "Upgrade check: %v\n", value)
		return
	}
	switch check["status"] {
	case "available":
		fmt.Fprintf(stdout, "Upgrade check: update available %v (current: %v)\n", check["latest_version"], check["current_version"])
	case "current":
		fmt.Fprintf(stdout, "Upgrade check: current (%v)\n", check["current_version"])
	case "failed":
		fmt.Fprintf(stdout, "Upgrade check: failed (%v)\n", check["error"])
	default:
		fmt.Fprintf(stdout, "Upgrade check: %v\n", check["status"])
	}
}
