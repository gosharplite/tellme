package mcp

import (
	"context"
	"errors"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/gosharplite/tellme/internal/config"
	domaintools "github.com/gosharplite/tellme/internal/domain/tools"
)

// Round 087 (ADR 0058) — the cache-aware prelude resolution.

var cachedObjSchema = []byte(`{"type":"object","properties":{}}`)

// memCache is an in-memory tools.MCPToolCache with a dial-independent save count.
type memCache struct {
	entries map[string]domaintools.MCPToolCacheEntry
	loadErr error
	saves   int
}

func (m *memCache) Load() (map[string]domaintools.MCPToolCacheEntry, error) {
	if m.loadErr != nil {
		return nil, m.loadErr
	}
	return m.entries, nil
}

func (m *memCache) Save(e map[string]domaintools.MCPToolCacheEntry) error {
	m.saves++
	m.entries = e
	return nil
}

// cacheFake is a tools.MCPClient that records dials via the factory.
type cacheFake struct {
	defs []domaintools.MCPToolDefinition
}

func (f *cacheFake) ListTools(context.Context) ([]domaintools.MCPToolDefinition, error) {
	return f.defs, nil
}
func (f *cacheFake) CallTool(context.Context, string, map[string]interface{}) (domaintools.MCPToolResult, error) {
	return domaintools.MCPToolResult{Text: "ok"}, nil
}
func (f *cacheFake) Close() error { return nil }

// recordingFactory counts dials per url and can be made to fail a url.
type recordingFactory struct {
	mu      sync.Mutex
	dials   []string
	failURL string
	defs    []domaintools.MCPToolDefinition
}

func (f *recordingFactory) new(_ context.Context, url, _ string, _, _ time.Duration) (domaintools.MCPClient, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.dials = append(f.dials, url)
	if f.failURL != "" && url == f.failURL {
		return nil, errors.New("boom")
	}
	return &cacheFake{defs: f.defs}, nil
}

func (f *recordingFactory) dialed() []string {
	f.mu.Lock()
	defer f.mu.Unlock()
	return append([]string(nil), f.dials...)
}

func noToken(context.Context) (string, error) { return "", nil }

func freshEntry(url string, defs ...string) domaintools.MCPToolCacheEntry {
	tools := make([]domaintools.MCPToolDefinition, 0, len(defs))
	for _, n := range defs {
		tools = append(tools, domaintools.MCPToolDefinition{Name: n, Description: "d", InputSchema: cachedObjSchema})
	}
	return domaintools.MCPToolCacheEntry{URL: url, Auth: "auto", FetchedAt: time.Now(), Tools: tools}
}

func offeredNames(ts []domaintools.Tool) []string {
	out := make([]string, 0, len(ts))
	for _, t := range ts {
		out = append(out, t.Name())
	}
	return out
}

// TestDiscoverCached_WarmFreshMakesNoDial — FR-001 / W1.
func TestDiscoverCached_WarmFreshMakesNoDial(t *testing.T) {
	f := &recordingFactory{}
	cache := &memCache{entries: map[string]domaintools.MCPToolCacheEntry{"shop": freshEntry("u-shop", "lookup_price")}}
	servers := map[string]config.MCPServerConfig{"shop": {URL: "u-shop"}}

	run := DiscoverCached(context.Background(), servers, time.Second, cache, time.Now, time.Hour, f.new, noToken)
	if got := f.dialed(); len(got) != 0 {
		t.Fatalf("a warm fresh cache must dial nothing; dialed %v", got)
	}
	names := offeredNames(run.Tools)
	if len(names) != 1 || names[0] != "mcp_shop_lookup_price" {
		t.Fatalf("the cached tool must be offered; got %v", names)
	}
	if !strings.Contains(run.Tools[0].Description(), `"mcp_shop_lookup_price"`) {
		t.Fatalf("the round-077 call-name note must survive the cached path; got %q", run.Tools[0].Description())
	}
	run.Close()
}

// TestDiscoverCached_ColdDiscoversAndWrites — FR-002 / W3.
func TestDiscoverCached_ColdDiscoversAndWrites(t *testing.T) {
	f := &recordingFactory{defs: []domaintools.MCPToolDefinition{{Name: "lookup_price", Description: "d", InputSchema: cachedObjSchema}}}
	cache := &memCache{}
	servers := map[string]config.MCPServerConfig{"shop": {URL: "u-shop"}}

	run := DiscoverCached(context.Background(), servers, time.Second, cache, time.Now, time.Hour, f.new, noToken)
	if got := f.dialed(); len(got) != 1 || got[0] != "u-shop" {
		t.Fatalf("a cold key must be dialed exactly once; dialed %v", got)
	}
	if cache.saves != 1 {
		t.Fatalf("a cold discovery must write the cache; saves=%d", cache.saves)
	}
	e, ok := cache.entries["shop"]
	if !ok || e.URL != "u-shop" || len(e.Tools) != 1 {
		t.Fatalf("the written entry is wrong: %+v", cache.entries)
	}
	if len(offeredNames(run.Tools)) != 1 {
		t.Fatalf("the discovered tool must be offered; got %v", offeredNames(run.Tools))
	}
	run.Close()
}

