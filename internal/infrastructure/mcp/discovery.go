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
	keys := enabledRemoteKeys(servers)
	if len(keys) == 0 {
		return run
	}
	results := discoverKeys(ctx, keys, servers, bound, newClient, resolveToken)
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

// enabledRemoteKeys returns the sorted keys of the enabled remote MCP servers.
func enabledRemoteKeys(servers map[string]config.MCPServerConfig) []string {
	keys := make([]string, 0, len(servers))
	for name, s := range servers {
		if s.IsRemote() && s.IsEnabled() {
			keys = append(keys, name)
		}
	}
	sort.Strings(keys)
	return keys
}

// discoverKeys probes the given server keys concurrently (one goroutine each,
// bounded by bound) and returns the results in the SAME order as keys.
func discoverKeys(ctx context.Context, keys []string, servers map[string]config.MCPServerConfig, bound time.Duration, newClient ClientFactory, resolveToken TokenSource) []serverResult {
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
	return results
}

// serverResult is one server's discovery outcome.
type serverResult struct {
	client   domaintools.MCPClient
	tools    []domaintools.Tool
	defs     []domaintools.MCPToolDefinition // the offered defs (normalized), in order
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
		res.defs = append(res.defs, dt)
		res.tools = append(res.tools, NewTool(key, dt, client, timeout))
	}
	return res
}

// CachedRun is the outcome of the cache-aware prelude (round 087; ADR 0058).
type CachedRun struct {
	Tools    []domaintools.Tool
	Warnings []string
	// Close tears down every opened client (the cold-discovery clients and the
	// lazy clients the cached tools bound).
	Close func()
	// Refresh is the post-answer stale refresh; nil when no key was stale. It
	// dials the stale keys (bounded, concurrent), rewrites their cache entries on
	// success (keeping the prior entry on failure), and returns any warnings.
	Refresh func() []string
}

// DiscoverCached is the round-087 (ADR 0058) cache-aware prompt-path prelude. For
// each enabled remote MCP server it prefers a cached entry whose declaration
// (url + auth) still matches the config:
//
//   - a FRESH entry is served with NO dial;
//   - a STALE entry is served with no pre-request dial and scheduled for a
//     post-answer Refresh;
//   - a COLD key (absent / declaration mismatch / corrupt cache) falls back to
//     exactly one bounded concurrent discovery and is written to the cache.
//
// The assembled tool list preserves the live order (native tools are appended by
// the caller; MCP tools appear in sorted server-key order, each server's tools in
// advertised order), so a warm cache offers an identical set/order to live
// discovery. Best-effort: a cache read/write failure degrades to live discovery
// and never fails the run.
func DiscoverCached(ctx context.Context, servers map[string]config.MCPServerConfig, bound time.Duration,
	cache domaintools.MCPToolCache, now func() time.Time, ttl time.Duration,
	newClient ClientFactory, resolveToken TokenSource) CachedRun {
	run := CachedRun{Close: func() {}}
	keys := enabledRemoteKeys(servers)
	if len(keys) == 0 {
		return run
	}
	loaded, _ := cache.Load() // best-effort: a read failure ⇒ every key cold

	byKey := make(map[string][]domaintools.Tool, len(keys))
	var clients []domaintools.MCPClient
	var cold, stale []string
	for _, key := range keys {
		cfg := servers[key]
		e, ok := loaded[key]
		if !ok || e.URL != cfg.URL || e.Auth != cfg.EffectiveAuth() {
			cold = append(cold, key)
			continue
		}
		// Serve the cached entry: bind a lazily-connecting client so an uneventful
		// turn dials nothing (round-032 TD1/R3 contract on failure).
		lc := NewLazyClient(cfg, ResolveMCPTimeout(cfg.Timeout), newClient, resolveToken)
		clients = append(clients, lc)
		byKey[key] = cachedTools(key, cfg, e, lc)
		if now().Sub(e.FetchedAt) >= ttl {
			stale = append(stale, key)
		}
	}

	coldTools, warns, coldClients, fresh := discoverColdKeys(ctx, cold, servers, bound, newClient, resolveToken, now)
	run.Warnings = append(run.Warnings, warns...)
	clients = append(clients, coldClients...)
	for k, ts := range coldTools {
		byKey[k] = ts
	}
	if len(fresh) > 0 {
		_ = cache.Save(mergeCache(loaded, fresh)) // best-effort
	}

	for _, key := range keys { // assemble in the LIVE order (sorted key order)
		run.Tools = append(run.Tools, byKey[key]...)
	}
	if len(clients) > 0 {
		cs := clients
		run.Close = func() {
			for _, c := range cs {
				_ = c.Close()
			}
		}
	}
	if len(stale) > 0 {
		run.Refresh = makeStaleRefresh(stale, servers, bound, cache, now, newClient, resolveToken)
	}
	return run
}

