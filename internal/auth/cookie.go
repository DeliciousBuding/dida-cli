package auth

import (
	"fmt"
	"path/filepath"
	"strings"
	"time"
	"unicode"

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
	token, err := NormalizeCookieToken(token)
	if err != nil {
		return nil, err
	}
	item := &CookieToken{Token: token, SavedAt: time.Now().UnixMilli()}
	store := NewTokenStore(CookiePath())
	if err := store.Save(item); err != nil {
		return nil, err
	}
	return item, nil
}

func NormalizeCookieToken(input string) (string, error) {
	token := strings.TrimSpace(input)
	if token == "" {
		return "", fmt.Errorf("empty cookie token")
	}
	lower := strings.ToLower(token)
	if strings.HasPrefix(lower, "cookie:") {
		return "", fmt.Errorf("paste only the Dida365 cookie named 't', not a full Cookie header")
	}
	if strings.Contains(token, ";") {
		return "", fmt.Errorf("paste only the Dida365 cookie named 't', not multiple cookies")
	}
	if strings.HasPrefix(token, "t=") {
		token = strings.TrimSpace(strings.TrimPrefix(token, "t="))
	}
	if token == "" {
		return "", fmt.Errorf("empty cookie token")
	}
	if strings.Contains(token, "=") {
		return "", fmt.Errorf("cookie token must be a single t cookie value")
	}
	for _, r := range token {
		if unicode.IsControl(r) || unicode.IsSpace(r) {
			return "", fmt.Errorf("cookie token contains invalid whitespace or control characters")
		}
	}
	return token, nil
}

func LoadCookieToken() (*CookieToken, error) {
	var item CookieToken
	store := NewTokenStore(CookiePath())
	if err := store.Load(&item); err != nil {
		return nil, err
	}
	if strings.TrimSpace(item.Token) == "" {
		return nil, fmt.Errorf("cookie token file has no token")
	}
	return &item, nil
}

func ClearCookieToken() error {
	store := NewTokenStore(CookiePath())
	if err := store.Clear(); err != nil {
		return err
	}
	if err := ClearBrowserLoginProfile(); err != nil {
		return err
	}
	return nil
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
