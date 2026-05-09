package openapi

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/DeliciousBuding/dida-cli/internal/config"
)

const (
	DefaultAuthBaseURL = "https://dida365.com"
	DefaultAPIBaseURL  = "https://api.dida365.com/open/v1"
	DefaultScopes      = "tasks:read tasks:write"
	DefaultTokenType   = "Bearer"
)

type OAuthToken struct {
	AccessToken  string `json:"access_token"`
	TokenType    string `json:"token_type,omitempty"`
	Scope        string `json:"scope,omitempty"`
	ExpiresIn    int64  `json:"expires_in,omitempty"`
	CreatedAt    int64  `json:"created_at"`
	RefreshToken string `json:"refresh_token,omitempty"`
}

type TokenResponse struct {
	OAuthToken
}

type ClientConfig struct {
	ClientID     string `json:"client_id"`
	ClientSecret string `json:"client_secret"`
	SavedAt      int64  `json:"saved_at"`
}

func TokenPath() string {
	return filepath.Join(config.DefaultDir(), "openapi-oauth.json")
}

func ClientConfigPath() string {
	return filepath.Join(config.DefaultDir(), "openapi-client.json")
}

func ResolveClientID(explicit string) (string, error) {
	if strings.TrimSpace(explicit) != "" {
		return strings.TrimSpace(explicit), nil
	}
	if value := strings.TrimSpace(os.Getenv("DIDA365_OPENAPI_CLIENT_ID")); value != "" {
		return value, nil
	}
	if cfg, err := LoadClientConfig(); err == nil && strings.TrimSpace(cfg.ClientID) != "" {
		return strings.TrimSpace(cfg.ClientID), nil
	}
	return "", fmt.Errorf("missing DIDA365_OPENAPI_CLIENT_ID")
}

func ResolveClientSecret(explicit string) (string, error) {
	if strings.TrimSpace(explicit) != "" {
		return strings.TrimSpace(explicit), nil
	}
	if value := strings.TrimSpace(os.Getenv("DIDA365_OPENAPI_CLIENT_SECRET")); value != "" {
		return value, nil
	}
	if cfg, err := LoadClientConfig(); err == nil && strings.TrimSpace(cfg.ClientSecret) != "" {
		return strings.TrimSpace(cfg.ClientSecret), nil
	}
	return "", fmt.Errorf("missing DIDA365_OPENAPI_CLIENT_SECRET")
}

func SaveClientConfig(clientID string, clientSecret string) (*ClientConfig, error) {
	clientID = strings.TrimSpace(clientID)
	clientSecret = strings.TrimSpace(clientSecret)
	if clientID == "" {
		return nil, fmt.Errorf("empty openapi client id")
	}
	if clientSecret == "" {
		return nil, fmt.Errorf("empty openapi client secret")
	}
	if err := os.MkdirAll(config.DefaultDir(), 0o700); err != nil {
		return nil, fmt.Errorf("create config dir: %w", err)
	}
	cfg := &ClientConfig{ClientID: clientID, ClientSecret: clientSecret, SavedAt: time.Now().Unix()}
	payload, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("encode openapi client config: %w", err)
	}
	if err := os.WriteFile(ClientConfigPath(), append(payload, '\n'), 0o600); err != nil {
		return nil, fmt.Errorf("write openapi client config: %w", err)
	}
	return cfg, nil
}

func LoadClientConfig() (*ClientConfig, error) {
	data, err := os.ReadFile(ClientConfigPath())
	if err != nil {
		return nil, err
	}
	var cfg ClientConfig
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("decode openapi client config: %w", err)
	}
	if strings.TrimSpace(cfg.ClientID) == "" {
		return nil, fmt.Errorf("openapi client config has no client id")
	}
	if strings.TrimSpace(cfg.ClientSecret) == "" {
		return nil, fmt.Errorf("openapi client config has no client secret")
	}
	return &cfg, nil
}

func ClearClientConfig() error {
	if err := os.Remove(ClientConfigPath()); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("remove openapi client config: %w", err)
	}
	return nil
}

