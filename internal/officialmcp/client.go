package officialmcp

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/DeliciousBuding/dida-cli/internal/config"
)

const DefaultURL = "https://mcp.dida365.com"
const ProtocolVersion = "2024-11-05"
const InitProtocolVersion = "2025-03-26"

type Client struct {
	URL        string
	Token      string
	HTTPClient *http.Client
	SessionID  string
}

type Tool struct {
	Name         string         `json:"name"`
	Description  string         `json:"description,omitempty"`
	InputSchema  map[string]any `json:"inputSchema,omitempty"`
	OutputSchema map[string]any `json:"outputSchema,omitempty"`
	Annotations  map[string]any `json:"annotations,omitempty"`
}

type TokenConfig struct {
	Token   string `json:"token"`
	SavedAt int64  `json:"saved_at"`
}

type rpcResponse struct {
	Result map[string]any `json:"result"`
	Error  *rpcError      `json:"error"`
}

type rpcError struct {
	Code    any    `json:"code"`
	Message string `json:"message"`
}

func ResolveToken(explicit string) (string, error) {
	if strings.TrimSpace(explicit) != "" {
		return explicit, nil
	}
	if value := strings.TrimSpace(os.Getenv("DIDA365_TOKEN")); value != "" {
		return value, nil
	}
	if cfg, err := LoadTokenConfig(); err == nil && strings.TrimSpace(cfg.Token) != "" {
		return strings.TrimSpace(cfg.Token), nil
	}
	return "", fmt.Errorf("missing official mcp token; set DIDA365_TOKEN or run dida official token set --token-stdin")
}

func TokenConfigPath() string {
	return filepath.Join(config.DefaultDir(), "official-mcp-token.json")
}

func SaveTokenConfig(token string) (*TokenConfig, error) {
	token = strings.TrimSpace(token)
	if token == "" {
		return nil, fmt.Errorf("empty official mcp token")
	}
	if err := os.MkdirAll(config.DefaultDir(), 0o700); err != nil {
		return nil, fmt.Errorf("create config dir: %w", err)
	}
	cfg := &TokenConfig{Token: token, SavedAt: time.Now().Unix()}
	payload, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("encode official mcp token config: %w", err)
	}
	if err := os.WriteFile(TokenConfigPath(), append(payload, '\n'), 0o600); err != nil {
		return nil, fmt.Errorf("write official mcp token config: %w", err)
	}
	return cfg, nil
}

func LoadTokenConfig() (*TokenConfig, error) {
	data, err := os.ReadFile(TokenConfigPath())
	if err != nil {
		return nil, err
	}
	var cfg TokenConfig
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("decode official mcp token config: %w", err)
	}
	if strings.TrimSpace(cfg.Token) == "" {
		return nil, fmt.Errorf("official mcp token config has no token")
	}
	return &cfg, nil
}

func ClearTokenConfig() error {
	if err := os.Remove(TokenConfigPath()); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("remove official mcp token config: %w", err)
	}
	return nil
}

func TokenConfigStatus() map[string]any {
	status := map[string]any{}
	if strings.TrimSpace(os.Getenv("DIDA365_TOKEN")) != "" {
		status["available"] = true
		status["source"] = "env"
		return status
	}
	cfg, err := LoadTokenConfig()
	if err != nil {
		status["available"] = false
		status["source"] = "missing"
		return status
	}
	status["available"] = true
	status["source"] = "config"
	status["token_preview"] = RedactForStatus(cfg.Token)
	if cfg.SavedAt > 0 {
		status["saved_at"] = time.Unix(cfg.SavedAt, 0).Format(time.RFC3339)
	}
	return status
}

func RedactForStatus(value string) string {
	value = strings.TrimSpace(value)
	if len(value) <= 8 {
		return "***"
	}
	return value[:4] + "..." + value[len(value)-4:]
}

func NewClient(token string) *Client {
	return &Client{
		URL:        DefaultURL,
		Token:      token,
		HTTPClient: http.DefaultClient,
	}
}

func (c *Client) Initialize(ctx context.Context, clientName string, clientVersion string) error {
	body := map[string]any{
		"jsonrpc": "2.0",
		"id":      1,
		"method":  "initialize",
		"params": map[string]any{
			"protocolVersion": InitProtocolVersion,
			"clientInfo": map[string]any{
				"name":    clientName,
				"version": clientVersion,
			},
			"capabilities": map[string]any{},
		},
	}
	resp, headers, err := c.post(ctx, body, false)
	if err != nil {
		return err
	}
	if resp.Error != nil {
		return fmt.Errorf("official mcp initialize failed: %s", resp.Error.Message)
	}
	if sessionID := headers.Get("Mcp-Session-Id"); sessionID != "" {
		c.SessionID = sessionID
	} else if sessionID := headers.Get("mcp-session-id"); sessionID != "" {
		c.SessionID = sessionID
	}
	return c.NotifyInitialized(ctx)
}

func (c *Client) NotifyInitialized(ctx context.Context) error {
	body := map[string]any{
		"jsonrpc": "2.0",
		"method":  "notifications/initialized",
		"params":  map[string]any{},
	}
	_, _, err := c.post(ctx, body, true)
	return err
}

