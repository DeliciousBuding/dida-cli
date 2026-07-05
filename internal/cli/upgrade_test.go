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
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"
)

func TestNormalizeVersion(t *testing.T) {
	cases := []struct{ input, want string }{
		{"v0.2.0", "0.2.0"},
		{"0.2.0", "0.2.0"},
		{"  v1.0.0  ", "1.0.0"},
		{"", ""},
		{"dev", "dev"},
	}
	for _, tc := range cases {
		got := normalizeVersion(tc.input)
		if got != tc.want {
			t.Errorf("normalizeVersion(%q) = %q, want %q", tc.input, got, tc.want)
		}
	}
}

func TestIsNewer(t *testing.T) {
	cases := []struct {
		latest, current string
		want            bool
	}{
		{"v0.3.0", "v0.2.0", true},
		{"v0.2.0", "v0.2.0", false},
		{"v1.0.0", "dev", true},
		{"", "v0.2.0", false},
		{"v0.2.0", "", false},
		{"v0.3.0", "0.2.0", true},
		{"v1.0.0", "v2.0.0", false},  // downgrade protection
		{"v0.2.0", "v0.3.0", false},  // downgrade protection
		{"v1.2.3", "v1.2.2", true},   // patch bump
		{"v1.2.0", "v1.1.9", true},   // minor bump
		{"v2.0.0", "v1.99.99", true}, // major bump
	}
	for _, tc := range cases {
		got := isNewer(tc.latest, tc.current)
		if got != tc.want {
			t.Errorf("isNewer(%q, %q) = %v, want %v", tc.latest, tc.current, got, tc.want)
		}
	}
}

func TestFindAsset(t *testing.T) {
	var platformAsset string
	if runtime.GOOS == "windows" {
		platformAsset = "dida_v0.3.0_windows_" + runtime.GOARCH + ".zip"
	} else {
		platformAsset = "dida_v0.3.0_" + runtime.GOOS + "_" + runtime.GOARCH + ".tar.gz"
	}
	release := &githubRelease{
		TagName: "v0.3.0",
		Assets: []githubAsset{
			{Name: platformAsset + ".sig", BrowserDownloadURL: "https://example.com/signature"},
			{Name: platformAsset + ".sha256", BrowserDownloadURL: "https://example.com/sha256"},
			{Name: "dida_v0.3.0_windows_amd64.zip", BrowserDownloadURL: "https://example.com/win"},
			{Name: "dida_v0.3.0_linux_amd64.tar.gz", BrowserDownloadURL: "https://example.com/linux"},
			{Name: platformAsset, BrowserDownloadURL: "https://example.com/archive"},
			{Name: "checksums.txt", BrowserDownloadURL: "https://example.com/checksums"},
		},
	}

	asset, _ := findAsset(release)
	if asset == nil {
		t.Fatalf("findAsset() = nil, want asset for %s/%s", runtime.GOOS, runtime.GOARCH)
	}
	if asset.Name != platformAsset {
		t.Fatalf("findAsset() = %q, want exact archive %q", asset.Name, platformAsset)
	}

	// verify it returns nil for missing platform
	emptyRelease := &githubRelease{
		TagName: "v0.3.0",
		Assets:  []githubAsset{{Name: "dida_v0.3.0_foo_bar.tar.gz"}},
	}
	asset, _ = findAsset(emptyRelease)
	if asset != nil {
		t.Fatalf("findAsset() with wrong platform = %v, want nil", asset)
	}
}

func TestFindChecksumsAsset(t *testing.T) {
	release := &githubRelease{
		Assets: []githubAsset{
			{Name: "dida_v0.3.0_linux_amd64.tar.gz"},
			{Name: "checksums.txt", BrowserDownloadURL: "https://example.com/cs"},
		},
	}
	asset := findChecksumsAsset(release)
	if asset == nil || asset.Name != "checksums.txt" {
		t.Fatalf("findChecksumsAsset() = %v", asset)
	}

	empty := &githubRelease{Assets: []githubAsset{{Name: "other.txt"}}}
	if findChecksumsAsset(empty) != nil {
		t.Fatalf("findChecksumsAsset() without checksums = non-nil")
	}
}

