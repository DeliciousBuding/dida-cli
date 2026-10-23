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

	"github.com/DeliciousBuding/dida-cli/internal/auth"
	"github.com/DeliciousBuding/dida-cli/internal/config"
)

const DefaultURL = "https://mcp.dida365.com"
const LatestProtocolVersion = "2026-07-28"
const LegacyProtocolVersion = "2025-11-25"

type Client struct {
	URL                string
	Token              string
	HTTPClient         *http.Client
	SessionID          string
	ProtocolVersion    string
	Stateless          bool
	ToolListCacheScope string
	ToolListTTLMs      int64
	clientName         string
	clientVersion      string
}

type Tool struct {
	Name         string         `json:"name"`
	Title        string         `json:"title,omitempty"`
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
	cfg := &TokenConfig{Token: token, SavedAt: time.Now().Unix()}
	store := auth.NewTokenStore(TokenConfigPath())
	if err := store.Save(cfg); err != nil {
		return nil, err
	}
	return cfg, nil
}

func LoadTokenConfig() (*TokenConfig, error) {
	var cfg TokenConfig
	store := auth.NewTokenStore(TokenConfigPath())
	if err := store.Load(&cfg); err != nil {
		return nil, err
	}
	if strings.TrimSpace(cfg.Token) == "" {
		return nil, fmt.Errorf("official mcp token config has no token")
	}
	return &cfg, nil
}

func ClearTokenConfig() error {
	store := auth.NewTokenStore(TokenConfigPath())
	return store.Clear()
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
	baseURL := os.Getenv("DIDA_MCP_BASE_URL")
	if baseURL == "" {
		baseURL = DefaultURL
	}
	return &Client{
		URL:   baseURL,
		Token: token,
		HTTPClient: &http.Client{
			Timeout: 60 * time.Second,
		},
	}
}

func (c *Client) Initialize(ctx context.Context, clientName string, clientVersion string) error {
	c.clientName = clientName
	c.clientVersion = clientVersion
	if err := c.discover(ctx); err == nil {
		c.Stateless = true
		c.ProtocolVersion = LatestProtocolVersion
		return nil
	}
	c.Stateless = false
	c.ProtocolVersion = LegacyProtocolVersion
	body := map[string]any{
		"jsonrpc": "2.0",
		"id":      1,
		"method":  "initialize",
		"params": map[string]any{
			"protocolVersion": LegacyProtocolVersion,
			"clientInfo": map[string]any{
				"name":    clientName,
				"version": clientVersion,
			},
			"capabilities": map[string]any{},
		},
	}
	resp, headers, err := c.post(ctx, body, requestOptions{})
	if err != nil {
		return err
	}
	if resp.Error != nil {
		return fmt.Errorf("official mcp initialize failed: %s", resp.Error.Message)
	}
	if negotiated := strings.TrimSpace(stringValue(resp.Result["protocolVersion"])); negotiated != "" {
		c.ProtocolVersion = negotiated
	}
	if sessionID := headers.Get("Mcp-Session-Id"); sessionID != "" {
		c.SessionID = sessionID
	} else if sessionID := headers.Get("mcp-session-id"); sessionID != "" {
		c.SessionID = sessionID
	}
	return c.NotifyInitialized(ctx)
}

type requestOptions struct {
	ProtocolVersion string
	Method          string
	ToolName        string
}

func (c *Client) discover(ctx context.Context) error {
	body := map[string]any{
		"jsonrpc": "2.0",
		"id":      1,
		"method":  "server/discover",
		"params": map[string]any{
			"_meta": c.requestMeta(),
		},
	}
	resp, _, err := c.post(ctx, body, requestOptions{
		ProtocolVersion: LatestProtocolVersion,
		Method:          "server/discover",
	})
	if err != nil {
		return err
	}
	if resp.Error != nil {
		return fmt.Errorf("official mcp server/discover failed: %s", resp.Error.Message)
	}
	versions, _ := resp.Result["supportedVersions"].([]any)
	for _, raw := range versions {
		if stringValue(raw) == LatestProtocolVersion {
			return nil
		}
	}
	return fmt.Errorf("official mcp does not advertise protocol %s", LatestProtocolVersion)
}

func (c *Client) NotifyInitialized(ctx context.Context) error {
	body := map[string]any{
		"jsonrpc": "2.0",
		"method":  "notifications/initialized",
		"params":  map[string]any{},
	}
	_, _, err := c.post(ctx, body, requestOptions{ProtocolVersion: c.ProtocolVersion})
	return err
}

