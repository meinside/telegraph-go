package telegraph

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

const (
	// maxResponseBytes caps how much of a response body is read. The largest
	// legitimate answers (a page with 64 KB of content, a 200-page list) are far below it.
	maxResponseBytes = 16 << 20
	// maxErrorBodyBytes caps the body quoted in an HTTPError.
	maxErrorBodyBytes = 512
)

type apiResponse struct {
	Ok     bool            `json:"ok"`
	Error  string          `json:"error,omitempty"`
	Result json.RawMessage `json:"result,omitempty"`
}

// call POSTs params to method (and path, when not empty) and decodes the result into T.
func call[T any](ctx context.Context, c *Client, method, path string, params url.Values) (result T, err error) {
	endpoint := c.baseURL + "/" + method
	if path != "" {
		endpoint += "/" + url.PathEscape(path)
	}

	c.logger.DebugContext(ctx, "telegraph request", "method", method, "path", path, "params", redact(params))

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, strings.NewReader(params.Encode()))
	if err != nil {
		return result, fmt.Errorf("telegraph: %s: build request: %w", method, err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	res, err := c.httpClient.Do(req)
	if err != nil {
		return result, fmt.Errorf("telegraph: %s: %w", method, err)
	}
	defer func() { _ = res.Body.Close() }()

	body, err := io.ReadAll(io.LimitReader(res.Body, maxResponseBytes+1))
	if err != nil {
		return result, fmt.Errorf("telegraph: %s: read response: %w", method, err)
	}
	if len(body) > maxResponseBytes {
		return result, fmt.Errorf("telegraph: %s: response exceeds %d bytes", method, maxResponseBytes)
	}

	statusOK := res.StatusCode >= 200 && res.StatusCode < 300

	var r apiResponse
	if err := json.Unmarshal(body, &r); err != nil {
		if !statusOK {
			return result, httpError(method, res.StatusCode, body)
		}
		return result, fmt.Errorf("telegraph: %s: decode response: %w", method, err)
	}
	if !r.Ok {
		if r.Error == "" && !statusOK {
			return result, httpError(method, res.StatusCode, body)
		}
		return result, &APIError{Method: method, Code: r.Error}
	}
	if !statusOK {
		return result, httpError(method, res.StatusCode, body)
	}

	if err := json.Unmarshal(r.Result, &result); err != nil {
		return result, fmt.Errorf("telegraph: %s: decode result: %w", method, err)
	}
	return result, nil
}

func httpError(method string, status int, body []byte) *HTTPError {
	if len(body) > maxErrorBodyBytes {
		body = body[:maxErrorBodyBytes]
	}
	return &HTTPError{Method: method, StatusCode: status, Body: string(body)}
}

// redact returns params with the access token hidden, for logging.
func redact(params url.Values) url.Values {
	if _, ok := params["access_token"]; !ok {
		return params
	}
	out := make(url.Values, len(params))
	for k, v := range params {
		out[k] = v
	}
	out["access_token"] = []string{"[REDACTED]"}
	return out
}