func TestVerifyChecksum(t *testing.T) {
	data := []byte("hello world")
	hash := sha256.Sum256(data)
	expectedHash := hex.EncodeToString(hash[:])

	checksums := []byte(expectedHash + "  test-archive.tar.gz\n")

	if err := verifyChecksum(data, checksums, "test-archive.tar.gz"); err != nil {
		t.Fatalf("verifyChecksum() error = %v", err)
	}

	// wrong hash
	badChecksums := []byte("0000000000000000000000000000000000000000000000000000000000000000  test-archive.tar.gz\n")
	if err := verifyChecksum(data, badChecksums, "test-archive.tar.gz"); err == nil {
		t.Fatalf("verifyChecksum() bad hash: error = nil")
	}

	// missing archive
	if err := verifyChecksum(data, checksums, "missing.tar.gz"); err == nil {
		t.Fatalf("verifyChecksum() missing: error = nil")
	}
}

func TestExtractBinary(t *testing.T) {
	t.Run("unsupported format", func(t *testing.T) {
		_, err := extractBinary([]byte("data"), "test.rar")
		if err == nil {
			t.Fatalf("extractBinary(.rar) error = nil")
		}
	})

	t.Run("zip", func(t *testing.T) {
		binary := buildTestZip(t, "dida_v0.3.0_test/dida", []byte("fake-binary"))
		got, err := extractBinary(binary, "dida_v0.3.0_test.zip")
		if err != nil {
			t.Fatalf("extractBinary(.zip) error = %v", err)
		}
		if string(got) != "fake-binary" {
			t.Fatalf("extractBinary(.zip) = %q", string(got))
		}
	})

	t.Run("tar.gz", func(t *testing.T) {
		binary := buildTestTarGz(t, "dida_v0.3.0_test/dida", []byte("fake-binary-tar"))
		got, err := extractBinary(binary, "dida_v0.3.0_test.tar.gz")
		if err != nil {
			t.Fatalf("extractBinary(.tar.gz) error = %v", err)
		}
		if string(got) != "fake-binary-tar" {
			t.Fatalf("extractBinary(.tar.gz) = %q", string(got))
		}
	})

	t.Run("zip missing binary", func(t *testing.T) {
		binary := buildTestZip(t, "dida_v0.3.0_test/other.txt", []byte("not-the-binary"))
		_, err := extractBinary(binary, "dida_v0.3.0_test.zip")
		if err == nil {
			t.Fatalf("extractBinary() missing dida: error = nil")
		}
	})
}

func TestUpgradeCheckModeWithMockServer(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/repos/DeliciousBuding/dida-cli/releases/latest" {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		_ = json.NewEncoder(w).Encode(githubRelease{
			TagName: "v99.0.0",
			Assets:  []githubAsset{},
		})
	}))
	defer server.Close()

	// Temporarily override the URL
	orig := releasesLatestURL
	releasesLatestURL = server.URL + "/repos/DeliciousBuding/dida-cli/releases/latest"
	defer func() { releasesLatestURL = orig }()

	origVersion := versionFromBuild
	versionFromBuild = "v1.0.0"
	defer func() { versionFromBuild = origVersion }()

	var stdout, stderr bytes.Buffer
	code := runUpgrade([]string{"--check"}, true, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("exit code = %d, stderr=%s", code, stderr.String())
	}
	if !strings.Contains(stdout.String(), `"needs_update": true`) {
		t.Fatalf("stdout missing needs_update: %s", stdout.String())
	}
	if !strings.Contains(stdout.String(), `"latest_version": "v99.0.0"`) {
		t.Fatalf("stdout missing latest_version: %s", stdout.String())
	}
}

