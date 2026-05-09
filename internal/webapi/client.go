package webapi

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"regexp"
	"strings"
)

const DefaultBaseURL = "https://api.dida365.com/api/v2"
const DefaultBaseURLV1 = "https://api.dida365.com/api/v1"
const DefaultMaxResponseBytes int64 = 16 << 20

type Client struct {
	BaseURL          string
	BaseURLV1        string
	HTTPClient       *http.Client
	Token            string
	UserAgent        string
	DeviceID         string
	MaxResponseBytes int64
}

func NewClient(token string) *Client {
	return &Client{
		BaseURL:          DefaultBaseURL,
		BaseURLV1:        DefaultBaseURLV1,
		HTTPClient:       http.DefaultClient,
		Token:            token,
		UserAgent:        "Mozilla/5.0 (Macintosh; Intel Mac OS X 10.15; rv:95.0) Gecko/20100101 Firefox/95.0",
		DeviceID:         randomDeviceID(),
		MaxResponseBytes: DefaultMaxResponseBytes,
	}
}

func (c *Client) Do(ctx context.Context, method string, path string, body any, out any) error {
	return c.doAtBaseURL(ctx, c.BaseURL, method, path, body, out)
}

func (c *Client) DoV1(ctx context.Context, method string, path string, body any, out any) error {
	return c.doAtBaseURL(ctx, c.BaseURLV1, method, path, body, out)
}

func (c *Client) doAtBaseURL(ctx context.Context, baseURL string, method string, path string, body any, out any) error {
	if strings.TrimSpace(c.Token) == "" {
		return fmt.Errorf("missing Dida web cookie token")
	}
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}
	if strings.TrimSpace(baseURL) == "" {
		baseURL = DefaultBaseURL
	}

	var reader io.Reader
	if body != nil {
		payload, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("encode request body: %w", err)
		}
		reader = bytes.NewReader(payload)
	}

	req, err := http.NewRequestWithContext(ctx, method, strings.TrimRight(baseURL, "/")+path, reader)
	if err != nil {
		return fmt.Errorf("build request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Cookie", "t="+c.Token)
	req.Header.Set("User-Agent", c.UserAgent)
	req.Header.Set("x-device", c.deviceHeader())

	httpClient := c.HTTPClient
	if httpClient == nil {
		httpClient = http.DefaultClient
	}
	resp, err := httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("request %s %s: %w", method, path, err)
	}
	defer resp.Body.Close()

	limit := c.MaxResponseBytes
	if limit <= 0 {
		limit = DefaultMaxResponseBytes
	}
	data, err := io.ReadAll(io.LimitReader(resp.Body, limit+1))
	if err != nil {
		return fmt.Errorf("read response: %w", err)
	}
	if int64(len(data)) > limit {
		return fmt.Errorf("dida web api %s %s response exceeded %d bytes", method, path, limit)
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		message := fmt.Sprintf("dida web api %s %s returned %d", method, path, resp.StatusCode)
		if os.Getenv("DIDA_DEBUG_API_ERRORS") == "1" {
			message += ": " + c.redactForError(string(data))
		}
		return fmt.Errorf("%s", message)
	}
	if out == nil || len(data) == 0 {
		return nil
	}
	if err := json.Unmarshal(data, out); err != nil {
		return fmt.Errorf("decode response: %w", err)
	}
	return nil
}

func (c *Client) deviceHeader() string {
	device := map[string]any{
		"platform":  "web",
		"os":        "OS X",
		"device":    "Firefox 95.0",
		"name":      "DidaCLI",
		"version":   4531,
		"id":        c.DeviceID,
		"channel":   "website",
		"campaign":  "",
		"websocket": "",
	}
	payload, _ := json.Marshal(device)
	return string(payload)
}

func randomDeviceID() string {
	var buf [10]byte
	if _, err := rand.Read(buf[:]); err != nil {
		return "649000000000000000000000"
	}
	return "6490" + hex.EncodeToString(buf[:])
}

func (c *Client) redactForError(value string) string {
	value = strings.TrimSpace(value)
	if c != nil && strings.TrimSpace(c.Token) != "" {
		value = strings.ReplaceAll(value, c.Token, "[REDACTED]")
	}
	value = redactSensitivePatterns(value)
	if len(value) > 500 {
		value = value[:500] + "..."
	}
	return value
}

func redactSensitivePatterns(value string) string {
	patterns := []*regexp.Regexp{
		regexp.MustCompile(`(?i)(cookie\s*:\s*)[^\r\n]+`),
		regexp.MustCompile(`(?i)(set-cookie\s*:\s*)[^\r\n]+`),
		regexp.MustCompile(`(?i)(authorization\s*:\s*bearer\s+)[A-Za-z0-9._~+\-/=]+`),
		regexp.MustCompile(`(?i)(["']?(?:token|access_token|refresh_token|cookie)["']?\s*[:=]\s*["']?)[^"',\s}]+`),
		regexp.MustCompile(`(?i)(\bt=)[^;\s"',}]+`),
	}
	for _, pattern := range patterns {
		value = pattern.ReplaceAllString(value, `${1}[REDACTED]`)
	}
	return value
}
