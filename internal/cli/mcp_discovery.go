package cli

import (
	"context"
	"sort"
	"sync"
	"time"

	"github.com/gosharplite/tellme/internal/config"
	domaintools "github.com/gosharplite/tellme/internal/domain/tools"
	"github.com/gosharplite/tellme/internal/infrastructure/di"
	"github.com/gosharplite/tellme/internal/infrastructure/mcp"
)

// mcpDiscoveryBound is the fixed, small per-server fast-fail deadline: it is
// INDEPENDENT of (and much smaller than) the per-tool call timeout, so an
// unreachable server is a non-event rather than a stall (FR-008/NFR-002). It also
// bounds the credential token source (FR-020).
const mcpDiscoveryBound = 3 * time.Second

// mcpClientFactory builds an MCP client for a server given the resolved
// Authorization header and the two bounding deadlines (the discovery fast-fail
// bound and the resolved tool-call timeout).
type mcpClientFactory func(ctx context.Context, url, authorization string, discoveryTimeout, callTimeout time.Duration) (domaintools.MCPClient, error)

// tokenResolver resolves a token from the external source for `gh`/`auto` auth
// (the injectable seam, FR-020).
type tokenResolver = mcp.TokenSource

// mcpDiscoveryConfig carries the injected discovery seams — the bound, the
// client factory, and the token resolver. It is a VALUE, not package-level
// mutable vars, so the seams are explicit at the call site and the discovery is
// `t.Parallel()`-safe (round-029/031 precedent; round-032 implementation-review
// F6). Tests construct their own config with fakes.
type mcpDiscoveryConfig struct {
	bound        time.Duration
	newClient    mcpClientFactory
	resolveToken tokenResolver
}

// defaultMCPDiscovery is the production discovery configuration: the confined
// SDK adapter + the bounded `gh` resolver.
var defaultMCPDiscovery = mcpDiscoveryConfig{
	bound:        mcpDiscoveryBound,
	newClient:    di.NewRemoteClient,
	resolveToken: di.NewGhTokenResolver(mcpDiscoveryBound),
}

// mcpRun is the prompt-path MCP augmentation of one run: the tools discovered
// from every enabled remote server (in deterministic order), the non-fatal
// warnings to surface on the diagnostic stream, and a close hook that tears down
// the clients at the end of the run.
type mcpRun struct {
	tools    []domaintools.Tool
	warnings []string
	close    func()
}

// discoverForRun runs the production discovery configuration.
func discoverForRun(ctx context.Context, servers map[string]config.MCPServerConfig) mcpRun {
	return defaultMCPDiscovery.discover(ctx, servers)
}

// discover discovers, per enabled remote MCP server, the tools it offers and
// returns them (sorted by server key) plus any warn+skip messages, bounded by
// d.bound. It never returns an error: a failed/unreachable server is a warning,
// not a run failure (FR-009). Discovery runs concurrently per server, so the
// total discovery-attributable delay is bounded by the single fixed deadline
// regardless of the number of servers (FR-010).
//
// A server marked `ENABLED: false` is not dialed at all (FR-011); a COMMAND
// (stdio) entry is excluded by validation (it is not remote).
func (d mcpDiscoveryConfig) discover(ctx context.Context, servers map[string]config.MCPServerConfig) mcpRun {
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
			results[i] = d.discoverServer(ctx, key, servers[key])
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
func (d mcpDiscoveryConfig) discoverServer(parent context.Context, key string, cfg config.MCPServerConfig) serverResult {
	ctx, cancel := context.WithTimeout(parent, d.bound)
	defer cancel()

	var res serverResult
	header, warn := mcp.ResolveAuthorization(ctx, cfg, d.resolveToken)
	if warn != "" {
		res.warnings = append(res.warnings, warn)
	}
	timeout := mcp.ResolveMCPTimeout(cfg.Timeout)
	client, err := d.newClient(ctx, cfg.URL, header, d.bound, timeout)
	if err != nil {
		res.warnings = append(res.warnings, mcp.UnreachableWarning(key))
		return res
	}
	defs, err := client.ListTools(ctx)
	if err != nil {
		_ = client.Close()
		res.warnings = append(res.warnings, mcp.UnreachableWarning(key))
		return res
	}
	res.client = client
	// Discovery is complete: move the adapter's per-request deadline from the
	// fast-fail bound to the tool-call timeout, so ListTools stayed on the bound
	// while the subsequent tool calls get the full resource-contract timeout
	// (round-032 principal-review TD).
	mcp.SwitchToCallTimeout(client)
	for _, dt := range defs {
		name := mcp.NamespacedName(key, dt.Name)
		if !mcp.ValidToolName(name) {
			res.warnings = append(res.warnings, mcp.NameSkippedWarning(key, dt.Name))
			continue
		}
		schema, serr := mcp.NormalizeMCPSchema(dt.InputSchema)
		if serr != nil {
			res.warnings = append(res.warnings, mcp.SchemaSkippedWarning(key, dt.Name))
			continue
		}
		dt.InputSchema = schema
		res.tools = append(res.tools, mcp.NewTool(key, dt, client, timeout))
	}
	return res
}