func TestUpgradeAlreadyUpToDate(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(githubRelease{
			TagName: "v1.0.0",
			Assets:  []githubAsset{},
		})
	}))
	defer server.Close()

	orig := releasesLatestURL
	releasesLatestURL = server.URL + "/repos/DeliciousBuding/dida-cli/releases/latest"
	defer func() { releasesLatestURL = orig }()

	origVersion := versionFromBuild
	versionFromBuild = "v1.0.0"
	defer func() { versionFromBuild = origVersion }()

	var stdout, stderr bytes.Buffer
	code := runUpgrade(nil, true, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("exit code = %d, stderr=%s", code, stderr.String())
	}
	if !strings.Contains(stdout.String(), `"needs_update": false`) {
		t.Fatalf("stdout: %s", stdout.String())
	}
}

func TestUpgradeNetworkError(t *testing.T) {
	orig := releasesLatestURL
	releasesLatestURL = "http://127.0.0.1:1/not-a-real-server"
	defer func() { releasesLatestURL = orig }()

	origVersion := versionFromBuild
	versionFromBuild = "v1.0.0"
	defer func() { versionFromBuild = origVersion }()

	var stdout, stderr bytes.Buffer
	code := runUpgrade(nil, true, &stdout, &stderr)
	if code != 1 {
		t.Fatalf("exit code = %d, want 1", code)
	}
	if !strings.Contains(stdout.String(), "network") {
		t.Fatalf("stdout missing network error type: %s", stdout.String())
	}
}

func TestUpgradeHelp(t *testing.T) {
	var stdout, stderr bytes.Buffer
	code := runUpgrade([]string{"--help"}, false, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("exit code = %d", code)
	}
	if !strings.Contains(stdout.String(), "--check") {
		t.Fatalf("help missing --check: %s", stdout.String())
	}
}

func TestUpgradeFullFlowIntegration(t *testing.T) {
	fakeBinary := []byte("#!/bin/sh\necho dida v99.0.0")
	var archiveData []byte
	var assetName string
	if runtime.GOOS == "windows" {
		assetName = "dida_v99.0.0_windows_" + runtime.GOARCH + ".zip"
		archiveData = buildTestZip(t, "dida_v99.0.0_windows_"+runtime.GOARCH+"/dida.exe", fakeBinary)
	} else {
		assetName = "dida_v99.0.0_" + runtime.GOOS + "_" + runtime.GOARCH + ".tar.gz"
		archiveData = buildTestTarGz(t, "dida_v99.0.0_"+runtime.GOOS+"_"+runtime.GOARCH+"/dida", fakeBinary)
	}

	archiveHash := sha256.Sum256(archiveData)
	checksumLine := hex.EncodeToString(archiveHash[:]) + "  " + assetName + "\n"

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/repos/DeliciousBuding/dida-cli/releases/latest":
			_ = json.NewEncoder(w).Encode(githubRelease{
				TagName: "v99.0.0",
				Assets: []githubAsset{
					{Name: assetName, BrowserDownloadURL: "http://" + r.Host + "/download/" + assetName},
					{Name: "checksums.txt", BrowserDownloadURL: "http://" + r.Host + "/download/checksums.txt"},
				},
			})
		case "/download/" + assetName:
			w.Header().Set("Content-Length", fmt.Sprintf("%d", len(archiveData)))
			w.Write(archiveData)
		case "/download/checksums.txt":
			w.Write([]byte(checksumLine))
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	orig := releasesLatestURL
	releasesLatestURL = server.URL + "/repos/DeliciousBuding/dida-cli/releases/latest"
	defer func() { releasesLatestURL = orig }()

	origVersion := versionFromBuild
	versionFromBuild = "v1.0.0"
	defer func() { versionFromBuild = origVersion }()

	var stdout, stderr bytes.Buffer
	code := runUpgrade([]string{"--check"}, true, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("--check exit code = %d, stdout=%s stderr=%s", code, stdout.String(), stderr.String())
	}
	if !strings.Contains(stdout.String(), `"needs_update": true`) {
		t.Fatalf("--check stdout missing needs_update: %s", stdout.String())
	}
}

