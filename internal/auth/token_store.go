package auth

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// TokenStore provides a unified save/load/clear interface for JSON token files.
// Each auth channel instantiates one with its own file path; the exported
// function signatures remain unchanged for backward compatibility.
type TokenStore struct {
	Path string
}

// NewTokenStore creates a TokenStore rooted at config.DefaultDir().
// The caller is responsible for providing the full path including filename.
func NewTokenStore(path string) *TokenStore {
	return &TokenStore{Path: path}
}

// Save marshals data with indentation and writes it atomically via a temp file
// followed by rename. The parent directory is created if missing.
func (s *TokenStore) Save(data any) error {
	dir := filepath.Dir(s.Path)
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return fmt.Errorf("create config dir: %w", err)
	}
	payload, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return fmt.Errorf("encode token: %w", err)
	}
	if err := os.WriteFile(s.Path, append(payload, '\n'), 0o600); err != nil {
		return fmt.Errorf("write token file: %w", err)
	}
	return nil
}

// Load reads and unmarshals the JSON token file into target, which must be a
// pointer to the expected struct type.
func (s *TokenStore) Load(target any) error {
	data, err := os.ReadFile(s.Path)
	if err != nil {
		return err
	}
	if err := json.Unmarshal(data, target); err != nil {
		return fmt.Errorf("decode token: %w", err)
	}
	return nil
}

// Clear removes the token file. It is idempotent: a missing file is not an error.
func (s *TokenStore) Clear() error {
	if err := os.Remove(s.Path); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("remove token file: %w", err)
	}
	return nil
}