// TestDiscoverCached_StaleServesThenRefreshes — FR-003 / W2 / EC-002.
func TestDiscoverCached_StaleServesThenRefreshes(t *testing.T) {
	f := &recordingFactory{defs: []domaintools.MCPToolDefinition{{Name: "lookup_price", Description: "d", InputSchema: cachedObjSchema}}}
	stale := freshEntry("u-shop", "lookup_price")
	stale.FetchedAt = time.Now().Add(-48 * time.Hour)
	cache := &memCache{entries: map[string]domaintools.MCPToolCacheEntry{"shop": stale}}
	servers := map[string]config.MCPServerConfig{"shop": {URL: "u-shop"}}

	run := DiscoverCached(context.Background(), servers, time.Second, cache, time.Now, time.Hour, f.new, noToken)
	if got := f.dialed(); len(got) != 0 {
		t.Fatalf("a stale entry must be served WITHOUT a pre-request dial; dialed %v", got)
	}
	if len(offeredNames(run.Tools)) != 1 {
		t.Fatalf("the stale-cached tool must still be offered; got %v", offeredNames(run.Tools))
	}
	if run.Refresh == nil {
		t.Fatalf("a stale entry must schedule a post-answer refresh")
	}
	run.Refresh() // post-answer
	if got := f.dialed(); len(got) != 1 || got[0] != "u-shop" {
		t.Fatalf("the refresh must dial the stale key exactly once; dialed %v", got)
	}
	if !cache.entries["shop"].FetchedAt.After(stale.FetchedAt) {
		t.Fatalf("the refresh must update the entry's fetched_at")
	}
	run.Close()
}

// TestDiscoverCached_FailedRefreshKeepsPrior — FR-003.
func TestDiscoverCached_FailedRefreshKeepsPrior(t *testing.T) {
	f := &recordingFactory{failURL: "u-shop"}
	stale := freshEntry("u-shop", "lookup_price")
	stale.FetchedAt = time.Now().Add(-48 * time.Hour)
	cache := &memCache{entries: map[string]domaintools.MCPToolCacheEntry{"shop": stale}}
	servers := map[string]config.MCPServerConfig{"shop": {URL: "u-shop"}}

	run := DiscoverCached(context.Background(), servers, time.Second, cache, time.Now, time.Hour, f.new, noToken)
	run.Refresh()
	if !cache.entries["shop"].FetchedAt.Equal(stale.FetchedAt) {
		t.Fatalf("a failed refresh must keep the prior entry")
	}
	run.Close()
}

// TestDiscoverCached_DeclarationMismatchIsCold — FR-004 / EC-001.
func TestDiscoverCached_DeclarationMismatchIsCold(t *testing.T) {
	f := &recordingFactory{defs: []domaintools.MCPToolDefinition{{Name: "lookup_price", Description: "d", InputSchema: cachedObjSchema}}}
	old := freshEntry("https://old.example/mcp", "lookup_price")
	cache := &memCache{entries: map[string]domaintools.MCPToolCacheEntry{"shop": old}}
	servers := map[string]config.MCPServerConfig{"shop": {URL: "u-shop"}}

	run := DiscoverCached(context.Background(), servers, time.Second, cache, time.Now, time.Hour, f.new, noToken)
	if got := f.dialed(); len(got) != 1 {
		t.Fatalf("a changed URL declaration must be cold; dialed %v", got)
	}
	run.Close()
}

// TestDiscoverCached_WarmOrderEqualsLive — FR-005 (the determinism pin).
func TestDiscoverCached_WarmOrderEqualsLive(t *testing.T) {
	schema := cachedObjSchema
	servers := map[string]config.MCPServerConfig{"alpha": {URL: "u-alpha"}, "shop": {URL: "u-shop"}}
	liveDefs := map[string][]domaintools.MCPToolDefinition{
		"u-alpha": {{Name: "a1", Description: "d", InputSchema: schema}},
		"u-shop":  {{Name: "lookup_price", Description: "d", InputSchema: schema}, {Name: "z2", Description: "d", InputSchema: schema}},
	}
	liveFactory := func(_ context.Context, url, _ string, _, _ time.Duration) (domaintools.MCPClient, error) {
		return &cacheFake{defs: liveDefs[url]}, nil
	}
	liveTools, _, _ := Discover(context.Background(), servers, time.Second, liveFactory, noToken)
	want := offeredNames(liveTools)

	cache := &memCache{entries: map[string]domaintools.MCPToolCacheEntry{
		"alpha": freshEntry("u-alpha", "a1"),
		"shop":  freshEntry("u-shop", "lookup_price", "z2"),
	}}
	run := DiscoverCached(context.Background(), servers, time.Second, cache, time.Now, time.Hour, (&recordingFactory{}).new, noToken)
	got := offeredNames(run.Tools)
	if strings.Join(got, ",") != strings.Join(want, ",") {
		t.Fatalf("the warm offer set/order must equal live discovery;\n got %v\nwant %v", got, want)
	}
	run.Close()
}