func TestUpgradeInstallFlowDownloadsVerifiesExtractsAndInstalls(t *testing.T) {
	fakeBinary := []byte("#!/bin/sh\necho dida v99.0.0")
	var archiveData []byte
	var assetName string
	if runtime.GOOS == "windows" {
		assetName = "dida_v99.0.0_windows_" + runtime.GOARCH + ".zip"
		archiveData = buildTestZip(t, "dida_v99.0.0_windows_"+runtime.GOARCH+"/dida.exe", fakeBinary)
	} else {
		assetName = "dida_v99.0.0_" + runtime.GOOS + "_" + runtime.GOARCH + ".tar.gz"
		archiveData = buildTestTarGz(t, "dida_v99.0.0_"+runtime.GOOS+"_"+runtime.GOARCH+"/dida", fakeBinary)
	}
	archiveHash := sha256.Sum256(archiveData)
	checksumLine := hex.EncodeToString(archiveHash[:]) + "  " + assetName + "\n"

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/repos/DeliciousBuding/dida-cli/releases/latest":
			_ = json.NewEncoder(w).Encode(githubRelease{
				TagName: "v99.0.0",
				Assets: []githubAsset{
					{Name: assetName, BrowserDownloadURL: "http://" + r.Host + "/download/" + assetName},
					{Name: "checksums.txt", BrowserDownloadURL: "http://" + r.Host + "/download/checksums.txt"},
				},
			})
		case "/download/" + assetName:
			w.Header().Set("Content-Length", fmt.Sprintf("%d", len(archiveData)))
			w.Write(archiveData)
		case "/download/checksums.txt":
			w.Write([]byte(checksumLine))
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	origURL := releasesLatestURL
	releasesLatestURL = server.URL + "/repos/DeliciousBuding/dida-cli/releases/latest"
	defer func() { releasesLatestURL = origURL }()

	origVersion := versionFromBuild
	versionFromBuild = "v1.0.0"
	defer func() { versionFromBuild = origVersion }()

	origReplace := replaceBinaryForUpgrade
	var installed []byte
	replaceBinaryForUpgrade = func(newBinary []byte) (replaceResult, error) {
		installed = append([]byte(nil), newBinary...)
		return replaceResult{Status: "installed"}, nil
	}
	defer func() { replaceBinaryForUpgrade = origReplace }()

	var stdout, stderr bytes.Buffer
	code := runUpgrade(nil, true, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("runUpgrade() code = %d stdout=%s stderr=%s", code, stdout.String(), stderr.String())
	}
	if string(installed) != string(fakeBinary) {
		t.Fatalf("installed binary = %q, want %q", string(installed), string(fakeBinary))
	}
	if !strings.Contains(stdout.String(), `"status": "installed"`) {
		t.Fatalf("stdout missing installed status: %s", stdout.String())
	}
}

func TestDownloadBytesProgressRetriesTransientHTTPStatus(t *testing.T) {
	origDelay := artifactDownloadRetryDelay
	artifactDownloadRetryDelay = 0
	defer func() { artifactDownloadRetryDelay = origDelay }()

	requests := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests++
		if requests < 3 {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.Write([]byte("ok"))
	}))
	defer server.Close()

	data, err := downloadBytesProgress(nil, server.URL, nil)
	if err != nil {
		t.Fatalf("downloadBytesProgress() error = %v", err)
	}
	if string(data) != "ok" {
		t.Fatalf("data = %q, want ok", string(data))
	}
	if requests != 3 {
		t.Fatalf("requests = %d, want 3", requests)
	}
}

