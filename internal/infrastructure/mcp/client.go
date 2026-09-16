package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"

	sdk "github.com/modelcontextprotocol/go-sdk/mcp"

	domaintools "github.com/gosharplite/tellme/internal/domain/tools"
)

// client is the remote (Streamable HTTP) tools.MCPClient adapter over the
// official MCP Go SDK. It is the ONLY production type that holds an SDK value —
// the SDK import is confined to this package by the verify-mcp-sdk-confinement
// gate (round-032 research Decision 2).
type client struct {
	session     *sdk.ClientSession
	rt          *boundedTransport
	callTimeout time.Duration
}

// clientVersion is the tellme implementation version advertised in the MCP
// handshake.
const clientVersion = "0.1"

// NewRemoteClient connects to a remote MCP server at url, adding the given
// Authorization header (empty = anonymous) to every request, and returns the
// tools.MCPClient adapter.
//
// Bounding (round-032 FR-008/FR-021): the SDK DETACHES the connection context
// from the caller's, so a server that never answers would otherwise hang
// unbounded. The adapter therefore applies a per-request deadline through its
// transport. The deadline starts at discoveryTimeout (the fixed fast-fail bound)
// and covers the connect/list phase; the caller switches it to the tool-call
// timeout via SwitchToCallTimeout once discovery is complete, so ListTools stays
// on the fast-fail bound while tool calls get the full resource-contract
// timeout.
func NewRemoteClient(ctx context.Context, url, authorization string, discoveryTimeout, callTimeout time.Duration) (domaintools.MCPClient, error) {
	rt := &boundedTransport{base: http.DefaultTransport, authorization: authorization, timeout: discoveryTimeout}
	httpClient := &http.Client{Transport: rt}
	c := sdk.NewClient(&sdk.Implementation{Name: "tellme", Version: clientVersion}, nil)
	transport := &sdk.StreamableClientTransport{
		Endpoint:   url,
		HTTPClient: httpClient,
		// request/response only: no persistent server→client SSE stream (keeps the
		// connection short-lived and the discovery bound meaningful).
		DisableStandaloneSSE: true,
		// no reconnect attempts: a transport failure is a failure, not a retry loop.
		MaxRetries: -1,
	}
	sess, err := c.Connect(ctx, transport, nil)
	if err != nil {
		return nil, fmt.Errorf("connect to MCP server: %w", err)
	}
	return &client{session: sess, rt: rt, callTimeout: callTimeout}, nil
}

// UseCallTimeout moves the adapter's per-request deadline to the tool-call
// timeout. It is called once discovery (connect + list) is complete.
func (c *client) UseCallTimeout() {
	if c.rt != nil {
		c.rt.setTimeout(c.callTimeout)
	}
}

// SwitchToCallTimeout moves a client's per-request deadline to the tool-call
// timeout when the adapter supports it (the discovery layer calls it after
// ListTools). A client without the capability is left unchanged.
func SwitchToCallTimeout(c domaintools.MCPClient) {
	if s, ok := c.(interface{ UseCallTimeout() }); ok {
		s.UseCallTimeout()
	}
}

// boundedTransport injects the resolved Authorization header (when non-empty)
// onto every outbound request and applies the CURRENT per-request deadline (the
// discovery bound during connect/list, the tool-call timeout afterwards). It
// never logs the header (FR-017).
type boundedTransport struct {
	base          http.RoundTripper
	authorization string

	mu      sync.Mutex
	timeout time.Duration
}

func (t *boundedTransport) setTimeout(d time.Duration) {
	t.mu.Lock()
	t.timeout = d
	t.mu.Unlock()
}

