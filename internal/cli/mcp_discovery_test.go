package cli

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/gosharplite/tellme/internal/config"
	domaintools "github.com/gosharplite/tellme/internal/domain/tools"
)

// fakeMCP is a test double for the tools.MCPClient port.
type fakeMCP struct {
	tools    []domaintools.MCPToolDefinition
	block    bool
	connects int
}

func (f *fakeMCP) ListTools(ctx context.Context) ([]domaintools.MCPToolDefinition, error) {
	if f.block {
		<-ctx.Done()
		return nil, ctx.Err()
	}
	return f.tools, nil
}

func (f *fakeMCP) CallTool(context.Context, string, map[string]interface{}) (domaintools.MCPToolResult, error) {
	return domaintools.MCPToolResult{}, nil
}

func (f *fakeMCP) Close() error { return nil }

func boolPtr(b bool) *bool { return &b }

// T027 [UNIT] — discovery: concurrent bounded probing, never-answer skip+warn,
// ENABLED skip (never dialed), and deterministic (sorted-by-server) order
// (FR-008/FR-011 / Decision 3). The seams are injected (F6).
func TestDiscoverForRun_NonStallSkipSortAndDisabled(t *testing.T) {
	objSchema := []byte(`{"type":"object","properties":{}}`)
	fakes := map[string]*fakeMCP{
		"u-alpha": {tools: []domaintools.MCPToolDefinition{{Name: "a1", InputSchema: objSchema}}},
		"u-shop":  {tools: []domaintools.MCPToolDefinition{{Name: "lookup_price", InputSchema: objSchema}}},
		"u-hf":    {block: true},
		"u-off":   {tools: []domaintools.MCPToolDefinition{{Name: "should_not_appear", InputSchema: objSchema}}},
	}
	d := mcpDiscoveryConfig{
		bound: 50 * time.Millisecond,
		newClient: func(_ context.Context, url, _ string, _, _ time.Duration) (domaintools.MCPClient, error) {
			f := fakes[url]
			if f != nil {
				f.connects++
			}
			return f, nil
		},
		resolveToken: func(context.Context) (string, error) { return "", nil },
	}

	servers := map[string]config.MCPServerConfig{
		"shop":  {URL: "u-shop"},
		"alpha": {URL: "u-alpha"},
		"hf":    {URL: "u-hf"},
		"off":   {URL: "u-off", Enabled: boolPtr(false)},
	}

	start := time.Now()
	run := d.discover(context.Background(), servers)
	defer run.close()
	elapsed := time.Since(start)

	var names []string
	for _, tool := range run.tools {
		names = append(names, tool.Name())
	}

	// sorted by server key: alpha before shop.
	if len(names) != 2 || names[0] != "mcp_alpha_a1" || names[1] != "mcp_shop_lookup_price" {
		t.Fatalf("expected sorted [mcp_alpha_a1 mcp_shop_lookup_price], got %v", names)
	}
	for _, n := range names {
		if strings.Contains(n, "off") {
			t.Fatalf("the disabled server's tool was offered: %v", names)
		}
	}
	if fakes["u-off"].connects != 0 {
		t.Fatalf("the disabled server was dialed %d time(s)", fakes["u-off"].connects)
	}
	warned := false
	for _, w := range run.warnings {
		if strings.Contains(w, "hf") {
			warned = true
		}
	}
	if !warned {
		t.Fatalf("the never-answering server was not warned: %v", run.warnings)
	}
	// the never-answer server must not stall beyond a small multiple of the bound.
	if elapsed > 2*time.Second {
		t.Fatalf("discovery stalled for %v (bound %v)", elapsed, d.bound)
	}
}

// TestDiscoverForRun_SkipsUnsafeNames pins F5: a server tool whose namespaced
// wire name cannot satisfy the tool-name grammar is skipped with a warning.
func TestDiscoverForRun_SkipsUnsafeNames(t *testing.T) {
	objSchema := []byte(`{"type":"object","properties":{}}`)
	fake := &fakeMCP{tools: []domaintools.MCPToolDefinition{
		{Name: "good_tool", InputSchema: objSchema},
		{Name: "bad tool\u00a0name", InputSchema: objSchema},
	}}
	d := mcpDiscoveryConfig{
		bound: 50 * time.Millisecond,
		newClient: func(context.Context, string, string, time.Duration, time.Duration) (domaintools.MCPClient, error) {
			return fake, nil
		},
		resolveToken: func(context.Context) (string, error) { return "", nil },
	}
	run := d.discover(context.Background(), map[string]config.MCPServerConfig{"shop": {URL: "u"}})
	defer run.close()
	if len(run.tools) != 1 || run.tools[0].Name() != "mcp_shop_good_tool" {
		t.Fatalf("expected only the safe tool, got %v", run.tools)
	}
	found := false
	for _, w := range run.warnings {
		if strings.Contains(w, "cannot be offered safely") {
			found = true
		}
	}
	if !found {
		t.Fatalf("expected a name-skip warning, got %v", run.warnings)
	}
}