// cachedTools rebuilds one cached entry's offered tools, bound to a lazy client.
// The cached NAME and SCHEMA are both re-validated (fold N-087-4): the name must
// satisfy the wire grammar and the schema is re-normalized (an absent/unsafe
// schema is skipped, exactly like the live path), so a hand-edited cache file is
// an untrusted input on BOTH axes, not just the name.
func cachedTools(key string, cfg config.MCPServerConfig, e domaintools.MCPToolCacheEntry, lc domaintools.MCPClient) []domaintools.Tool {
	to := ResolveMCPTimeout(cfg.Timeout)
	tools := make([]domaintools.Tool, 0, len(e.Tools))
	for _, def := range e.Tools {
		if !ValidToolName(NamespacedName(key, def.Name)) {
			continue
		}
		schema, err := NormalizeMCPSchema(def.InputSchema)
		if err != nil {
			continue
		}
		def.InputSchema = schema
		tools = append(tools, NewTool(key, def, lc, to))
	}
	return tools
}

// discoverColdKeys probes the cold keys (bounded, concurrent) and returns their
// tools keyed by server, their warnings, their clients, and the cache entries to
// persist (only for keys that answered, so a failed key is never cached).
func discoverColdKeys(ctx context.Context, cold []string, servers map[string]config.MCPServerConfig, bound time.Duration, newClient ClientFactory, resolveToken TokenSource, now func() time.Time) (map[string][]domaintools.Tool, []string, []domaintools.MCPClient, map[string]domaintools.MCPToolCacheEntry) {
	byKey := make(map[string][]domaintools.Tool, len(cold))
	entries := make(map[string]domaintools.MCPToolCacheEntry, len(cold))
	var warns []string
	var clients []domaintools.MCPClient
	if len(cold) == 0 {
		return byKey, warns, clients, entries
	}
	results := discoverKeys(ctx, cold, servers, bound, newClient, resolveToken)
	for i, r := range results {
		warns = append(warns, r.warnings...)
		if r.client == nil {
			continue
		}
		key := cold[i]
		clients = append(clients, r.client)
		byKey[key] = r.tools
		entries[key] = domaintools.MCPToolCacheEntry{
			URL:       servers[key].URL,
			Auth:      servers[key].EffectiveAuth(),
			FetchedAt: now(),
			Tools:     r.defs,
		}
	}
	return byKey, warns, clients, entries
}

// mergeCache returns loaded with fresh overriding same-key entries.
func mergeCache(loaded, fresh map[string]domaintools.MCPToolCacheEntry) map[string]domaintools.MCPToolCacheEntry {
	merged := make(map[string]domaintools.MCPToolCacheEntry, len(loaded)+len(fresh))
	for k, v := range loaded {
		merged[k] = v
	}
	for k, v := range fresh {
		merged[k] = v
	}
	return merged
}

// makeStaleRefresh builds the post-answer refresh: dial the stale keys (bounded,
// concurrent), rewrite their entries on success, keep the prior entry on failure.
func makeStaleRefresh(stale []string, servers map[string]config.MCPServerConfig, bound time.Duration, cache domaintools.MCPToolCache, now func() time.Time, newClient ClientFactory, resolveToken TokenSource) func() []string {
	keys := append([]string(nil), stale...)
	return func() []string {
		ctx, cancel := context.WithTimeout(context.Background(), bound)
		defer cancel()
		results := discoverKeys(ctx, keys, servers, bound, newClient, resolveToken)
		cur, _ := cache.Load()
		if cur == nil {
			cur = map[string]domaintools.MCPToolCacheEntry{}
		}
		var warns []string
		changed := false
		for i, r := range results {
			warns = append(warns, r.warnings...)
			if r.client == nil {
				continue // failed refresh: keep the prior entry
			}
			_ = r.client.Close()
			key := keys[i]
			cur[key] = domaintools.MCPToolCacheEntry{
				URL:       servers[key].URL,
				Auth:      servers[key].EffectiveAuth(),
				FetchedAt: now(),
				Tools:     r.defs,
			}
			changed = true
		}
		if changed {
			_ = cache.Save(cur)
		}
		return warns
	}
}
