package cli

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"runtime"
)

var versionFromBuild = "dev"

func runUpgrade(args []string, jsonOut bool, stdout io.Writer, stderr io.Writer) int {
	if hasHelpFlag(args) {
		printUpgradeHelp(stdout)
		return 0
	}

	checkOnly := hasFlag(args, "--check")
	metadataClient := &http.Client{Timeout: metadataDownloadTimeout}
	artifactClient := &http.Client{Timeout: artifactDownloadTimeout}

	release, info, err := latestUpgradeMetadata(versionFromBuild, metadataClient)
	if err != nil {
		return failTyped("upgrade", "network", fmt.Sprintf("check for updates failed: %v", err), "check your internet connection and try again", jsonOut, stdout, stderr)
	}

	current := versionFromBuild

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
	checksumsAsset := findChecksumsAsset(release)
	if checksumsAsset == nil {
		return failTyped("upgrade", "checksum", "checksums.txt not found in release assets", "the release may be incomplete; try again later or download manually from "+info.ReleaseURL, jsonOut, stdout, stderr)
	}

	var progressWriter io.Writer
	if !jsonOut {
		progressWriter = stderr
	}

	type downloadResult struct {
		kind string
		data []byte
		err  error
	}
	downloadCtx, cancelDownloads := context.WithCancel(context.Background())
	defer cancelDownloads()
	results := make(chan downloadResult, 2)
	go func() {
		data, err := downloadBytesProgressContext(downloadCtx, artifactClient, asset.BrowserDownloadURL, progressWriter)
		results <- downloadResult{kind: "archive", data: data, err: err}
	}()
	go func() {
		data, err := downloadBytesProgressContext(downloadCtx, artifactClient, checksumsAsset.BrowserDownloadURL, nil)
		results <- downloadResult{kind: "checksums", data: data, err: err}
	}()

	var archiveData []byte
	var checksums []byte
	for completed := 0; completed < 2; completed++ {
		result := <-results
		if result.err != nil {
			cancelDownloads()
			if result.kind == "checksums" {
				return failTyped("upgrade", "download", fmt.Sprintf("download checksums.txt failed: %v", result.err), "", jsonOut, stdout, stderr)
			}
			return failTyped("upgrade", "download", fmt.Sprintf("download failed: %v", result.err), "try again or download manually from "+info.ReleaseURL, jsonOut, stdout, stderr)
		}
		switch result.kind {
		case "archive":
			archiveData = result.data
		case "checksums":
			checksums = result.data
		}
	}
	cancelDownloads()

	if !jsonOut {
		fmt.Fprintln(stdout, "Verifying checksum...")
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
	replace, err := replaceBinaryForUpgrade(binary)
	if err != nil {
		return failTyped("upgrade", "install", fmt.Sprintf("replace binary failed: %v", err), "try running with elevated permissions or download manually", jsonOut, stdout, stderr)
	}

	if jsonOut {
		data := map[string]any{"previous_version": current, "new_version": release.TagName, "asset": asset.Name, "status": replace.Status}
		if replace.Message != "" {
			data["message"] = replace.Message
		}
		return writeJSON(stdout, envelope{
			OK: true, Command: "upgrade",
			Data: data,
		})
	}
	if replace.Status == "scheduled" {
		fmt.Fprintf(stdout, "Update scheduled: %s -> %s\n", current, release.TagName)
		if replace.Message != "" {
			fmt.Fprintln(stdout, replace.Message)
		}
	} else {
		fmt.Fprintf(stdout, "Updated %s -> %s\n", current, release.TagName)
	}
	return 0
}
