package webapi

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/textproto"
	"net/url"
	"os"
	"strings"
)

func (c *Client) UploadCommentAttachment(ctx context.Context, projectID string, taskID string, fileName string, contentType string, file io.Reader) (map[string]any, error) {
	path := "/attachment/upload/comment/" + url.PathEscape(projectID) + "/" + url.PathEscape(taskID)
	return c.doV1MultipartFile(ctx, path, "file", fileName, contentType, file)
}

func (c *Client) doV1MultipartFile(ctx context.Context, path string, fieldName string, fileName string, contentType string, file io.Reader) (map[string]any, error) {
	if strings.TrimSpace(c.Token) == "" {
		return nil, fmt.Errorf("missing Dida web cookie token")
	}
	if strings.TrimSpace(c.BaseURLV1) == "" {
		c.BaseURLV1 = DefaultBaseURLV1
	}
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}
	if strings.TrimSpace(contentType) == "" {
		contentType = "application/octet-stream"
	}

	pr, pw := io.Pipe()
	writer := multipart.NewWriter(pw)
	go func() {
		defer pw.Close()
		header := make(textproto.MIMEHeader)
		header.Set("Content-Disposition", fmt.Sprintf(`form-data; name="%s"; filename="%s"`, escapeMultipartValue(fieldName), escapeMultipartValue(fileName)))
		header.Set("Content-Type", contentType)
		part, err := writer.CreatePart(header)
		if err != nil {
			_ = pw.CloseWithError(err)
			return
		}
		if _, err := io.Copy(part, file); err != nil {
			_ = pw.CloseWithError(err)
			return
		}
		if err := writer.Close(); err != nil {
			_ = pw.CloseWithError(err)
			return
		}
	}()

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, strings.TrimRight(c.BaseURLV1, "/")+path, pr)
	if err != nil {
		return nil, fmt.Errorf("build request: %w", err)
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("Cookie", "t="+c.Token)
	req.Header.Set("User-Agent", c.UserAgent)
	req.Header.Set("x-device", c.deviceHeader())

	httpClient := c.HTTPClient
	if httpClient == nil {
		httpClient = http.DefaultClient
	}
	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request %s %s: %w", http.MethodPost, path, err)
	}
	defer resp.Body.Close()

	limit := c.MaxResponseBytes
	if limit <= 0 {
		limit = DefaultMaxResponseBytes
	}
	data, err := io.ReadAll(io.LimitReader(resp.Body, limit+1))
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}
	if int64(len(data)) > limit {
		return nil, fmt.Errorf("dida web api %s %s response exceeded %d bytes", http.MethodPost, path, limit)
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, &APIError{
			Method:      http.MethodPost,
			Path:        path,
			StatusCode:  resp.StatusCode,
			BodySnippet: c.redactForError(string(data)),
			IncludeBody: os.Getenv("DIDA_DEBUG_API_ERRORS") == "1",
		}
	}
	var out map[string]any
	if len(data) > 0 {
		if err := json.Unmarshal(data, &out); err != nil {
			return nil, fmt.Errorf("decode response: %w", err)
		}
	}
	return out, nil
}

func escapeMultipartValue(value string) string {
	return strings.NewReplacer("\\", "\\\\", `"`, "\\\"").Replace(value)
}
