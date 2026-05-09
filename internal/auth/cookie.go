package auth

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/DeliciousBuding/dida-cli/internal/config"
)

type CookieToken struct {
	Token   string `json:"token"`
	SavedAt int64  `json:"saved_at"`
}

func CookiePath() string {
	return filepath.Join(config.DefaultDir(), "cookie.json")
}

func SaveCookieToken(token string) (*CookieToken, error) {
	token = strings.TrimSpace(token)
	if token == "" {
		return nil, fmt.Errorf("empty cookie token")
	}
	if err := os.MkdirAll(config.DefaultDir(), 0o700); err != nil {
		return nil, fmt.Errorf("create config dir: %w", err)
	}
	item := &CookieToken{Token: token, SavedAt: time.Now().UnixMilli()}
	payload, err := json.MarshalIndent(item, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("encode token: %w", err)
	}
	if err := os.WriteFile(CookiePath(), append(payload, '\n'), 0o600); err != nil {
		return nil, fmt.Errorf("write cookie token: %w", err)
	}
	return item, nil
}

func LoadCookieToken() (*CookieToken, error) {
	payload, err := os.ReadFile(CookiePath())
	if err != nil {
		return nil, err
	}
	var item CookieToken
	if err := json.Unmarshal(payload, &item); err != nil {
		return nil, fmt.Errorf("decode cookie token: %w", err)
	}
	if strings.TrimSpace(item.Token) == "" {
		return nil, fmt.Errorf("cookie token file has no token")
	}
	return &item, nil
}

func CookieStatus() map[string]any {
	item, err := LoadCookieToken()
	status := map[string]any{
		"path": CookiePath(),
	}
	if err != nil {
		status["available"] = false
		status["message"] = "missing"
		return status
	}
	status["available"] = true
	status["saved_at"] = time.UnixMilli(item.SavedAt).Format(time.RFC3339)
	status["token_length"] = len(item.Token)
	status["token_preview"] = RedactToken(item.Token)
	return status
}

func RedactToken(token string) string {
	token = strings.TrimSpace(token)
	if len(token) <= 8 {
		return "***"
	}
	return token[:4] + "..." + token[len(token)-4:]
}
