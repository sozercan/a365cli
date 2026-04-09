package mcp

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"strconv"
	"strings"
	"sync/atomic"
	"time"

	"github.com/sozercan/a365cli/internal/config"
	"github.com/sozercan/a365cli/internal/version"
)

// TokenProvider returns a bearer token for authorization.
type TokenProvider func(ctx context.Context) (string, error)

// VerboseLogger is called with request/response details when verbose mode is on.
type VerboseLogger func(format string, args ...any)

// Client is a lightweight MCP JSON-RPC client that speaks HTTP+SSE to agent365.
type Client struct {
	endpoint       string
	tokenProvider  TokenProvider
	httpClient     *http.Client
	sessionID      string
	nextID         atomic.Int64
	verbose        VerboseLogger
	maxRetries     int
	retryBaseDelay time.Duration
}

// NewClient creates a new MCP client for the given endpoint.
func NewClient(endpoint string, tokenProvider TokenProvider) *Client {
	return &Client{
		endpoint:      endpoint,
		tokenProvider: tokenProvider,
		httpClient: &http.Client{
			Transport: &http.Transport{
				DialContext:           (&net.Dialer{Timeout: 30 * time.Second}).DialContext,
				TLSHandshakeTimeout:   10 * time.Second,
				ResponseHeaderTimeout: 60 * time.Second,
			},
		},
		maxRetries:     2,
		retryBaseDelay: time.Second,
	}
}

// SetVerbose enables verbose logging of MCP requests and responses.
func (c *Client) SetVerbose(logger VerboseLogger) {
	c.verbose = logger
}

func (c *Client) logf(format string, args ...any) {
	if c.verbose != nil {
		c.verbose(format, args...)
	}
}

func (c *Client) nextRequestID() int {
	return int(c.nextID.Add(1))
}

// Initialize performs the MCP initialize handshake and stores the session ID.
// If a valid cached session exists for this endpoint, the handshake is skipped.
func (c *Client) Initialize(ctx context.Context) error {
	// Try to reuse a cached session (best-effort).
	if sid, ok := LoadSession(c.endpoint); ok {
		c.sessionID = sid
		c.logf("--- MCP reusing cached session for %s", c.endpoint)
		return nil
	}

	if err := c.doInitialize(ctx); err != nil {
		return err
	}

	// Cache the session for future invocations (best-effort).
	if c.sessionID != "" {
		SaveSession(c.endpoint, c.sessionID)
	}

	return nil
}

// doInitialize performs the raw MCP initialize handshake.
func (c *Client) doInitialize(ctx context.Context) error {
	params := InitializeParams{
		ProtocolVersion: "2024-11-05",
		ClientInfo: ClientInfo{
			Name:    "a365",
			Version: version.Version,
		},
	}

	req := JSONRPCRequest{
		JSONRPC: "2.0",
		ID:      c.nextRequestID(),
		Method:  "initialize",
		Params:  params,
	}

	resp, err := c.doRequest(ctx, req)
	if err != nil {
		return fmt.Errorf("initialize: %w", err)
	}
	if resp.Error != nil {
		return fmt.Errorf("initialize: %w", resp.Error)
	}

	return nil
}

// CallTool invokes a named MCP tool and returns the response.
// If the request fails with a session-related error, the cached session is
// cleared and the client re-initializes before retrying the call once.
func (c *Client) CallTool(ctx context.Context, name string, args map[string]any) (*JSONRPCResponse, error) {
	resp, err := c.callToolOnce(ctx, name, args)
	if isSessionError(resp, err) {
		c.logf("--- MCP session invalid, re-initializing for %s", c.endpoint)
		c.sessionID = ""
		ClearSession(c.endpoint)
		if initErr := c.doInitialize(ctx); initErr != nil {
			return nil, initErr
		}
		if c.sessionID != "" {
			SaveSession(c.endpoint, c.sessionID)
		}
		return c.callToolOnce(ctx, name, args)
	}
	return resp, err
}

func (c *Client) callToolOnce(ctx context.Context, name string, args map[string]any) (*JSONRPCResponse, error) {
	req := JSONRPCRequest{
		JSONRPC: "2.0",
		ID:      c.nextRequestID(),
		Method:  "tools/call",
		Params: ToolCallParams{
			Name:      name,
			Arguments: args,
		},
	}

	return c.doRequest(ctx, req)
}

