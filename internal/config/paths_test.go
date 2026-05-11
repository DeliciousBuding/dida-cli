package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDefaultDir_EnvOverride(t *testing.T) {
	t.Setenv("DIDA_CONFIG_DIR", "/custom/path")
	got := DefaultDir()
	if got != "/custom/path" {
		t.Errorf("DefaultDir() with DIDA_CONFIG_DIR=/custom/path = %q, want %q", got, "/custom/path")
	}
}

func TestDefaultDir_FallbackOnMissingHome(t *testing.T) {
	t.Setenv("DIDA_CONFIG_DIR", "")
	// UserHomeDir should work in test env; just verify the result is under home.
	got := DefaultDir()
	home, _ := os.UserHomeDir()
	want := filepath.Join(home, ".dida-cli")
	if got != want {
		t.Errorf("DefaultDir() = %q, want %q", got, want)
	}
}

func TestDefaultDir_EnvTakesPrecedence(t *testing.T) {
	t.Setenv("DIDA_CONFIG_DIR", "override")
	got := DefaultDir()
	if got != "override" {
		t.Errorf("DefaultDir() = %q, want %q", got, "override")
	}
}
