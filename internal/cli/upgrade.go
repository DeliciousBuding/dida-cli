package cli

import (
	"archive/tar"
	"archive/zip"
	"bytes"
	"compress/gzip"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

const (
	defaultRepo             = "DeliciousBuding/dida-cli"
	githubAPIBase           = "https://api.github.com"
	metadataDownloadTimeout = 30 * time.Second
	artifactDownloadTimeout = 120 * time.Second
	artifactDownloadMax     = 200 << 20
	artifactDownloadRetries = 2
)

var releasesLatestURL = githubAPIBase + "/repos/" + defaultRepo + "/releases/latest"
var scheduleWindowsReplacement = scheduleWindowsReplacementCommand
var artifactDownloadRetryDelay = 200 * time.Millisecond
var replaceBinaryForUpgrade = replaceBinary

type githubRelease struct {
	TagName string        `json:"tag_name"`
	Assets  []githubAsset `json:"assets"`
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
	return fetchLatestReleaseForVersion(httpClient, currentVersionForUpgrade())
}

func fetchLatestReleaseForVersion(httpClient *http.Client, version string) (*githubRelease, error) {
	if httpClient == nil {
		httpClient = &http.Client{Timeout: metadataDownloadTimeout}
	}
	req, err := http.NewRequest("GET", releasesLatestURL, nil)
	if err != nil {
		return nil, fmt.Errorf("build request: %w", err)
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("User-Agent", "DidaCLI/"+strings.TrimSpace(version))

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

func latestUpgradeMetadata(current string, httpClient *http.Client) (*githubRelease, upgradeInfo, error) {
	release, err := fetchLatestReleaseForVersion(httpClient, current)
	if err != nil {
		return nil, upgradeInfo{}, err
	}
	info := upgradeInfo{
		CurrentVersion: current,
		LatestVersion:  release.TagName,
		NeedsUpdate:    isNewer(release.TagName, current),
		ReleaseURL:     fmt.Sprintf("https://github.com/%s/releases/tag/%s", defaultRepo, release.TagName),
	}
	return release, info, nil
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
	if c == "dev" {
		return true
	}
	return compareSemver(l, c) > 0
}

func compareSemver(a, b string) int {
	ap := parseSemverParts(a)
	bp := parseSemverParts(b)
	for i := 0; i < 3; i++ {
		if ap[i] > bp[i] {
			return 1
		}
		if ap[i] < bp[i] {
			return -1
		}
	}
	return 0
}

func parseSemverParts(v string) [3]int {
	var parts [3]int
	fields := strings.SplitN(v, ".", 3)
	for i, f := range fields {
		if i >= 3 {
			break
		}
		n := 0
		for _, ch := range f {
			if ch >= '0' && ch <= '9' {
				n = n*10 + int(ch-'0')
			} else {
				break
			}
		}
		parts[i] = n
	}
	return parts
}

func findAsset(release *githubRelease) (*githubAsset, string) {
	ext := ".tar.gz"
	if runtime.GOOS == "windows" {
		ext = ".zip"
	}
	name := fmt.Sprintf("dida_%s_%s_%s%s", release.TagName, runtime.GOOS, runtime.GOARCH, ext)
	for _, asset := range release.Assets {
		if asset.Name == name {
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
	return downloadBytesProgress(httpClient, url, nil)
}

func downloadBytesProgress(httpClient *http.Client, url string, progress io.Writer) ([]byte, error) {
	return downloadBytesProgressContext(context.Background(), httpClient, url, progress)
}

func downloadBytesProgressContext(ctx context.Context, httpClient *http.Client, url string, progress io.Writer) ([]byte, error) {
	if httpClient == nil {
		httpClient = &http.Client{Timeout: artifactDownloadTimeout}
	}

	var lastErr error
	for attempt := 0; attempt <= artifactDownloadRetries; attempt++ {
		data, err := downloadBytesProgressOnce(ctx, httpClient, url, progress)
		if err == nil {
			return data, nil
		}
		lastErr = err
		if ctx.Err() != nil || attempt == artifactDownloadRetries || !isRetryableDownloadError(err) {
			return nil, err
		}
		delay := retryDelayForDownloadError(err, attempt)
		if delay <= 0 {
			continue
		}
		timer := time.NewTimer(delay)
		select {
		case <-ctx.Done():
			timer.Stop()
			return nil, ctx.Err()
		case <-timer.C:
		}
	}
	return nil, lastErr
}

func downloadBytesProgressOnce(ctx context.Context, httpClient *http.Client, url string, progress io.Writer) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}
	resp, err := httpClient.Do(req)
	if err != nil {
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}
		return nil, retryableDownloadError{err: err}
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return nil, downloadHTTPError{status: resp.StatusCode, retryAfter: retryAfterDelay(resp.Header.Get("Retry-After"))}
	}
	if resp.ContentLength > artifactDownloadMax {
		return nil, fmt.Errorf("download exceeds %d bytes", artifactDownloadMax)
	}
	var reader io.Reader = io.LimitReader(resp.Body, artifactDownloadMax+1)
	if progress != nil && resp.ContentLength > 0 {
		reader = &progressReader{r: reader, total: resp.ContentLength, w: progress}
	}
	data, err := io.ReadAll(reader)
	if err != nil {
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}
		return nil, retryableDownloadError{err: err}
	}
	if len(data) > artifactDownloadMax {
		return nil, fmt.Errorf("download exceeds %d bytes", artifactDownloadMax)
	}
	return data, nil
}

type downloadHTTPError struct {
	status     int
	retryAfter time.Duration
}

func (e downloadHTTPError) Error() string {
	return fmt.Sprintf("download returned HTTP %d", e.status)
}

type retryableDownloadError struct {
	err error
}

func (e retryableDownloadError) Error() string {
	return e.err.Error()
}

func (e retryableDownloadError) Unwrap() error {
	return e.err
}

func isRetryableDownloadError(err error) bool {
	var httpErr downloadHTTPError
	if errors.As(err, &httpErr) {
		return httpErr.status == http.StatusTooManyRequests || httpErr.status >= 500
	}
	var retryable retryableDownloadError
	return errors.As(err, &retryable)
}

func retryDelayForDownloadError(err error, attempt int) time.Duration {
	var httpErr downloadHTTPError
	if errors.As(err, &httpErr) && httpErr.retryAfter > 0 {
		return httpErr.retryAfter
	}
	return artifactDownloadRetryDelay * time.Duration(1<<attempt)
}

func retryAfterDelay(value string) time.Duration {
	value = strings.TrimSpace(value)
	if value == "" {
		return 0
	}
	seconds, err := parseIntStrict(value)
	if err == nil && seconds > 0 {
		return time.Duration(seconds) * time.Second
	}
	when, err := http.ParseTime(value)
	if err != nil {
		return 0
	}
	delay := time.Until(when)
	if delay < 0 {
		return 0
	}
	return delay
}

type progressReader struct {
	r       io.Reader
	w       io.Writer
	total   int64
	read    int64
	lastPct int
}

func (p *progressReader) Read(buf []byte) (int, error) {
	n, err := p.r.Read(buf)
	p.read += int64(n)
	pct := int(p.read * 100 / p.total)
	if pct != p.lastPct && pct%10 == 0 {
		fmt.Fprintf(p.w, "  %d%% (%d / %d bytes)\n", pct, p.read, p.total)
		p.lastPct = pct
	}
	return n, err
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

type replaceResult struct {
	Status  string
	Message string
}

func replaceBinary(newBinary []byte) (replaceResult, error) {
	exePath, err := os.Executable()
	if err != nil {
		return replaceResult{}, fmt.Errorf("find executable: %w", err)
	}
	exePath, err = filepath.EvalSymlinks(exePath)
	if err != nil {
		return replaceResult{}, fmt.Errorf("resolve symlinks: %w", err)
	}

	if runtime.GOOS == "windows" {
		return replaceBinaryWindows(exePath, newBinary)
	}
	return replaceBinaryUnix(exePath, newBinary)
}

func replaceBinaryUnix(exePath string, newBinary []byte) (replaceResult, error) {
	tmpPath := exePath + ".new"
	if err := os.WriteFile(tmpPath, newBinary, 0o755); err != nil {
		return replaceResult{}, fmt.Errorf("write new binary: %w", err)
	}
	if err := os.Rename(tmpPath, exePath); err != nil {
		os.Remove(tmpPath)
		return replaceResult{}, fmt.Errorf("replace binary: %w", err)
	}
	return replaceResult{Status: "installed"}, nil
}

func replaceBinaryWindows(exePath string, newBinary []byte) (replaceResult, error) {
	newPath := exePath + ".new"
	if err := os.WriteFile(newPath, newBinary, 0o755); err != nil {
		return replaceResult{}, fmt.Errorf("write new binary: %w", err)
	}
	scriptPath, err := writeWindowsReplacementScript(exePath, newPath, os.Getpid())
	if err != nil {
		os.Remove(newPath)
		return replaceResult{}, err
	}
	if err := scheduleWindowsReplacement(scriptPath); err != nil {
		os.Remove(newPath)
		os.Remove(scriptPath)
		return replaceResult{}, fmt.Errorf("schedule deferred replacement: %w", err)
	}
	return replaceResult{
		Status:  "scheduled",
		Message: "replacement will run after the current dida.exe process exits",
	}, nil
}

func writeWindowsReplacementScript(exePath string, newPath string, pid int) (string, error) {
	scriptPath := filepath.Join(os.TempDir(), fmt.Sprintf("dida-upgrade-%d-%d.cmd", pid, time.Now().UnixNano()))
	script := buildWindowsReplacementScript(exePath, newPath, scriptPath, pid)
	if err := os.WriteFile(scriptPath, []byte(script), 0o600); err != nil {
		return "", fmt.Errorf("write deferred replacement script: %w", err)
	}
	return scriptPath, nil
}

func buildWindowsReplacementScript(exePath string, newPath string, scriptPath string, pid int) string {
	return fmt.Sprintf(`@echo off
setlocal
set "TARGET=%s"
set "SOURCE=%s"
set "SELF=%s"
set "PID=%d"
:wait
tasklist /FI "PID eq %%PID%%" 2^>NUL | find "%%PID%%" ^>NUL
if not errorlevel 1 (
  timeout /T 1 /NOBREAK ^>NUL
  goto wait
)
move /Y "%%SOURCE%%" "%%TARGET%%" ^>NUL
if errorlevel 1 exit /B 1
del "%%SELF%%" ^>NUL 2^>NUL
`, batchSetValue(exePath), batchSetValue(newPath), batchSetValue(scriptPath), pid)
}

func batchSetValue(value string) string {
	return strings.ReplaceAll(value, "%", "%%")
}

func scheduleWindowsReplacementCommand(scriptPath string) error {
	cmd := exec.Command("cmd", "/C", "start", "", "/B", scriptPath)
	if err := cmd.Start(); err != nil {
		return err
	}
	return cmd.Process.Release()
}