// ListTools queries the server for available tools.
func (c *Client) ListTools(ctx context.Context) (*JSONRPCResponse, error) {
	req := JSONRPCRequest{
		JSONRPC: "2.0",
		ID:      c.nextRequestID(),
		Method:  "tools/list",
	}

	return c.doRequest(ctx, req)
}

// ListToolsCached returns tools from the session cache if available and fresh.
// On a cache miss, it calls ListTools and caches the result for future use.
func (c *Client) ListToolsCached(ctx context.Context) ([]ToolInfo, error) {
	if tools := LoadTools(c.endpoint); len(tools) > 0 {
		c.logf("--- MCP reusing cached tools for %s", c.endpoint)
		return tools, nil
	}

	resp, err := c.ListTools(ctx)
	if err != nil {
		return nil, err
	}
	if resp.Error != nil {
		return nil, resp.Error
	}
	if resp.Result == nil {
		return nil, fmt.Errorf("tools/list: empty result")
	}

	tools := resp.Result.Tools
	// Best-effort cache — don't fail if caching errors.
	SaveTools(c.endpoint, tools)

	return tools, nil
}

// isSessionError returns true when the response or error indicates the server
// rejected the request due to an invalid or expired session.
func isSessionError(resp *JSONRPCResponse, err error) bool {
	if err != nil {
		msg := err.Error()
		if strings.Contains(msg, "HTTP 401") || strings.Contains(msg, "HTTP 403") {
			return true
		}
		if strings.Contains(msg, "session") || strings.Contains(msg, "Session") {
			return true
		}
	}
	if resp != nil && resp.Error != nil {
		msg := strings.ToLower(resp.Error.Message)
		if strings.Contains(msg, "session") || strings.Contains(msg, "invalid session") {
			return true
		}
	}
	return false
}

// doRequest sends a JSON-RPC request and parses the SSE response.
func (c *Client) doRequest(ctx context.Context, rpcReq JSONRPCRequest) (*JSONRPCResponse, error) {
	body, err := json.Marshal(rpcReq)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	c.logf(">>> MCP %s %s\n%s", rpcReq.Method, c.endpoint, string(body))

	if err := config.ValidateEndpointURL(c.endpoint); err != nil {
		return nil, fmt.Errorf("invalid endpoint %q: %w", c.endpoint, err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, c.endpoint, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Accept", "application/json, text/event-stream")
	httpReq.Header.Set("User-Agent", fmt.Sprintf("a365/%s (Go)", version.Version))

	if c.sessionID != "" {
		httpReq.Header.Set("Mcp-Session-Id", c.sessionID)
	}

	if c.tokenProvider == nil {
		return nil, fmt.Errorf("token provider not configured")
	}

	token, err := c.tokenProvider(ctx)
	if err != nil {
		return nil, fmt.Errorf("get token: %w", err)
	}
	httpReq.Header.Set("Authorization", "Bearer "+token)

	resp, err := c.doHTTPWithRetry(ctx, httpReq, body)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// Store session ID from response header
	if sid := resp.Header.Get("Mcp-Session-Id"); sid != "" {
		c.sessionID = sid
	}

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(bodyBytes))
	}

	contentType := resp.Header.Get("Content-Type")

	var rpcResp *JSONRPCResponse
	if strings.HasPrefix(contentType, "text/event-stream") {
		requestID := rpcReq.ID
		rpcResp, err = parseSSE(resp.Body, &requestID)
	} else {
		// Plain JSON response
		rpcResp = &JSONRPCResponse{}
		err = json.NewDecoder(resp.Body).Decode(rpcResp)
	}

	if err != nil {
		return nil, err
	}

	if rpcResp != nil {
		respBytes, _ := json.MarshalIndent(rpcResp, "", "  ")
		c.logf("<<< MCP response\n%s", string(respBytes))
	}

	return rpcResp, nil
}

// isRetryableStatus returns true for HTTP status codes that warrant a retry.
func isRetryableStatus(code int) bool {
	switch code {
	case http.StatusTooManyRequests, // 429
		http.StatusBadGateway,         // 502
		http.StatusServiceUnavailable, // 503
		http.StatusGatewayTimeout:     // 504
		return true
	}
	return false
}

