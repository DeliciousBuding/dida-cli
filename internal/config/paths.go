package config

import (
	"os"
	"path/filepath"
)

func DefaultDir() string {
	if value := os.Getenv("DIDA_CONFIG_DIR"); value != "" {
		return value
	}
	home, err := os.UserHomeDir()
	if err != nil || home == "" {
		return ".dida-cli"
	}
	return filepath.Join(home, ".dida-cli")
}