func TestDownloadBytesProgressRetriesRateLimitStatus(t *testing.T) {
	origDelay := artifactDownloadRetryDelay
	artifactDownloadRetryDelay = 0
	defer func() { artifactDownloadRetryDelay = origDelay }()

	requests := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests++
		if requests == 1 {
			w.Header().Set("Retry-After", "0")
			w.WriteHeader(http.StatusTooManyRequests)
			return
		}
		w.Write([]byte("ok"))
	}))
	defer server.Close()

	data, err := downloadBytesProgress(nil, server.URL, nil)
	if err != nil {
		t.Fatalf("downloadBytesProgress() error = %v", err)
	}
	if string(data) != "ok" {
		t.Fatalf("data = %q, want ok", string(data))
	}
	if requests != 2 {
		t.Fatalf("requests = %d, want 2", requests)
	}
}

func TestRunUpgradeReturnsChecksumDownloadErrorWithoutWaitingForSlowArchive(t *testing.T) {
	origDelay := artifactDownloadRetryDelay
	artifactDownloadRetryDelay = 0
	defer func() { artifactDownloadRetryDelay = origDelay }()

	var assetName string
	if runtime.GOOS == "windows" {
		assetName = "dida_v99.0.0_windows_" + runtime.GOARCH + ".zip"
	} else {
		assetName = "dida_v99.0.0_" + runtime.GOOS + "_" + runtime.GOARCH + ".tar.gz"
	}

	archiveCanceled := make(chan struct{})
	archiveStarted := make(chan struct{})
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/repos/DeliciousBuding/dida-cli/releases/latest":
			_ = json.NewEncoder(w).Encode(githubRelease{
				TagName: "v99.0.0",
				Assets: []githubAsset{
					{Name: assetName, BrowserDownloadURL: "http://" + r.Host + "/download/" + assetName},
					{Name: "checksums.txt", BrowserDownloadURL: "http://" + r.Host + "/download/checksums.txt"},
				},
			})
		case "/download/" + assetName:
			w.Header().Set("Content-Length", "100")
			closeOnce(archiveStarted)
			select {
			case <-r.Context().Done():
				close(archiveCanceled)
			case <-time.After(5 * time.Second):
				w.Write([]byte("slow archive"))
			}
		case "/download/checksums.txt":
			<-archiveStarted
			w.WriteHeader(http.StatusInternalServerError)
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	origURL := releasesLatestURL
	releasesLatestURL = server.URL + "/repos/DeliciousBuding/dida-cli/releases/latest"
	defer func() { releasesLatestURL = origURL }()

	origVersion := versionFromBuild
	versionFromBuild = "v1.0.0"
	defer func() { versionFromBuild = origVersion }()

	start := time.Now()
	var stdout, stderr bytes.Buffer
	code := runUpgrade(nil, true, &stdout, &stderr)
	elapsed := time.Since(start)
	if code != 1 {
		t.Fatalf("runUpgrade() code = %d, want 1 stdout=%s stderr=%s", code, stdout.String(), stderr.String())
	}
	if elapsed > 2*time.Second {
		t.Fatalf("runUpgrade() took %s, want fast checksum failure", elapsed)
	}
	if !strings.Contains(stdout.String(), "download checksums.txt failed") {
		t.Fatalf("stdout missing checksums download error: %s", stdout.String())
	}
	select {
	case <-archiveCanceled:
	case <-time.After(2 * time.Second):
		t.Fatalf("archive request was not canceled")
	}
}