func ClientConfigStatus() map[string]any {
	status := map[string]any{}
	cfg, err := LoadClientConfig()
	if err != nil {
		status["available"] = false
		status["message"] = "missing"
		return status
	}
	status["available"] = true
	status["client_id_preview"] = redactToken(cfg.ClientID)
	status["client_secret_available"] = true
	if cfg.SavedAt > 0 {
		status["saved_at"] = time.Unix(cfg.SavedAt, 0).Format(time.RFC3339)
	}
	return status
}

func AuthorizationURL(clientID string, redirectURI string, scope string, state string) string {
	values := url.Values{}
	values.Set("client_id", clientID)
	values.Set("scope", scope)
	values.Set("state", state)
	values.Set("redirect_uri", redirectURI)
	values.Set("response_type", "code")
	return DefaultAuthBaseURL + "/oauth/authorize?" + values.Encode()
}

func ExchangeCode(ctx context.Context, clientID string, clientSecret string, code string, redirectURI string, scope string) (*TokenResponse, error) {
	values := url.Values{}
	values.Set("grant_type", "authorization_code")
	values.Set("code", code)
	values.Set("scope", scope)
	values.Set("redirect_uri", redirectURI)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, DefaultAuthBaseURL+"/oauth/token", strings.NewReader(values.Encode()))
	if err != nil {
		return nil, fmt.Errorf("build token request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Authorization", "Basic "+basicAuth(clientID, clientSecret))
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("exchange oauth code: %w", err)
	}
	defer resp.Body.Close()
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read token response: %w", err)
	}
	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("oauth token exchange returned HTTP %d: %s", resp.StatusCode, summarizeBody(string(data)))
	}
	var out TokenResponse
	if err := json.Unmarshal(data, &out); err != nil {
		return nil, fmt.Errorf("decode token response: %w", err)
	}
	out.CreatedAt = time.Now().Unix()
	if out.TokenType == "" {
		out.TokenType = DefaultTokenType
	}
	return &out, nil
}

func SaveToken(token *TokenResponse) error {
	if token == nil || strings.TrimSpace(token.AccessToken) == "" {
		return fmt.Errorf("empty oauth token")
	}
	if err := os.MkdirAll(config.DefaultDir(), 0o700); err != nil {
		return fmt.Errorf("create config dir: %w", err)
	}
	payload, err := json.MarshalIndent(token, "", "  ")
	if err != nil {
		return fmt.Errorf("encode oauth token: %w", err)
	}
	if err := os.WriteFile(TokenPath(), append(payload, '\n'), 0o600); err != nil {
		return fmt.Errorf("write oauth token: %w", err)
	}
	return nil
}

func LoadToken() (*TokenResponse, error) {
	data, err := os.ReadFile(TokenPath())
	if err != nil {
		return nil, err
	}
	var out TokenResponse
	if err := json.Unmarshal(data, &out); err != nil {
		return nil, fmt.Errorf("decode oauth token: %w", err)
	}
	if strings.TrimSpace(out.AccessToken) == "" {
		return nil, fmt.Errorf("oauth token file has no access token")
	}
	return &out, nil
}

func ClearToken() error {
	if err := os.Remove(TokenPath()); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("remove oauth token: %w", err)
	}
	return nil
}

func TokenStatus() map[string]any {
	status := map[string]any{}
	token, err := LoadToken()
	if err != nil {
		status["available"] = false
		status["message"] = "missing"
		return status
	}
	status["available"] = true
	status["token_type"] = token.TokenType
	status["scope"] = token.Scope
	status["created_at"] = time.Unix(token.CreatedAt, 0).Format(time.RFC3339)
	status["token_preview"] = redactToken(token.AccessToken)
	if token.ExpiresIn > 0 {
		status["expires_in"] = token.ExpiresIn
	}
	return status
}

func basicAuth(user string, pass string) string {
	raw := user + ":" + pass
	return base64.StdEncoding.EncodeToString([]byte(raw))
}

func summarizeBody(body string) string {
	body = strings.TrimSpace(body)
	if len(body) > 300 {
		return body[:300] + "..."
	}
	return body
}

func redactToken(token string) string {
	token = strings.TrimSpace(token)
	if len(token) <= 8 {
		return "***"
	}
	return token[:4] + "..." + token[len(token)-4:]
}

func RedactForStatus(value string) string {
	return redactToken(value)
}
