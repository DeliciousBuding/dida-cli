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
	"strings"
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
	return "", fmt.Errorf("missing DIDA365_TOKEN; set it to the official API token from dida365 account settings")
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