func (c *Client) Tools(ctx context.Context) ([]Tool, error) {
	tools := []Tool{}
	cursor := ""
	seenCursors := map[string]bool{}
	for page := 0; page < 100; page++ {
		params := map[string]any{}
		if c.Stateless {
			params["_meta"] = c.requestMeta()
		}
		if cursor != "" {
			params["cursor"] = cursor
		}
		body := map[string]any{
			"jsonrpc": "2.0",
			"id":      2,
			"method":  "tools/list",
			"params":  params,
		}
		resp, _, err := c.post(ctx, body, c.methodOptions("tools/list", ""))
		if err != nil {
			return nil, err
		}
		if resp.Error != nil {
			return nil, fmt.Errorf("official mcp tools/list failed: %s", resp.Error.Message)
		}
		c.ToolListCacheScope = stringValue(resp.Result["cacheScope"])
		c.ToolListTTLMs = int64Value(resp.Result["ttlMs"])
		rawTools, _ := resp.Result["tools"].([]any)
		for _, raw := range rawTools {
			if item, ok := raw.(map[string]any); ok {
				tools = append(tools, Tool{
					Name:         stringValue(item["name"]),
					Title:        stringValue(item["title"]),
					Description:  stringValue(item["description"]),
					InputSchema:  mapValue(item["inputSchema"]),
					OutputSchema: mapValue(item["outputSchema"]),
					Annotations:  mapValue(item["annotations"]),
				})
			}
		}
		nextCursor := stringValue(resp.Result["nextCursor"])
		if nextCursor == "" {
			return tools, nil
		}
		if seenCursors[nextCursor] {
			return nil, fmt.Errorf("official mcp tools/list repeated cursor %q", nextCursor)
		}
		seenCursors[nextCursor] = true
		cursor = nextCursor
	}
	return nil, fmt.Errorf("official mcp tools/list exceeded 100 pages")
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
	if c.Stateless {
		body["params"].(map[string]any)["_meta"] = c.requestMeta()
	}
	resp, _, err := c.post(ctx, body, c.methodOptions("tools/call", name))
	if err != nil {
		return nil, err
	}
	if resp.Error != nil {
		return nil, fmt.Errorf("official mcp tools/call failed: %s", resp.Error.Message)
	}
	if resultType := stringValue(resp.Result["resultType"]); resultType != "" && resultType != "complete" {
		return nil, fmt.Errorf("official mcp tools/call returned resultType %q; multi-round-trip interactions are not supported", resultType)
	}
	if isError, _ := resp.Result["isError"].(bool); isError {
		return nil, fmt.Errorf("official mcp tool %s failed: %s", name, toolErrorMessage(resp.Result))
	}
	return unwrapToolResult(resp.Result), nil
}

func (c *Client) methodOptions(method string, toolName string) requestOptions {
	if !c.Stateless {
		return requestOptions{ProtocolVersion: c.ProtocolVersion}
	}
	return requestOptions{
		ProtocolVersion: c.ProtocolVersion,
		Method:          method,
		ToolName:        toolName,
	}
}

func (c *Client) requestMeta() map[string]any {
	return map[string]any{
		"io.modelcontextprotocol/protocolVersion":    LatestProtocolVersion,
		"io.modelcontextprotocol/clientCapabilities": map[string]any{},
		"io.modelcontextprotocol/clientInfo": map[string]any{
			"name":    c.clientName,
			"version": c.clientVersion,
		},
	}
}

func (c *Client) post(ctx context.Context, payload map[string]any, options requestOptions) (rpcResponse, http.Header, error) {
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
	if options.ProtocolVersion != "" {
		req.Header.Set("MCP-Protocol-Version", options.ProtocolVersion)
	}
	if options.Method != "" {
		req.Header.Set("Mcp-Method", options.Method)
	}
	if options.ToolName != "" {
		req.Header.Set("Mcp-Name", options.ToolName)
	}
	if !c.Stateless && c.SessionID != "" {
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

func int64Value(value any) int64 {
	switch number := value.(type) {
	case float64:
		return int64(number)
	case int64:
		return number
	case int:
		return int64(number)
	default:
		return 0
	}
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

func toolErrorMessage(result map[string]any) string {
	if content, ok := result["content"].([]any); ok {
		for _, raw := range content {
			item, ok := raw.(map[string]any)
			if !ok || stringValue(item["type"]) != "text" {
				continue
			}
			if text := strings.TrimSpace(stringValue(item["text"])); text != "" {
				if len(text) > 4000 {
					return text[:4000] + "..."
				}
				return text
			}
		}
	}
	return "tool returned isError=true without a textual message"
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
