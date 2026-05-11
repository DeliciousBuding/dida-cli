package auth

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRedactToken(t *testing.T) {
	tests := []struct {
		name  string
		token string
		want  string
	}{
		{name: "short", token: "abcd", want: "***"},
		{name: "long", token: "1234567890abcdef", want: "1234...cdef"},
		{name: "trim", token: "  1234567890abcdef  ", want: "1234...cdef"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := RedactToken(tt.token); got != tt.want {
				t.Fatalf("RedactToken() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestNormalizeCookieToken(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    string
		wantErr bool
	}{
		{name: "value only", input: "abc123", want: "abc123"},
		{name: "named t cookie", input: " t=abc123 ", want: "abc123"},
		{name: "full header rejected", input: "Cookie: t=abc123", wantErr: true},
		{name: "multiple cookies rejected", input: "t=abc123; other=secret", wantErr: true},
		{name: "embedded whitespace rejected", input: "abc 123", wantErr: true},
		{name: "empty rejected", input: " t= ", wantErr: true},
		{name: "unexpected key rejected", input: "other=abc123", wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NormalizeCookieToken(tt.input)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("NormalizeCookieToken() error = nil, want error")
				}
				return
			}
			if err != nil {
				t.Fatalf("NormalizeCookieToken() error = %v", err)
			}
			if got != tt.want {
				t.Fatalf("NormalizeCookieToken() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestClearCookieTokenRemovesBrowserProfile(t *testing.T) {
	configDir := t.TempDir()
	browserDir := filepath.Join(t.TempDir(), "browser")
	t.Setenv("DIDA_CONFIG_DIR", configDir)
	t.Setenv("DIDA_BROWSER_PROFILE_DIR", browserDir)

	if _, err := SaveCookieToken("abc123"); err != nil {
		t.Fatalf("SaveCookieToken() error = %v", err)
	}
	profileFile := filepath.Join(BrowserLoginProfileDir(), "session.txt")
	if err := os.MkdirAll(filepath.Dir(profileFile), 0o700); err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}
	if err := os.WriteFile(profileFile, []byte("cookie cache"), 0o600); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}
	if err := ClearCookieToken(); err != nil {
		t.Fatalf("ClearCookieToken() error = %v", err)
	}
	if _, err := os.Stat(CookiePath()); !os.IsNotExist(err) {
		t.Fatalf("cookie path still exists or unexpected error: %v", err)
	}
	if _, err := os.Stat(BrowserLoginProfileDir()); !os.IsNotExist(err) {
		t.Fatalf("browser profile still exists or unexpected error: %v", err)
	}
}

func TestValidateBrowserProfileRemovalTargetRejectsUnsafePaths(t *testing.T) {
	temp := t.TempDir()
	if _, err := validateBrowserProfileRemovalTarget(filepath.Join(temp, "dida-web-login")); err != nil {
		t.Fatalf("safe profile path rejected: %v", err)
	}
	if _, err := validateBrowserProfileRemovalTarget(filepath.Join(temp, "not-profile")); err == nil {
		t.Fatalf("unexpected profile basename accepted")
	}
	if home, err := os.UserHomeDir(); err == nil {
		if _, err := validateBrowserProfileRemovalTarget(filepath.Join(home, "dida-web-login")); err == nil {
			t.Fatalf("home-level profile path accepted")
		}
	}
}

func TestSaveCookieTokenRejectsEmpty(t *testing.T) {
	t.Setenv("DIDA_CONFIG_DIR", t.TempDir())
	_, err := SaveCookieToken("")
	if err == nil {
		t.Fatalf("SaveCookieToken(\"\") error = nil")
	}
	_, err = SaveCookieToken("Cookie: t=abc")
	if err == nil {
		t.Fatalf("SaveCookieToken with Cookie: prefix error = nil")
	}
}

func TestSaveAndLoadCookieToken(t *testing.T) {
	t.Setenv("DIDA_CONFIG_DIR", t.TempDir())
	saved, err := SaveCookieToken("my_secret_token")
	if err != nil {
		t.Fatalf("SaveCookieToken() error = %v", err)
	}
	if saved.Token != "my_secret_token" {
		t.Fatalf("saved.Token = %q", saved.Token)
	}
	if saved.SavedAt == 0 {
		t.Fatalf("saved.SavedAt = 0")
	}

	loaded, err := LoadCookieToken()
	if err != nil {
		t.Fatalf("LoadCookieToken() error = %v", err)
	}
	if loaded.Token != "my_secret_token" {
		t.Fatalf("loaded.Token = %q", loaded.Token)
	}
}

func TestLoadCookieTokenMissingFile(t *testing.T) {
	t.Setenv("DIDA_CONFIG_DIR", t.TempDir())
	_, err := LoadCookieToken()
	if err == nil {
		t.Fatalf("LoadCookieToken() on missing file: error = nil")
	}
}

func TestLoadCookieTokenEmptyToken(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("DIDA_CONFIG_DIR", dir)
	data := []byte(`{"token":"","saved_at":12345}`)
	if err := os.WriteFile(CookiePath(), data, 0o600); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}
	_, err := LoadCookieToken()
	if err == nil {
		t.Fatalf("LoadCookieToken() empty token: error = nil")
	}
	if !strings.Contains(err.Error(), "no token") {
		t.Fatalf("error = %v, want 'no token'", err)
	}
}

func TestLoadCookieTokenInvalidJSON(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("DIDA_CONFIG_DIR", dir)
	if err := os.WriteFile(CookiePath(), []byte("not json"), 0o600); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}
	_, err := LoadCookieToken()
	if err == nil {
		t.Fatalf("LoadCookieToken() invalid JSON: error = nil")
	}
	if !strings.Contains(err.Error(), "decode") {
		t.Fatalf("error = %v, want decode error", err)
	}
}

func TestCookieStatusMissing(t *testing.T) {
	t.Setenv("DIDA_CONFIG_DIR", t.TempDir())
	status := CookieStatus()
	if status["available"] != false {
		t.Fatalf("available = %v, want false", status["available"])
	}
	if status["message"] != "missing" {
		t.Fatalf("message = %v, want missing", status["message"])
	}
	if _, ok := status["path"].(string); !ok {
		t.Fatalf("path should be a string")
	}
}

func TestCookieStatusAvailable(t *testing.T) {
	t.Setenv("DIDA_CONFIG_DIR", t.TempDir())
	_, err := SaveCookieToken("test_cookie_value_12345")
	if err != nil {
		t.Fatalf("SaveCookieToken() error = %v", err)
	}
	status := CookieStatus()
	if status["available"] != true {
		t.Fatalf("available = %v, want true", status["available"])
	}
	if status["token_length"] != len("test_cookie_value_12345") {
		t.Fatalf("token_length = %v", status["token_length"])
	}
	preview, _ := status["token_preview"].(string)
	if !strings.Contains(preview, "...") {
		t.Fatalf("token_preview = %q, want redacted", preview)
	}
	// Ensure token value is not leaked
	all, _ := json.Marshal(status)
	if strings.Contains(string(all), "test_cookie_value_12345") {
		t.Fatalf("CookieStatus leaked full token: %s", string(all))
	}
}

func TestClearCookieTokenNoFile(t *testing.T) {
	t.Setenv("DIDA_CONFIG_DIR", t.TempDir())
	t.Setenv("DIDA_BROWSER_PROFILE_DIR", t.TempDir())
	// Clear when no file exists should succeed
	if err := ClearCookieToken(); err != nil {
		t.Fatalf("ClearCookieToken() on missing file: error = %v", err)
	}
}