func TestUpgradeMissingChecksumsFails(t *testing.T) {
	var assetName string
	if runtime.GOOS == "windows" {
		assetName = "dida_v99.0.0_windows_" + runtime.GOARCH + ".zip"
	} else {
		assetName = "dida_v99.0.0_" + runtime.GOOS + "_" + runtime.GOARCH + ".tar.gz"
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/repos/DeliciousBuding/dida-cli/releases/latest":
			_ = json.NewEncoder(w).Encode(githubRelease{
				TagName: "v99.0.0",
				Assets: []githubAsset{
					{Name: assetName, BrowserDownloadURL: "http://" + r.Host + "/download/" + assetName},
				},
			})
		case "/download/" + assetName:
			w.Write([]byte("fake-archive-data"))
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	orig := releasesLatestURL
	releasesLatestURL = server.URL + "/repos/DeliciousBuding/dida-cli/releases/latest"
	defer func() { releasesLatestURL = orig }()

	origVersion := versionFromBuild
	versionFromBuild = "v1.0.0"
	defer func() { versionFromBuild = origVersion }()

	var stdout, stderr bytes.Buffer
	code := runUpgrade(nil, true, &stdout, &stderr)
	if code == 0 {
		t.Fatalf("expected failure when checksums.txt missing, got exit 0")
	}
	if !strings.Contains(stdout.String(), "checksum") {
		t.Fatalf("expected checksum error type, got: %s", stdout.String())
	}
}

func TestUpgradeChecksumMismatchFails(t *testing.T) {
	var assetName string
	if runtime.GOOS == "windows" {
		assetName = "dida_v99.0.0_windows_" + runtime.GOARCH + ".zip"
	} else {
		assetName = "dida_v99.0.0_" + runtime.GOOS + "_" + runtime.GOARCH + ".tar.gz"
	}
	archiveData := []byte("fake-archive-data")
	badChecksum := "0000000000000000000000000000000000000000000000000000000000000000  " + assetName + "\n"

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/repos/DeliciousBuding/dida-cli/releases/latest":
			_ = json.NewEncoder(w).Encode(githubRelease{
				TagName: "v99.0.0",
				Assets: []githubAsset{
					{Name: assetName, BrowserDownloadURL: "http://" + r.Host + "/download/" + assetName},
					{Name: "checksums.txt", BrowserDownloadURL: "http://" + r.Host + "/download/checksums.txt"},
				},
			})
		case "/download/" + assetName:
			w.Write(archiveData)
		case "/download/checksums.txt":
			w.Write([]byte(badChecksum))
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	orig := releasesLatestURL
	releasesLatestURL = server.URL + "/repos/DeliciousBuding/dida-cli/releases/latest"
	defer func() { releasesLatestURL = orig }()

	origVersion := versionFromBuild
	versionFromBuild = "v1.0.0"
	defer func() { versionFromBuild = origVersion }()

	var stdout, stderr bytes.Buffer
	code := runUpgrade(nil, true, &stdout, &stderr)
	if code == 0 {
		t.Fatalf("expected failure on checksum mismatch, got exit 0")
	}
	if !strings.Contains(stdout.String(), "checksum") {
		t.Fatalf("expected checksum error, got: %s", stdout.String())
	}
}

func TestProgressReader(t *testing.T) {
	data := bytes.Repeat([]byte("x"), 1000)
	var progress bytes.Buffer
	pr := &progressReader{r: bytes.NewReader(data), total: 1000, w: &progress}
	buf := make([]byte, 100)
	for {
		_, err := pr.Read(buf)
		if err != nil {
			break
		}
	}
	if !strings.Contains(progress.String(), "100%") {
		t.Fatalf("progress output missing 100%%: %s", progress.String())
	}
}

func TestReplaceBinaryWindowsUsesDeferredInstaller(t *testing.T) {
	dir := t.TempDir()
	exePath := filepath.Join(dir, "dida.exe")
	if err := os.WriteFile(exePath, []byte("old"), 0o755); err != nil {
		t.Fatal(err)
	}

	origSchedule := scheduleWindowsReplacement
	var scheduledScript string
	scheduleWindowsReplacement = func(scriptPath string) error {
		scheduledScript = scriptPath
		return nil
	}
	defer func() { scheduleWindowsReplacement = origSchedule }()

	result, err := replaceBinaryWindows(exePath, []byte("new"))
	if err != nil {
		t.Fatalf("replaceBinaryWindows() error = %v", err)
	}
	if result.Status != "scheduled" {
		t.Fatalf("replace result status = %q, want scheduled", result.Status)
	}

	current, err := os.ReadFile(exePath)
	if err != nil {
		t.Fatal(err)
	}
	if string(current) != "old" {
		t.Fatalf("running exe should not be replaced synchronously, got %q", string(current))
	}
	newData, err := os.ReadFile(exePath + ".new")
	if err != nil {
		t.Fatalf("new binary was not staged: %v", err)
	}
	if string(newData) != "new" {
		t.Fatalf("staged binary = %q, want new", string(newData))
	}
	if scheduledScript == "" {
		t.Fatalf("replacement script was not scheduled")
	}
	script, err := os.ReadFile(scheduledScript)
	if err != nil {
		t.Fatalf("scheduled script missing: %v", err)
	}
	scriptText := string(script)
	for _, want := range []string{"tasklist", "move /Y", exePath, exePath + ".new"} {
		if !strings.Contains(scriptText, want) {
			t.Fatalf("script missing %q:\n%s", want, scriptText)
		}
	}
	_ = os.Remove(scheduledScript)
}

func TestReplaceBinaryWindowsCleansStagedFilesWhenSchedulingFails(t *testing.T) {
	dir := t.TempDir()
	exePath := filepath.Join(dir, "dida.exe")
	if err := os.WriteFile(exePath, []byte("old"), 0o755); err != nil {
		t.Fatal(err)
	}

	origSchedule := scheduleWindowsReplacement
	var scheduledScript string
	scheduleWindowsReplacement = func(scriptPath string) error {
		scheduledScript = scriptPath
		return fmt.Errorf("scheduler failed")
	}
	defer func() { scheduleWindowsReplacement = origSchedule }()

	_, err := replaceBinaryWindows(exePath, []byte("new"))
	if err == nil {
		t.Fatalf("replaceBinaryWindows() error = nil, want scheduler failure")
	}
	if _, err := os.Stat(exePath + ".new"); !os.IsNotExist(err) {
		t.Fatalf("staged binary should be removed, stat error = %v", err)
	}
	if scheduledScript == "" {
		t.Fatalf("expected scheduler to receive script path")
	}
	if _, err := os.Stat(scheduledScript); !os.IsNotExist(err) {
		t.Fatalf("script should be removed after scheduling failure, stat error = %v", err)
	}
}

func TestBuildWindowsReplacementScriptEscapesPercent(t *testing.T) {
	script := buildWindowsReplacementScript(`X:\Pkg\Dida%CLI\dida.exe`, `X:\Pkg\Dida%CLI\dida.exe.new`, `X:\Pkg\dida%upgrade.cmd`, 1234)
	for _, want := range []string{`Dida%%CLI`, `dida%%upgrade.cmd`, `PID=1234`} {
		if !strings.Contains(script, want) {
			t.Fatalf("script missing escaped value %q:\n%s", want, script)
		}
	}
}

func closeOnce(ch chan struct{}) {
	select {
	case <-ch:
	default:
		close(ch)
	}
}

// helpers

func buildTestZip(t *testing.T, name string, content []byte) []byte {
	t.Helper()
	var buf bytes.Buffer
	w := zip.NewWriter(&buf)
	f, err := w.Create(name)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := f.Write(content); err != nil {
		t.Fatal(err)
	}
	w.Close()
	return buf.Bytes()
}

func buildTestTarGz(t *testing.T, name string, content []byte) []byte {
	t.Helper()
	var buf bytes.Buffer
	gw := gzip.NewWriter(&buf)
	tw := tar.NewWriter(gw)
	hdr := &tar.Header{Name: name, Mode: 0o755, Size: int64(len(content))}
	if err := tw.WriteHeader(hdr); err != nil {
		t.Fatal(err)
	}
	if _, err := tw.Write(content); err != nil {
		t.Fatal(err)
	}
	tw.Close()
	gw.Close()
	return buf.Bytes()
}