// retryDelay computes the backoff duration for the given attempt, respecting
// the Retry-After header for 429 responses when present.
func retryDelay(attempt int, resp *http.Response, base time.Duration) time.Duration {
	if resp != nil && resp.StatusCode == http.StatusTooManyRequests {
		if ra := resp.Header.Get("Retry-After"); ra != "" {
			if secs, err := strconv.Atoi(ra); err == nil && secs > 0 {
				return time.Duration(secs) * time.Second
			}
		}
	}
	return base * (1 << attempt) // base, 2*base, ...
}

// doHTTPWithRetry executes the HTTP request, retrying on transient failures
// with exponential backoff. body is the raw request payload used to reset the
// request body on each retry.
func (c *Client) doHTTPWithRetry(ctx context.Context, httpReq *http.Request, body []byte) (*http.Response, error) {
	var resp *http.Response
	var err error

	for attempt := 0; attempt <= c.maxRetries; attempt++ {
		// Reset the request body for each attempt (consumed by prior Do call).
		httpReq.Body = io.NopCloser(bytes.NewReader(body))
		httpReq.GetBody = func() (io.ReadCloser, error) {
			return io.NopCloser(bytes.NewReader(body)), nil
		}

		resp, err = c.httpClient.Do(httpReq)

		// On network error, decide whether to retry.
		if err != nil {
			if attempt < c.maxRetries {
				delay := c.retryBaseDelay * (1 << attempt)
				c.logf("retrying request (attempt %d/%d) after %v...", attempt+1, c.maxRetries, delay)
				select {
				case <-ctx.Done():
					return nil, ctx.Err()
				case <-time.After(delay):
				}
				continue
			}
			return nil, fmt.Errorf("HTTP request: %w", err)
		}

		// Successful status or non-retryable error status — return immediately.
		if !isRetryableStatus(resp.StatusCode) {
			return resp, nil
		}

		// Retryable status code — close the body and retry if we have attempts left.
		if attempt < c.maxRetries {
			delay := retryDelay(attempt, resp, c.retryBaseDelay)
			resp.Body.Close()
			c.logf("retrying request (attempt %d/%d) after %v...", attempt+1, c.maxRetries, delay)
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(delay):
			}
			continue
		}

		// Out of retries — return the last (failed) response for the caller to handle.
		return resp, nil
	}

	// Should not be reached, but just in case.
	return resp, err
}

// parseSSE reads an SSE stream and extracts the JSON-RPC response for requestID.
func parseSSE(r io.Reader, requestID *int) (*JSONRPCResponse, error) {
	scanner := bufio.NewScanner(r)

	// Increase buffer size for large responses.
	buf := make([]byte, 0, 1024*1024)
	scanner.Buffer(buf, 10*1024*1024)

	var dataLines []string

	for scanner.Scan() {
		line := scanner.Text()

		if strings.HasPrefix(line, "data:") {
			// Handle both "data: payload" (with space) and "data:payload" (without)
			payload := strings.TrimPrefix(line, "data:")
			payload = strings.TrimPrefix(payload, " ")
			dataLines = append(dataLines, payload)
			continue
		}

		// Empty line = end of event
		if line == "" && len(dataLines) > 0 {
			data := strings.Join(dataLines, "\n")
			dataLines = nil

			var resp JSONRPCResponse
			if err := json.Unmarshal([]byte(data), &resp); err != nil {
				// Skip non-JSON events (e.g., keep-alive)
				continue
			}
			if resp.Result == nil && resp.Error == nil {
				continue
			}
			if requestID != nil && resp.ID != *requestID {
				continue
			}
			return &resp, nil
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("read SSE: %w", err)
	}

	// If we collected data but didn't hit an empty line (stream ended)
	if len(dataLines) > 0 {
		data := strings.Join(dataLines, "\n")
		var resp JSONRPCResponse
		if err := json.Unmarshal([]byte(data), &resp); err != nil {
			return nil, fmt.Errorf("parse final SSE data: %w", err)
		}
		if resp.Result == nil && resp.Error == nil {
			return nil, fmt.Errorf("no MCP response in SSE stream")
		}
		if requestID != nil && resp.ID != *requestID {
			return nil, fmt.Errorf("no MCP response for request %d in SSE stream", *requestID)
		}
		return &resp, nil
	}

	return nil, fmt.Errorf("no MCP response in SSE stream")
}