func (c *Client) Tools(ctx context.Context) ([]Tool, error) {
	body := map[string]any{
		"jsonrpc": "2.0",
		"id":      2,
		"method":  "tools/list",
		"params":  map[string]any{},
	}
	resp, _, err := c.post(ctx, body, true)
	if err != nil {
		return nil, err
	}
	if resp.Error != nil {
		return nil, fmt.Errorf("official mcp tools/list failed: %s", resp.Error.Message)
	}
	rawTools, _ := resp.Result["tools"].([]any)
	tools := make([]Tool, 0, len(rawTools))
	for _, raw := range rawTools {
		if item, ok := raw.(map[string]any); ok {
			tools = append(tools, Tool{
				Name:         stringValue(item["name"]),
				Description:  stringValue(item["description"]),
				InputSchema:  mapValue(item["inputSchema"]),
				OutputSchema: mapValue(item["outputSchema"]),
				Annotations:  mapValue(item["annotations"]),
			})
		}
	}
	return tools, nil
}

func (c *Client) ToolSchema(ctx context.Context, name string) (*Tool, error) {
	tools, err := c.Tools(ctx)
	if err != nil {
		return nil, err
	}
	for _, tool := range tools {
		if tool.Name == name {
			item := tool
			return &item, nil
		}
	}
	return nil, fmt.Errorf("official mcp tool %q not found", name)
}

func (c *Client) CallTool(ctx context.Context, name string, arguments map[string]any) (any, error) {
	body := map[string]any{
		"jsonrpc": "2.0",
		"id":      3,
		"method":  "tools/call",
		"params": map[string]any{
			"name":      name,
			"arguments": arguments,
		},
	}
	resp, _, err := c.post(ctx, body, true)
	if err != nil {
		return nil, err
	}
	if resp.Error != nil {
		return nil, fmt.Errorf("official mcp tools/call failed: %s", resp.Error.Message)
	}
	return unwrapToolResult(resp.Result), nil
}

func (c *Client) post(ctx context.Context, payload map[string]any, includeProtocol bool) (rpcResponse, http.Header, error) {
	data, err := json.Marshal(payload)
	if err != nil {
		return rpcResponse{}, nil, fmt.Errorf("encode request: %w", err)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.URL, bytes.NewReader(data))
	if err != nil {
		return rpcResponse{}, nil, fmt.Errorf("build request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+c.Token)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json, text/event-stream")
	if includeProtocol {
		req.Header.Set("MCP-Protocol-Version", ProtocolVersion)
	}
	if c.SessionID != "" {
		req.Header.Set("Mcp-Session-Id", c.SessionID)
	}
	httpClient := c.HTTPClient
	if httpClient == nil {
		httpClient = http.DefaultClient
	}
	resp, err := httpClient.Do(req)
	if err != nil {
		return rpcResponse{}, nil, fmt.Errorf("official mcp request failed: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusUnauthorized || resp.StatusCode == http.StatusForbidden {
		return rpcResponse{}, resp.Header, fmt.Errorf("official mcp rejected token with HTTP %d", resp.StatusCode)
	}
	if resp.StatusCode >= 400 {
		return rpcResponse{}, resp.Header, fmt.Errorf("official mcp returned HTTP %d", resp.StatusCode)
	}
	if payload["id"] == nil && (resp.StatusCode == http.StatusNoContent || resp.ContentLength == 0) {
		return rpcResponse{}, resp.Header, nil
	}
	parsed, err := parseRPCResponse(resp)
	if err != nil {
		return rpcResponse{}, resp.Header, err
	}
	return parsed, resp.Header, nil
}

func parseRPCResponse(resp *http.Response) (rpcResponse, error) {
	if strings.Contains(resp.Header.Get("Content-Type"), "text/event-stream") {
		return parseSSEResponse(resp.Body)
	}
	var out rpcResponse
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return rpcResponse{}, fmt.Errorf("decode official mcp response: %w", err)
	}
	return out, nil
}

func parseSSEResponse(body io.Reader) (rpcResponse, error) {
	scanner := bufio.NewScanner(body)
	buffer := make([]string, 0, 8)
	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			if len(buffer) == 0 {
				continue
			}
			payload := strings.Join(buffer, "\n")
			var out rpcResponse
			if err := json.Unmarshal([]byte(payload), &out); err == nil {
				return out, nil
			}
			buffer = buffer[:0]
			continue
		}
		if strings.HasPrefix(line, "data:") {
			buffer = append(buffer, strings.TrimSpace(strings.TrimPrefix(line, "data:")))
		}
	}
	if err := scanner.Err(); err != nil {
		return rpcResponse{}, fmt.Errorf("read official mcp sse response: %w", err)
	}
	return rpcResponse{}, fmt.Errorf("empty official mcp response")
}

func stringValue(value any) string {
	text, _ := value.(string)
	return text
}

func mapValue(value any) map[string]any {
	item, _ := value.(map[string]any)
	return item
}

func unwrapToolResult(result map[string]any) any {
	if structured, ok := result["structuredContent"]; ok && structured != nil {
		return unwrapEnvelope(structured)
	}
	if content, ok := result["content"].([]any); ok && len(content) == 1 {
		if item, ok := content[0].(map[string]any); ok && stringValue(item["type"]) == "text" {
			text := stringValue(item["text"])
			var decoded any
			if err := json.Unmarshal([]byte(text), &decoded); err == nil {
				return unwrapEnvelope(decoded)
			}
			return text
		}
	}
	return unwrapEnvelope(result)
}

func unwrapEnvelope(value any) any {
	for {
		item, ok := value.(map[string]any)
		if !ok || len(item) != 1 {
			return value
		}
		if next, ok := item["result"]; ok {
			value = next
			continue
		}
		if next, ok := item["results"]; ok {
			value = next
			continue
		}
		return value
	}
}
