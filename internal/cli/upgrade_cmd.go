package cli

import (
	"fmt"
	"io"
	"net/http"
	"runtime"
	"time"
)

var versionFromBuild = "dev"

func runUpgrade(args []string, jsonOut bool, stdout io.Writer, stderr io.Writer) int {
	if hasHelpFlag(args) {
		printUpgradeHelp(stdout)
		return 0
	}

	checkOnly := hasFlag(args, "--check")
	httpClient := &http.Client{Timeout: 30 * time.Second}

	release, err := fetchLatestRelease(httpClient)
	if err != nil {
		return failTyped("upgrade", "network", fmt.Sprintf("check for updates failed: %v", err), "check your internet connection and try again", jsonOut, stdout, stderr)
	}

	current := versionFromBuild
	info := upgradeInfo{
		CurrentVersion: current,
		LatestVersion:  release.TagName,
		NeedsUpdate:    isNewer(release.TagName, current),
		ReleaseURL:     fmt.Sprintf("https://github.com/%s/releases/tag/%s", defaultRepo, release.TagName),
	}

	if !info.NeedsUpdate {
		if jsonOut {
			return writeJSON(stdout, envelope{OK: true, Command: "upgrade", Data: info})
		}
		fmt.Fprintf(stdout, "Already up to date: %s\n", current)
		return 0
	}

	if checkOnly {
		if jsonOut {
			return writeJSON(stdout, envelope{OK: true, Command: "upgrade", Data: info})
		}
		fmt.Fprintf(stdout, "New version available: %s (current: %s)\n", release.TagName, current)
		fmt.Fprintf(stdout, "Run 'dida upgrade' to update.\n")
		return 0
	}

	asset, _ := findAsset(release)
	if asset == nil {
		return failTyped("upgrade", "not_found",
			fmt.Sprintf("no release asset found for %s/%s", runtime.GOOS, runtime.GOARCH),
			"check https://github.com/"+defaultRepo+"/releases", jsonOut, stdout, stderr)
	}

	if !jsonOut {
		fmt.Fprintf(stdout, "Downloading %s (%s -> %s)...\n", asset.Name, current, release.TagName)
	}

	var progressWriter io.Writer
	if !jsonOut {
		progressWriter = stderr
	}
	archiveData, err := downloadBytesProgress(httpClient, asset.BrowserDownloadURL, progressWriter)
	if err != nil {
		return failTyped("upgrade", "download", fmt.Sprintf("download failed: %v", err), "try again or download manually from "+info.ReleaseURL, jsonOut, stdout, stderr)
	}

	checksumsAsset := findChecksumsAsset(release)
	if checksumsAsset == nil {
		return failTyped("upgrade", "checksum", "checksums.txt not found in release assets", "the release may be incomplete; try again later or download manually from "+info.ReleaseURL, jsonOut, stdout, stderr)
	}
	if !jsonOut {
		fmt.Fprintln(stdout, "Verifying checksum...")
	}
	checksums, err := downloadBytes(httpClient, checksumsAsset.BrowserDownloadURL)
	if err != nil {
		return failTyped("upgrade", "download", fmt.Sprintf("download checksums.txt failed: %v", err), "", jsonOut, stdout, stderr)
	}
	if err := verifyChecksum(archiveData, checksums, asset.Name); err != nil {
		return failTyped("upgrade", "checksum", err.Error(), "the download may be corrupted; try again", jsonOut, stdout, stderr)
	}

	if !jsonOut {
		fmt.Fprintln(stdout, "Extracting binary...")
	}
	binary, err := extractBinary(archiveData, asset.Name)
	if err != nil {
		return failTyped("upgrade", "extract", fmt.Sprintf("extract failed: %v", err), "", jsonOut, stdout, stderr)
	}

	if !jsonOut {
		fmt.Fprintln(stdout, "Installing...")
	}
	if err := replaceBinary(binary); err != nil {
		return failTyped("upgrade", "install", fmt.Sprintf("replace binary failed: %v", err), "try running with elevated permissions or download manually", jsonOut, stdout, stderr)
	}

	if jsonOut {
		return writeJSON(stdout, envelope{
			OK: true, Command: "upgrade",
			Data: map[string]any{"previous_version": current, "new_version": release.TagName, "asset": asset.Name},
		})
	}
	fmt.Fprintf(stdout, "Updated %s -> %s\n", current, release.TagName)
	return 0
}
