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
	"strings"
)

const DefaultBaseURL = "https://api.dida365.com/api/v2"

type Client struct {
	BaseURL    string
	HTTPClient *http.Client
	Token      string
	UserAgent  string
	DeviceID   string
}

func NewClient(token string) *Client {
	return &Client{
		BaseURL:    DefaultBaseURL,
		HTTPClient: http.DefaultClient,
		Token:      token,
		UserAgent:  "Mozilla/5.0 (Macintosh; Intel Mac OS X 10.15; rv:95.0) Gecko/20100101 Firefox/95.0",
		DeviceID:   randomDeviceID(),
	}
}

func (c *Client) Do(ctx context.Context, method string, path string, body any, out any) error {
	if strings.TrimSpace(c.Token) == "" {
		return fmt.Errorf("missing Dida web cookie token")
	}
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}

	var reader io.Reader
	if body != nil {
		payload, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("encode request body: %w", err)
		}
		reader = bytes.NewReader(payload)
	}

	req, err := http.NewRequestWithContext(ctx, method, strings.TrimRight(c.BaseURL, "/")+path, reader)
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

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("read response: %w", err)
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("dida web api %s %s returned %d: %s", method, path, resp.StatusCode, redactForError(string(data)))
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

func redactForError(value string) string {
	value = strings.TrimSpace(value)
	if len(value) > 500 {
		value = value[:500] + "..."
	}
	return value
}