// TestDiscoverCached_IgnoresEntryForUnknownServer — EC-003.
func TestDiscoverCached_IgnoresEntryForUnknownServer(t *testing.T) {
	f := &recordingFactory{defs: []domaintools.MCPToolDefinition{{Name: "t", Description: "d", InputSchema: cachedObjSchema}}}
	cache := &memCache{entries: map[string]domaintools.MCPToolCacheEntry{"gone": freshEntry("u-gone", "x")}}
	servers := map[string]config.MCPServerConfig{"shop": {URL: "u-shop"}}

	run := DiscoverCached(context.Background(), servers, time.Second, cache, time.Now, time.Hour, f.new, noToken)
	if got := f.dialed(); len(got) != 1 || got[0] != "u-shop" {
		t.Fatalf("only the configured key may be dialed; dialed %v", got)
	}
	if _, ok := cache.entries["gone"]; !ok {
		t.Fatalf("an entry for a server not in the config must be preserved, not dropped")
	}
	run.Close()
}

// TestDiscoverCached_NoServersIsInert — EC-004.
func TestDiscoverCached_NoServersIsInert(t *testing.T) {
	f := &recordingFactory{}
	run := DiscoverCached(context.Background(), nil, time.Second, &memCache{}, time.Now, time.Hour, f.new, noToken)
	if len(f.dialed()) != 0 || len(run.Tools) != 0 {
		t.Fatalf("no MCP servers ⇒ no dial, no tools; dialed=%v tools=%v", f.dialed(), offeredNames(run.Tools))
	}
}

// TestDiscoverCached_CorruptCacheIsCold — FR-008.
func TestDiscoverCached_CorruptCacheIsCold(t *testing.T) {
	f := &recordingFactory{defs: []domaintools.MCPToolDefinition{{Name: "t", Description: "d", InputSchema: cachedObjSchema}}}
	cache := &memCache{loadErr: errors.New("corrupt")}
	servers := map[string]config.MCPServerConfig{"shop": {URL: "u-shop"}}

	run := DiscoverCached(context.Background(), servers, time.Second, cache, time.Now, time.Hour, f.new, noToken)
	if got := f.dialed(); len(got) != 1 {
		t.Fatalf("a corrupt cache must degrade to live discovery; dialed %v", got)
	}
	if cache.saves != 1 {
		t.Fatalf("the cold discovery must overwrite the corrupt cache; saves=%d", cache.saves)
	}
	run.Close()
}

// TestLazyClient_ConnectFailureIsRecoverable — FR-007 (the round-032 TD1/R3 contract).
func TestLazyClient_ConnectFailureIsRecoverable(t *testing.T) {
	f := &recordingFactory{failURL: "u-shop"}
	lc := NewLazyClient(config.MCPServerConfig{URL: "u-shop"}, time.Second, f.new, noToken)
	res, err := lc.CallTool(context.Background(), "lookup_price", nil)
	if err != nil {
		t.Fatalf("a call-time connect failure must NOT return an error; got %v", err)
	}
	if !strings.HasPrefix(res.Text, "error: ") {
		t.Fatalf("a connect failure must fold into a recoverable `error: …` result; got %q", res.Text)
	}
}

// TestLazyClient_NoConnectWhenUnused — FR-001 (a cached tool never dials unless called).
func TestLazyClient_NoConnectWhenUnused(t *testing.T) {
	f := &recordingFactory{}
	cache := &memCache{entries: map[string]domaintools.MCPToolCacheEntry{"shop": freshEntry("u-shop", "lookup_price")}}
	servers := map[string]config.MCPServerConfig{"shop": {URL: "u-shop"}}
	run := DiscoverCached(context.Background(), servers, time.Second, cache, time.Now, time.Hour, f.new, noToken)
	run.Close()
	if len(f.dialed()) != 0 {
		t.Fatalf("closing an unused lazy client must not dial; dialed %v", f.dialed())
	}
}
