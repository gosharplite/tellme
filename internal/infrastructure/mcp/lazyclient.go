package mcp

import (
	"context"
	"sync"
	"time"

	"github.com/gosharplite/tellme/internal/config"
	domaintools "github.com/gosharplite/tellme/internal/domain/tools"
)

// lazyClient is the round-087 (ADR 0058 D4) client a CACHED tool binds: it
// defers both the credential resolution (so no `gh` spawn happens for a tool the
// model never calls) and the connect until the first CallTool. A cache hit that
// executes no tool therefore makes ZERO MCP connections, which is what makes the
// warm prelude local.
//
// It satisfies tools.MCPClient. The connect is bounded by the CALL timeout (a
// tool call deserves the full resource-contract timeout, not the 3 s discovery
// fast-fail bound), and any connect failure is folded into the recoverable
// nil-error `error: …` result — the round-032 TD1/R3 contract, so a cached tool
// against a dead server never aborts the turn.
type lazyClient struct {
	cfg          config.MCPServerConfig
	callTimeout  time.Duration
	newClient    ClientFactory
	resolveToken TokenSource

	mu     sync.Mutex
	inner  domaintools.MCPClient
	err    error
	listed bool
}

// NewLazyClient builds a client that connects on first use (round 087; ADR 0058).
func NewLazyClient(cfg config.MCPServerConfig, callTimeout time.Duration, newClient ClientFactory, resolveToken TokenSource) domaintools.MCPClient {
	return &lazyClient{cfg: cfg, callTimeout: callTimeout, newClient: newClient, resolveToken: resolveToken}
}

// connect resolves the credential and establishes the session exactly once.
func (c *lazyClient) connect(ctx context.Context) (domaintools.MCPClient, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.inner != nil {
		return c.inner, nil
	}
	if c.err != nil {
		return nil, c.err
	}
	header, _ := ResolveAuthorization(ctx, c.cfg, c.resolveToken)
	inner, err := c.newClient(ctx, c.cfg.URL, header, c.callTimeout, c.callTimeout)
	if err != nil {
		c.err = err
		return nil, err
	}
	c.inner = inner
	return inner, nil
}

// ListTools connects and delegates. It is not exercised on the cached path (the
// tool set is already known), but is provided for interface completeness.
func (c *lazyClient) ListTools(ctx context.Context) ([]domaintools.MCPToolDefinition, error) {
	cl, err := c.connect(ctx)
	if err != nil {
		return nil, err
	}
	return cl.ListTools(ctx)
}

// warmToolsList issues ONE tools/list on the session so the client SDK caches the
// tool definitions (round 088; ADR 0059). This is REQUIRED for `x-mcp-header`
// routing: the SDK resolves a tool's argument-header annotations from its
// `tools/list` cache (mcp/client.go `lookupTool`), so a session that never listed
// the tools sends the arguments in the body instead of as `Mcp-Param-*` headers,
// and a server that requires them (e.g. the GitHub MCP server) rejects the call.
// Best-effort: a warm-up failure is ignored and the call is still attempted.
func (c *lazyClient) warmToolsList(ctx context.Context, cl domaintools.MCPClient) {
	c.mu.Lock()
	if c.listed {
		c.mu.Unlock()
		return
	}
	c.mu.Unlock()
	_, _ = cl.ListTools(ctx) // populate the SDK tools cache; ignore a failure
	c.mu.Lock()
	c.listed = true
	c.mu.Unlock()
}

// CallTool connects on first use, warms the session's tools/list cache (so
// header-routed arguments are emitted as `Mcp-Param-*`; ADR 0059), then
// delegates. A connect failure is a recoverable result.
func (c *lazyClient) CallTool(ctx context.Context, name string, args map[string]interface{}) (domaintools.MCPToolResult, error) {
	cl, err := c.connect(ctx)
	if err != nil {
		return domaintools.MCPToolResult{Text: "error: " + err.Error()}, nil
	}
	c.warmToolsList(ctx, cl)
	return cl.CallTool(ctx, name, args)
}

// Close closes the underlying session if one was ever opened.
func (c *lazyClient) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.inner != nil {
		return c.inner.Close()
	}
	return nil
}
