package cli

import (
	"archive/tar"
	"archive/zip"
	"bytes"
	"compress/gzip"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

const (
	defaultRepo   = "DeliciousBuding/dida-cli"
	githubAPIBase = "https://api.github.com"
)

var releasesLatestURL = githubAPIBase + "/repos/" + defaultRepo + "/releases/latest"

type githubRelease struct {
	TagName string         `json:"tag_name"`
	Assets  []githubAsset  `json:"assets"`
}

type githubAsset struct {
	Name               string `json:"name"`
	BrowserDownloadURL string `json:"browser_download_url"`
}

type upgradeInfo struct {
	CurrentVersion string `json:"current_version"`
	LatestVersion  string `json:"latest_version"`
	NeedsUpdate    bool   `json:"needs_update"`
	ReleaseURL     string `json:"release_url"`
}

func fetchLatestRelease(httpClient *http.Client) (*githubRelease, error) {
	if httpClient == nil {
		httpClient = &http.Client{Timeout: 30 * time.Second}
	}
	req, err := http.NewRequest("GET", releasesLatestURL, nil)
	if err != nil {
		return nil, fmt.Errorf("build request: %w", err)
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("User-Agent", "DidaCLI/"+currentVersionForUpgrade())

	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetch latest release: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("GitHub API returned HTTP %d", resp.StatusCode)
	}
	var release githubRelease
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return nil, fmt.Errorf("decode release: %w", err)
	}
	return &release, nil
}

func currentVersionForUpgrade() string {
	return strings.TrimSpace(versionFromBuild)
}

func normalizeVersion(v string) string {
	v = strings.TrimSpace(v)
	v = strings.TrimPrefix(v, "v")
	return v
}

func isNewer(latest, current string) bool {
	l := normalizeVersion(latest)
	c := normalizeVersion(current)
	if l == "" || c == "" {
		return false
	}
	// dev versions are always considered older
	if c == "dev" {
		return true
	}
	return l != c
}

func findAsset(release *githubRelease) (*githubAsset, string) {
	pattern := fmt.Sprintf("dida_%s_%s_%s", release.TagName, runtime.GOOS, runtime.GOARCH)
	for _, asset := range release.Assets {
		if strings.HasPrefix(asset.Name, pattern) {
			return &asset, release.TagName
		}
	}
	return nil, ""
}

func findChecksumsAsset(release *githubRelease) *githubAsset {
	for _, asset := range release.Assets {
		if asset.Name == "checksums.txt" {
			return &asset
		}
	}
	return nil
}

func downloadBytes(httpClient *http.Client, url string) ([]byte, error) {
	if httpClient == nil {
		httpClient = &http.Client{Timeout: 120 * time.Second}
	}
	resp, err := httpClient.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("download returned HTTP %d", resp.StatusCode)
	}
	return io.ReadAll(io.LimitReader(resp.Body, 200<<20)) // 200MB max
}

func verifyChecksum(data []byte, checksums []byte, archiveName string) error {
	lines := strings.Split(string(checksums), "\n")
	for _, line := range lines {
		parts := strings.Fields(line)
		if len(parts) != 2 {
			continue
		}
		if parts[1] == archiveName {
			hash := sha256.Sum256(data)
			actual := hex.EncodeToString(hash[:])
			if actual != parts[0] {
				return fmt.Errorf("checksum mismatch: got %s, want %s", actual, parts[0])
			}
			return nil
		}
	}
	return fmt.Errorf("archive %q not found in checksums.txt", archiveName)
}

func extractBinary(data []byte, assetName string) ([]byte, error) {
	if strings.HasSuffix(assetName, ".zip") {
		return extractFromZip(data)
	}
	if strings.HasSuffix(assetName, ".tar.gz") {
		return extractFromTarGz(data)
	}
	return nil, fmt.Errorf("unsupported archive format: %s", assetName)
}

func extractFromZip(data []byte) ([]byte, error) {
	r, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		return nil, fmt.Errorf("open zip: %w", err)
	}
	for _, f := range r.File {
		if f.FileInfo().IsDir() {
			continue
		}
		name := filepath.Base(f.Name)
		if name == "dida" || name == "dida.exe" {
			rc, err := f.Open()
			if err != nil {
				return nil, err
			}
			defer rc.Close()
			return io.ReadAll(rc)
		}
	}
	return nil, fmt.Errorf("dida binary not found in zip archive")
}

func extractFromTarGz(data []byte) ([]byte, error) {
	gz, err := gzip.NewReader(bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("open gzip: %w", err)
	}
	defer gz.Close()
	tr := tar.NewReader(gz)
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
		if hdr.FileInfo().IsDir() {
			continue
		}
		name := filepath.Base(hdr.Name)
		if name == "dida" || name == "dida.exe" {
			return io.ReadAll(tr)
		}
	}
	return nil, fmt.Errorf("dida binary not found in tar.gz archive")
}

func replaceBinary(newBinary []byte) error {
	exePath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("find executable: %w", err)
	}
	exePath, err = filepath.EvalSymlinks(exePath)
	if err != nil {
		return fmt.Errorf("resolve symlinks: %w", err)
	}

	if runtime.GOOS == "windows" {
		return replaceBinaryWindows(exePath, newBinary)
	}
	return replaceBinaryUnix(exePath, newBinary)
}

func replaceBinaryUnix(exePath string, newBinary []byte) error {
	tmpPath := exePath + ".new"
	if err := os.WriteFile(tmpPath, newBinary, 0o755); err != nil {
		return fmt.Errorf("write new binary: %w", err)
	}
	if err := os.Rename(tmpPath, exePath); err != nil {
		os.Remove(tmpPath)
		return fmt.Errorf("replace binary: %w", err)
	}
	return nil
}

func replaceBinaryWindows(exePath string, newBinary []byte) error {
	oldPath := exePath + ".old"
	os.Remove(oldPath) // cleanup from previous update

	if err := os.Rename(exePath, oldPath); err != nil {
		return fmt.Errorf("rename current binary: %w", err)
	}
	if err := os.WriteFile(exePath, newBinary, 0o755); err != nil {
		os.Rename(oldPath, exePath) // rollback
		return fmt.Errorf("write new binary: %w", err)
	}
	os.Remove(oldPath) // cleanup
	return nil
}
