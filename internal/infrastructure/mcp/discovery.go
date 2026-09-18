package mcp

import (
	"context"
	"sort"
	"sync"
	"time"

	"github.com/gosharplite/tellme/internal/config"
	domaintools "github.com/gosharplite/tellme/internal/domain/tools"
)

// ClientFactory builds an MCP client for a server given the resolved
// Authorization header and the two bounding deadlines (the discovery fast-fail
// bound and the resolved tool-call timeout). It is the injectable seam that keeps
// this package free of a direct internal/infrastructure/di import (round-044).
type ClientFactory func(ctx context.Context, url, authorization string, discoveryTimeout, callTimeout time.Duration) (domaintools.MCPClient, error)

// Discover performs the round-032 prompt-path MCP discovery (relocated here from
// internal/cli by round 044 / ADR 0013): per enabled remote server it discovers
// the offered tools (concurrently, each bounded by bound) and returns them sorted
// by server key, plus any warn+skip messages and a close hook that tears down the
// discovered clients. It never returns an error: a failed/unreachable server is a
// warning, not a run failure (FR-009).
//
// The two infrastructure seams (the client factory and the credential token
// source) are passed IN — the composition root binds them — so this package
// imports neither internal/infrastructure/di nor the SDK in any new way, and
// verify-mcp-sdk-confinement stays green.
func Discover(ctx context.Context, servers map[string]config.MCPServerConfig, bound time.Duration, newClient ClientFactory, resolveToken TokenSource) ([]domaintools.Tool, []string, func()) {
	run := discover(ctx, servers, bound, newClient, resolveToken)
	return run.tools, run.warnings, run.close
}

// mcpRun is the prompt-path MCP augmentation of one run.
type mcpRun struct {
	tools    []domaintools.Tool
	warnings []string
	close    func()
}

// discover discovers, per enabled remote MCP server, the tools it offers and
// returns them (sorted by server key) plus any warn+skip messages, bounded by
// bound. Discovery runs concurrently per server, so the total discovery-
// attributable delay is bounded by the single fixed deadline regardless of the
// number of servers (FR-010). A server marked `ENABLED: false` is not dialed at
// all (FR-011); a COMMAND (stdio) entry is excluded by validation.
func discover(ctx context.Context, servers map[string]config.MCPServerConfig, bound time.Duration, newClient ClientFactory, resolveToken TokenSource) mcpRun {
	run := mcpRun{close: func() {}}
	keys := make([]string, 0, len(servers))
	for name, s := range servers {
		if s.IsRemote() && s.IsEnabled() {
			keys = append(keys, name)
		}
	}
	if len(keys) == 0 {
		return run
	}
	sort.Strings(keys)

	results := make([]serverResult, len(keys))
	var wg sync.WaitGroup
	for i, key := range keys {
		wg.Add(1)
		go func(i int, key string) {
			defer wg.Done()
			results[i] = discoverServer(ctx, key, servers[key], bound, newClient, resolveToken)
		}(i, key)
	}
	wg.Wait()

	var clients []domaintools.MCPClient
	for _, r := range results {
		if r.client != nil {
			clients = append(clients, r.client)
		}
		run.warnings = append(run.warnings, r.warnings...)
		run.tools = append(run.tools, r.tools...)
	}
	if len(clients) > 0 {
		run.close = func() {
			for _, c := range clients {
				_ = c.Close()
			}
		}
	}
	return run
}

// serverResult is one server's discovery outcome.
type serverResult struct {
	client   domaintools.MCPClient
	tools    []domaintools.Tool
	warnings []string
}

// discoverServer probes one server under the fixed fast-fail bound: resolve the
// credential, connect, list the tools, and normalize each schema (an unsafe one
// is skipped with a warning). Any failure is a warning, never fatal.
func discoverServer(parent context.Context, key string, cfg config.MCPServerConfig, bound time.Duration, newClient ClientFactory, resolveToken TokenSource) serverResult {
	ctx, cancel := context.WithTimeout(parent, bound)
	defer cancel()

	var res serverResult
	header, warn := ResolveAuthorization(ctx, cfg, resolveToken)
	if warn != "" {
		res.warnings = append(res.warnings, warn)
	}
	timeout := ResolveMCPTimeout(cfg.Timeout)
	client, err := newClient(ctx, cfg.URL, header, bound, timeout)
	if err != nil {
		res.warnings = append(res.warnings, UnreachableWarningWithHint(key, err))
		return res
	}
	defs, err := client.ListTools(ctx)
	if err != nil {
		_ = client.Close()
		res.warnings = append(res.warnings, UnreachableWarningWithHint(key, err))
		return res
	}
	res.client = client
	// Discovery is complete: move the adapter's per-request deadline from the
	// fast-fail bound to the tool-call timeout (round-032 principal-review TD).
	SwitchToCallTimeout(client)
	for _, dt := range defs {
		name := NamespacedName(key, dt.Name)
		if !ValidToolName(name) {
			res.warnings = append(res.warnings, NameSkippedWarning(key, dt.Name))
			continue
		}
		schema, serr := NormalizeMCPSchema(dt.InputSchema)
		if serr != nil {
			res.warnings = append(res.warnings, SchemaSkippedWarning(key, dt.Name))
			continue
		}
		dt.InputSchema = schema
		res.tools = append(res.tools, NewTool(key, dt, client, timeout))
	}
	return res
}
