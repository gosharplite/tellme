package mcp

import (
	"context"
	"errors"
	"path/filepath"
	"testing"
	"time"

	"github.com/gosharplite/tellme/internal/config"
	domaintools "github.com/gosharplite/tellme/internal/domain/tools"
)

// Round 088 (ADR 0059) unit pins — the lazy client's tools/list warm-up (header
// routing) and the cache's per-mode workspace location.

// warmFake records ListTools and CallTool invocations; ListTools can be made to
// fail (EC-001).
type warmFake struct {
	lists   int
	calls   int
	listErr error
	result  string
}

func (f *warmFake) ListTools(context.Context) ([]domaintools.MCPToolDefinition, error) {
	f.lists++
	return nil, f.listErr
}

func (f *warmFake) CallTool(context.Context, string, map[string]interface{}) (domaintools.MCPToolResult, error) {
	f.calls++
	return domaintools.MCPToolResult{Text: f.result}, nil
}

func (f *warmFake) Close() error { return nil }

func clientReturning(f *warmFake) ClientFactory {
	return func(context.Context, string, string, time.Duration, time.Duration) (domaintools.MCPClient, error) {
		return f, nil
	}
}

// TestLazyClient_WarmsToolsListOnceBeforeCall — FR-001: the first call issues
// exactly one tools/list (so the SDK can route x-mcp-header args), and the
// warm-up is memoized (a second call does not re-list).
func TestLazyClient_WarmsToolsListOnceBeforeCall(t *testing.T) {
	f := &warmFake{result: "ok"}
	lc := NewLazyClient(config.MCPServerConfig{URL: "u"}, time.Second, clientReturning(f), noToken)
	if f.lists != 0 {
		t.Fatalf("a lazy client must not list before a call; lists=%d", f.lists)
	}
	if _, err := lc.CallTool(context.Background(), "t", nil); err != nil {
		t.Fatalf("CallTool: %v", err)
	}
	if f.lists != 1 || f.calls != 1 {
		t.Fatalf("the first call must warm once then call; lists=%d calls=%d", f.lists, f.calls)
	}
	if _, err := lc.CallTool(context.Background(), "t", nil); err != nil {
		t.Fatalf("CallTool 2: %v", err)
	}
	if f.lists != 1 {
		t.Fatalf("the warm-up must be memoized; lists=%d", f.lists)
	}
	if f.calls != 2 {
		t.Fatalf("both calls must reach the tool; calls=%d", f.calls)
	}
}

// TestLazyClient_WarmFailureDoesNotBlockCall — EC-001: a best-effort warm-up
// failure does not stop the call from being attempted.
func TestLazyClient_WarmFailureDoesNotBlockCall(t *testing.T) {
	f := &warmFake{listErr: errors.New("boom"), result: "ok"}
	lc := NewLazyClient(config.MCPServerConfig{URL: "u"}, time.Second, clientReturning(f), noToken)
	res, err := lc.CallTool(context.Background(), "t", nil)
	if err != nil {
		t.Fatalf("a warm-up failure must not surface an error; got %v", err)
	}
	if res.Text != "ok" || f.calls != 1 {
		t.Fatalf("the call must still be attempted; result=%q calls=%d", res.Text, f.calls)
	}
}

// TestLazyClient_NoWarmWhenUnused — FR-002: constructing and discarding a lazy
// client makes no connection and issues no tools/list.
func TestLazyClient_NoWarmWhenUnused(t *testing.T) {
	f := &warmFake{}
	lc := NewLazyClient(config.MCPServerConfig{URL: "u"}, time.Second, clientReturning(f), noToken)
	_ = lc.Close()
	if f.lists != 0 || f.calls != 0 {
		t.Fatalf("an unused lazy client must dial/list nothing; lists=%d calls=%d", f.lists, f.calls)
	}
}

// TestFileToolCache_LivesInTheGivenWorkspace — FR-003: the store path is
// <workspace>/mcp-toolcache.json (round 088 / ADR 0059).
func TestFileToolCache_LivesInTheGivenWorkspace(t *testing.T) {
	ws := t.TempDir()
	c := NewFileToolCache(ws).(*fileToolCache)
	want := filepath.Join(ws, MCPToolCacheFileName)
	if c.path != want {
		t.Fatalf("the cache path must be the workspace one; got %q want %q", c.path, want)
	}
}
