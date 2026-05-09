package auth

import (
	"os"
	"path/filepath"
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