// RoundTrip applies the per-request deadline. IMPORTANT: the deadline's cancel
// func is tied to the response body's Close (NOT deferred inside RoundTrip),
// because net/http's RoundTrip returns once the response HEADERS are decoded —
// the body is read lazily by the caller. A `defer cancel()` here would cancel
// the request the instant headers arrived, so any chunked / streaming body read
// afterwards would fail with `context canceled` (round-032 principal review
// BLOCKER).
func (t *boundedTransport) RoundTrip(r *http.Request) (*http.Response, error) {
	t.mu.Lock()
	d := t.timeout
	t.mu.Unlock()

	ctx := r.Context()
	var cancel context.CancelFunc
	if d > 0 {
		ctx, cancel = context.WithTimeout(ctx, d)
	}
	r = r.Clone(ctx)
	if t.authorization != "" {
		r.Header.Set("Authorization", t.authorization)
	}

	resp, err := t.base.RoundTrip(r)
	if err != nil {
		if cancel != nil {
			cancel()
		}
		return nil, err
	}
	if cancel != nil {
		resp.Body = &cancelBody{ReadCloser: resp.Body, cancel: cancel}
	}
	return resp, nil
}

// cancelBody releases the per-request deadline's cancel func when the response
// body is closed, so the context stays live for the whole body read.
type cancelBody struct {
	io.ReadCloser
	cancel context.CancelFunc
}

func (b *cancelBody) Close() error {
	defer b.cancel()
	return b.ReadCloser.Close()
}

// ListTools lists the server's tools as network-free domain definitions.
func (c *client) ListTools(ctx context.Context) ([]domaintools.MCPToolDefinition, error) {
	res, err := c.session.ListTools(ctx, &sdk.ListToolsParams{})
	if err != nil {
		return nil, err
	}
	defs := make([]domaintools.MCPToolDefinition, 0, len(res.Tools))
	for _, t := range res.Tools {
		raw, err := schemaToRaw(t.InputSchema)
		if err != nil {
			return nil, fmt.Errorf("tool %q: %w", t.Name, err)
		}
		defs = append(defs, domaintools.MCPToolDefinition{
			Name:        t.Name,
			Description: t.Description,
			InputSchema: raw,
		})
	}
	return defs, nil
}

// CallTool calls the named tool. Per the port's TD1/R3 invariant, EVERY
// call-time failure (tool-level error, transport failure, or a call against a
// closed client) is returned as a nil-error MCPToolResult carrying the error
// text, so the loop takes the recoverable path and the run never aborts
// (FR-018). A nil args map is normalised to {} before the wire (TD2/R2).
func (c *client) CallTool(ctx context.Context, name string, args map[string]interface{}) (domaintools.MCPToolResult, error) {
	if args == nil {
		args = map[string]interface{}{}
	}
	res, err := c.session.CallTool(ctx, &sdk.CallToolParams{Name: name, Arguments: args})
	if err != nil {
		return domaintools.MCPToolResult{Text: "error: " + err.Error()}, nil
	}
	text := contentText(res)
	if res.IsError && text == "" {
		text = "error: the tool reported a failure"
	}
	return domaintools.MCPToolResult{Text: text}, nil
}

// Close closes the session (idempotent at the SDK layer).
func (c *client) Close() error {
	if c.session == nil {
		return nil
	}
	return c.session.Close()
}

// schemaToRaw returns the server's input schema as raw JSON. From the client the
// SDK exposes it as the default JSON marshaling of the wire schema (a
// map[string]any), so it is re-marshaled; a json.RawMessage passes through.
func schemaToRaw(v any) (json.RawMessage, error) {
	switch s := v.(type) {
	case nil:
		return nil, nil
	case json.RawMessage:
		return s, nil
	default:
		b, err := json.Marshal(s)
		if err != nil {
			return nil, fmt.Errorf("marshal input schema: %w", err)
		}
		return b, nil
	}
}

// contentText extracts the concatenated text of a call result's content blocks.
// A result carrying only structured content falls back to its JSON form.
func contentText(res *sdk.CallToolResult) string {
	var b []byte
	if res == nil {
		return ""
	}
	for _, c := range res.Content {
		if tc, ok := c.(*sdk.TextContent); ok {
			b = append(b, tc.Text...)
		}
	}
	if len(b) == 0 && res.StructuredContent != nil {
		if out, err := json.Marshal(res.StructuredContent); err == nil {
			return string(out)
		}
	}
	return string(b)
}
